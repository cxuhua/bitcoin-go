package net

import (
	"log"
	"net"
	"testing"

	ping "github.com/cxuhua/go-ping"
)

func TestPing(t *testing.T) {
	pinger, err := ping.NewPinger("127.0.0.1")
	if err != nil {
		panic(err)
	}
	pinger.Count = 3
	pinger.Run()                 // blocks until finished
	stats := pinger.Statistics() // get send/receive/rtt stats
	log.Println(stats)
}

func TestVersionMessage(t *testing.T) {
	//ips, err := net.LookupIP("seed.bitcoin.sipa.be")
	//if err != nil {
	//	panic(err)
	//}
	//for _, v := range ips {
	//	log.Println(v)
	//}
	m := NewMsgVersion("192.168.0.1", "124.189.37.118")
	b, err := ToMessageBytes(m)
	if err != nil {
		panic(err)
	}

	c, err := net.Dial("tcp4", "124.189.37.118:8333")
	if err != nil {
		panic(err)
	}
	//ioutil.WriteFile("f:\\a.dat", buf.Bytes(), os.ModePerm)
	x, err := c.Write(b)
	if err != nil {
		panic(err)
	}
	log.Println(x, err)
	hh, pr, err := FromMessageBytes(c)
	log.Println(hh, err, hh.Command == NMT_VERSION)
	msg := &MsgVersion{}
	msg.Read(pr)
	log.Println(msg)
}
