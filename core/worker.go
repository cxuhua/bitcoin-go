package core

import (
	"bitcoin/db"
	"context"
	"log"
	"sync"
	"time"
)

const (
	WorkerQueueSize = 64
)

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

func processBlock(db db.DbImp, c *Client, m *MsgBlock) {

}

func processTX(db db.DbImp, c *Client, m *MsgTX) {

}

func processHeaders(db db.DbImp, c *Client, m *MsgHeaders) {

}

func processInv(db db.DbImp, c *Client, m *MsgINV) {

}

func doWorker(ctx context.Context, wg *sync.WaitGroup, i int) {
	defer wg.Done()
	mfx := func(db db.DbImp) error {
		log.Println("start worker unit", i)
		defer func() {
			if err := recover(); err != nil {
				log.Println("[worker error]:", err)
			}
		}()
		for {
			select {
			case unit := <-WorkerQueue:
				cmd := unit.m.Command()
				switch cmd {
				case NMT_INV:
					processInv(db, unit.c, unit.m.(*MsgINV))
				case NMT_BLOCK:
					processBlock(db, unit.c, unit.m.(*MsgBlock))
				case NMT_TX:
					processTX(db, unit.c, unit.m.(*MsgTX))
				case NMT_HEADERS:
					processHeaders(db, unit.c, unit.m.(*MsgHeaders))
				}
			case <-ctx.Done():
				log.Println("stop worker unit", i, ctx.Err())
				return ctx.Err()
			}
		}
	}
	for ctx.Err() != context.Canceled {
		time.Sleep(time.Second * 3)
		err := db.UseSession(ctx, func(db db.DbImp) error {
			return mfx(db)
		})
		log.Println("db session end, return err", err)
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
