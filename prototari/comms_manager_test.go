package prototari

import (
	"testing"

	"github.com/stretchr/testify/mock"
)

func TestCommsManager(t *testing.T) {
	t.Run("Send a broadcast message", func(t *testing.T) {
		var (
			broadcastConn = new(fakeConn)
			manager       = MakeManager(broadcastConn)
			doneCh        = make(chan struct{})
		)

		broadcastConn.
			On("Write", []byte(discoveryMessage)).
			Run(func(args mock.Arguments) {
				doneCh <- struct{}{}
			}).
			Return(len(discoveryMessage), nil)

		go manager.StartDiscovery()
		defer manager.StopDiscovery()

		<-doneCh

		broadcastConn.AssertExpectations(t)
	})
}
