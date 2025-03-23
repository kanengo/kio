//go:build linux

package kio

import (
	"io"

	"golang.org/x/sys/unix"
)

func (c *Conn) processIo(ev unix.EpollEvent) error {
	// First check for any unexpected non-IO events.
	// For these events we just close the connection directly.
	if ev.Events&(ErrEvents|unix.EPOLLRDHUP) != 0 && ev.Events&ReadEvents == 0 {
		return c.el.close(c, io.EOF)
	}

	if ev.Events&(WriteEvents|ErrEvents) != 0 {
		//write
	}

	if ev.Events&(ReadEvents|ErrEvents) != 0 {
		return c.el.read(c)
	}

	// Ultimately, check for EPOLLRDHUP, this event indicates that the remote has
	// either closed connection or shut down the writing half of the connection.
	if ev.Events&unix.EPOLLRDHUP != 0 && c.opened {
		if ev.Events&unix.EPOLLIN == 0 { // unreadable EPOLLRDHUP, close the connection directly
			return c.el.close(c, io.EOF)
		}
		// Received the event of EPOLLIN|EPOLLRDHUP, but the previous eventloop.read
		// failed to drain the socket buffer, so we ensure to get it done this time.

		//read
		return c.el.read(c)
	}

	return nil
}
