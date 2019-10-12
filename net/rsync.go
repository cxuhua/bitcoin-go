package net

import (
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
	IpChan = make(chan net.IP, 1024)
	OutIps = &ClientMap{nodes: map[string]*Client{}}
	InIps  = &ClientMap{nodes: map[string]*Client{}}
)

func processMsg(c *Client, msg MsgIO) {
	//cmd := msg.Command()
	//switch cmd {
	//case NMT_PONG:
	//	mp := msg.(*MsgPong)
	//	nodes.SetPing(c.IP, mp.Ping())
	//}
	//if cmd == NMT_PONG {
	//
	//}
	//log.Println(c.IP, "recv", cmd)
}

func startconnect(ip net.IP) {
	c := NewClientWithIP(ClientTypeOut, ip)
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
			processMsg(c, msg)
		},
		OnWrite: func(msg MsgIO) {

		},
		OnError: func(err interface{}) {
			//log.Println(c.IP, "close err ", err)
		},
	})
	c.Run()
}

func printStatus() {
	log.Println("Out Count=", OutIps.Len(), "In Count=", InIps.Len(), "Conn Queue=", len(IpChan))
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
