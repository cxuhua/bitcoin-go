package net

import (
	"bitcoin/util"
	"context"
	"fmt"
	"log"
	"net"
	"time"
)

const (
	ClientTypeIn  = 0x1
	ClientTypeOut = 0x2
)

type Client struct {
	net.Conn
	ctx       context.Context
	cancel    context.CancelFunc
	Type      int
	WC        chan MsgIO
	RC        chan *NetMessage
	IP        net.IP
	Port      uint16
	connected bool
	try       int
	ptimer    *time.Timer
	Acked     bool //is recv verack
	VerInfo   *MsgVersion
	ping      int
}

func (c *Client) processMsg(m *NetMessage) error {
	if c.Acked {
		m.Header.Ver = c.VerInfo.Ver
	}
	log.Println("CMD=", m.Header.Command, "PAYLOAD Size=", m.Header.PayloadLen)
	switch m.Header.Command {
	case NMT_VERSION:
		msg := &MsgVersion{}
		m.Full(msg)
		c.VerInfo = msg
	case NMT_VERACK:
		c.WC <- NewMsgVerAck()
		c.WC <- NewMsgPing()
		c.Acked = true
		c.OnReady()
	case NMT_PING:
		c.WC <- NewMsgPong()
	case NMT_HEADERS:
		mp := NewMsgHeaders()
		m.Full(mp)
	case NMT_PONG:
		pong := NewMsgPong()
		m.Full(pong)
		c.ping = pong.Ping()
		c.OnPing()
	case NMT_SENDHEADERS:
		mp := NewMsgSendHeaders()
		m.Full(mp)
	case NMT_SENDCMPCT:
		mp := NewMsgSendCmpct()
		m.Full(mp)
	case NMT_GETHEADERS:
		mp := NewMsgGetHeaders()
		m.Full(mp)
	case NMT_FEEFILTER:
		mp := NewMsgFeeFilter()
		m.Full(mp)
		log.Println("feerate = ", mp.FeeRate)
	case NMT_INV:
		mp := NewMsgINV()
		m.Full(mp)
		//if len(mp.Invs) > 0 {
		//	log.Println(hex.EncodeToString(mp.Invs[0].Hash[:]))
		//}
	case NMT_NOTFOUND:
		mp := NewMsgNotFound()
		m.Full(mp)
		log.Println("NMT_NOTFOUND", mp)
	case NMT_TX:
		mp := NewMsgTX()
		m.Full(mp)
		log.Println(mp.Tx)
	case NMT_BLOCK:
		mp := NewMsgBlock()
		m.Full(mp)
	case NMT_ADDR:
		am := NewMsgAddr()
		m.Full(am)
	case NMT_REJECT:
		mp := NewMsgReject()
		m.Full(mp)
		log.Println("NMT_REJECT", mp.Message, mp.Reason)
	default:
		log.Println(m.Header.Command, " not process")
	}
	return nil
}

func (c *Client) OnReady() {
	c.WC <- NewMsgGetAddr()
}

func (c *Client) OnPing() {

}

func (c *Client) OnClosed() {

}

func (c *Client) IsConnected() bool {
	return c.connected
}

func (c *Client) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.IP.String(), c.Port)
}

func (c *Client) OnConnected() {
	if c.Type == ClientTypeOut {
		addr := c.Conn.LocalAddr()
		c.WC <- NewMsgVersion(addr.String(), c.GetAddr())
	}
}

func (c *Client) stop() {
	err := recover()
	if err == nil {
		err = c.ctx.Err()
	} else {
		err = fmt.Errorf("err = %v , ctx err = %v", err, c.ctx.Err())
	}
	if err != nil {
		log.Println("client run finished ", err)
	}
	if c.connected {
		c.Close()
		c.connected = false
	}
	c.OnClosed()
}

func (c *Client) run() {
	//defer c.stop()
	for !c.connected {
		err := c.Connect()
		if err != nil {
			log.Println("connect error", err)
			time.Sleep(time.Second * 3)
			c.try--
		}
		if c.try <= 0 && !c.connected {
			log.Println("try connect failed,try == 0")
			c.cancel()
			return
		}
		c.OnConnected()
	}
	go func() {
		for {
			m, err := ReadMsg(c)
			if err != nil {
				log.Println("read message error", err)
				c.cancel()
				break
			}
			c.RC <- m
		}
	}()
	//not recv ack timeout
	vertimer := time.NewTimer(time.Second * 5)
	for {
		select {
		case wp := <-c.WC:
			err := WriteMsg(c, wp)
			if err != nil {
				log.Println("write msg error", err)
				c.cancel()
			}
		case rp := <-c.RC:
			err := c.processMsg(rp)
			if err != nil {
				log.Println("process msg error", err)
				c.cancel()
			}
		case <-vertimer.C:
			if !c.Acked {
				c.cancel()
			}
		case <-c.ptimer.C:
			if !c.Acked {
				break
			}
			c.WC <- NewMsgPing()
			c.ptimer.Reset(time.Second * 60)
		case <-c.ctx.Done():
			log.Println("done! close socket")
			return
		}
	}
}

func (c *Client) Stop() {
	c.cancel()
}

func (c *Client) Run() {
	go c.run()
}

func (c *Client) Connect() error {
	addr := fmt.Sprintf("%s:%d", c.IP, c.Port)
	conn, err := net.DialTimeout("tcp", addr, time.Second*5)
	if err != nil {
		return err
	}
	c.Conn = conn
	c.connected = true
	return nil
}

func NewClient(typ int, addr string) *Client {
	ip, port := util.ParseAddr(addr)
	c := &Client{}
	c.connected = false
	c.IP = ip
	c.Port = port
	c.Type = typ
	c.WC = make(chan MsgIO, 10)
	c.RC = make(chan *NetMessage, 10)
	c.try = 3
	c.ptimer = time.NewTimer(time.Second * 60)
	c.ctx, c.cancel = context.WithCancel(context.Background())
	return c
}
