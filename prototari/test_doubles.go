package prototari

import (
	"net"

	"github.com/stretchr/testify/mock"
)

type fakeBroadcastConn struct {
	mock.Mock
	writeChan chan<- []byte
	readChan  <-chan []byte

	written [][]byte
	read    [][]byte
}

func (fb *fakeBroadcastConn) Write(b []byte) (int, error) {
	args := fb.Called(b)
	if args.Error(1) == nil && fb.writeChan != nil {
		fb.writeChan <- b
		fb.written = append(fb.written, b)
	}

	return args.Int(0), args.Error(1)
}

func (fb *fakeBroadcastConn) Read(b []byte) (int, *net.UDPAddr, error) {
	args := fb.Called(b)
	if args.Error(2) == nil && fb.readChan != nil {
		messageBytes := <-fb.readChan
		n := copy(b, messageBytes)
		fb.read = append(fb.read, messageBytes)

		return n, args.Get(1).(*net.UDPAddr), args.Error(2)
	}

	return args.Int(0), args.Get(1).(*net.UDPAddr), args.Error(2)
}
