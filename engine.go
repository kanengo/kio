package kio

import (
	"net"
	"runtime"
)

type Config struct {
	ListenPort  int
	ListenAddrs []string
	PollerNum   int
	ThreadLock  bool
	AsyncRead   bool

	Listen func(network, address string) (net.Listener, error)
}

const (
	NetworkTCP        = "tcp"
	NetworkTCP4       = "tcp4"
	NetworkTCP6       = "tcp6"
	NetworkUDP        = "udp"
	NetworkUDP4       = "udp4"
	NetworkUDP6       = "udp6"
	NetworkUNIX       = "unix"
	NetworkUNIXGRAM   = "unixgram"
	NetworkUNIXPACKET = "unixpacket"
)

type Engine struct {
	Config

	NetWork   string
	listeners []*poller
	pollers   []*poller
}

//go:norace
func NewEngine(config Config) *Engine {
	if config.Listen == nil {
		config.Listen = net.Listen
	}
	if config.PollerNum <= 0 {
		config.PollerNum = runtime.NumCPU()
	}
	return &Engine{
		Config: config,
	}
}

//go:norace
func (e *Engine) Start() error {

	return nil
}
