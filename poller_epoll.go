//go:build linux

package kio

import (
	"log/slog"
	"net"
	"strconv"
	"sync"
	"syscall"
	_ "unsafe"
)

const (
	// EPOLLET .
	EPOLLET = 0x80000000
)

type poller struct {
	*Engine
	IsListener  bool
	listenerFds []int

	connections map[int]*Conn
	mux         sync.RWMutex
	efd         int
	evfd        int
}

func newPoller(e *Engine, isListener bool) (p *poller, err error) {
	var efd int
	efd, err = syscall.EpollCreate1(0)
	if err != nil {
		slog.Error("start listener failed", "err", err)
		return nil, err
	}

	defer func() {
		if err != nil {
			syscall.Close(efd)
			for _, fd := range p.listenerFds {
				syscall.Close(fd)
			}
		}
	}()
	if isListener {
		domain := syscall.AF_INET6
		sotype := syscall.SOCK_STREAM
		proto := syscall.IPPROTO_IP

		isIpv6 := true
		switch p.NetWork {
		case NetworkTCP, NetworkTCP4, NetworkTCP6:
			if p.NetWork == NetworkTCP4 {
				domain = syscall.AF_INET
				isIpv6 = false
			}
		case NetworkUDP, NetworkUDP4, NetworkUDP6:
			if p.NetWork == NetworkUDP4 {
				domain = syscall.AF_INET
				isIpv6 = false
			}
			sotype = syscall.SOCK_DGRAM
		case NetworkUNIX:
			isIpv6 = false
			domain = syscall.AF_UNIX
			sotype = syscall.SOCK_STREAM
		}

		for _, addr := range p.ListenAddrs {
			var listenerFd int
			listenerFd, err = syscall.Socket(domain, sotype, proto)
			if err != nil {
				break
			}
			err = syscall.SetNonblock(listenerFd, true)
			if err != nil {
				break
			}
			if p.NetWork == NetworkTCP || p.NetWork == NetworkUDP {
				err = syscall.SetsockoptInt(listenerFd, syscall.AF_INET6, syscall.IPV6_V6ONLY, 0)
				if err != nil {
					break
				}
			}
			err = syscall.SetsockoptInt(listenerFd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
			if err != nil {
				break
			}
			var host, port string
			host, port, err = net.SplitHostPort(addr)
			if err != nil {
				break
			}
			var portInt int
			portInt, err = strconv.Atoi(port)
			if err != nil {
				break
			}
			var sockAddr syscall.Sockaddr
			if p.NetWork == NetworkUNIX {
				sa := &syscall.SockaddrUnix{
					Name: addr,
				}
				sockAddr = sa
			} else {
				if isIpv6 {
					sa := &syscall.SockaddrInet6{
						Port: portInt,
					}
					copy(sa.Addr[:], net.ParseIP(host).To16())
					sockAddr = sa
				} else {
					sa := &syscall.SockaddrInet4{
						Port: portInt,
					}
					copy(sa.Addr[:], net.ParseIP(host).To4())
					sockAddr = sa
				}
			}
			err = syscall.Bind(listenerFd, sockAddr)
			if err != nil {
				break
			}
			err = syscall.Listen(listenerFd, listenerBacklog())
			if err != nil {
				break
			}
			p.listenerFds = append(p.listenerFds, listenerFd)
		}

		if err != nil {
			slog.Error("start listener failed", "err", err)
			return nil, err
		}
	}

	r0, _, e0 := syscall.Syscall(syscall.SYS_EVENTFD2, 0, syscall.O_NONBLOCK, 0)
	if e0 != 0 {
		slog.Error("Syscall", "err", e0)
		syscall.Close(efd)
		return nil, err
	}
	err = syscall.EpollCtl(efd, syscall.EPOLL_CTL_ADD, int(r0), &syscall.EpollEvent{
		Events: syscall.EPOLLIN,
		Fd:     int32(r0),
	})
	if err != nil {
		syscall.Close(int(r0))
		return nil, err
	}
	for _, listenerFd := range p.listenerFds {
		err = syscall.EpollCtl(efd, syscall.EPOLL_CTL_ADD, listenerFd, &syscall.EpollEvent{
			Events: syscall.EPOLLIN | EPOLLET,
			Fd:     int32(listenerFd),
		})
		if err != nil {
			syscall.Close(efd)
			for _, fd := range p.listenerFds {
				syscall.Close(fd)
			}
			return nil, err
		}
	}

	return &poller{
		Engine:      e,
		IsListener:  isListener,
		listenerFds: make([]int, 0, len(e.ListenAddrs)),

		efd:  efd,
		evfd: efd,
	}, nil
}

func (p *poller) start() {

}

//go:linkname listenerBacklog net.listenerBacklog
func listenerBacklog() int
