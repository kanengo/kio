//go:build linux

package kio

import (
	"testing"
)

func TestNewPoller(t *testing.T) {
	e := NewEngine(Config{
		PollerNum:     1,
		ThreadPoolNum: 1,
		ListenAddrs: []Address{
			{
				Network: NetworkTCP,
				Address: "[::]:8080",
			},
		},
	})
	p, err := newPoller(e, true)
	if err != nil {
		t.Fatal(err)
	}
	defer p.close()
}
