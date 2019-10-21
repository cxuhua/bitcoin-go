package main

import (
	"bitcoin/core"
	"io/ioutil"
	"log"
	"testing"
)

func TestError(t *testing.T) {
	db := core.DB()
	defer db.Close()
	if err := core.G.Init(); err != nil {
		panic(err)
	}
	data, err := ioutil.ReadFile("f:\\blocks\\000000000000018f5ee13ecf9e9595356148c097a2fb5825169fde3f48e8eb8a")
	if err != nil {
		panic(err)
	}
	h := core.NewNetHeader(data)
	m := &core.MsgBlock{}
	m.Read(h)
	m.Height = core.G.LastHeight() + 1
	err = m.Check()
	log.Println(err)
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
