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

	sendConn *net.UDPConn
	readConn *net.UDPConn
}

func (conn *UDPBroadcastConn) Connect() {
	if conn.isConnected {
		return
	}

	privIP, broadIP, err := GetPrivateIPAndBroadcastAddr()
	if err != nil {
		// Not much we can do here.
		// A network protocol can't work without a network.
		log.Fatal(err)
	}

	conn.localAddr = &net.UDPAddr{
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
	conn.broadcastAddr = broadcastAddr

	sendConn, err := net.DialUDP("udp", nil, conn.broadcastAddr)
	if err != nil {
		// Not much we can do here.
		// If a UDP address can't be dialed, the protocol can't work.
		log.Fatal(err)
	}
	conn.sendConn = sendConn

	readAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", BroadcastPort))
	if err != nil {
		log.Fatal(err)
	}
	readUdpConn, err := net.ListenUDP("udp", readAddr)
	if err != nil {
		// Not much we can do here.
		// If a UDP address can't be dialed, the protocol can't work.
		panic(err)
	}
	conn.readConn = readUdpConn

	conn.isConnected = true
}

func (conn UDPBroadcastConn) LocalAddr() *net.UDPAddr {
	return conn.localAddr
}

func (conn UDPBroadcastConn) Write(b []byte) (int, error) {
	return conn.sendConn.Write(b)
}

func (conn UDPBroadcastConn) Read(b []byte) (int, *net.UDPAddr, error) {
	conn.readConn.SetReadDeadline(time.Now().Add(connReadTimeout))
	return conn.readConn.ReadFromUDP(b)
}

func (conn *UDPBroadcastConn) Close() {
	if !conn.isConnected {
		return
	}

	conn.sendConn.Close()
	conn.readConn.Close()

	conn.sendConn = nil
	conn.readConn = nil
	conn.localAddr = nil
	conn.broadcastAddr = nil

	conn.isConnected = false
}

// UDPUnicastConn is an implementation of the UnicastConn interface that uses
// the UDP connection-less protocol to send and receive messages.
type UDPUnicastConn struct {
	localAddr   *net.UDPAddr
	isConnected bool
	readConn    *net.UDPConn
}

func (conn *UDPUnicastConn) Connect() {
	if conn.isConnected {
		return
	}

	privIP, _, err := GetPrivateIPAndBroadcastAddr()
	if err != nil {
		// Not much we can do here.
		// A network protocol can't work without a network.
		log.Fatal(err)
	}

	conn.localAddr = &net.UDPAddr{
		IP:   privIP,
		Port: BroadcastPort,
	}

	readAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", UnicastPort))
	if err != nil {
		log.Fatal(err)
	}
	readUdpConn, err := net.ListenUDP("udp", readAddr)
	if err != nil {
		// Not much we can do here.
		// If a UDP address can't be dialed, the protocol can't work.
		panic(err)
	}
	conn.readConn = readUdpConn

	conn.isConnected = true
}

func (conn UDPUnicastConn) LocalAddr() *net.UDPAddr {
	return conn.localAddr
}

func (conn UDPUnicastConn) Write(b []byte, to *net.UDPAddr) (int, error) {
	sendConn, err := net.DialUDP("upd", nil, to)
	if err != nil {
		return 0, nil
	}

	return sendConn.Write(b)
}

func (conn UDPUnicastConn) Read(b []byte) (int, *net.UDPAddr, error) {
	conn.readConn.SetReadDeadline(time.Now().Add(connReadTimeout))
	return conn.readConn.ReadFromUDP(b)
}

func (conn *UDPUnicastConn) Close() {
	if !conn.isConnected {
		return
	}

	conn.readConn.Close()

	conn.readConn = nil
	conn.localAddr = nil

	conn.isConnected = false
}
