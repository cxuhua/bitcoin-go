package main

import (
	"bitcoin/net"
	"testing"
	"time"
)

func TestRunClient(t *testing.T) {
	c := net.NewClient(net.ClientTypeOut, "101.201.211.33:8333")
	c.Run()
	time.Sleep(time.Second * 3)
	//m := NewMsgGetBlocks()
	m := net.NewMsgGetData()
	m.Add(&net.Inventory{
		Type: net.MSG_BLOCK,
		ID:   net.NewHexBHash("000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f").Swap(),
	})
	c.WC <- m

	time.Sleep(time.Hour)
}
