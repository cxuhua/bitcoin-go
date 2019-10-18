package core

import (
	"bitcoin/store"
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

func (l *BlockHeaderList) Front() *BlockHeader {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.lv.Len() == 0 {
		return nil
	}
	el := l.lv.Front()
	return el.Value.(*BlockHeader)
}

func (l *BlockHeaderList) Push(h *BlockHeader) {
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

func processBlock(wid int, mdb store.DbImp, c *Client, m *MsgBlock) error {
	return mdb.Transaction(func(sdb store.DbImp) error {
		G.Lock()
		defer G.Unlock()
		if bh := m.ToBlockHeader(); !G.IsNextBlock(bh) {
			return fmt.Errorf("can't link prev block,ignore block %v", NewHashID(bh.Hash))
		} else if err := m.CheckBlock(G.LastBlock(), sdb); err != nil {
			return fmt.Errorf("check block error %v,ignore save", err)
		} else {
			if sdb.HasBK(bh.Hash) {
				return errors.New("block exists,ignore save ,hash=" + NewHashID(bh.Hash).String())
			}
			if err := sdb.SetBK(bh.Hash, bh); err != nil {
				return fmt.Errorf("DB setbk error %v", err)
			}
			if err := m.SaveTXS(sdb); err != nil {
				return err
			}
			if err := AvailableBlockComing(sdb, m); err != nil {
				return err
			}
			Headers.Remove()
			G.SetLastBlock(bh)
			if c != nil {
				Notice <- c
				hv := fmt.Sprintf("%.3f", float32(bh.Height)/float32(c.VerInfo.Height))
				log.Println("Work", wid, "save block:", m.Hash, "height=", bh.Height, "finish=", hv, "from", c.Key(), "OK")
			}
		}
		return nil
	})
}

func processTX(wid int, db store.DbImp, c *Client, m *MsgTX) error {
	//log.Println("Work id", wid, "recv tx=", m.Tx.Hash)
	//TxsMap.Set(&m.Tx)
	return nil
}

func processHeaders(wid int, db store.DbImp, c *Client, m *MsgHeaders) error {
	for _, v := range m.Headers {
		bh := v.ToBlockHeader()
		if G.IsNextHeader(bh) {
			Headers.Push(bh)
			log.Println("get block header", NewHashID(bh.Hash))
		}
	}
	if Headers.Len() > 0 {
		Notice <- c
	}
	return nil
}

func processInv(wid int, db store.DbImp, c *Client, m *MsgINV) error {
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
