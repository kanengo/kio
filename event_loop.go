package kio

import (
	"errors"
	"io"
	"os"
	"sync"
	"sync/atomic"

	"golang.org/x/sys/unix"
)

const (
	defaultBufSize = 1024 * 64
)

type eventLoop struct {
	e      *Engine
	poller *poller

	listeners map[int]*listener

	mux         sync.RWMutex
	connections map[int]*Conn
	connNum     atomic.Int64

	readBufSize  int
	maxReadChunk int

	buffer []byte
}

//go:norace
func newEventLoop(e *Engine, isListener bool) (el *eventLoop, err error) {
	p, err := newPoller(e, isListener)
	if err != nil {
		return
	}

	el = &eventLoop{
		e:            e,
		poller:       p,
		connections:  make(map[int]*Conn),
		listeners:    make(map[int]*listener),
		maxReadChunk: 1024 * 1024,
	}
	if e.Config.ReadBufSize <= 0 {
		e.Config.ReadBufSize = defaultBufSize
	}
	el.buffer = make([]byte, e.Config.ReadBufSize)

	if isListener {
		defer func() {
			if err != nil {
				p.close()
			}
		}()
		for _, addr := range e.Config.ListenAddrs {
			var listenerFd int
			var isIpv6 bool
			var sa unix.Sockaddr
			sa, isIpv6, err = GetSockAddr(addr)
			if err != nil {
				return
			}

			listenerFd, err = unix.Socket(addr.normalize(isIpv6))
			if err != nil {
				return
			}
			err = unix.SetsockoptInt(listenerFd, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
			if err != nil {
				return
			}
			if err = unix.SetsockoptInt(listenerFd, unix.SOL_SOCKET, unix.SO_RCVBUF, MaxSocktReadBufSize); err != nil {
				return
			}
			if err = unix.SetsockoptInt(listenerFd, unix.SOL_SOCKET, unix.SO_SNDBUF, MaxSocktSendBufSize); err != nil {
				return
			}
			if err = unix.SetNonblock(listenerFd, true); err != nil {
				return
			}

			if isIpv6 {
				if addr.isIpv6Only() {
					err = unix.SetsockoptInt(listenerFd, unix.IPPROTO_IPV6, unix.IPV6_V6ONLY, 1)
					if err != nil {
						return
					}
				}
			}

			err = unix.Bind(listenerFd, sa)
			if err != nil {
				return
			}
			err = unix.Listen(listenerFd, listenerBacklog())
			if err != nil {
				return
			}
			if err = p.addRead(listenerFd, false); err != nil {
				return
			}

			el.listeners[listenerFd] = &listener{
				name: addr.Name,
				fd:   listenerFd,
			}
		}
	}

	return
}

//go:norace
func (el *eventLoop) addConn(conn *Conn) error {
	if err := el.poller.addReadWrite(conn.fd, true); err != nil {
		return err
	}
	el.mux.Lock()
	el.connections[conn.fd] = conn
	el.mux.Unlock()
	el.connNum.Add(1)

	conn.opened = true

	eh := el.e.eventHandlers[conn.Name]
	if eh != nil {
		eh.OnOpen(conn)
	}

	return nil
}

//go:norace
func (el *eventLoop) accept(ev unix.EpollEvent) error {
	nfd, sa, err := unix.Accept4(int(ev.Fd), unix.SOCK_NONBLOCK|unix.SOCK_CLOEXEC)
	switch err {
	case nil:
	case unix.EAGAIN, unix.EINTR, unix.ECONNRESET, unix.ECONNABORTED:
		// ECONNRESET or ECONNABORTED could indicate that a socket
		// in the Accept queue was closed before we Accept()ed it.
		// It's a silly error, let's retry it.
		return nil
	default:
		return err
	}

	// 创建新连接
	conn := &Conn{
		Name:       el.listeners[int(ev.Fd)].name,
		fd:         nfd,
		RemoteAddr: SockaddrToString(sa),
	}

	if err := el.e.addConn(conn); err != nil {
		unix.Close(nfd)
		return err
	}

	return nil
}

func (el *eventLoop) getConn(fd int) (*Conn, bool) {
	el.mux.RLock()
	c, ok := el.connections[fd]
	el.mux.RUnlock()
	return c, ok
}

//go:norace
func (el *eventLoop) handleReadWrite(ev unix.EpollEvent) error {
	conn, ok := el.getConn(int(ev.Fd))
	if !ok {
		return nil
	}
	return conn.processIo(ev)
}

func (el *eventLoop) read(c *Conn) (err error) {
	if !c.opened {
		return nil
	}

	var recv int
	chunk := el.maxReadChunk
	handler := el.e.eventHandlers[c.Name]
	for {
		n, err := unix.Read(c.fd, el.buffer)
		if err != nil || n == 0 {
			if errors.Is(err, unix.EINTR) {
				continue
			}
			if errors.Is(err, unix.EAGAIN) {
				return nil
			}
			if n == 0 {
				err = io.EOF
			}

			return el.close(c, os.NewSyscallError("read", err))
		}
		recv += n
		if handler != nil {
			handler.OnData(c, el.buffer[:n])
		}

		if recv >= chunk || n < len(el.buffer) {
			break
		}
	}

	return nil
}

//go:norace
func (el *eventLoop) start() {

}

func (el *eventLoop) close(c *Conn, err error) error {
	el.mux.Lock()
	delete(el.connections, c.fd)
	el.mux.Unlock()
	return c.close(err)
}

//go:norace
func (el *eventLoop) stop() {
	el.poller.close()
}
