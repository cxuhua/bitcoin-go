package core

import (
	"container/list"
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

const (
	WorkerQueueSize = 64
)

type BlockHeaderList struct {
	mu sync.Mutex
	lv *list.List
}

func (l *BlockHeaderList) Remove() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.lv.Len() == 0 {
		return
	}
	e := l.lv.Front()
	l.lv.Remove(e)
}

func (l *BlockHeaderList) Front() *BHeader {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.lv.Len() == 0 {
		return nil
	}
	el := l.lv.Front()
	return el.Value.(*BHeader)
}

func (l *BlockHeaderList) Push(h *BHeader) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.lv.PushBack(h)
}

func (l *BlockHeaderList) Len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.lv.Len()
}

func (l *BlockHeaderList) Pop() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.lv.Len() == 0 {
		return
	}
	l.lv.Remove(l.lv.Front())
}

func NewBlockHeaderList() *BlockHeaderList {
	return &BlockHeaderList{
		lv: list.New(),
	}
}

type WorkerUnit struct {
	m MsgIO
	c *Client
}

func NewWorkerUnit(m MsgIO, c *Client) *WorkerUnit {
	return &WorkerUnit{c: c, m: m}
}

var (
	WorkerQueue = make(chan *WorkerUnit, WorkerQueueSize)
)

func processBlock(wid int, c *Client, m *MsgBlock) error {
	G.Lock()
	defer G.Unlock()
	if !G.IsNextBlock(m) {
		return fmt.Errorf("can't link prev block,ignore block %v", m.Hash)
	}
	if err := m.Check(); err != nil {
		return fmt.Errorf("check block error %w", err)
	}
	if err := m.Connect(true); err != nil {
		return fmt.Errorf("DB save block error %w", err)
	}
	Headers.Remove()
	G.SetBestBlock(m)
	if c != nil {
		Notice <- c
		hv := fmt.Sprintf("%.3f", float32(m.Height)/float32(c.VerInfo.Height))
		log.Println("Work", wid, "save block:", m.Hash, "height=", m.Height, "finish=", hv, "from", c.Key(), "OK")
	}
	return nil
}

func processTX(wid int, c *Client, m *MsgTX) error {
	//log.Println("Work id", wid, "recv tx=", m.Tx.Hash)
	//TxsMap.Set(&m.Tx)
	return nil
}

func processHeaders(wid int, c *Client, m *MsgHeaders) error {
	for _, v := range m.Headers {
		if G.IsNextHeader(v) {
			Headers.Push(v)
		}
	}
	if Headers.Len() > 0 {
		Notice <- c
	}
	return nil
}

func processInv(wid int, c *Client, m *MsgINV) error {
	//log.Println("Work id", wid, "recv inv")
	tm := NewMsgGetData()
	for _, v := range m.Invs {
		switch v.Type {
		case MSG_TX:
			//log.Println("get inv TX ", v.ID, " start get TX data")
			tm.AddHash(v.Type, v.ID[:])
		case MSG_BLOCK:
			tm.AddHash(v.Type, v.ID[:])
		case MSG_FILTERED_BLOCK:
		case MSG_CMPCT_BLOCK:
		}
	}
	if len(tm.Invs) > 0 {
		c.WriteMsg(tm)
	}
	return nil
}

func processGetHeaders(wid int, c *Client, m *MsgGetHeaders) error {
	//log.Println("Work id", wid, "recv headers")
	return nil
}

func doWorker(ctx context.Context, wg *sync.WaitGroup, i int) {
	defer wg.Done()
	mfx := func() error {
		log.Println("start worker unit", i)
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[Recovery] %s panic recovered:%s\n", err, stack(3))
			}
		}()
		for {
			var err error = nil
			select {
			case unit := <-WorkerQueue:
				cmd := unit.m.Command()
				switch cmd {
				case NMT_INV:
					err = processInv(i, unit.c, unit.m.(*MsgINV))
				case NMT_BLOCK:
					err = processBlock(i, unit.c, unit.m.(*MsgBlock))
				case NMT_TX:
					err = processTX(i, unit.c, unit.m.(*MsgTX))
				case NMT_HEADERS:
					err = processHeaders(i, unit.c, unit.m.(*MsgHeaders))
				case NMT_GETHEADERS:
					err = processGetHeaders(i, unit.c, unit.m.(*MsgGetHeaders))
				}
			case <-ctx.Done():
				err = fmt.Errorf("recv done worker exit %w", ctx.Err())
			}
			if err != nil {
				return err
			}
		}
	}
	for {
		time.Sleep(time.Second * 3)
		err := mfx()
		if err != nil {
			log.Println(err)
		}
		if errors.Is(err, context.Canceled) {
			break
		}
	}
}

func StartWorker(ctx context.Context, num int) {
	wg := &sync.WaitGroup{}
	log.Println("start worker num = ", num)
	for i := 0; i < num; i++ {
		wg.Add(1)
		go doWorker(ctx, wg, i)
	}
	wg.Wait()
	log.Println("stop worker num = ", num)
}
