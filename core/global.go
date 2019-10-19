package core

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"runtime"
	"sync"
)

type Global struct {
	mu   sync.Mutex
	best *MsgBlock
}

func (g *Global) Lock() {
	g.mu.Lock()
}

func (g *Global) Unlock() {
	g.mu.Unlock()
}

func (g *Global) IsNextHeader(bh *BHeader) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return bh.Prev.Equal(g.best.Hash)
}

func (g *Global) LastBlock() *MsgBlock {
	return g.best
}

func (g *Global) LastHeight() uint32 {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.best == nil {
		return 0
	}
	return g.best.Height
}

func (g *Global) LastHash() HashID {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.best.Hash
}

func (g *Global) IsNextBlock(m *MsgBlock) bool {
	if m.IsGenesis() {
		return true
	}
	if m.Prev.Equal(g.best.Hash) {
		m.Height = g.best.Height + 1
		return true
	}
	return false
}

func (g *Global) SetBestBlock(m *MsgBlock) {
	g.best = m
}

func (g *Global) IsRequestGenesis() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.best == nil
}

func (g *Global) Init() error {
	if best, err := LoadBestBlock(); err == nil {
		g.best = best
		log.Println("load best block", best.Hash, "height=", best.Height)
	} else {
		log.Println("database empty,start download genesis block")
	}
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
