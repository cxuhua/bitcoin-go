package core

import (
	"bitcoin/config"
	"bitcoin/store"
	"errors"
)

const (
	COIN      = Amount(100000000)
	MAX_MONEY = Amount(21000000 * COIN)
)

func GetCoinbaseReward(h int) Amount {
	conf := config.GetConfig()
	halvings := h / conf.SubHalving
	if halvings >= 64 {
		return 0
	}
	n := 50 * COIN
	n >>= halvings
	return n
}

type Amount int64

func (a Amount) IsRange() bool {
	return a >= 0 && a < MAX_MONEY
}

//record address value
type Moneys struct {
	Id    []byte `bson:"_id"`   //tx id + out index 32+4
	Addr  string `bson:"addr"`  //bitcoin address
	Value Amount `bson:"value"` //outvalue
	Lose  Amount `bson:"lose"`  //miss amount
}

//when,txid same,bitcoin miss
func (m *Moneys) LoseMoney() *Moneys {
	m.Id[0] = store.MT_IO_LOSE
	m.Lose = m.Value
	m.Value = 0
	return m
}

func (m Moneys) TxId() HashID {
	if len(m.Id) != 37 {
		panic(errors.New("id error"))
	}
	id := HashID{}
	copy(id[1:], m.Id[:33])
	return id
}

func (m Moneys) GetTx(db store.DbImp) (*TX, error) {
	return LoadTX(m.TxId(), db)
}

func (m Moneys) OutIdx() uint32 {
	if len(m.Id) != 37 {
		panic(errors.New("id error"))
	}
	return ByteOrder.Uint32(m.Id[33:])
}

func NewMoneys() *Moneys {
	l := len(HashID{})
	m := &Moneys{}
	m.Id = make([]byte, l+4)
	return m
}

func NewMoneyId(txid HashID, idx uint32, io byte) []byte {
	l := len(HashID{})
	id := make([]byte, 37)
	id[0] = io
	copy(id[1:l], txid[:])
	ByteOrder.PutUint32(id[l+1:], idx)
	return id
}
