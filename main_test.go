package main

import (
	"bitcoin/net"
	"log"
	"testing"
	"time"
)

func TestRunClient(t *testing.T) {
	c := net.NewClient(net.ClientTypeOut, "127.0.0.1:8333")
	c.Run()
	time.Sleep(time.Second * 3)
	//m := NewMsgGetBlocks()
	m := net.NewMsgGetData()
	m.Add(&net.Inventory{
		Type: net.MSG_TX,
		ID:   net.NewHexBHash("e38ae39eb2ad68fba8d3ba384925ae5a1c1ebdc5aca81e43a9a84ee4ec6236c6"),
	})
	c.WriteMsg(m)

	time.Sleep(time.Hour)
}

func TestCopy(t *testing.T) {
	slice1 := []int{1, 2, 3, 4, 5}
	slice2 := []int{5, 4, 3, 4, 6, 7, 8, 8}

	i := copy(slice2, slice1)
	log.Println(i)
}
