package kio

import (
	"testing"
	"time"
)

func TestEngineNew(t *testing.T) {
	e, err := NewEngine(Config{
		PollerNum:     2,
		ThreadPoolNum: 2,
		ListenAddrs: []Address{
			{
				Network: NetworkTCP,
				Address: "127.0.0.1:8080",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()

	time.Sleep(10 * time.Second)
}
