package core

import (
	"bitcoin/store"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"runtime"
	"sync"
)

type Global struct {
	mu sync.Mutex
	lb *BlockHeader
	hb *BlockHeader
}

func (g *Global) Lock() {
	g.mu.Lock()
}

func (g *Global) Unlock() {
	g.mu.Unlock()
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

func (g *Global) LastBlock() *BlockHeader {
	return g.lb
}

func (g *Global) LastHeight() uint32 {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.lb == nil {
		return 0
	}
	return g.lb.Height
}

func (g *Global) LastHash() []byte {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.lb.Hash
}

func (g *Global) IsNextBlock(bh *BlockHeader) bool {
	if g.lb == nil && bh.IsGenesis() {
		return true
	}
	if g.lb == nil {
		return false
	}
	ok := bytes.Equal(bh.Prev[:], g.lb.Hash[:])
	if ok {
		bh.Height = g.lb.Height + 1
	}
	return ok
}

func (g *Global) HasLast() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.lb != nil
}

func (g *Global) SetLastBlock(v *BlockHeader) {
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

func stack(skip int) []byte {
	buf := new(bytes.Buffer)
	var lines [][]byte
	var lastFile string
	for i := skip; ; i++ { // Skip the expected number of frames
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc)
		if file != lastFile {
			data, err := ioutil.ReadFile(file)
			if err != nil {
				continue
			}
			lines = bytes.Split(data, []byte{'\n'})
			lastFile = file
		}
		fmt.Fprintf(buf, "\t%s: %s\n", function(pc), source(lines, line))
	}
	return buf.Bytes()
}

var (
	dunno     = []byte("???")
	centerDot = []byte("·")
	dot       = []byte(".")
	slash     = []byte("/")
)

func source(lines [][]byte, n int) []byte {
	n-- // in stack trace, lines are 1-indexed but our array is 0-indexed
	if n < 0 || n >= len(lines) {
		return dunno
	}
	return bytes.TrimSpace(lines[n])
}

func function(pc uintptr) []byte {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return dunno
	}
	name := []byte(fn.Name())
	if lastSlash := bytes.LastIndex(name, slash); lastSlash >= 0 {
		name = name[lastSlash+1:]
	}
	if period := bytes.Index(name, dot); period >= 0 {
		name = name[period+1:]
	}
	name = bytes.Replace(name, centerDot, dot, -1)
	return name
}
