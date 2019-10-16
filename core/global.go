package core

import (
	"bitcoin/store"
	"sync"
)

type Global struct {
	mu      sync.Mutex
	lnblock *BlockHeader
}

func (g *Global) LastBlock() *BlockHeader {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.lnblock
}

func (g *Global) SetLastBlock(v *BlockHeader) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.lnblock = v
}

func (g *Global) Init(db store.DbImp) error {
	bh := &BlockHeader{}
	if err := db.GetBK(store.NewestBK, bh); err == nil {
		g.lnblock = bh
	}
	return nil
}

var (
	G = &Global{}
)
