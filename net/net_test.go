package net

import (
	"log"
	"net"
	"testing"
)

func TestVersionMessage(t *testing.T) {
	//ips, err := net.LookupIP("seed.bitcoin.sipa.be")
	//if err != nil {
	//	panic(err)
	//}
	//for _, v := range ips {
	//	log.Println(v)
	//}
	m := NewMsgVersion("192.168.0.1", "124.189.37.118")

	c, err := net.Dial("tcp4", "124.189.37.118:8333")
	if err != nil {
		panic(err)
	}
	//ioutil.WriteFile("f:\\a.dat", buf.Bytes(), os.ModePerm)
	err = WriteMsg(c, m)
	if err != nil {
		panic(err)
	}
	log.Println(err)
	hh, pr, err := ReadMsg(c)
	log.Println(hh, err, hh.Command == NMT_VERSION)
	msg := &MsgVersion{}
	msg.Read(pr)

	hh, pr, err = ReadMsg(c)
	log.Println(hh.Command)

}
