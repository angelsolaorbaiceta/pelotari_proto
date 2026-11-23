package prototari

import (
	"fmt"
	"log"
	"net"
	"time"
)

// UDPBroadcastConn is an implementation of the BroadcastConn interface that uses
// the UDP connection-less protocol to send and receive messages.
type UDPBroadcastConn struct {
	localAddr     *net.UDPAddr
	broadcastAddr *net.UDPAddr

	isConnected bool

	conn *net.UDPConn
}

func (c *UDPBroadcastConn) Connect() {
	if c.isConnected {
		return
	}

	privIP, broadIP, err := GetPrivateIPAndBroadcastAddr()
	if err != nil {
		// Not much we can do here.
		// A network protocol can't work without a network.
		log.Fatal(err)
	}

	c.localAddr = &net.UDPAddr{
		IP:   privIP,
		Port: BroadcastPort,
	}
	broadcastAddr, err := net.ResolveUDPAddr(
		"udp",
		fmt.Sprintf("%s:%d", broadIP, BroadcastPort),
	)
	if err != nil {
		// Not much we can do here.
		// If a UDP address can't be dialed, the protocol can't work.
		log.Fatal(err)
	}
	c.broadcastAddr = broadcastAddr

	udpConn, err := net.ListenUDP("udp", broadcastAddr)
	if err != nil {
		log.Fatal(err)
	}
	c.conn = udpConn

	c.isConnected = true
}

func (c UDPBroadcastConn) LocalAddr() *net.UDPAddr {
	return c.localAddr
}

func (c UDPBroadcastConn) Write(b []byte) (int, error) {
	return c.conn.WriteToUDP(b, c.broadcastAddr)
}

func (c UDPBroadcastConn) Read(b []byte) (int, *net.UDPAddr, error) {
	c.conn.SetReadDeadline(time.Now().Add(connReadTimeout))
	return c.conn.ReadFromUDP(b)
}

func (c *UDPBroadcastConn) Close() {
	if !c.isConnected {
		return
	}

	c.conn.Close()

	c.conn = nil
	c.localAddr = nil
	c.broadcastAddr = nil

	c.isConnected = false
}

// UDPUnicastConn is an implementation of the UnicastConn interface that uses
// the UDP connection-less protocol to send and receive messages.
type UDPUnicastConn struct {
	localAddr   *net.UDPAddr
	isConnected bool
	conn        *net.UDPConn
}

func (c *UDPUnicastConn) Connect() {
	if c.isConnected {
		return
	}

	privIP, _, err := GetPrivateIPAndBroadcastAddr()
	if err != nil {
		// Not much we can do here.
		// A network protocol can't work without a network.
		log.Fatal(err)
	}

	c.localAddr = &net.UDPAddr{
		IP:   privIP,
		Port: UnicastPort,
	}

	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", UnicastPort))
	if err != nil {
		log.Fatal(err)
	}
	udpConn, err := net.ListenUDP("udp", addr)
	if err != nil {
		// Not much we can do here.
		// If a UDP address can't be dialed, the protocol can't work.
		panic(err)
	}
	c.conn = udpConn

	c.isConnected = true
}

func (c UDPUnicastConn) LocalAddr() *net.UDPAddr {
	return c.localAddr
}

func (c UDPUnicastConn) Write(b []byte, to *net.UDPAddr) (int, error) {
	return c.conn.WriteToUDP(b, to)
}

func (c UDPUnicastConn) Read(b []byte) (int, *net.UDPAddr, error) {
	c.conn.SetReadDeadline(time.Now().Add(connReadTimeout))
	return c.conn.ReadFromUDP(b)
}

func (c *UDPUnicastConn) Close() {
	if !c.isConnected {
		return
	}

	c.conn.Close()

	c.conn = nil
	c.localAddr = nil

	c.isConnected = false
}
