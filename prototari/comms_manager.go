package prototari

import "net"

// A CommsManager is the central authority in the Pelotari protocol.
// It deals with discovering and registering peers, as well as sending periodic
// heartbeats to those peers from whom it hasn't heard anything in a specified
// amount of time.
type CommsManager struct {
	broadcastConn net.Conn
}

func MakeManager(broadcastConn net.Conn) *CommsManager {
	return &CommsManager{
		broadcastConn: broadcastConn,
	}
}

func (manager *CommsManager) StartDiscovery() {
	manager.broadcastConn.Write([]byte(discoveryMessage))
}

func (manager *CommsManager) StopDiscovery() {}
