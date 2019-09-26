package net

import (
	"net"
	"testing"
)

func TestVersionMessage(t *testing.T) {
	c := NewClient(ClientTypeOut, net.ParseIP("172.81.183.236"), 8333)
	c.run()
}
