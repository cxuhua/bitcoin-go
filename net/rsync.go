package net

import (
	"context"
	"log"
	"time"
)

var (
	IpChan = make(chan *SeedItem, 1024)
)

func processMsg(c *Client, msg MsgIO) {
	cmd := msg.Command()
	//switch cmd {
	//case NMT_PONG:
	//	mp := msg.(*MsgPong)
	//	nodes.SetPing(c.IP, mp.Ping())
	//}
	if cmd == NMT_PONG {
		log.Println(c.IP, "recv", cmd)
	}
}

func startconnect(item *SeedItem) {
	c := NewClientWithIP(ClientTypeOut, item.Ip)
	c.SetListener(&ClientListener{
		OnConnected: func() {
			nodes.SetConnected(item.Ip, c)
		},
		OnClosed: func() {
			nodes.SetConnected(item.Ip, nil)
		},
		OnLoop: func() {

		},
		OnMessage: func(msg MsgIO) {
			processMsg(c, msg)
		},
		OnWrite: func(msg MsgIO) {

		},
	})
	c.Run()
}

func StartRsync(ctx context.Context) {
	defer func() {
		MWG.Done()
	}()
	MWG.Add(1)
	mfx := func() {
		log.Println("resync start")
		defer func() {
			if err := recover(); err != nil {
				log.Println("[RSYNC error]:", err)
			}
		}()
		for {
			select {
			case item := <-IpChan:
				startconnect(item)
			case <-ctx.Done():
				log.Println("rsync stop", ctx.Err())
				return
			}
		}
	}
	for ctx.Err() != context.Canceled {
		time.Sleep(time.Second * 3)
		mfx()
	}
}
