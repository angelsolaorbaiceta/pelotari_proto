package prototari

import (
	"net"
	"time"
)

// A Peer is another computer running the same protocol, with whom this computer
// can talk to and receive messages from.
type Peer struct {
	// The peer's IP address inside the private network.
	IP net.IP

	// The time when the last message from the peer was received.
	LastSeen time.Time

	// The number of heartbeats the peer hasn't responded to.
	MissedHeartbeats int
}

func MakePeer(IP net.IP) Peer {
	return Peer{
		IP:               IP,
		LastSeen:         time.Now(),
		MissedHeartbeats: 0,
	}
}

func (p Peer) Address() *net.UDPAddr {
	return &net.UDPAddr{
		IP:   p.IP,
		Port: UnicastPort,
	}
}

// Equal returs whether this and other peer are the same.
// Peers are identified by their IP address.
func (p Peer) Equal(other Peer) bool {
	return p.IP.Equal(other.IP)
}
