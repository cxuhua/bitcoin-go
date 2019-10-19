package config

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
	BIP16Exception string
	BIP34Height    int
	BIP34Hash      string
	BIP65Height    int
	BIP66Height    int
	CSVHeight      int
	SegwitHeight   int

	b58prefixs map[int][]byte
	//210000
	SubHalving int
	//bech32 address prefix
	Bech32HRP string
	//
	GenesisBlock string

	PowLimit                string
	PowTargetTimespan       int
	PowTargetSpacing        int
	MinerConfirmationWindow int
}

//DifficultyAdjustmentInterval
func (c Config) DiffAdjusInterval() int {
	return c.PowTargetTimespan / c.PowTargetSpacing
}

func (c Config) Base58Prefix(idx int) []byte {
	return c.b58prefixs[idx]
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

	c.MaxInConn = 5

	c.MaxOutConn = 5

	c.b58prefixs = map[int][]byte{}
	c.MsgStart = []byte{0xF9, 0xBE, 0xB4, 0xD9}

	c.PowLimit = "00000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
	c.PowTargetTimespan = 14 * 24 * 60 * 60 // two weeks
	c.PowTargetSpacing = 10 * 60
	c.MinerConfirmationWindow = c.PowTargetTimespan / c.PowTargetSpacing
	c.GenesisBlock = "000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f"

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

	c.BIP16Exception = "00000000000002dc756eebf4f49723ed8d30cc28a5f108eb94b1ba88ac4f9c22"
	c.BIP34Height = 227931
	c.BIP34Hash = "000000000000024b89b42a942fe0d9fea3bb44ab7bd1b19115dd6a759c0808b8"
	c.BIP65Height = 388381  // 000000000000000004c2b624ed5d7756c508d90fd0da2c7c679febfa6c4735f0
	c.BIP66Height = 363725  // 00000000000000000379eaa19dce8c9b722d46ae6a57c2f1a988119488b50931
	c.CSVHeight = 419328    // 000000000000000004a1b34462cb8aeebd5799177f7a29cf28f2d1961716b5b5
	c.SegwitHeight = 481824 // 0000000000000000001c8018d9cb3b742ef25114f27563e3fc4a1902167f9893
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
