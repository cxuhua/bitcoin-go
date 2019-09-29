package net

import (
	"bytes"
	"encoding/hex"
	"io/ioutil"
	"log"
	"testing"
	"time"
)

func TestVersionMessage(t *testing.T) {
	c := NewClient(ClientTypeOut, "101.201.211.33:8333")
	c.Run()
	time.Sleep(time.Second * 3)
	//m := NewMsgGetBlocks()
	m := NewMsgGetData()
	m.Add(&Inventory{
		Type: MSG_BLOCK,
		ID:   NewHexBHash("00000000000000000011c1e53fdf22c430c55153130e8d707a5e286a9635c0ec").Swap(),
	})
	c.WC <- m

	time.Sleep(time.Hour)
}
func TestTXWNoWitness(t *testing.T) {
	data, err := hex.DecodeString("0100000001ec8abe57241834aab404a43288ce8c57061dfc47fe68312e43e2baddb01dfcfe010000006b483045022100b0b297b340fc5f42640b1dd9c96d58a4c19616045b7c921e0a0177eabec1f06402201f5efe9c400b1b5005f82ba57dd653abfcda4974bbabe00dc88e1e5b7c30c74b01210346aa8f3d640d829c7d6d40243d5e052a316ed5186bebed8da300e3a2e676ce3ffdffffff011ca0bf06000000001976a914975aa8ed4bd63ed0f388f40b4a3533425e42909f88ac00000000")
	if err != nil {
		panic(err)
	}
	h := NewNetHeader(NMT_BLOCK, data)
	tx := &TX{}
	tx.Read(h)
	if !h.IsEOF() {
		t.Errorf("read bytes remain")
	}
	if len(tx.Ins) != 1 {
		t.Errorf("txin error")
	}
	if len(tx.Outs) != 1 {
		t.Errorf("txout error")
	}
	if tx.LockTime != 0x0 {
		t.Errorf("lock time error")
	}

	oh := NewNetHeader(NMT_UNKNNOW, []byte{})
	tx.Write(oh)
	if !bytes.Equal(data, oh.Payload) {
		t.Errorf("write tx error")
	}
}

func TestTXWithWitness(t *testing.T) {
	data, err := hex.DecodeString("020000000001021773acdad1335ff086c7d068d048d26b557590fd2a4cfcd0d899dfb76fdd2aed0800000017160014eb4d74cc8dd8c821ae4a380146e37080a0a05ac5feffffff2440863590028c4de882197fb3fa3991705bcb5ffc0d92b3d09a71420f70cc4901000000171600142bda22f8efbeb2beae50589476300c80a5a5e305feffffff023a510d000000000017a914417eae8bf2973643457cd4773738024f93801b8e8710cd0e000000000017a914edb15ac37715e7ba7d47f9d7940f7fef1d177e4c8702483045022100c6580b974a2b481d8be8f21ae4b598e90716c324957fb8a70c278057d2e5e2c602203685b442a3dd26f89700e25237f5f0cc4df402c4863e354977b56210c44a1dda01210305d54457a0ff506e86c94225b9f643a9d0d89741ee2ed8cc68ed77ef4a65f0190247304402202e3f31fd55ab33deb95e1cc1c95a05ed6b45ea9a988e33790b67dda3e8fd705d0220110a1023d4b95a46201342ad786da50066088b1eab5ed2ac98b1aa4577257836012103e6221cd805d9e9b68cf7844daab7478fee13d9a979f830803cc0fb3bb30c2213311c0900")
	if err != nil {
		panic(err)
	}
	h := NewNetHeader(NMT_BLOCK, data)
	tx := &TX{}
	tx.Read(h)
	if !h.IsEOF() {
		t.Errorf("read bytes remain")
	}
	if len(tx.Ins) != 2 {
		t.Errorf("txin error")
	}
	if len(tx.Outs) != 2 {
		t.Errorf("txout error")
	}
	if tx.LockTime != 0x91c31 {
		t.Errorf("lock time error")
	}
	oh := NewNetHeader(NMT_UNKNNOW, []byte{})
	tx.Write(oh)
	if !bytes.Equal(data, oh.Payload) {
		t.Errorf("write tx error")
	}
}

func TestBlockData(t *testing.T) {
	data, err := ioutil.ReadFile("block.dat")
	if err != nil {
		panic(err)
	}
	h := NewNetHeader(NMT_BLOCK, data)
	m := NewMsgBlock()
	m.Read(h)
	log.Println(m.HashID().Swap())
	for _, v := range m.Txs {
		log.Println(v)
	}
}
