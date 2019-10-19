package main

import (
	"bitcoin/core"
	"log"
	"testing"
)

func TestLoadKey(t *testing.T) {
	defer core.DB().Close()

	last, err := core.LoadBestBlock()
	if err != nil {
		panic(err)
	}
	log.Println("last block", last.Height, last.Hash)

	m, err := core.LoadBlock(core.NewHashID("00000000fb11ef25014e02b315285a22f80c8f97689d7e36d723317defaabe5b"))
	if err != nil {
		panic(err)
	}
	log.Println(m.Height == 104)

	iter := core.DB().NewIterator(nil, nil)
	for iter.Next() {
		key := iter.Key()
		if key[0] == core.TPrefixBlock {
			vk := core.TBlock(iter.Value())
			bk := core.TBlockKey{}
			copy(bk[:], key)
			log.Println(bk, vk.Height())
		} else if key[0] == core.TPrefixTxId {
			tk := core.TTxKey{}
			copy(tk[:], key)
			log.Println(tk)
		} else if key[0] == core.TPrefixAddress {
			v := core.TAddrValue{}
			copy(v[:], iter.Value())
			log.Println(core.TAddrKey(key), v)
		}
	}
}

func TestRunClient(t *testing.T) {
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
