package main

import (
	"bitcoin/core"
	"bitcoin/store"
	"context"
	"encoding/base64"
	"encoding/hex"
	"log"
	"testing"
)

func TestXX(t *testing.T) {
	b, _ := base64.StdEncoding.DecodeString("mC25hwpeMNjwsqTrzMWFK1oeJBPpJ0xJR7/sa9qpudcAAAAA")
	log.Println(hex.EncodeToString(b))
}

func TestRunClient(t *testing.T) {
	store.UseSession(context.Background(), func(db store.DbImp) error {
		return core.G.Init(db)
	})
	c := core.NewClient(core.ClientTypeOut, "192.168.31.198:8333")
	c.Sync(&core.ClientListener{
		OnConnected: func() {
			log.Println("OnConnected")
		},
		OnClosed: func() {
			log.Println("OnClosed", c.Err)
		},
		OnLoop: func() {
			//
		},
		OnWrite: func(m core.MsgIO) {
			log.Println("send message:", m.Command())
		},
		OnMessage: func(m core.MsgIO) {
			cmd := m.Command()
			log.Println(cmd)
			if cmd == core.NMT_GETHEADERS {
				mp := m.(*core.MsgGetHeaders)
				for _, v := range mp.Blocks {
					log.Println("GetHeaders", v)
				}
			}
		},
	})
}
