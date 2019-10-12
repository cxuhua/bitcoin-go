package net

import (
	"bitcoin/config"
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

type SeedItem struct {
	Ip        net.IP
	Ping      int
	Connected bool
	Available bool
	Ignore    bool
	Ver       *MsgVersion
}

func (si *SeedItem) IsNeedCheck() bool {
	if si.Connected {
		return false
	}
	if si.Ignore {
		return false
	}
	if !si.Available {
		return false
	}
	return true
}

type SeedMap struct {
	nodes map[string]*SeedItem
	mutex sync.Mutex
}

func (s *SeedMap) SetConnected(ip string, pv bool, ver *MsgVersion) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if v, b := s.nodes[ip]; b {
		v.Connected = pv
		v.Ver = ver
	}
}

func (s *SeedMap) SetPing(ip string, pv int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if v, b := s.nodes[ip]; b {
		v.Ping = pv
	}
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

func (s *SeedMap) Remove(ip string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.nodes, ip)
}

var (
	nodes = &SeedMap{
		nodes: map[string]*SeedItem{},
	}
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
			nodes.Add(v.String())
		}
		log.Println("LOOKUP", v, "Count=", len(ips))
	}
}

func runcheckip(conf *config.Config) {
	num := 0
	idx := 0
	for _, v := range nodes.nodes {
		if !v.IsNeedCheck() {
			continue
		}
		num++
	}
	for _, v := range nodes.nodes {
		if !v.IsNeedCheck() {
			continue
		}
		idx++
		addr := fmt.Sprintf("%s:%d", v.Ip, conf.ListenPort)

		timeout := 10
		c := NewClient(ClientTypeOut, addr)
		c.SetTry(1)
		c.Sync(&ClientListener{
			OnConnected: func() {

			},
			OnClosed: func() {
				//log.Println("Closed")
			},
			OnMessage: func(msg MsgIO) {
				cmd := msg.Command()
				if cmd == NMT_VERSION {
					v.Ver = msg.(*MsgVersion)
					v.Available = true
					c.Close()
					log.Println(idx, "/", num, "Check network OK", addr, "VER=", v.Ver.Ver)
				}
			},
			OnLoop: func() {
				timeout--
				if timeout <= 0 {
					v.Ignore = true
					c.Close()
				}
			},
		})
	}
	for _, v := range nodes.nodes {
		if v.Available {
			log.Println(v, " available")
		}
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
