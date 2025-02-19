package prototari

import (
	"net"
)

// A BroadcastConn is the connection used to send and receive broadcast messages
// into the local private network.
type BroadcastConn interface {
	// LocalAddr returns the local network address.
	LocalAddr() *net.UDPAddr

	// Write sends a payload as a broadcast UDP message to the local network.
	// It returns the number of sent bytes.
	Write(b []byte) (int, error)

	// Read receives broadcast messages from the local network.
	// The number of read bytes and the UDP address of the sender are returned.
	Read(b []byte) (int, *net.UDPAddr, error)

	// Close closes the broadcast connections.
	Close()
}

// UnicastConn is the connection used to send and receive unicast messages
// to known peers in the local private network.
type UnicastConn interface {
	// LocalAddr returns the local network address.
	LocalAddr() *net.UDPAddr

	// Write sends a payload to a given peer in the local network.
	// It returns the number of sent bytes.
	Write(b []byte, to *net.UDPAddr) (int, error)

	// Read receives unicast messages from peers on the local network.
	// The number of read bytes and the UDP address of the sender are returned.
	Read(b []byte) (int, *net.UDPAddr, error)

	// Close closes the unicast connections.
	Close()
}
