package net

import (
	"bitcoin/config"
	"context"
	"log"
	"net"
	"sync"
	"time"
)

type ClientMap struct {
	mu    sync.Mutex
	nodes map[string]*Client
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
	IpChan   = make(chan net.IP, 1024)
	OutIps   = &ClientMap{nodes: map[string]*Client{}}
	InIps    = &ClientMap{nodes: map[string]*Client{}}
	RecvAddr = make(chan *MsgAddr, 10)
)

func startconnect(ip net.IP, port int) {
	conf := config.GetConfig()
	if !ip.IsGlobalUnicast() {
		return
	}
	c := NewClientWithIPPort(ClientTypeOut, ip, uint16(port))
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
			case NMT_BLOCK, NMT_TX, NMT_HEADERS:
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
}

func processAddrs(addr *MsgAddr) {
	for _, v := range addr.Addrs {
		if v.Service&NODE_NETWORK != 0 {
			continue
		}
		startconnect(v.IpAddr, int(v.Port))
	}
}

func StartDispatch(ctx context.Context) {
	conf := config.GetConfig()
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
				startconnect(ip, conf.ListenPort)
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
