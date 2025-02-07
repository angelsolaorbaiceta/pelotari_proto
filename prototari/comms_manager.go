package prototari

import (
	"errors"
	"log"
	"net"
	"sync"
	"time"
)

// A CommsManager is the central authority in the Pelotari protocol.
// It deals with discovering and registering peers, as well as sending periodic
// heartbeats to those peers from whom it hasn't heard anything in a specified
// amount of time.
type CommsManager struct {
	broadcaster BroadcastConn
	unicaster   UnicastConn

	config Config

	peers      map[string]Peer
	peersMutex sync.RWMutex

	isRunning bool
	done      chan struct{}
	wg        sync.WaitGroup
}

func MakeManager(
	broadcaster BroadcastConn,
	unicaster UnicastConn,
	config Config,
) *CommsManager {
	return &CommsManager{
		broadcaster: broadcaster,
		unicaster:   unicaster,
		config:      config,
		peers:       make(map[string]Peer, config.MaxPeers),
		isRunning:   false,
	}
}

// Peers returns a slice of the currently registered peers.
func (m *CommsManager) Peers() []Peer {
	m.peersMutex.RLock()
	defer m.peersMutex.RUnlock()

	peers := make([]Peer, 0, len(m.peers))
	for _, peer := range m.peers {
		peers = append(peers, peer)
	}

	return peers
}

// Start begins the peer discovery mechanism and listens for incoming messages
// from registered peers.
//
// The discovery mechanism consists of three different tasks that run
// concurrently:
//
//  1. Broadcasting: Broadcasts the discovery message every few seconds.
//  2. Responding: Listens to broadcasts from other peers and resonds with a
//     response message.
//  3. Handshake: Sends a unicast message to confirm adding a peer.
func (m *CommsManager) Start() {
	if m.isRunning {
		return
	}

	m.isRunning = true
	m.done = make(chan struct{})
	m.wg = sync.WaitGroup{}
	m.wg.Add(3)

	// 1. Broadcasting
	go func() {
		defer m.wg.Done()

		for {
			select {
			case <-m.done:
				return
			default:
				_, err := m.broadcaster.Write([]byte(discoveryMessage))
				if err != nil {
					log.Println("Sending a broadcast message failed")
				}

				select {
				case <-time.After(m.config.BroadcastInterval):
					continue
				case <-m.done:
					return
				}
			}
		}
	}()

	// 2. Responding to broadcast messages
	go func() {
		defer m.wg.Done()

		var (
			buff = make([]byte, 128)
			myIp = m.broadcaster.LocalAddr().IP
		)

		for {
			select {
			case <-m.done:
				return
			default:
				n, addr, err := m.broadcaster.Read(buff)
				if err != nil {
					log.Printf("Error reading broadcast: %v", err)
					continue
				}

				// Ignore our own broadcast messages
				if myIp.Equal(addr.IP) {
					continue
				}

				if string(buff[:n]) == discoveryMessage {
					peerAddr := *addr
					peerAddr.Port = UnicastPort

					m.unicaster.Write([]byte(responseMessage), &peerAddr)
				}
			}
		}
	}()

	// 3. Unicast messages (including handshake)
	go func() {
		defer m.wg.Done()

		buff := make([]byte, 1024)

		for {
			select {
			case <-m.done:
				return
			default:
				n, addr, err := m.unicaster.Read(buff)
				if err != nil {
					log.Printf("Error reading unicast: %v", err)
					continue
				}

				message := buff[:n]
				if string(message) == responseMessage {
					m.completeHandshake(addr)
				} else if string(message) == confirmationMessage {
					peer := MakePeer(addr.IP)
					// TODO: handle error
					m.registerPeer(peer)
				}
			}
		}
	}()
}

// completeHandshake is called by the broadcaster to add the responder as a peer
// and send the confirmation message that completes the handshake.
//
// If the maximum number of peers is already reached, it does nothing.
func (m *CommsManager) completeHandshake(peerAddr *net.UDPAddr) {
	peer := MakePeer(peerAddr.IP)
	if err := m.registerPeer(peer); err != nil {
		return
	}

	_, err := m.unicaster.Write([]byte(confirmationMessage), peer.Address())
	if err != nil {
		// TODO: handle error. Retry or fail to register peer
		return
	}
}

func (m *CommsManager) registerPeer(peer Peer) error {
	m.peersMutex.Lock()
	defer m.peersMutex.Unlock()

	if len(m.peers) >= m.config.MaxPeers {
		return errors.New("max peers registered")
	}

	m.peers[string(peer.IP)] = peer

	return nil
}

func (m *CommsManager) Stop() {
	if !m.isRunning {
		return
	}

	close(m.done)
	m.wg.Wait()
	m.isRunning = false
}
