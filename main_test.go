package main

import (
	"bitcoin/net"
	"bytes"
	"log"
	"testing"
)

func TestRunClient(t *testing.T) {
	c := net.NewClient(net.ClientTypeOut, "54.36.172.26:8333")
	c.Sync(&net.ClientListener{
		OnConnected: func() {
			log.Println("OnConnected")
		},
		OnClosed: func() {
			log.Println("OnClosed")
		},
		OnLoop: func() {
			//
		},
		OnMessage: func(m net.MsgIO) {
			log.Println(m.Command())
			if m.Command() == net.NMT_VERACK {
				c.WriteMsg(net.NewMsgSendHeaders())
			}
		},
	})
}

type X [20]byte

func TestCopy(t *testing.T) {
	var a []byte
	var b []byte
	log.Println(bytes.Compare(a, b))
}
