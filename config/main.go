package config

import (
	"fmt"
	"net"
)

const (
	PUBKEY_ADDRESS = iota
	SCRIPT_ADDRESS
	SECRET_KEY
	EXT_PUBLIC_KEY
	EXT_SECRET_KEY
)

type Config struct {
	LocalIP    string
	ListenAddr string
	//default port
	ListenPort int
	//max connected me client
	MaxInConn int
	//max connect to count
	MaxOutConn int
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

	b58prefixs map[int][]byte
	//210000
	SubHalving int
	//bech32 address prefix
	Bech32HRP string
}

func (c Config) Base58Prefix(idx int) []byte {
	return c.b58prefixs[idx]
}

func (c Config) GetLocalAddr() string {
	return net.JoinHostPort(c.LocalIP, fmt.Sprintf("%d", c.ListenPort))
}

var (
	config *Config = nil
)

//main network config
func GetConfig() *Config {
	if config != nil {
		return config
	}

	c := &Config{Id: "main"}

	c.LocalIP = "192.168.31.198"
	c.ListenAddr = "0.0.0.0"
	c.ListenPort = 8333

	c.MaxInConn = 50

	c.MaxOutConn = 10

	c.b58prefixs = map[int][]byte{}
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
	//
	c.b58prefixs[PUBKEY_ADDRESS] = []byte{0}
	c.b58prefixs[SCRIPT_ADDRESS] = []byte{5}
	c.b58prefixs[SECRET_KEY] = []byte{128}
	c.b58prefixs[EXT_PUBLIC_KEY] = []byte{0x04, 0x88, 0xB2, 0x1E}
	c.b58prefixs[EXT_SECRET_KEY] = []byte{0x04, 0x88, 0xAD, 0xE4}
	//
	c.SubHalving = 210000
	c.Bech32HRP = "bc"
	config = c
	return config
}
