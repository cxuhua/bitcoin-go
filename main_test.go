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
		ID:   net.NewHexBHash("0000000000000000002a2451180749294cd74058e0a0dd37cc19ad0ee66e77ff").Swap(),
	})
	c.WC <- m

	time.Sleep(time.Hour)
}
