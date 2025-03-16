package kio

import "net"

type Conn struct {
	Name       string
	Fd         int
	RemoteAddr net.Addr
}
