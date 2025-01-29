package prototari

import (
	"net"
	"time"
)

// A Peer is another computer running the same protocol, with whom this computer
// can talk to and receive messages from.
type Peer struct {
	// The peer's IP address inside the private network.
	ip net.IP

	// The time when the last message from the peer was received.
	lastSeen time.Time

	// The UDP "connection" to send messages to this peer.
	conn *net.UDPConn

	// The number of heartbeats the peer hasn't responded to.
	missedHeartbeats int
}
