package config

import (
	"strconv"
	"strings"
)

type Config struct {
	//network id
	Id string
	//dns seed names
	Seeds []string
	//message start 4bytes
	MsgStart []byte
	//subver
	SubVer string
	//local listen ip port
	LocalAddr string //ip:port
	//
	SegwitHeight uint
}

func (c Config) GetLocalAddr() (string, uint16) {
	ip := ""
	pv := uint16(8333)
	ss := strings.Split(c.LocalAddr, ":")
	if len(ss) > 0 {
		ip = ss[0]
	}
	if len(ss) > 1 {
		v, err := strconv.ParseInt(ss[1], 10, 32)
		if err != nil {
			panic(err)
		}
		pv = uint16((v))
	}
	return ip, pv
}

var (
	config *Config = nil
)

//main network config
func GetConfig() *Config {
	if config != nil {
		return config
	}
	c := &Config{}
	c.Id = "main"
	c.MsgStart = []byte{0xF9, 0xBE, 0xB4, 0xD9}
	c.Seeds = []string{
		"seed.bitcoin.sipa.be",
		"dnsseed.bluematt.me",
		"dnsseed.bitcoin.dashjr.org",
		"seed.bitcoinstats.com",
		"seed.bitcoin.jonasschnelli.ch",
		"seed.btc.petertodd.org",
		"seed.bitcoin.sprovoost.nl",
		"dnsseed.emzy.de",
	}
	c.SubVer = "/golang:0.1.0/"
	c.LocalAddr = "192.168.31.198:8333"
	c.SegwitHeight = 481824
	config = c
	return config
}
