package net

import (
	"bitcoin/config"
	"context"
	"log"
	"net"
	"sync"
	"time"
)

var (
	MWG = sync.WaitGroup{}
)

//run lookup
func runlookup(conf *config.Config) {
	for _, v := range conf.Seeds {
		ips, err := net.LookupIP(v)
		if err != nil {
			log.Println("lookup ip error ", v, err)
			continue
		}
		for _, v := range ips {
			IpChan <- v
		}
		log.Println("LOOKUP", v, "Count=", len(ips))
	}
}

//启动
func StartLookUp(ctx context.Context) {
	defer MWG.Done()
	MWG.Add(1)
	mfx := func() {
		log.Println("lookup start")
		defer func() {
			if err := recover(); err != nil {
				log.Println("[LOOKUP error]:", err)
			}
		}()
		conf := config.GetConfig()
		runlookup(conf)
		ltimer := time.NewTimer(time.Minute * 10)
		for {
			select {
			case <-ltimer.C:
				ltimer.Reset(time.Minute * 10)
			case <-ctx.Done():
				log.Println("looup stop", ctx.Err())
				return
			}
		}
	}
	for ctx.Err() != context.Canceled {
		time.Sleep(time.Second * 3)
		mfx()
	}
}
