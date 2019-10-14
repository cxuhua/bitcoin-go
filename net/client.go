package net

import (
	"bitcoin/config"
	"bitcoin/util"
	"context"
	"fmt"
	"log"
	"net"
	"time"
)

type ClientType byte

const (
	ClientTypeIn  = ClientType(0x1)
	ClientTypeOut = ClientType(0x2)
)

type ClientListener struct {
	OnConnected func()
	OnClosed    func()
	OnLoop      func()
	OnMessage   func(m MsgIO)
	OnWrite     func(m MsgIO)
	OnError     func(err interface{})
}

type defaultLister struct {
	lis *ClientListener
}

func (l *defaultLister) OnError(err interface{}) {
	if l.lis != nil && l.lis.OnError != nil {
		l.lis.OnError(err)
	}
}

func (l *defaultLister) OnClosed() {
	if l.lis != nil && l.lis.OnClosed != nil {
		l.lis.OnClosed()
	}
}

func (l *defaultLister) OnConnected() {
	if l.lis != nil && l.lis.OnConnected != nil {
		l.lis.OnConnected()
	}
}

func (l *defaultLister) OnLoop() {
	if l.lis != nil && l.lis.OnLoop != nil {
		l.lis.OnLoop()
	}
}

func (l *defaultLister) OnMessage(m MsgIO) {
	if l.lis != nil && l.lis.OnMessage != nil {
		l.lis.OnMessage(m)
	}
}

func (l *defaultLister) OnWrite(m MsgIO) {
	if l.lis != nil && l.lis.OnWrite != nil {
		l.lis.OnWrite(m)
	}
}

type Client struct {
	net.Conn
	ctx       context.Context
	cancel    context.CancelFunc
	Type      ClientType
	wc        chan MsgIO
	rc        chan *NetHeader
	IP        net.IP
	Port      uint16
	connected bool
	try       int
	Acked     bool //is recv verack
	VerInfo   *MsgVersion
	listener  *defaultLister
	Ping      int
	Err       interface{}
	FeeRate   Amount //trans fee
}

func (c *Client) SetListener(lis *ClientListener) {
	c.listener = &defaultLister{lis: lis}
}

//sync running
func (c *Client) Sync(lis *ClientListener) {
	c.SetListener(lis)
	c.run()
}

func (c *Client) WriteMsg(m MsgIO) {
	c.wc <- m
}

func (c *Client) processMsg(m *NetHeader) {
	if c.Acked {
		m.Ver = c.VerInfo.Ver
	}
	var msg MsgIO = nil
	switch m.Command {
	case NMT_VERSION:
		mp := &MsgVersion{}
		msg = m.Full(mp)
		c.VerInfo = mp
	case NMT_VERACK:
		mp := &MsgVerAck{}
		msg = m.Full(mp)
		c.WriteMsg(NewMsgVerAck())
		c.Acked = true
		c.OnReady()
	case NMT_PING:
		mp := &MsgPing{}
		msg = m.Full(mp)
		np := NewMsgPong()
		np.Timestamp = mp.Timestamp
		c.WriteMsg(np)
	case NMT_HEADERS:
		mp := NewMsgHeaders()
		msg = m.Full(mp)
	case NMT_PONG:
		mp := NewMsgPong()
		msg = m.Full(mp)
		c.OnPong(mp)
	case NMT_SENDHEADERS:
		mp := NewMsgSendHeaders()
		msg = m.Full(mp)
	case NMT_SENDCMPCT:
		mp := NewMsgSendCmpct()
		msg = m.Full(mp)
	case NMT_GETHEADERS:
		mp := NewMsgGetHeaders()
		msg = m.Full(mp)
	case NMT_FEEFILTER:
		mp := NewMsgFeeFilter()
		msg = m.Full(mp)
		c.FeeRate = Amount(mp.FeeRate)
	case NMT_INV:
		mp := NewMsgINV()
		msg = m.Full(mp)
	case NMT_NOTFOUND:
		mp := NewMsgNotFound()
		msg = m.Full(mp)
	case NMT_TX:
		mp := NewMsgTX()
		msg = m.Full(mp)
	case NMT_BLOCK:
		mp := NewMsgBlock()
		msg = m.Full(mp)
	case NMT_ADDR:
		am := NewMsgAddr()
		msg = m.Full(am)
	case NMT_REJECT:
		mp := NewMsgReject()
		msg = m.Full(mp)
	case NMT_ALERT:
		mp := NewMsgAlert()
		msg = m.Full(mp)
	case NMT_MERKLEBLOCK:
		mp := NewMsgMerkleBlock()
		msg = m.Full(mp)
	default:
		log.Println(m.Command, " not process", c.IP)
		return
	}
	c.listener.OnMessage(msg)
}

func (c *Client) OnReady() {
	c.WriteMsg(NewMsgPing())
}

func (c *Client) OnPong(msg *MsgPong) {
	c.Ping = msg.Ping()
}

func (c *Client) OnClosed() {
	if c.Type == ClientTypeOut {
		OutIps.Del(c)
	} else if c.Type == ClientTypeIn {
		InIps.Del(c)
	}
	c.listener.OnClosed()
}

func (c *Client) IsConnected() bool {
	return c.connected
}

func (c *Client) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.IP.String(), c.Port)
}

func (c *Client) OnConnected() {
	if c.Type == ClientTypeOut {
		OutIps.Set(c)
		addr := c.Conn.LocalAddr()
		mp := NewMsgVersion(addr.String(), c.GetAddr())
		c.WriteMsg(mp)
	} else if c.Type == ClientTypeIn {
		InIps.Set(c)
	}
	c.listener.OnConnected()
}

func (c *Client) SetTry(v int) *Client {
	c.try = v
	return c
}

func (c *Client) OnError(err interface{}) {
	c.Err = err
	c.listener.OnError(err)
}

func (c *Client) Key() string {
	return net.JoinHostPort(c.IP.String(), fmt.Sprintf("%d", c.Port))
}

func (c *Client) stop() {
	if err := recover(); err != nil {
		c.OnError(err)
	}
	if c.connected {
		close(c.wc)
		close(c.rc)
		c.Close()
		c.connected = false
	}
	c.Acked = false
	c.OnClosed()
}

func (c *Client) init() {
	//init write msg queue
	c.wc = make(chan MsgIO, 10)
	//init read msg queue
	c.rc = make(chan *NetHeader, 10)
}

func (c *Client) run() {
	defer c.stop()
	c.init()
	//connect host
	for c.Type == ClientTypeOut && !c.connected {
		err := c.Connect()
		if err != nil {
			c.try--
			if c.try > 0 {
				time.Sleep(time.Second)
				continue
			}
		}
		if !c.connected {
			c.OnError(fmt.Errorf("connect error %v", err))
			c.cancel()
			return
		}
		c.OnConnected()
	}
	//start loop readmsg
	go func() {
		defer func() {
			if err := recover(); err != nil {
				c.OnError(err)
			}
			c.cancel()
		}()
		for {
			m, err := ReadMsg(c)
			if err != nil {
				panic(fmt.Errorf("read msg error %v", err))
			}
			c.rc <- m
		}
	}()
	//not recv ack timeout
	vtimer := time.NewTimer(time.Second * 5)
	//loop timer
	ltimer := time.NewTimer(time.Second)
	//
	ptimer := time.NewTimer(time.Second * 60)
	//start loop process timer msg
	for {
		select {
		case wp := <-c.wc:
			err := WriteMsg(c, wp)
			if err != nil {
				panic(fmt.Errorf("write msg error %v", err))
			}
			c.listener.OnWrite(wp)
		case rp := <-c.rc:
			c.processMsg(rp)
		case <-vtimer.C:
			if !c.Acked {
				c.cancel()
			}
		case <-ltimer.C:
			c.OnLoop()
			ltimer.Reset(time.Second)
		case <-ptimer.C:
			if !c.Acked {
				break
			}
			c.WriteMsg(NewMsgPing())
			ptimer.Reset(time.Second * 60)
		case <-c.ctx.Done():
			return
		}
	}
}

func (c *Client) OnLoop() {
	c.listener.OnLoop()
}

func (c *Client) Stop() {
	c.Err = fmt.Errorf("client stop,will close")
	c.cancel()
}

func (c *Client) Run() {
	go c.run()
}

func (c *Client) Connect() error {
	conn, err := net.DialTimeout("tcp", c.Key(), time.Second*10)
	if err != nil {
		return err
	}
	c.Conn = conn
	c.connected = true
	return nil
}

func NewClientWithIPPort(typ ClientType, ip net.IP, port uint16) *Client {
	c := &Client{}
	c.connected = false
	c.IP = ip
	c.Port = port
	c.Type = typ
	c.try = 3
	c.ctx, c.cancel = context.WithCancel(context.Background())
	c.listener = &defaultLister{}
	return c
}

func NewClientWithIP(typ ClientType, ip net.IP) *Client {
	conf := config.GetConfig()
	return NewClientWithIPPort(typ, ip, uint16(conf.ListenPort))
}

func NewClient(typ ClientType, addr string) *Client {
	ip, port := util.ParseAddr(addr)
	return NewClientWithIPPort(typ, ip, port)
}
