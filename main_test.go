package main

import (
	"bitcoin/net"
	"bytes"
	"log"
	"testing"
	"time"
)

func TestRunClient(t *testing.T) {
	c := net.NewClient(net.ClientTypeOut, "127.0.0.1:8333")
	c.Run()
	time.Sleep(time.Second * 30)
	if c.IsConnected() {
		//m := NewMsgGetBlocks()
		m := net.NewMsgGetData()
		m.Add(&net.Inventory{
			Type: net.MSG_TX,
			ID:   net.NewHexBHash("e38ae39eb2ad68fba8d3ba384925ae5a1c1ebdc5aca81e43a9a84ee4ec6236c6"),
		})
		c.WriteMsg(m)

		time.Sleep(time.Hour)
	} else {
		log.Println("connected failed ")
	}
}

type X [20]byte

func TestCopy(t *testing.T) {
	var a []byte
	var b []byte
	log.Println(bytes.Equal(a, b))
}
