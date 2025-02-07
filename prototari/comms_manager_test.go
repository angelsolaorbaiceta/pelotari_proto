package prototari

import (
	"io"
	"log"
	"net"
	"testing"
	"time"

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
		responderBroadAddr = net.UDPAddr{
			IP:   []byte(responderIP),
			Port: 46799,
		}
		broadcasterUniAddr = net.UDPAddr{
			IP:   []byte(broadcasterIP),
			Port: 24567,
		}
		responderUniAddr = net.UDPAddr{
			IP:   []byte(responderIP),
			Port: 14567,
		}
	)

	originalOutput := log.Writer()
	log.SetOutput(io.Discard)
	defer func() { log.SetOutput(originalOutput) }()

	t.Run("Successful handshake", func(t *testing.T) {
		var (
			writtenMsgsChan      = make(chan fakeMsgRecord) // For synchronization
			broadCommsChan       = make(chan fakeMsgRecord)
			broadToRespCommsChan = make(chan fakeMsgRecord)
			respToBroadCommsChan = make(chan fakeMsgRecord)

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
				localAddr: &responderBroadAddr,
			}
			broadcasterUnicConn = fakeUnicastConn{
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
				makeTestingConfig(),
			)
			responder = MakeManager(
				&responderBroadConn,
				&responderUnicConn,
				makeTestingConfig(),
			)
			got, want fakeMsgRecord
		)

		broadcaster.Start()
		responder.Start()
		defer func() {
			close(broadCommsChan)
			close(broadToRespCommsChan)
			close(respToBroadCommsChan)
			close(writtenMsgsChan)

			broadcaster.Stop()
			responder.Stop()
		}()

		// Wait for the broadcast message to be sent by the broadcaster
		// BROADCASTER --> EVERYONE
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

		// Wait for the response message to be sent by the responder
		// RESPONDER --> BROADCASTER
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

		// Wait for the confirmatio message to be sent by the broadcaster
		// BROADCASTER --> RESPONDER
		got = <-writtenMsgsChan
		want = fakeMsgRecord{
			IsUnicast: true,
			From:      &broadcasterUniAddr,
			To: &net.UDPAddr{
				IP:   []byte(responderIP),
				Port: UnicastPort,
			},
			Payload: []byte(confirmationMessage),
		}
		assert.Equal(t, want, got)

		// Check that the peer is correctly registered in the broadcaster
		wantPeer := Peer{
			IP: []byte(responderIP),
		}
		gotPeer := broadcaster.Peers()[0]
		assert.True(t, wantPeer.Equal(gotPeer))
	})

	t.Run("Broadcaster ignores its own messages", func(t *testing.T) {
		// The broadcaster writes to and reads from the same channel.
		// The broadcaster should ignore its own message and not respond to it.
		// There should be no unicast response from it.
		var (
			broadCh         = make(chan fakeMsgRecord)
			writtenMsgsChan = make(chan fakeMsgRecord)
			broadConn       = fakeBroadcastConn{
				writeChan: broadCh,
				readChan:  broadCh,
				localAddr: &broadcasterBroadAddr,
				written:   writtenMsgsChan,
			}
			unicConn = fakeUnicastConn{
				writeChan: nil,
				readChan:  nil,
				localAddr: &broadcasterUniAddr,
				written:   writtenMsgsChan,
			}
			broadcaster = MakeManager(&broadConn, &unicConn, makeTestingConfig())

			want, got fakeMsgRecord
		)

		broadcaster.Start()
		defer func() {
			close(broadCh)
			close(writtenMsgsChan)
			broadcaster.Stop()
		}()

		// Wait for the broadcast message to be sent by the broadcaster
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

		// Make sure that no other message is sent
		select {
		case msg := <-writtenMsgsChan:
			assert.FailNow(t, "A message was sent", string(msg.Payload))
		case <-time.After(100 * time.Millisecond):
			// Test passes. No message received in the timeout.
		}
	})
}
