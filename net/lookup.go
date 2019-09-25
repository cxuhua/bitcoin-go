package net

import (
	"bitcoin/config"
	"context"
	"log"
	"net"
	"sync"
	"time"
)

type SeedItem struct {
	Ip        net.IP
	Ping      int
	Connected bool
}

type SeedMap struct {
	nodes map[string]*SeedItem
	mutex sync.Mutex
}

func (s *SeedMap) Add(ip string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if _, b := s.nodes[ip]; b {
		return
	}
	s.nodes[ip] = &SeedItem{
		Ip:        net.ParseIP(ip),
		Ping:      0,
		Connected: false,
	}
}

var (
	nodes = &SeedMap{
		nodes: map[string]*SeedItem{},
	}
)

//run lookup
func runlookup(conf *config.Config) {
	for _, v := range conf.Seeds {
		log.Println("LOOKUP ", v)
		ips, err := net.LookupIP(v)
		if err != nil {
			log.Println("lookup ip error ", v, err)
			continue
		}
		for _, v := range ips {
			nodes.Add(v.String())
		}
	}
	for _, v := range nodes.nodes {
		log.Println("NODE = ", v.Ip)
	}
}

func runcheckip(conf *config.Config) {
	for _, v := range nodes.nodes {
		log.Println(v.Ip)
	}
}

//启动
func StartLookUp(ctx context.Context) {
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
		ctimer := time.NewTimer(time.Second)
		for {
			select {
			case <-ltimer.C:
				runlookup(conf)
				ltimer.Reset(time.Minute * 10)
			case <-ctimer.C:
				runcheckip(conf)
				ctimer.Reset(time.Second)
			case <-ctx.Done():
				log.Println("looup stop", ctx.Err())
				return
			}
		}
	}
	for ctx.Err() == nil {
		time.Sleep(time.Second * 3)
		mfx()
	}
}
