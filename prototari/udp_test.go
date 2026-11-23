package prototari

import (
	"errors"
	"net"
	"runtime"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests exercise UDPUnicastConn and UDPBroadcastConn against real kernel
// sockets on the loopback interface. The CommsManager tests use in-memory fakes,
// so they can't catch bugs in how the real connections talk to the network (e.g.
// writing on an unconnected socket without a destination address, which fails
// with EDESTADDRREQ, or sending from an ephemeral port instead of the protocol
// port).
//
// Connect() discovers the machine's private IP, which isn't available
// everywhere (CI runners, containers). The sockets are opened directly here
// instead, mirroring what Connect() is expected to do: listen on the protocol
// port and send from that same socket. As a consequence, these tests don't
// cover Connect() itself.

// ioTimeout bounds every blocking read so a broken implementation makes the test
// fail quickly instead of hanging until the global test timeout.
const ioTimeout = 1 * time.Second

var (
	loopbackIP = net.IPv4(127, 0, 0, 1)

	// loopbackBroadcastIP is the broadcast address of 127.0.0.0/8. The lo
	// interface has no IFF_BROADCAST flag, but Linux installs a broadcast route
	// for it in the local routing table (see `ip route show table local`), so
	// datagrams sent here are delivered to every socket bound to the port.
	loopbackBroadcastIP = net.IPv4(127, 255, 255, 255)
)

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

// newLoopbackBroadcastConn returns a UDPBroadcastConn listening on
// BroadcastPort and broadcasting to the loopback broadcast address. The
// connection is closed when the test ends.
func newLoopbackBroadcastConn(t *testing.T) *UDPBroadcastConn {
	t.Helper()

	if runtime.GOOS != "linux" {
		// Other systems don't necessarily route 127.255.255.255 as a broadcast.
		t.Skipf("loopback broadcast isn't supported on %s", runtime.GOOS)
	}

	udpConn, err := net.ListenUDP("udp", &net.UDPAddr{Port: BroadcastPort})
	if errors.Is(err, syscall.EADDRINUSE) {
		// Most likely a Pelotari instance is running on this machine.
		t.Skipf("port %d already in use, skipping: %v", BroadcastPort, err)
	}
	require.NoError(t, err)

	conn := &UDPBroadcastConn{
		localAddr:     &net.UDPAddr{IP: loopbackIP, Port: BroadcastPort},
		broadcastAddr: &net.UDPAddr{IP: loopbackBroadcastIP, Port: BroadcastPort},
		isConnected:   true,
		conn:          udpConn,
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

func TestUDPBroadcastConn(t *testing.T) {
	t.Run("Write broadcasts from BroadcastPort and the broadcast is received", func(t *testing.T) {
		var (
			conn    = newLoopbackBroadcastConn(t)
			payload = []byte(discoveryMessage)
		)

		n, err := conn.Write(payload)
		require.NoError(t, err)
		assert.Equal(t, len(payload), n)

		// The broadcast is delivered to every socket bound to BroadcastPort,
		// including the sender's own. That's why the CommsManager filters out
		// broadcasts coming from its own IP.
		buff := make([]byte, 64)
		n, from, err := conn.Read(buff)
		require.NoError(t, err, "the broadcast was never received on BroadcastPort")

		assert.Equal(t, payload, buff[:n])
		assert.True(t, loopbackIP.Equal(from.IP), "want sender IP %s, got %s", loopbackIP, from.IP)
		assert.Equal(t, BroadcastPort, from.Port, "broadcast wasn't sent from BroadcastPort")
	})

	t.Run("Read receives broadcasts sent by other hosts", func(t *testing.T) {
		var (
			conn     = newLoopbackBroadcastConn(t)
			peer     = newLoopbackPeer(t)
			peerAddr = peer.LocalAddr().(*net.UDPAddr)
			payload  = []byte(discoveryMessage)
		)

		_, err := peer.WriteToUDP(
			payload,
			&net.UDPAddr{IP: loopbackBroadcastIP, Port: BroadcastPort},
		)
		require.NoError(t, err)

		buff := make([]byte, 64)
		n, from, err := conn.Read(buff)
		require.NoError(t, err, "the broadcast was never received on BroadcastPort")

		assert.Equal(t, payload, buff[:n])
		assert.True(t, peerAddr.IP.Equal(from.IP), "want sender IP %s, got %s", peerAddr.IP, from.IP)
		assert.Equal(t, peerAddr.Port, from.Port)
	})
}
