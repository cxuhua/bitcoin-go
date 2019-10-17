package core

import (
	"bitcoin/store"
	"bytes"
	"log"
	"sync"
)

type Global struct {
	mu sync.Mutex
	lb *BlockHeader
	hb *BlockHeader
}

func (g *Global) IsNextHeader(bh *BlockHeader) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.hb == nil {
		g.hb = g.lb
	}
	ok := bytes.Equal(bh.Prev[:], g.hb.Hash[:])
	if ok {
		g.hb = bh
	}
	return ok
}

func (g *Global) IsNext(bh *BlockHeader) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	ok := bytes.Equal(bh.Prev[:], g.lb.Hash[:])
	if ok {
		bh.Height = g.lb.Height + 1
	}
	return ok
}

func (g *Global) LastBlock() *BlockHeader {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.lb
}

func (g *Global) SetLastBlock(v *BlockHeader) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.lb = v
}

func (g *Global) Init(db store.DbImp) error {
	//get last block
	bh := &BlockHeader{}
	if err := db.GetBK(store.NewestBK, bh); err == nil {
		G.SetLastBlock(bh)
		log.Println("last block height", bh.Height, "hash=", NewHashID(bh.Hash))
	}
	//get not download block
	return nil
}

var (
	G = &Global{}
)
