package kio

import (
	"fmt"
	"log/slog"
	"net"
	"strconv"

	"golang.org/x/sys/unix"
)

const (
	MaxSocktSendBufSize = 1024 * 1024 * 4
	MaxSocktReadBufSize = 1024 * 1024 * 4
)

const (
	NetworkTCP  = "tcp" //contains both ipv4 and ipv6
	NetworkTCP4 = "tcp4"
	NetworkTCP6 = "tcp6"
	NetworkUDP  = "udp" // both ipv4 and ipv6
	NetworkUDP4 = "udp4"
	NetworkUDP6 = "udp6"
	NetworkUNIX = "unix"
)

type Address struct {
	Network string
	Address string
	Name    string

	sa unix.Sockaddr
}

type Config struct {
	ListenAddrs   []Address
	PollerNum     int
	ThreadPoolNum int
	logger        *slog.Logger
}

//go:norace
func (a Address) normalize(isIpv6 bool) (domain, sotype, proto int) {
	domain = unix.AF_INET6
	sotype = unix.SOCK_STREAM
	proto = unix.IPPROTO_IP

	switch a.Network {
	case NetworkTCP, NetworkTCP4, NetworkTCP6:
		if a.Network == NetworkTCP4 || !isIpv6 {
			domain = unix.AF_INET
		}
		proto = unix.IPPROTO_TCP
	case NetworkUDP, NetworkUDP4, NetworkUDP6:
		if a.Network == NetworkUDP4 || !isIpv6 {
			domain = unix.AF_INET
		}
		sotype = unix.SOCK_DGRAM
		proto = unix.IPPROTO_UDP
	case NetworkUNIX:
		domain = unix.AF_UNIX
		sotype = unix.SOCK_STREAM
	}
	fmt.Println(domain, sotype, proto)
	return domain, sotype | unix.SOCK_CLOEXEC | unix.SOCK_NONBLOCK, proto
}

//go:norace
func (a Address) isIpv6Only() bool {
	return a.Network == NetworkTCP6 || a.Network == NetworkUDP6
}

//go:norace
func (a Address) HostPort() (host string, port int, err error) {
	host, portStr, err := net.SplitHostPort(a.Address)
	if err != nil {
		return
	}
	port, err = strconv.Atoi(portStr)
	return host, port, err
}
