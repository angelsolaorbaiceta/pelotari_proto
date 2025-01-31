package prototari

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommsManager(t *testing.T) {
	// In this test we configure two broadcasters that communicate through channels.
	// Whatever the broadcaster writes in the broadcast connection, the responder
	// receives, but the opposite is not true. This is to emulate the discovery
	// from one peer to another in a single direction, which simplifies the test.
	var (
		broadcasterIP        = "192.168.0.10"
		responderIP          = "192.168.0.20"
		broadcasterBroadAddr = net.UDPAddr{
			IP:   []byte(broadcasterIP),
			Port: 45678,
		}
		responderUniAddr = net.UDPAddr{
			IP:   []byte(responderIP),
			Port: 14567,
		}
		broadcasterUniAddr = net.UDPAddr{
			IP:   []byte(responderIP),
			Port: 24567,
		}

		// A channel used for synchronization of the goroutines and assert that
		// the expected messages have been exchanged between peers.
		writtenMsgsChan = make(chan fakeMsgRecord)

		broadCommsChan       = make(chan fakeMsgRecord)
		broadcasterBroadConn = fakeBroadcastConn{
			writeChan: broadCommsChan,
			readChan:  nil,
			written:   writtenMsgsChan,
			localAddr: &broadcasterBroadAddr,
		}
		responderBroadConn = fakeBroadcastConn{
			writeChan: nil, // Responder shouldn't broadcast anything
			written:   nil,
			readChan:  broadCommsChan,
			localAddr: nil,
		}

		broadToRespCommsChan = make(chan fakeMsgRecord)
		respToBroadCommsChan = make(chan fakeMsgRecord)
		broadcasterUnicConn  = fakeUnicastConn{
			writeChan: broadToRespCommsChan,
			readChan:  respToBroadCommsChan,
			written:   writtenMsgsChan,
			localAddr: &broadcasterUniAddr,
		}
		responderUnicConn = fakeUnicastConn{
			writeChan: respToBroadCommsChan,
			readChan:  broadToRespCommsChan,
			written:   writtenMsgsChan,
			localAddr: &responderUniAddr,
		}

		broadcaster = MakeManager(
			&broadcasterBroadConn,
			&broadcasterUnicConn,
			MakeDefaultConfig(),
		)
		responder = MakeManager(
			&responderBroadConn,
			&responderUnicConn,
			MakeDefaultConfig(),
		)
	)

	t.Run("Successful handshake", func(t *testing.T) {
		var got, want fakeMsgRecord

		go broadcaster.Start()
		go responder.Start()
		defer func() {
			broadcaster.Stop()
			responder.Stop()
		}()

		// Wait for the broadcast message to be sent
		got = <-writtenMsgsChan
		want = fakeMsgRecord{
			IsUnicast: false,
			From:      &broadcasterBroadAddr,
			To: &net.UDPAddr{
				IP:   []byte(fakeBroadcastAddr),
				Port: BroadcastPort,
			},
			Payload: []byte(discoveryMessage),
		}
		assert.Equal(t, want, got)

		// Wait for the broadcast message to be read
		got = <-writtenMsgsChan
		want = fakeMsgRecord{
			IsUnicast: true,
			From:      &responderUniAddr,
			To: &net.UDPAddr{
				IP:   []byte(broadcasterIP),
				Port: UnicastPort,
			},
			Payload: []byte(responseMessage),
		}
		assert.Equal(t, want, got)
	})
}
