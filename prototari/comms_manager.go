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

	peersCh    chan []Peer
	peers      map[string]Peer
	peersMutex sync.RWMutex

	isRunning bool
	done      chan struct{}
	wg        sync.WaitGroup
}

// MakeUDPManager returns an instance of a CommsManager with the broadcaster
// and unicaster connected and ready to send UDP messages.
func MakeUDPManager(config Config) *CommsManager {
	var (
		broadcaster = UDPBroadcastConn{}
		unicaster   = UDPUnicastConn{}
	)

	broadcaster.Connect()
	unicaster.Connect()

	return MakeManager(
		&broadcaster,
		&unicaster,
		config,
	)
}

// MakeManager returns an instance of a CommsManager with the passed in
// broadcaster and unicaster. The underlying connections of the messagers
// have to be connected by the client using this factory.
func MakeManager(
	broadcaster BroadcastConn,
	unicaster UnicastConn,
	config Config,
) *CommsManager {
	return &CommsManager{
		broadcaster: broadcaster,
		unicaster:   unicaster,
		config:      config,
		peersCh:     make(chan []Peer, 1),
		peers:       make(map[string]Peer, config.MaxPeers),
		isRunning:   false,
	}
}

// PeersCh returns a channel of the registered peers.
// The channel is a buffered channel with a capacity of one.
// The CommsManager sends the most up-to-date list of registered peers.
func (m *CommsManager) PeersCh() <-chan []Peer {
	return m.peersCh
}

// NOfPeers returns the currently registered number of peers.
func (m *CommsManager) NOfPeers() int {
	m.peersMutex.RLock()
	defer m.peersMutex.RUnlock()

	return len(m.peers)
}

// hasPeer checks if a peer with a given IP is registered.
func (m *CommsManager) hasPeer(IP net.IP) bool {
	m.peersMutex.RLock()
	defer m.peersMutex.RUnlock()

	_, ok := m.peers[string(IP)]
	return ok
}

// Start begins the peer discovery and heartbeat mechanisms and listens for
// incoming messages from registered peers.
func (m *CommsManager) Start() {
	if m.isRunning {
		return
	}

	m.isRunning = true
	m.done = make(chan struct{})
	m.wg = sync.WaitGroup{}
	m.wg.Add(3)

	go m.startBroadcasting()
	go m.startRespondingToBroadcasts()
	go m.startListeningToUnicast()
}

func (m *CommsManager) startBroadcasting() {
	defer func() {
		m.wg.Done()
		log.Println("[Close] Broadcasting goroutine done!")
	}()

	for {
		select {
		case <-m.done:
			return
		default:
			if m.NOfPeers() < m.config.MaxPeers {
				_, err := m.broadcaster.Write([]byte(discoveryMessage))
				if err != nil {
					log.Println("Sending a broadcast message failed")
				}
			}

			select {
			case <-time.After(m.config.BroadcastInterval):
				continue
			case <-m.done:
				return
			}
		}
	}
}

func (m *CommsManager) startRespondingToBroadcasts() {
	defer func() {
		m.wg.Done()
		log.Println("[Close] Broadcaster responder goroutine done!")
	}()

	var (
		buff = make([]byte, 128)
		myIP = m.broadcaster.LocalAddr().IP
	)

	for {
		select {
		case <-m.done:
			return
		default:
			n, addr, err := m.broadcaster.Read(buff)
			if err != nil {
				continue
			}

			// Ignore our own broadcast messages
			if myIP.Equal(addr.IP) {
				continue
			}

			if string(buff[:n]) == discoveryMessage && !m.hasPeer(addr.IP) {
				peerAddr := *addr
				peerAddr.Port = UnicastPort

				_, err := m.unicaster.Write([]byte(responseMessage), &peerAddr)
				if err != nil {
					log.Printf("Couldn't send response to %s: %s\n", peerAddr.IP, err)
				}
			}
		}
	}
}

func (m *CommsManager) startListeningToUnicast() {
	defer func() {
		m.wg.Done()
		log.Println("[Close] Unicaster goroutine done!")
	}()

	buff := make([]byte, 1024)

	for {
		select {
		case <-m.done:
			return
		default:
			n, addr, err := m.unicaster.Read(buff)
			if err != nil {
				continue
			}

			message := buff[:n]
			if string(message) == responseMessage {
				// TODO: handle error
				m.completeHandshake(addr)
			} else if string(message) == confirmationMessage {
				peer := MakePeer(addr.IP)
				// TODO: handle error
				m.registerPeer(peer)
			}
		}
	}
}

// completeHandshake is called by the broadcaster to add the responder as a peer
// and send the confirmation message that completes the handshake.
//
// It returns an error if the maximum number of peers are already registered.
func (m *CommsManager) completeHandshake(peerAddr *net.UDPAddr) error {
	peer := MakePeer(peerAddr.IP)
	if err := m.registerPeer(peer); err != nil {
		return err
	}

	_, err := m.unicaster.Write([]byte(confirmationMessage), peer.Address())

	return err
}

// registerPeer attempts to register a peer and sends a message to the peers
// channel with the new registered peers.
//
// It returns an error if the maximum number of peers are already registered.
func (m *CommsManager) registerPeer(peer Peer) error {
	m.peersMutex.Lock()
	defer m.peersMutex.Unlock()

	if len(m.peers) >= m.config.MaxPeers {
		return errors.New("max peers registered")
	}

	m.peers[string(peer.IP)] = peer

	// Send the new peers to the peers channel, replacing the last sent
	// value, if any.
	peers := make([]Peer, 0, len(m.peers))
	for _, peer := range m.peers {
		peers = append(peers, peer)
	}

	select {
	case <-m.peersCh:
		// There was an unread item, just discarded
		m.peersCh <- peers
	default:
		m.peersCh <- peers
	}

	return nil
}

// Stop signals all the CommsManager goroutines to stop.
func (m *CommsManager) Stop() {
	if !m.isRunning {
		return
	}

	close(m.done)
	m.wg.Wait()
	m.isRunning = false
}

// Close stops the communications (if they weren't already) and closes the
// broadcast and unicast connections.
func (m *CommsManager) Close() {
	m.Stop()
	m.broadcaster.Close()
	m.unicaster.Close()
}
