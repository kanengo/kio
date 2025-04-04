package kio

import (
	"fmt"
	"net"
	"testing"
	"time"
)

func TestParseIp(t *testing.T) {
	addr := "[::]:8080"

	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("SplitHostPort failed: %v", err)
	}
	fmt.Printf("Host: %s, Port: %s\n", host, port)

	ta, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		t.Fatalf("ResolveTCPAddr failed: %v", err)
	}
	fmt.Println("-----------------------------")

	fmt.Println(ta.IP, ta.Port, ta.IP.To16(), ta.IP.To4())
}

// func TestLinkListenerBacklog(t *testing.T) {
// 	fmt.Println(listenerBacklog())
// }

func TestPrintSomething(t *testing.T) {
	fmt.Println(8 & 7)
	fmt.Println(1 << 10)
	fmt.Println(int64(time.Now().UnixNano() % 5))
}
