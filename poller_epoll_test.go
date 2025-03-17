//go:build linux

package kio

import (
	"testing"
)

func TestNewPoller(t *testing.T) {

	p, err := newPoller(nil, true)
	if err != nil {
		t.Fatal(err)
	}
	defer p.close()
}
