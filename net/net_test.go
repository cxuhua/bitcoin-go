package net

import (
	"testing"
)

func TestVersionMessage(t *testing.T) {
	c := NewClient(ClientTypeOut, "172.81.183.236:8333")
	c.run()
}
