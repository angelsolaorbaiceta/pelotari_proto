package prototari

import (
	"net"
)

const fakeBroadcastAddr = "192.168.0.255"

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
	if fb.writeChan == nil {
		return 0, nil
	}

	msg := fakeMsgRecord{
		IsUnicast: false,
		Payload:   b,
		From:      fb.localAddr,
		To: &net.UDPAddr{
			IP:   []byte(fakeBroadcastAddr),
			Port: BroadcastPort,
		},
	}
	fb.written <- msg
	fb.writeChan <- msg

	return len(b), nil
}

func (fb *fakeBroadcastConn) Read(b []byte) (int, *net.UDPAddr, error) {
	message := <-fb.readChan
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
	fu.written <- msg
	fu.writeChan <- msg

	return len(b), nil
}

func (fu *fakeUnicastConn) Read(b []byte) (int, *net.UDPAddr, error) {
	message := <-fu.readChan
	n := copy(b, message.Payload)

	return n, message.From, nil
}
