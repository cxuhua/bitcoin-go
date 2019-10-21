package core

import (
	"fmt"
	"io/ioutil"
	"testing"
)

type testtxcacher struct {
	items map[HashID]interface{}
}

func (t *testtxcacher) Set(id HashID, v interface{}) (interface{}, error) {
	t.items[id] = v
	return v, nil
}

func (t *testtxcacher) Del(id HashID) {
	delete(t.items, id)
}

func (t *testtxcacher) Get(id HashID) (interface{}, error) {
	data, err := ioutil.ReadFile(fmt.Sprintf("../dat/tx%v.dat", id))
	if err != nil {
		return nil, err
	}
	h := NewNetHeader(data)
	tx := &TX{}
	tx.Read(h)
	return t.Set(id, tx)
}

func NewTestFileCacher() ICacher {
	return &testtxcacher{
		items: map[HashID]interface{}{},
	}
}

//8d5bc6ff636d9cfb3a3b37cc2ad7681e5ba8078d8c7eb4a47531d75c18c8487f
func TestP2WPKHSign(t *testing.T) {
	Txs.Push(NewTestFileCacher())
	defer Txs.Pop()
	id := NewHashID("8d5bc6ff636d9cfb3a3b37cc2ad7681e5ba8078d8c7eb4a47531d75c18c8487f")
	tx2, err := LoadTx(id)
	if err != nil {
		t.Errorf("load tx error %v", err)
	}
	for i, v := range tx2.Outs {
		if i == 0 && v.Script.GetAddress() != "395MUYNnnhaUDhm4VDSKn7jtafQbU5kXRB" {
			t.Errorf("get out %d address", i)
		}
	}
	if err := VerifyTX(tx2, 0); err != nil {
		t.Errorf("verify tx error %v", err)
	}
}

func TestP2SHMSIGSign(t *testing.T) {
	Txs.Push(NewTestFileCacher())
	defer Txs.Pop()
	id := NewHashID("c7f04832fc99b87a0140da2377ec81d1e1a062ed72f507f84533e572db1f6d15")
	tx2, err := LoadTx(id)
	if err != nil {
		panic(err)
	}
	for i, v := range tx2.Outs {
		if i == 0 && v.Script.GetAddress() != "3AAhq47sBv78RWNTWF5vsAeDdWmA2EqV88" {
			panic(fmt.Errorf("get out %d address", i))
		}
		if i == 1 && v.Script.GetAddress() != "3BMEXQxztwFkN3E6FSf3VuGNTeUQzG41Vf" {
			panic(fmt.Errorf("get out %d address", i))
		}
	}
	if err := VerifyTX(tx2, 0); err != nil {
		t.Errorf("Verify test failed  err=%v", err)
	}
}

func TestP2WSHMSIGSign(t *testing.T) {
	Txs.Push(NewTestFileCacher())
	defer Txs.Pop()
	id := NewHashID("2cc59f3c646b3917ed9b5224f71b335a2eab70ca4610a01dee90c2536d35d940")
	tx2, err := LoadTx(id)
	if err != nil {
		panic(err)
	}
	for i, v := range tx2.Outs {
		if i == 0 && v.Script.GetAddress() != "3EMvHQQrqHuX8vDBtW6SATSdVYPX2Yc529" {
			panic(fmt.Errorf("get out %d address", i))
		}
		if i == 1 && v.Script.GetAddress() != "bc1qwqdg6squsna38e46795at95yu9atm8azzmyvckulcc7kytlcckxswvvzej" {
			panic(fmt.Errorf("get out %d address", i))
		}
	}
	if err := VerifyTX(tx2, 0); err != nil {
		t.Errorf("Verify test failed  err=%v", err)
	}
}

func TestP2SHWPKHSign(t *testing.T) {
	Txs.Push(NewTestFileCacher())
	defer Txs.Pop()
	id := NewHashID("0ae88f93be14b77994da8ebb948e817e6fbb98d66c0091366e46df0663ea3813")
	tx2, err := LoadTx(id)
	if err != nil {
		panic(err)
	}
	for i, v := range tx2.Outs {
		if i == 0 && v.Script.GetAddress() != "3GDiJ4gRqnzAws1bFvkBwimh8Pykx5cUPi" {
			panic(fmt.Errorf("get out %d address", i))
		}
		if i == 1 && v.Script.GetAddress() != "3FAX1sAtk1NTVpjLuNYJEp9D489ZvrRsvW" {
			panic(fmt.Errorf("get out %d address", i))
		}
	}
	if err := VerifyTX(tx2, 0); err != nil {
		t.Errorf("Verify test failed  err=%v", err)
	}
}

func TestP2PKSign(t *testing.T) {
	Txs.Push(NewTestFileCacher())
	defer Txs.Pop()
	id := NewHashID("80d417567b5a032465474052cca4dc38c57f6d5dc10dc7519b2ca20ac7d5512b")
	tx2, err := LoadTx(id)
	if err != nil {
		panic(err)
	}
	for i, v := range tx2.Outs {
		if i == 0 && v.Script.GetAddress() != "1VayNert3x1KzbpzMGt2qdqrAThiRovi8" {
			panic(fmt.Errorf("get out %d address", i))
		}
		if i == 1 && v.Script.GetAddress() != "1AvxGSFo8sVJKkfwHhtt6stHyuKUyLaKZp" {
			panic(fmt.Errorf("get out %d address", i))
		}
	}
	if err := VerifyTX(tx2, 0); err != nil {
		t.Errorf("Verify test failed  err=%v", err)
	}
}

func TestP2PKHSign(t *testing.T) {
	Txs.Push(NewTestFileCacher())
	defer Txs.Pop()
	id := NewHashID("78470577b25f58e0b18fd21e57eb64c10eb66272a856208440362103de0f31da")
	tx2, err := LoadTx(id)
	if err != nil {
		panic(err)
	}
	for i, v := range tx2.Outs {
		if i == 0 && v.Script.GetAddress() != "1MX1S4dniXHPJdySEszvM42nYryJx6NPgG" {
			panic(fmt.Errorf("get out %d address", i))
		}
		if i == 1 && v.Script.GetAddress() != "1k2saXX9kkxcSpx5W9yDMbuzRLsGfE71W" {
			panic(fmt.Errorf("get out %d address", i))
		}
	}
	if err := VerifyTX(tx2, 0); err != nil {
		t.Errorf("Verify test failed  err=%v", err)
	}
}
