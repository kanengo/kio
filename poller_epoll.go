//go:build linux

package kio

import (
	"errors"
	"log/slog"
	"os"
	_ "unsafe"

	"github.com/kanengo/kio/errorx"
	"golang.org/x/sys/unix"
)

const (
	// InitPollEventsCap represents the initial capacity of poller event-list.
	InitPollEventsCap = 128
	// MaxPollEventsCap is the maximum limitation of events that the poller can process.
	MaxPollEventsCap = 1024
	// MinPollEventsCap is the minimum limitation of events that the poller can process.
	MinPollEventsCap = 32

	ReadEvents        = unix.EPOLLIN | unix.EPOLLPRI
	WriteEvents       = unix.EPOLLOUT
	ReadWriteEvents   = ReadEvents | WriteEvents
	ErrEvents         = unix.EPOLLERR | unix.EPOLLHUP
	EdgeTriggerEvents = unix.EPOLLET | unix.EPOLLRDHUP
)

type poller struct {
	e *Engine

	epfd int
	evfd int
}

type eventList struct {
	size   int
	events []unix.EpollEvent
}

//go:norace
func newEventList(size int) *eventList {
	return &eventList{
		size:   size,
		events: make([]unix.EpollEvent, size),
	}
}

//go:norace
func (el *eventList) expand() {
	if newSize := el.size << 1; newSize <= MaxPollEventsCap {
		el.size = newSize
		el.events = make([]unix.EpollEvent, newSize)
	}
}

//go:norace
func (el *eventList) shrink() {
	if newSize := el.size >> 1; newSize >= MinPollEventsCap {
		el.size = newSize
		el.events = make([]unix.EpollEvent, newSize)
	}
}

//go:norace
func newPoller(e *Engine, isListener bool) (p *poller, err error) {
	var efd int
	efd, err = unix.EpollCreate1(unix.EPOLL_CLOEXEC)
	if err != nil {
		slog.Error("start listener failed", "err", err)
		return nil, err
	}
	defer func() {
		if err != nil {
			unix.Close(efd)
		}
	}()

	p = &poller{
		e:    e,
		epfd: efd,
	}

	var e0 int
	e0, err = unix.Eventfd(0, unix.EFD_NONBLOCK|unix.EFD_CLOEXEC)
	if err != nil {
		return
	}
	err = p.addRead(e0, true)
	if err != nil {
		unix.Close(e0)
		return
	}
	p.evfd = e0

	return p, nil
}

//go:norace
func (p *poller) polling(handler func(ev unix.EpollEvent) error) error {
	el := newEventList(InitPollEventsCap)

	msec := -1
	for {
		n, err := unix.EpollWait(p.epfd, el.events, msec)
		if err != nil {
			p.e.logger.Error("epoll wait failed", "err", err)
			return err
		}
		for i := range n {
			ev := el.events[i]
			if err := handler(ev); err != nil {
				p.e.logger.Error("handle event failed", "err", err)
				if errors.Is(err, errorx.ErrorAcceptSocket) {
					return err
				}
			}
		}
	}
}

func (p *poller) close() {
	unix.Close(p.evfd)
	unix.Close(p.epfd)
}

func (p *poller) addRead(fd int, et bool) error {
	var ev unix.EpollEvent
	ev.Events = ReadEvents | ErrEvents
	if et {
		ev.Events |= EdgeTriggerEvents
	}
	return os.NewSyscallError("epoll_ctl add", unix.EpollCtl(p.epfd, unix.EPOLL_CTL_ADD, fd, &ev))
}

func (p *poller) addWrite(fd int, et bool) error {
	var ev unix.EpollEvent
	ev.Events = WriteEvents | ErrEvents
	if et {
		ev.Events |= EdgeTriggerEvents
	}
	return os.NewSyscallError("epoll_ctl add", unix.EpollCtl(p.epfd, unix.EPOLL_CTL_ADD, fd, &ev))
}

func (p *poller) addReadWrite(fd int, et bool) error {
	var ev unix.EpollEvent
	ev.Events = ReadEvents | WriteEvents | ErrEvents
	if et {
		ev.Events |= EdgeTriggerEvents
	}
	return os.NewSyscallError("epoll_ctl add", unix.EpollCtl(p.epfd, unix.EPOLL_CTL_ADD, fd, &ev))
}

//go:linkname listenerBacklog net.listenerBacklog
func listenerBacklog() int
