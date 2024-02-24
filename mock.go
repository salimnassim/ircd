package ircd

import (
	"net"
	"time"
)

type connMock struct {
	deadline time.Time
}

func (c *connMock) Read(b []byte) (n int, err error)   { return 0, nil }
func (c *connMock) SetDeadline(t time.Time) error      { c.deadline = t; return nil }
func (c *connMock) SetReadDeadline(t time.Time) error  { c.deadline = t; return nil }
func (c *connMock) SetWriteDeadline(t time.Time) error { c.deadline = t; return nil }
func (c *connMock) Write(b []byte) (int, error)        { return 0, nil }
func (c *connMock) Close() error                       { return nil }
func (C *connMock) LocalAddr() net.Addr {
	return &net.TCPAddr{
		IP:   net.IPv4(0x7F, 0x00, 0x00, 0x01),
		Port: 5000,
	}
}
func (C *connMock) RemoteAddr() net.Addr {
	return &net.TCPAddr{
		IP:   net.IPv4(0x7F, 0x00, 0x00, 0x01),
		Port: 5001,
	}
}
