package core

import (
	"bitcoin/config"
	"bitcoin/util"
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

func (p IPPort) Equal(v IPPort) bool {
	return p.Key() == v.Key()
}

func (p IPPort) IsEnable() bool {
	return p.ip.IsGlobalUnicast()
}

func (p IPPort) Key() string {
	return net.JoinHostPort(p.ip.String(), fmt.Sprintf("%d", p.port))
}

type AddrState int

const (
	AddrStateOpen  = AddrState(1)
	AddrStateClose = 2
	AddrStatePush  = 3
)

type AddrElement struct {
	IP        IPPort
	LastTime  time.Time
	WriteTime time.Time
	ReadTime  time.Time
	OpenTime  time.Time
	CloseTime time.Time
	State     AddrState
	Err       interface{}
}

type AddrMap struct {
	mu  sync.Mutex
	ips map[string]*AddrElement
}

func NewAddrMap() *AddrMap {
	return &AddrMap{ips: map[string]*AddrElement{}}
}

func (a *AddrMap) Ips() []IPPort {
	a.mu.Lock()
	defer a.mu.Unlock()
	ips := []IPPort{}
	now := time.Now()
	for _, v := range a.ips {
		if now.Sub(v.OpenTime) < 5*time.Minute {
			continue
		}
		if v.State != AddrStateClose {
			continue
		}
		ips = append(ips, v.IP)
	}
	return ips
}

func (a *AddrMap) SetError(ip IPPort, err interface{}) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if v, b := a.ips[ip.Key()]; b {
		v.Err = err
	}
}

func (a *AddrMap) Iter(f func(a *AddrElement)) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, v := range a.ips {
		f(v)
	}
}

func (a *AddrMap) Set(ip IPPort) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.ips[ip.Key()] = &AddrElement{
		IP:       ip,
		LastTime: time.Now(),
		State:    AddrStatePush,
	}
}

func (a *AddrMap) UpRead(ip IPPort) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if v, b := a.ips[ip.Key()]; b {
		v.ReadTime = time.Now()
	}
}

func (a *AddrMap) UpWrite(ip IPPort) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if v, b := a.ips[ip.Key()]; b {
		v.WriteTime = time.Now()
	}
}

func (a *AddrMap) Update(c *Client) {
	a.mu.Lock()
	defer a.mu.Unlock()
	v, b := a.ips[c.Key()]
	if !b {
		return
	}
	v.LastTime = time.Now()
	if v.LastTime.Sub(v.ReadTime) > time.Minute*30 {
		c.Stop()
	} else if v.LastTime.Sub(v.WriteTime) > time.Minute*30 {
		c.Stop()
	}
}
func (a *AddrMap) IsConnect(ip IPPort) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	v, b := a.ips[ip.Key()]
	return !b || v.State == AddrStateClose
}

func (a *AddrMap) Open(ip IPPort) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if v, b := a.ips[ip.Key()]; b {
		v.State = AddrStateOpen
		v.Err = nil
		v.OpenTime = time.Now()
	}
}

func (a *AddrMap) Close(ip IPPort) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if v, b := a.ips[ip.Key()]; b {
		v.State = AddrStateClose
		v.CloseTime = time.Now()
	}
}

func (a *AddrMap) Len() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return len(a.ips)
}

func (a *AddrMap) Has(ip IPPort) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	_, ok := a.ips[ip.Key()]
	return ok
}

func (a *AddrMap) Del(ip IPPort) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.ips, ip.Key())
}

type ClientMap struct {
	mu    sync.Mutex
	nodes map[string]*Client
}

func NewClientMap() *ClientMap {
	return &ClientMap{nodes: map[string]*Client{}}
}

//find Fastest networdk
func (m *ClientMap) Fastest(num int, stop bool) []*Client {
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
	//close slowest
	if stop && len(ds) > num {
		for _, v := range ds[num:] {
			v.Stop()
		}
	}
	if num > len(ds) {
		num = len(ds)
	}
	return ds[:num]
}

func (m *ClientMap) Iter(f func(c *Client) bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, v := range m.nodes {
		if f(v) {
			break
		}
	}
}

func (m *ClientMap) Has(c *Client) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.nodes[c.Key()]
	return ok
}

func (m *ClientMap) AnyWrite(msg MsgIO) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	ds := []*Client{}
	for _, v := range m.nodes {
		ds = append(ds, v)
	}
	sort.Slice(ds, func(i, j int) bool {
		return ds[i].Ping < ds[j].Ping
	})
	if len(ds) == 0 {
		return false
	}
	nv := uint16(0)
	util.SetRandInt(&nv)
	nv = nv % uint16(len(ds))
	ds[nv].WriteMsg(msg)
	return true
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
	Addrs    = NewAddrMap()
)

func startconnect(ip IPPort) {
	c := NewClientWithIPPort(ClientTypeOut, ip)
	c.SetListener(&ClientListener{
		OnConnected: func() {
			Addrs.Open(c.IP)
		},
		OnClosed: func() {
			Addrs.Close(c.IP)
			//log.Println(c.Key(), "Closed error = ", c.Err)
		},
		OnLoop: func() {
			Addrs.Update(c)
		},
		OnMessage: func(msg MsgIO) {
			cmd := msg.Command()
			switch cmd {
			case NMT_HEADERS, NMT_GETHEADERS:
				WorkerQueue <- NewWorkerUnit(msg, c)
			case NMT_BLOCK, NMT_TX, NMT_INV:
				WorkerQueue <- NewWorkerUnit(msg, c)
			case NMT_ADDR:
				RecvAddr <- msg.(*MsgAddr)
			}
			Addrs.UpRead(c.IP)
		},
		OnWrite: func(msg MsgIO) {
			Addrs.UpWrite(c.IP)
		},
		OnError: func(err interface{}) {
			//log.Println(c.IP, "close err ", err)
			Addrs.SetError(c.IP, err)
		},
	})
	log.Println("start connect :", c.Key())
	c.Run()
}

func checkStatus(conf *config.Config) {
	//log.Println("Out Count=", OutIps.Len(), "Addrs Count=", Addrs.Len())
	//ds := OutIps.Fastest(conf.MaxOutConn, true)
	//for i, v := range ds {
	//	log.Println("Fastest ip", i, v.IP.ip.String(), v.IP.port, v.Ping)
	//}
	OutIps.Fastest(conf.MaxOutConn, true)
}

func processAddrs(addr *MsgAddr) {
	for _, v := range addr.Addrs {
		if v.Service&NODE_NETWORK != 0 {
			continue
		}
		ip := IPPort{v.IpAddr, int(v.Port)}
		Addrs.Set(ip)
	}
}

const (
	NoticeSaveHeadersOK = 1
)

var (
	Notice = make(chan int, 10)
	Blocks = NewBlockHeaderList()
)

func syncBlock(conf *config.Config) {
	if OutIps.Len() == 0 {
		return
	}
	if G.LastBlock() == nil {
		m := NewMsgGetData()
		m.Add(Inventory{
			Type: MSG_BLOCK,
			ID:   NewHashID(conf.GenesisBlock),
		})
		OutIps.AnyWrite(m)
	} else if Blocks.Len() == 0 {
		m := NewMsgGetHeaders()
		m.AddHash(G.LastBlock().Hash)
		OutIps.AnyWrite(m)
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
		conf := config.GetConfig()
		ctimer := time.NewTimer(time.Second * 5)
		stimer := time.NewTimer(time.Second * 10)
		for {
			select {
			case v := <-Notice:
				switch v {
				case NoticeSaveHeadersOK:
					stimer.Reset(time.Millisecond * 50)
				}
			case addrs := <-RecvAddr:
				processAddrs(addrs)
			case <-stimer.C:
				syncBlock(conf)
				stimer.Reset(time.Second * 10)
			case <-ctimer.C:
				checkStatus(conf)
				ctimer.Reset(time.Second * 5)
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
