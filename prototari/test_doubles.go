package prototari

import (
	"io"
	"net"
	"time"
)

const (
	fakeBroadcastAddr = "192.168.0.255"
	failReadTimeout   = 100 * time.Millisecond
)

type fakeMsgRecord struct {
	IsUnicast bool
	From      *net.UDPAddr
	To        *net.UDPAddr
	Payload   []byte
}

type fakeBroadcastConn struct {
	writeChan chan<- fakeMsgRecord
	readChan  <-chan fakeMsgRecord

	localAddr *net.UDPAddr

	written chan<- fakeMsgRecord
}

func (fg *fakeBroadcastConn) LocalAddr() *net.UDPAddr {
	return fg.localAddr
}

func (fb *fakeBroadcastConn) Write(b []byte) (int, error) {
	msg := fakeMsgRecord{
		IsUnicast: false,
		Payload:   b,
		From:      fb.localAddr,
		To: &net.UDPAddr{
			IP:   []byte(fakeBroadcastAddr),
			Port: BroadcastPort,
		},
	}

	if fb.writeChan != nil {
		fb.writeChan <- msg
		fb.written <- msg
		return len(b), nil
	}

	return 0, nil
}

func (fb *fakeBroadcastConn) Read(b []byte) (int, *net.UDPAddr, error) {
	if fb.readChan == nil {
		time.Sleep(failReadTimeout)
		return 0, nil, io.EOF
	}

	// The goroutine might get stuck here waiting for a new message to be sent
	// to the channel. We want to finish gracefully when the read channel is
	// closed, and so the ok is handled to return an error.
	message, ok := <-fb.readChan
	if !ok {
		return 0, nil, io.EOF
	}
	n := copy(b, message.Payload)
	return n, message.From, nil
}

type fakeUnicastConn struct {
	writeChan chan<- fakeMsgRecord
	readChan  <-chan fakeMsgRecord

	localAddr *net.UDPAddr

	written chan<- fakeMsgRecord
}

func (fu *fakeUnicastConn) LocalAddr() *net.UDPAddr {
	return fu.localAddr
}

func (fu *fakeUnicastConn) Write(b []byte, to *net.UDPAddr) (int, error) {
	msg := fakeMsgRecord{
		IsUnicast: true,
		Payload:   b,
		From:      fu.localAddr,
		To:        to,
	}

	if fu.writeChan != nil {
		fu.writeChan <- msg
		fu.written <- msg
		return len(b), nil
	}

	return 0, nil
}

func (fu *fakeUnicastConn) Read(b []byte) (int, *net.UDPAddr, error) {
	if fu.readChan == nil {
		time.Sleep(failReadTimeout)
		return 0, nil, io.EOF
	}

	// The goroutine might get stuck here waiting for a new message to be sent
	// to the channel. We want to finish gracefully when the read channel is
	// closed, and so the ok is handled to return an error.
	message, ok := <-fu.readChan
	if !ok {
		return 0, nil, io.EOF
	}
	n := copy(b, message.Payload)
	return n, message.From, nil
}
