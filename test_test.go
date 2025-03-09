package kio

import (
	"fmt"
	"net"
	"testing"
)

func TestParseIp(t *testing.T) {
	addr := "[::]:8080"

	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("SplitHostPort failed: %v", err)
	}
	fmt.Printf("Host: %s, Port: %s", host, port)
}

func TestLinkListenerBacklog(t *testing.T) {
	fmt.Println(listenerBacklog())
}
