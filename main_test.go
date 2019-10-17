package main

import (
	"bitcoin/core"
	"log"
	"testing"
)

func TestRunClient(t *testing.T) {
	c := core.NewClient(core.ClientTypeOut, "47.97.62.19:8333")
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
			if cmd == core.NMT_VERACK {
				//getdata(c, "000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f")
				//m := core.NewMsgGetHeaders()
				//m.Blocks = make([]core.HashID, 1)
				//m.Blocks[0] = core.NewHashID("000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f")
				//c.WriteMsg(m)
				d := core.NewMsgSendHeaders()
				c.WriteMsg(d)

				m := core.NewMsgSendCmpct()
				m.Ver = 1
				m.Inter = 1
				c.WriteMsg(m)
				//m := core.NewMsgGetBlocks()
				//m.AddHashID(core.NewHashID("000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f"))
				//m.Stop = core.NewHashID("000000006a625f06636b8bb6ac7b960a8d03705d1ace08b1a19da3fdcc99ddbd")
				//c.WriteMsg(m)

				//m := core.NewMsgGetHeaders()
				//m.AddHashID(core.NewHashID("000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f"))
				//m.Stop = core.NewHashID("000000006a625f06636b8bb6ac7b960a8d03705d1ace08b1a19da3fdcc99ddbd")
				//c.WriteMsg(m)
			} else if cmd == core.NMT_BLOCK {
				d := m.(*core.MsgBlock)
				log.Println(d.Hash)
				//getdata(c, d.Prev.String())
			} else if cmd == core.NMT_GETHEADERS {
				//m := m.(*core.MsgGetHeaders)
				//log.Println(m)
			} else if cmd == core.NMT_INV {
				m := m.(*core.MsgINV)
				for _, v := range m.Invs {
					if v.Type != core.MSG_TX {
						log.Println(v.Type, v.ID)
					}
				}
			} else if cmd == core.NMT_HEADERS {
				m := m.(*core.MsgHeaders)
				for _, v := range m.Headers {
					log.Println(v.Hash, v.Count)
				}
			} else if cmd == core.NMT_MERKLEBLOCK {
				m := m.(*core.MsgMerkleBlock)
				a, b, c := m.Extract()
				log.Println(a, b, c)
			} else if cmd == core.NMT_CMPCTBLOCK {
				m := m.(*core.MsgCmpctBlock)
				log.Println(m.Header.Merkle)
			}
		},
	})
}
