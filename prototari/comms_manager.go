package prototari

import (
	"fmt"
	"net"
	"time"
)

// A BroadcastConn is the connection used to send and receive broadcast messages
// into the local private network.
type BroadcastConn interface {
	Write(buff []byte) (int, error)
	Read(buff []byte) (int, *net.UDPAddr, error)
}

// A CommsManager is the central authority in the Pelotari protocol.
// It deals with discovering and registering peers, as well as sending periodic
// heartbeats to those peers from whom it hasn't heard anything in a specified
// amount of time.
type CommsManager struct {
	broadcaster BroadcastConn
	config      Config
	peers       map[net.Addr]Peer
}

func MakeManager(
	broadcaster BroadcastConn,
	config Config,
) *CommsManager {
	return &CommsManager{
		broadcaster: broadcaster,
		config:      config,
		peers:       make(map[net.Addr]Peer, config.MaxPeers),
	}
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
			// TODO: error handling
			manager.broadcaster.Write([]byte(discoveryMessage))
			time.Sleep(manager.config.BroadcastInterval)
		}
	}()

	// 2. Responding
	go func() {
		buff := make([]byte, 128)

		for {
			n, _, err := manager.broadcaster.Read(buff)
			if err != nil {
				// TODO: error handling
			}

			message := string(buff[:n])
			fmt.Println(message)
		}
	}()
}

func (manager *CommsManager) Stop() {}
