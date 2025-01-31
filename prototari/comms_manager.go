package prototari

import (
	"errors"
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
	config      Config
	peers       map[string]Peer
	peersMutex  sync.RWMutex
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
	}
}

// Peers returns a slice of the currently registered peers.
func (manager CommsManager) Peers() []Peer {
	manager.peersMutex.RLock()
	defer manager.peersMutex.RUnlock()

	peers := make([]Peer, 0, len(manager.peers))
	for _, peer := range manager.peers {
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
func (manager *CommsManager) Start() {
	// 1. Broadcasting
	go func() {
		for {
			_, err := manager.broadcaster.Write([]byte(discoveryMessage))
			if err != nil {
				// TODO: error handling
			}
			time.Sleep(manager.config.BroadcastInterval)
		}
	}()

	// 2. Responding to broadcast messages
	go func() {
		buff := make([]byte, 128)

		for {
			n, respAddr, err := manager.broadcaster.Read(buff)
			if err != nil {
				// TODO: error handling. Take io.EOF into account.
			}

			// TODO: ignore own messages
			message := string(buff[:n])
			if message == discoveryMessage {
				peerAddr := *respAddr
				peerAddr.Port = UnicastPort

				manager.unicaster.Write([]byte(responseMessage), &peerAddr)
			}
		}
	}()

	// 3. Unicast messages (including handshake)
	go func() {
		buff := make([]byte, 1024)

		for {
			n, addr, err := manager.unicaster.Read(buff)
			if err != nil {
				// TODO: handle error. Take io.EOF into account.
			}

			message := buff[:n]
			if string(message) == responseMessage {
				manager.completeHandshake(addr)
			}
		}
	}()
}

// completeHandshake is called by the broadcaster to add the responder as a peer
// and send the confirmation message that completes the handshake.
//
// If the maximum number of peers is already reached, it does nothing.
func (manager *CommsManager) completeHandshake(peerAddr *net.UDPAddr) {
	peer := MakePeer(peerAddr.IP)
	if err := manager.registerPeer(peer); err != nil {
		return
	}

	_, err := manager.unicaster.Write([]byte(confirmationMessage), peer.Address())
	if err != nil {
		// TODO: handle error
		return
	}
}

func (manager *CommsManager) registerPeer(peer Peer) error {
	manager.peersMutex.Lock()
	defer manager.peersMutex.Unlock()

	if len(manager.peers) >= manager.config.MaxPeers {
		return errors.New("Max peers registered")
	}

	manager.peers[string(peer.IP)] = peer

	return nil
}

func (manager *CommsManager) Stop() {}
