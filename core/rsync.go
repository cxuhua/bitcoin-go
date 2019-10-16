package core

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

type IPPort struct {
	ip   net.IP
	port int
}

func (p IPPort) IsEnable() bool {
	return p.ip.IsGlobalUnicast()
}

func (p IPPort) Key() string {
	return net.JoinHostPort(p.ip.String(), fmt.Sprintf("%d", p.port))
}

type AddrMap struct {
	mu  sync.Mutex
	ips map[string]int64
}

func NewAddrMap() *AddrMap {
	return &AddrMap{ips: map[string]int64{}}
}

type ClientMap struct {
	mu    sync.Mutex
	nodes map[string]*Client
}

func NewClientMap() *ClientMap {
	return &ClientMap{nodes: map[string]*Client{}}
}

//find Fastest networdk
func (m *ClientMap) Fastest(num int) []*Client {
	m.mu.Lock()
	defer m.mu.Unlock()
	ds := []*Client{}
	for _, v := range m.nodes {
		if v.Ping > 0 {
			ds = append(ds, v)
		}
	}
	sort.Slice(ds, func(i, j int) bool {
		return ds[i].Ping < ds[j].Ping
	})
	if num > len(ds) {
		num = len(ds)
	}
	return ds[:num]
}

func (m *ClientMap) All(f func(c *Client)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, v := range m.nodes {
		f(v)
	}
}

func (m *ClientMap) Has(c *Client) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.nodes[c.Key()]
	return ok
}

func (m *ClientMap) Len() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.nodes)
}

func (m *ClientMap) Set(c *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nodes[c.Key()] = c
}

func (m *ClientMap) Del(c *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.nodes, c.Key())
}

var (
	IpChan   = make(chan IPPort, 1024)
	OutIps   = NewClientMap()
	InIps    = NewClientMap()
	RecvAddr = make(chan *MsgAddr, 10)
)

func startconnect(ip IPPort) {
	conf := config.GetConfig()
	if !ip.IsEnable() {
		return
	}
	c := NewClientWithIPPort(ClientTypeOut, ip)
	//don't connect self
	if c.Key() == conf.GetLocalAddr() {
		return
	}
	//don't repeat connect
	if OutIps.Has(c) {
		return
	}
	c.SetListener(&ClientListener{
		OnConnected: func() {

		},
		OnClosed: func() {
			log.Println(c.Key(), "Closed error = ", c.Err)
		},
		OnLoop: func() {

		},
		OnMessage: func(msg MsgIO) {
			cmd := msg.Command()
			switch cmd {
			case NMT_BLOCK, NMT_TX, NMT_HEADERS, NMT_INV:
				WorkerQueue <- NewWorkerUnit(msg, c)
			case NMT_ADDR:
				RecvAddr <- msg.(*MsgAddr)
			}
		},
		OnWrite: func(msg MsgIO) {

		},
		OnError: func(err interface{}) {
			//log.Println(c.IP, "close err ", err)
		},
	})
	log.Println("start connect", c.IP, c.Port)
	c.Run()
}

func printStatus() {
	log.Println("Out Count=", OutIps.Len(), "In Count=", InIps.Len(), "Conn Queue=", len(IpChan))
	ds := OutIps.Fastest(5)
	for _, v := range ds {
		log.Println(v.IP, v.Ping)
	}
}

func processAddrs(addr *MsgAddr) {
	for _, v := range addr.Addrs {
		if v.Service&NODE_NETWORK != 0 {
			continue
		}
		startconnect(IPPort{v.IpAddr, int(v.Port)})
	}
}

func StartDispatch(ctx context.Context) {
	defer func() {
		MWG.Done()
	}()
	MWG.Add(1)
	mfx := func() {
		log.Println("dispatch start")
		defer func() {
			if err := recover(); err != nil {
				log.Println("[dispatch error]:", err)
			}
		}()
		stimer := time.NewTimer(time.Second * 5)
		for {
			select {
			case addrs := <-RecvAddr:
				processAddrs(addrs)
			case <-stimer.C:
				printStatus()
				stimer.Reset(time.Second * 5)
			case ip := <-IpChan:
				startconnect(ip)
			case <-ctx.Done():
				log.Println("dispatch stop", ctx.Err())
				return
			}
		}
	}
	for ctx.Err() != context.Canceled {
		time.Sleep(time.Second * 3)
		mfx()
	}
}
