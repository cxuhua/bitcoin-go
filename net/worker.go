package net

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

func processBlock(db db.DbImp, block *MsgBlock) {

}

func processTX(db db.DbImp, tx *MsgTX) {

}

func processHeaders(db db.DbImp, headers *MsgHeaders) {

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
				case NMT_BLOCK:
					processBlock(db, unit.m.(*MsgBlock))
				case NMT_TX:
					processTX(db, unit.m.(*MsgTX))
				case NMT_HEADERS:
					processHeaders(db, unit.m.(*MsgHeaders))
				}
			case <-ctx.Done():
				log.Println("stop worker unit", i, ctx.Err())
				return ctx.Err()
			}
		}
	}
	for ctx.Err() != context.Canceled {
		time.Sleep(time.Second * 3)
		db.UseSession(ctx, func(db db.DbImp) error {
			return mfx(db)
		})
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
