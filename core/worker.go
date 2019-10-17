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
	ch *BlockHeader
	ct time.Time
}

func (l *BlockHeaderList) Front() *BlockHeader {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.lv.Len() == 0 {
		return nil
	}
	now := time.Now()
	fv := l.lv.Front().Value.(*BlockHeader)
	if l.ch == fv && now.Sub(l.ct) < time.Second*30 {
		return nil
	}
	l.ch = fv
	l.ct = time.Now()
	return l.ch
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
func (l *BlockHeaderList) Load(db store.DbImp) {
	l.mu.Lock()
	defer l.mu.Unlock()
	db.GetBK(store.ListSyncBK, store.IterFunc(func(cur *mongo.Cursor) {
		hv := &BlockHeader{}
		err := cur.Decode(hv)
		if err != nil {
			return
		}
		l.lv.PushBack(hv)
	}))
}

func (l *BlockHeaderList) Pop() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.lv.Len() > 0 {
		l.lv.Remove(l.lv.Front())
		l.ch = nil
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
	if bh.IsGenesis() {
		if err := db.SetBK(bh.Hash, bh); err != nil {
			return fmt.Errorf("DB save block header error %v", err)
		}
		G.SetLastBlock(bh)
	} else {
		v := store.SetValue{"count": bh.Count}
		if err := db.SetBK(bh.Hash, v); err != nil {
			return fmt.Errorf("DB set block tx count error %v", err)
		}
		Headers.Pop()
	}
	Notice <- NoticeRecvBlock
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
		Headers.Push(bh)
		G.SetLastBlock(bh)
		log.Println("save blockheader", NewHashID(bh.Hash), bh.Height)
	}
	if Headers.Len() > 0 {
		Notice <- NoticeSaveHeaders
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
