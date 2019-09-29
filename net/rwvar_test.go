package net

import (
	"bitcoin/script"
	"bytes"
	"encoding/hex"
	"io/ioutil"
	"log"
	"testing"

	"golang.org/x/crypto/ripemd160"
)

func TestHash(t *testing.T) {
	log.Println(ripemd160.New().Sum([]byte{}))
}

func TestRWVarInt(t *testing.T) {
	buf := &bytes.Buffer{}
	vs := []uint64{254, 0xFD, 0xFFFF, 0xFFFF1, 0xFFFFFFFF, 0xFFFFFFFF1, 0xFFFFFFFFFF1}
	for _, v1 := range vs {
		buf.Reset()
		l1 := WriteVarInt(buf, v1)
		v2, l2 := ReadVarInt(buf)
		if l1 != l2 || v1 != v2 {
			t.Fatalf("test v1=%X v2=%X error,l1=%d,l2=%d", v1, v2, l1, l2)
		}
	}
}

func GetTestScript() *script.Script {
	data, err := ioutil.ReadFile("../dat/tx.dat")
	if err != nil {
		panic(err)
	}
	h := &MessageHeader{}
	pr := bytes.NewReader(data)
	m := NewMsgTX()
	m.Read(h, pr)
	for _, v := range m.Tx.Outs {
		return script.NewScript(v.Script.Bytes())
	}
	return nil
}

func TestTX(t *testing.T) {
	s := "010000000001025376ff679fc3e7530774cc62c118536bad95eadc96dc827812d98fe9aba64325000000001716001404a71ae9b8dddcf2b3b7d8284b361fd2ca7ec1d2ffffffff18eaeecdda7daa58dcb294b47be8ed17e53c67dd2aa256daba93e2301e78bf670100000017160014abb8dd40a17594a9fb1329dea22de37d6551ba75ffffffff02b8bbb8010000000017a91489bef25f2c8c261238a5916f410e5a6ab8477ca887656bdd010000000017a914a83facf2031751acbdccf5cadc174dc8983dbde48702483045022100aa68ba92a34a73572e20319ee1551a6bc3b326d1d8c4ef0e58a449fbb3fbe9ba022034562f8ea8e310ab84856182feee3327c37638536a418f2b6de24f77a83ff4640121029a36dc704accfb49eea080e56d451a27531e34effa96005a261db042ca128ff00247304402201e4d3c448a866bbc852dac95065170f79bc17dbc3a94777392b01cb106105b59022037bca2d40d528f83641a391f8aec3c2878d66bcc65266d501e6e869b58c28000012102e3364b889623f33204f6f41cd6c73dc005ff55b0ff56fdb90107fa78eb97339500000000"
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	tx := &TX{}
	tx.Read(bytes.NewReader(b))
	stack := script.NewStack()
	log.Println(tx.Ins[0].Output.Hash)
	i1 := tx.Ins[0].Script
	o1 := tx.Outs[0].Script
	o, err := i1.Eval(stack, nil, 0, script.SIG_VER_BASE)
	if err != nil {
		panic(err)
	}
	o, err = o1.Eval(stack, nil, 0, script.SIG_VER_BASE)
	if err != nil {
		panic(err)
	}
	log.Println(o)
}

func TestScript(t *testing.T) {
	s := script.NewScriptHex("76a9148fd139bb39ced713f231c58a4d07bf6954d1c20188ac")
	log.Printf(hex.EncodeToString(*s))
	stack := script.NewStack()
	s.Eval(stack, nil, 0, script.SIG_VER_BASE)
}

//tx.dat
func TestMsgTX(t *testing.T) {
	data, err := ioutil.ReadFile("../dat/tx.dat")
	if err != nil {
		panic(err)
	}
	h := &MessageHeader{}
	pr := bytes.NewReader(data)
	m := NewMsgTX()
	m.Read(h, pr)

	stack := script.NewStack()
	i1 := m.Tx.Ins[0].Script
	o1 := m.Tx.Outs[0].Script
	o, err := i1.Eval(stack, nil, 0, script.SIG_VER_BASE)
	if err != nil {
		panic(err)
	}
	o, err = o1.Eval(stack, nil, 0, script.SIG_VER_BASE)
	if err != nil {
		panic(err)
	}
	log.Println(o)

	for _, v := range m.Tx.Outs {
		log.Println(len(v.Script.Bytes()))
	}

}
