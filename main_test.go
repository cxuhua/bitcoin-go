package main

import (
	"bitcoin/core"
	"encoding/hex"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
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
	data, err := ioutil.ReadFile("f:\\blocks\\000000002732d387256b57cabdcb17767e3d30e220ea73f844b1907c0b5919ea")
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

func readvarint(h *core.NetHeader) uint64 {
	n := uint64(0)
	for {
		ch := h.ReadUint8()
		n = (n << 7) | uint64(ch&0x7F)
		if ch&0x80 != 0 {
			n++
		} else {
			break
		}
	}
	return n
}

func TestLoadKey(t *testing.T) {

	db, err := leveldb.OpenFile("/Volumes/backup/bitcoin/datadir/blocks/index", nil)
	if err != nil {
		panic(err)
	}
	rang := util.BytesPrefix([]byte{0x66})
	iter := db.NewIterator(rang, nil)
	//i := 0
	for iter.Next() {

		log.Println(hex.EncodeToString(iter.Key()), hex.EncodeToString(iter.Value()))
		h := core.NewNetHeader(iter.Value())
		log.Printf("%x", readvarint(h))
		log.Printf("%x", readvarint(h))
		log.Printf("%x", readvarint(h))
		log.Printf("%x", readvarint(h))
		log.Printf("%x", readvarint(h))

	}
	db.Close()
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
