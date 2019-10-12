package net

import (
	"context"
	"log"
	"sync"
	"time"
)

func doWorker(ctx context.Context, wg *sync.WaitGroup, i int) {
	mfx := func() {
		log.Println("start worker unit", i)
		defer func() {
			if err := recover(); err != nil {
				log.Println("[worker error]:", err)
			}
		}()
		//conf := config.GetConfig()
		for {
			select {
			case <-ctx.Done():
				log.Println("stop worker unit", i, ctx.Err())
				wg.Done()
				return
			}
		}
	}
	for ctx.Err() != context.Canceled {
		time.Sleep(time.Second * 3)
		mfx()
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
