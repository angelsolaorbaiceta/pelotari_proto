package prototari

import (
	"net"
	"time"

	"github.com/stretchr/testify/mock"
)

type fakeConn struct {
	mock.Mock
}

func (conn *fakeConn) Read(b []byte) (n int, err error) {
	return 0, nil
}

func (conn *fakeConn) Write(b []byte) (n int, err error) {
	args := conn.Called(b)
	return args.Int(0), args.Error(1)
}

func (conn *fakeConn) Close() error {
	return nil
}

func (conn *fakeConn) LocalAddr() net.Addr {
	return nil
}

func (conn *fakeConn) RemoteAddr() net.Addr {
	return nil
}

func (conn *fakeConn) SetDeadline(t time.Time) error {
	return nil
}

func (conn *fakeConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (conn *fakeConn) SetWriteDeadline(t time.Time) error {
	return nil
}
