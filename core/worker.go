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

	"go.mongodb.org/mongo-driver/mongo"
)

const (
	WorkerQueueSize = 64
)

type BlockHeaderList struct {
	mu sync.Mutex
	lv *list.List
}

func (l *BlockHeaderList) Front() *BlockHeader {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.lv.Len() == 0 {
		return nil
	}
	el := l.lv.Front()
	fv := el.Value.(*BlockHeader)
	l.lv.Remove(el)
	return fv
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

//load not download block header
func (l *BlockHeaderList) Load(db store.DbImp) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return db.GetBK(store.ListSyncBK, store.IterFunc(func(cur *mongo.Cursor) error {
		hv := &BlockHeader{}
		err := cur.Decode(hv)
		if err != nil {
			return err
		}
		l.lv.PushBack(hv)
		return nil
	}))
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

func processBlock(wid int, db store.DbImp, c *Client, m *MsgBlock) error {
	log.Println("Work id", wid, "recv block:", m.Hash, " from", c.Key())
	//check block and block txs
	if err := m.CheckBlock(db); err != nil {
		return fmt.Errorf("check block error %v,ignore save", err)
	}
	//save block
	bh := m.ToBlockHeader()
	if bh.IsGenesis() {
		err := db.Transaction(func(sdb store.DbImp) error {
			if err := sdb.SetBK(bh.Hash, bh); err != nil {
				return fmt.Errorf("DB save block header error %v", err)
			}
			return m.SaveTXS(sdb)
		})
		if err != nil {
			return err
		}
		G.SetLastBlock(bh)
		Notice <- c
		return nil
	}
	return db.Transaction(func(sdb store.DbImp) error {
		if !sdb.HasBK(bh.Hash) {
			lb := G.LastBlock()
			if !bytes.Equal(bh.Prev, lb.Hash) {
				return errors.New("recv block,can't link prev block")
			}
			bh.Height = lb.Height + 1
			if err := sdb.SetBK(bh.Hash, bh); err != nil {
				return fmt.Errorf("DB setbk error %v", err)
			}
			if err := m.SaveTXS(sdb); err != nil {
				return err
			}
			G.SetLastBlock(bh)
			return nil
		} else {
			v := store.SetValue{"count": bh.Count}
			if err := sdb.SetBK(bh.Hash, v); err != nil {
				return fmt.Errorf("DB set block tx count error %v", err)
			}
			if err := m.SaveTXS(sdb); err != nil {
				return err
			}
			Notice <- c
			return nil
		}
	})
}

func processTX(wid int, db store.DbImp, c *Client, m *MsgTX) error {
	//log.Println("Work id", wid, "recv tx=", m.Tx.Hash)
	TxsMap.Set(&m.Tx)
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
		Headers.Push(bh)
		G.SetLastBlock(bh)
		log.Println("save blockheader", NewHashID(bh.Hash), bh.Height)
	}
	if Headers.Len() > 0 {
		Notice <- c
	}
	return nil
}

func processInv(wid int, db store.DbImp, c *Client, m *MsgINV) error {
	//log.Println("Work id", wid, "recv inv")
	hm := NewMsgGetHeaders()
	tm := NewMsgGetData()
	for _, v := range m.Invs {
		switch v.Type {
		case MSG_TX:
			//log.Println("get inv TX ", v.ID, " start get TX data")
			tm.AddHash(v.Type, v.ID[:])
		case MSG_BLOCK:
			log.Println("get inv block id", v.ID, " start get block headers")
			tm.AddHash(v.Type, v.ID[:])
		case MSG_FILTERED_BLOCK:
		case MSG_CMPCT_BLOCK:
		}
	}
	if len(hm.Blocks) > 0 {
		c.WriteMsg(hm)
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
