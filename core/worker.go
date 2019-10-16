package core

import (
	"bitcoin/store"
	"bytes"
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
	ch *BHeader
	ct time.Time
}

func (l *BlockHeaderList) has(hv HashID) (*list.Element, bool) {
	for e := l.lv.Front(); e != nil; e = e.Next() {
		if cv := e.Value.(*BHeader); cv.Hash.Equal(hv) {
			return e, true
		}
	}
	return nil, false
}

func (l *BlockHeaderList) PushMany(hs []*BHeader) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, v := range hs {
		if _, ok := l.has(v.Hash); ok {
			continue
		}
		l.lv.PushBack(v)
	}
}

func (l *BlockHeaderList) PushOne(h *BHeader) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if _, ok := l.has(h.Hash); !ok {
		l.lv.PushBack(h)
	}
}

func (l *BlockHeaderList) Len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.lv.Len()
}

func (l *BlockHeaderList) Remove(hv HashID) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if e, ok := l.has(hv); ok {
		l.lv.Remove(e)
	}
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

func processBlock(wid int, db store.DbImp, c *Client, m *MsgBlock) error {
	log.Println("Work id", wid, "recv block:", m.Hash)
	bh := m.ToBlockHeader()
	if err := db.SetBK(bh.Hash, bh); err != nil {
		return fmt.Errorf("DB save block header error %v", err)
	}
	G.SetLastBlock(bh)
	return nil
}

func processTX(wid int, db store.DbImp, c *Client, m *MsgTX) error {
	//log.Println("Work id", wid, "recv tx")
	return nil
}

func processHeaders(wid int, db store.DbImp, c *Client, m *MsgHeaders) error {
	for _, v := range m.Headers {
		lb := G.LastBlock()
		bh := v.ToBlockHeader()
		if !bytes.Equal(bh.Prev, lb.Hash) {
			continue
		}
		bh.Height = lb.Height + 1
		if err := db.SetBK(bh.Hash, bh); err != nil {
			return fmt.Errorf("DB setbk error %v", err)
		}
		G.SetLastBlock(bh)
		log.Println("save blockheader", NewHashID(bh.Hash), bh.Height)
	}
	if len(m.Headers) > 0 {
		Notice <- NoticeSaveHeadersOK
	}
	return nil
}

func processInv(wid int, db store.DbImp, c *Client, m *MsgINV) error {
	//log.Println("Work id", wid, "recv inv")
	return nil
}

func processGetHeaders(wid int, db store.DbImp, c *Client, m *MsgGetHeaders) error {
	//log.Println("Work id", wid, "recv headers")
	return nil
}

func doWorker(ctx context.Context, wg *sync.WaitGroup, i int) {
	defer wg.Done()
	mfx := func(db store.DbImp) error {
		log.Println("start worker unit", i)
		defer func() {
			if err := recover(); err != nil {
				log.Println("[worker error]:", err)
			}
		}()
		for {
			var err error = nil
			select {
			case unit := <-WorkerQueue:
				cmd := unit.m.Command()
				switch cmd {
				case NMT_INV:
					err = processInv(i, db, unit.c, unit.m.(*MsgINV))
				case NMT_BLOCK:
					err = processBlock(i, db, unit.c, unit.m.(*MsgBlock))
				case NMT_TX:
					err = processTX(i, db, unit.c, unit.m.(*MsgTX))
				case NMT_HEADERS:
					err = processHeaders(i, db, unit.c, unit.m.(*MsgHeaders))
				case NMT_GETHEADERS:
					err = processGetHeaders(i, db, unit.c, unit.m.(*MsgGetHeaders))
				}
			case <-ctx.Done():
				err = fmt.Errorf("recv done worker exit %v", ctx.Err())
			}
			if err != nil {
				return err
			}
		}
	}
	for {
		time.Sleep(time.Second * 3)
		err := store.UseSession(ctx, func(db store.DbImp) error {
			return mfx(db)
		})
		log.Println("store session end, return err:", err)
		if errors.Is(context.Canceled, err) {
			return
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
