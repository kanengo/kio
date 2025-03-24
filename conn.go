package kio

import (
	"net"
	"sync/atomic"

	"github.com/kanengo/ku/bufferx/ring"
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

	inBuffer *ring.Buffer
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
