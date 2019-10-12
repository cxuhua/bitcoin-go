package net

import (
	"bitcoin/config"
	"context"
	"fmt"
	"log"
	"net"
	"sort"
	"sync"
	"time"
)

var (
	MWG = sync.WaitGroup{}
)

type SeedItem struct {
	Ip        net.IP
	Connected bool
	Available bool
	Ignore    bool
	Ver       *MsgVersion
	Client    *Client
}

func (si SeedItem) String() string {
	return si.Ip.String() + fmt.Sprintf(" A=%v I=%v C=%v", si.Available, si.Ignore, si.Connected)
}

func (si *SeedItem) IsNeedCheck() bool {
	if si.Connected {
		return false
	}
	if si.Ignore {
		return false
	}
	if si.Available {
		return false
	}
	return true
}

type SeedMap struct {
	nodes map[string]*SeedItem
	mutex sync.Mutex
}

func (s *SeedMap) SetConnected(ip net.IP, c *Client) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if v, b := s.nodes[ip.String()]; b {
		if c != nil {
			v.Connected = true
		} else {
			v.Connected = false
		}
		v.Client = c
	}
}

func (s *SeedMap) SetAvailable(ip net.IP, pv bool, ver *MsgVersion) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if v, b := s.nodes[ip.String()]; b {
		v.Available = pv
		v.Ver = ver
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
		Connected: false,
	}
}

func (s *SeedMap) Remove(ip net.IP) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.nodes, ip.String())
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
	log.Println("CHECK all seed host")
	runcheckip(conf)
	ds := []*SeedItem{}
	for _, v := range nodes.nodes {
		if v.Available {
			ds = append(ds, v)
		}
	}
	//save available host
	nodes = &SeedMap{
		nodes: map[string]*SeedItem{},
	}
	for _, v := range ds {
		nodes.nodes[v.Ip.String()] = v
	}
	for _, v := range ds {
		IpChan <- v
	}
	log.Println(len(nodes.nodes), "seed host available")
}

func runcheckip(conf *config.Config) {
	wg := sync.WaitGroup{}
	for _, v := range nodes.nodes {
		if !v.IsNeedCheck() {
			continue
		}
		timeout := 10
		c := NewClientWithIP(ClientTypeOut, v.Ip)
		c.SetTry(1)
		wg.Add(1)
		go c.Sync(&ClientListener{
			OnClosed: func() {
				wg.Done()
			},
			OnMessage: func(msg MsgIO) {
				cmd := msg.Command()
				if cmd == NMT_VERSION {
					nodes.SetAvailable(c.IP, true, msg.(*MsgVersion))
					c.Stop()
				}
			},
			OnLoop: func() {
				timeout--
				if timeout <= 0 {
					v.Ignore = true
					c.Stop()
				}
			},
		})
	}
	wg.Wait()
}

func checkconnping(conf *config.Config) {
	nodes.mutex.Lock()
	ds := []*SeedItem{}
	for _, v := range nodes.nodes {
		if !v.Connected {
			continue
		}
		if v.Client.Ping > 0 {
			ds = append(ds, v)
		}
	}
	nodes.mutex.Unlock()
	if len(ds) <= conf.MaxOutConn {
		return
	}
	sort.Slice(ds, func(i, j int) bool {
		return ds[i].Client.Ping < ds[j].Client.Ping
	})
	//keep conf.MaxOutConn conn
	for _, v := range ds[conf.MaxOutConn:] {
		v.Client.Stop()
		nodes.mutex.Lock()
		delete(nodes.nodes, v.Ip.String())
		nodes.mutex.Unlock()
	}
	log.Println("keep", len(nodes.nodes), "seed host connection")
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
		ctimer := time.NewTimer(time.Second * 10)
		for {
			select {
			case <-ltimer.C:
				ltimer.Reset(time.Minute * 10)
			case <-ctimer.C:
				checkconnping(conf)
				ctimer.Reset(time.Second * 10)
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
