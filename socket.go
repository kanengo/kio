//go:build linux || freebsd || darwin

package kio

import (
	"net"

	"golang.org/x/sys/unix"
)

func GetSockAddr(addr Address) (sa unix.Sockaddr, isIpv6 bool, err error) {
	if addr.Network == NetworkUNIX {
		sa = &unix.SockaddrUnix{
			Name: addr.Address,
		}
		return
	}

	isIpv6 = true
	switch addr.Network {
	case NetworkTCP, NetworkTCP4, NetworkTCP6:
		var ta *net.TCPAddr
		ta, err = net.ResolveTCPAddr(addr.Network, addr.Address)
		if err != nil {
			return
		}
		if ta.IP.To4() != nil {
			isIpv6 = false
			sa4 := &unix.SockaddrInet4{
				Port: ta.Port,
			}
			copy(sa4.Addr[:], ta.IP.To4())
			sa = sa4
		} else {
			sa6 := &unix.SockaddrInet6{
				Port: ta.Port,
			}
			copy(sa6.Addr[:], ta.IP.To16())
			sa = sa6
		}
	default: //udp
		var ua *net.UDPAddr
		ua, err = net.ResolveUDPAddr(addr.Network, addr.Address)
		if err != nil {
			return
		}
		if ua.IP.To4() != nil {
			isIpv6 = false
			sa4 := &unix.SockaddrInet4{
				Port: ua.Port,
			}
			copy(sa4.Addr[:], ua.IP.To4())
			sa = sa4
		} else {
			sa6 := &unix.SockaddrInet6{
				Port: ua.Port,
			}
			copy(sa6.Addr[:], ua.IP.To16())
			sa = sa6
		}
	}

	return
}

// SockaddrToString 将 unix.Sockaddr 转换为对应的 IP 地址字符串
// 对于 IPv4 和 IPv6 地址，返回格式为 "IP:Port"
// 对于 Unix 域套接字，返回套接字路径
func SockaddrToString(sa unix.Sockaddr) net.Addr {
	if sa == nil {
		return nil
	}

	switch sa := sa.(type) {
	case *unix.SockaddrInet4:
		return &net.TCPAddr{
			IP:   sa.Addr[:],
			Port: int(sa.Port),
		}
	case *unix.SockaddrInet6:
		return &net.TCPAddr{
			IP:   sa.Addr[:],
			Port: int(sa.Port),
		}
	case *unix.SockaddrUnix:
		return &net.UnixAddr{
			Net:  "unix",
			Name: sa.Name,
		}
	default:
		return nil
	}
}
