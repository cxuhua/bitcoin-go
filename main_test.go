package main

import (
	"bitcoin/net"
	"bitcoin/util"
	"encoding/hex"
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

type X [20]byte

func TestCopy(t *testing.T) {
	d, _ := hex.DecodeString("86d6bcb9d5a7e172d8f29b3b11ac62ab5c8d227209f9f54053494784873fc3")
	log.Println(hex.EncodeToString(util.HASH160(d)))
}
