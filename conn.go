package kio

import (
	"net"
	"sync/atomic"

	"golang.org/x/sys/unix"
)

type Conn struct {
	Name       string
	fd         int
	RemoteAddr net.Addr

	opened bool

	closed atomic.Bool

	closeErr error

	el *eventLoop

	buffer []byte
}

func (c *Conn) close(err error) error {
	if !c.closed.CompareAndSwap(false, true) {
		return nil
	}
	c.opened = false

	err = unix.Close(c.fd)
	c.closeErr = err

	return nil
}
