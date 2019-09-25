package config

type Config struct {
	//network id
	Id string
	//dns seed names
	Seeds []string
	//message start 4bytes
	MsgStart []byte
	//subver
	SubVer string
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
	config = c
	return config
}
