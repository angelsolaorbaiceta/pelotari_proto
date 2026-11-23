package prototari

import (
	"errors"
	"net"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests exercise UDPUnicastConn against real kernel sockets on the
// loopback interface. The CommsManager tests use in-memory fakes, so they can't
// catch bugs in how the real connection talks to the network (e.g. writing on an
// unconnected socket without a destination address, which fails with
// EDESTADDRREQ, or sending from an ephemeral port instead of UnicastPort).
//
// UDPUnicastConn.Connect() discovers the machine's private IP, which isn't
// available everywhere (CI runners, containers). The socket is listened on
// directly here instead, mirroring what Connect() does.

// ioTimeout bounds every blocking read so a broken implementation makes the test
// fail quickly instead of hanging until the global test timeout.
const ioTimeout = 1 * time.Second

var loopbackIP = net.IPv4(127, 0, 0, 1)

// newLoopbackUnicastConn returns a UDPUnicastConn listening on UnicastPort, the
// same way Connect() sets it up, but without relying on a private IP being
// present. The connection is closed when the test ends.
func newLoopbackUnicastConn(t *testing.T) *UDPUnicastConn {
	t.Helper()

	udpConn, err := net.ListenUDP("udp", &net.UDPAddr{Port: UnicastPort})
	if errors.Is(err, syscall.EADDRINUSE) {
		// Most likely a Pelotari instance is running on this machine.
		t.Skipf("port %d already in use, skipping: %v", UnicastPort, err)
	}
	require.NoError(t, err)

	conn := &UDPUnicastConn{
		localAddr:   &net.UDPAddr{IP: loopbackIP, Port: UnicastPort},
		isConnected: true,
		conn:        udpConn,
	}
	t.Cleanup(conn.Close)

	return conn
}

// newLoopbackPeer returns a plain UDP socket bound to an ephemeral port on the
// loopback interface. It plays the role of a remote peer.
func newLoopbackPeer(t *testing.T) *net.UDPConn {
	t.Helper()

	peer, err := net.ListenUDP("udp4", &net.UDPAddr{IP: loopbackIP, Port: 0})
	require.NoError(t, err)
	t.Cleanup(func() { peer.Close() })

	return peer
}

func TestUDPUnicastConn(t *testing.T) {
	t.Run("Write delivers the payload to the given address", func(t *testing.T) {
		var (
			conn     = newLoopbackUnicastConn(t)
			peer     = newLoopbackPeer(t)
			peerAddr = peer.LocalAddr().(*net.UDPAddr)
			payload  = []byte(responseMessage)
		)

		// The peer listens on an ephemeral port, not UnicastPort. If Write
		// ignored its destination, the datagram couldn't reach it.
		n, err := conn.Write(payload, peerAddr)
		require.NoError(t, err)
		assert.Equal(t, len(payload), n)

		buff := make([]byte, 64)
		require.NoError(t, peer.SetReadDeadline(time.Now().Add(ioTimeout)))
		n, from, err := peer.ReadFromUDP(buff)
		require.NoError(t, err, "the peer never received the datagram")

		assert.Equal(t, payload, buff[:n])

		// The protocol mandates all unicast traffic flows from and to
		// UnicastPort, so replies can be addressed to the sender's port and
		// stateful firewalls see a single flow.
		assert.Equal(t, UnicastPort, from.Port, "datagram wasn't sent from UnicastPort")
	})

	t.Run("Read returns the payload and the sender's address", func(t *testing.T) {
		var (
			conn     = newLoopbackUnicastConn(t)
			peer     = newLoopbackPeer(t)
			peerAddr = peer.LocalAddr().(*net.UDPAddr)
			payload  = []byte(confirmationMessage)
		)

		_, err := peer.WriteToUDP(payload, &net.UDPAddr{IP: loopbackIP, Port: UnicastPort})
		require.NoError(t, err)

		buff := make([]byte, 64)
		n, from, err := conn.Read(buff)
		require.NoError(t, err, "the datagram wasn't received on UnicastPort")

		assert.Equal(t, payload, buff[:n])
		assert.True(t, peerAddr.IP.Equal(from.IP), "want sender IP %s, got %s", peerAddr.IP, from.IP)
		assert.Equal(t, peerAddr.Port, from.Port)
	})

	t.Run("Many writes reuse the same socket", func(t *testing.T) {
		// Regression test for #5: each Write used to open a new socket that was
		// never closed. Every datagram must come from the same source address.
		const nMsgs = 100

		var (
			conn     = newLoopbackUnicastConn(t)
			peer     = newLoopbackPeer(t)
			peerAddr = peer.LocalAddr().(*net.UDPAddr)
			buff     = make([]byte, 64)
		)

		for i := range nMsgs {
			_, err := conn.Write([]byte("ping"), peerAddr)
			require.NoError(t, err, "write #%d failed", i)

			require.NoError(t, peer.SetReadDeadline(time.Now().Add(ioTimeout)))
			_, from, err := peer.ReadFromUDP(buff)
			require.NoError(t, err, "datagram #%d never arrived", i)
			require.Equal(t, UnicastPort, from.Port, "datagram #%d sent from another socket", i)
		}
	})
}
