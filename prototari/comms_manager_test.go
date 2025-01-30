package prototari

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCommsManager(t *testing.T) {
	// In this test we configure two broadcasters that communicate through a
	// channel. Whatever broadcasterBroadConn writes, responderBroadConn receives,
	// but the opposite is not true: responderBroadConn doesn't write anywhere.
	// This is to emulate the discovery from one peer to another in a single
	// direction, which simplifies the test.
	var (
		commsChan            = make(chan []byte)
		broadcasterBroadConn = fakeBroadcastConn{writeChan: commsChan, readChan: nil}
		responderBroadConn   = fakeBroadcastConn{writeChan: nil, readChan: commsChan}

		broadcaster = MakeManager(&broadcasterBroadConn, MakeDefaultConfig())
		responder   = MakeManager(&responderBroadConn, MakeDefaultConfig())

		// The IP-port the broadcaster uses to send broadcast messages
		broadcasterAddr = net.UDPAddr{
			IP:   []byte("192.168.0.10"),
			Port: 45678,
		}
		// The IP-port the receiver uses to send broadcast messages
		receiverAddr = net.UDPAddr{
			IP:   []byte("192.168.0.20"),
			Port: 45678,
		}
	)

	t.Run("Successful handshake", func(t *testing.T) {
		doneCh := make(chan struct{})

		broadcasterBroadConn.
			On("Write", []byte(discoveryMessage)).
			Run(func(args mock.Arguments) {
				doneCh <- struct{}{}
			}).
			Return(discoveryMessageLen, nil)
		// The broadcaster's reads from the broadcast connection are ignored
		broadcasterBroadConn.
			On("Read", mock.Anything).
			Return(0, &receiverAddr, nil)

			// The receiver's writes to the broadcast connection are ignored
		responderBroadConn.
			On("Write", mock.Anything).
			Return(0, nil)
		responderBroadConn.
			On("Read", mock.Anything).
			Run(func(args mock.Arguments) {
				doneCh <- struct{}{}
			}).
			Return(discoveryMessageLen, &broadcasterAddr, nil)

		go broadcaster.Start()
		go responder.Start()
		defer func() {
			broadcaster.Stop()
			responder.Stop()
		}()

		// Wait for the broadcast message to be sent
		<-doneCh
		// Wait for the broadcast message to be read
		<-doneCh

		broadcasterBroadConn.AssertExpectations(t)
		responderBroadConn.AssertExpectations(t)

		// Assert that the broadcaster sent the discovery message "pelotari?"
		wantWrittenMsgs := [][]byte{[]byte(discoveryMessage)}
		assert.ElementsMatch(t, broadcasterBroadConn.written, wantWrittenMsgs)

		// Assert that the responder read the discovery message "pelotari?"
		wantReadMsgs := wantWrittenMsgs
		assert.ElementsMatch(t, responderBroadConn.read, wantReadMsgs)
	})
}
