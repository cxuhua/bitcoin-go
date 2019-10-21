package main

import (
	"bitcoin/core"
	"bitcoin/util"
	"encoding/binary"
	"encoding/hex"
	"io/ioutil"
	"log"
	"testing"
)

func TestBB(t *testing.T) {
	b4 := []byte{0x1d, 0x01, 0x00, 0x00}
	log.Println(binary.LittleEndian.Uint32(b4))
}

func TestError(t *testing.T) {
	//db := core.DB()
	//defer db.Close()
	//if err := core.G.Init(); err != nil {
	//	panic(err)
	//}
	data, err := ioutil.ReadFile("f:\\blocks\\000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f")
	if err != nil {
		panic(err)
	}

	log.Println(hex.EncodeToString(util.HASH256(data)))

	//h := core.NewNetHeader(data)
	//m := &core.MsgBlock{}
	//m.Read(h)
	//m.Height = core.G.LastHeight() + 1
	//err = m.Check()
	//log.Println(err)
}

func TestLoadKey(t *testing.T) {
	db := core.DB()
	defer db.Close()
	if err := core.G.Init(); err != nil {
		panic(err)
	}
	eles := core.ListAddrValues("127YYnp1jvgAX3vCB22WUUsuyTYfAeSQHh")
	for _, ele := range eles {
		log.Println(ele.TAddrKey, ele.GetValue())
	}
	//b, err := core.LoadHeightBlock(20)
	//if err != nil {
	//	panic(err)
	//}
	//if b.Height != 20 {
	//	t.Errorf("load error")
	//}
	//iter := core.DB().NewIterator(nil, nil)
	//for iter.Next() {
	//	key := iter.Key()
	//	if key[0] == core.TPrefixBlock {
	//		vk := core.TBlock(iter.Value())
	//		bk := core.TBlockKey{}
	//		copy(bk[:], key)
	//		log.Println(bk, vk.Height())
	//	} else if key[0] == core.TPrefixTxId {
	//		tk := core.TTxKey{}
	//		copy(tk[:], key)
	//		log.Println(tk)
	//	} else if key[0] == core.TPrefixAddress {
	//		v := core.TAddrValue{}
	//		copy(v[:], iter.Value())
	//		log.Println(core.TAddrKey(key), v)
	//	} else if key[0] == core.TPrefixHeight {
	//		log.Println(core.NewHashID(iter.Value()))
	//	}
	//}
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

func TestLoadBlocks(t *testing.T) {

}
