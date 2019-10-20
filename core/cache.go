package core

import (
	"errors"
	"github.com/patrickmn/go-cache"
	"time"
)

type TxCacher interface {
	Del(id HashID)
	Get(id HashID) (*TX, error)
	Set(id HashID, tx *TX) (*TX, error)
}

type BlockCacher interface {
	Del(id HashID)
	Get(id HashID) (*MsgBlock, error)
	Set(id HashID, bl *MsgBlock) (*MsgBlock, error)
}

var (
	//tx cacher
	Txs TxCacher = &txcacherdb{}
	//block cacher
	Bxs BlockCacher = &blockcacherdb{}
)

var (
	CacheNotFoundErr = errors.New("cache not found")
	//tx cache
	txs = cache.New(time.Minute*10, time.Minute*30)
	//block cache
	bxs = cache.New(time.Minute*5, time.Minute*20)
)

type txcacherdb struct {
}

func (db *txcacherdb) Only() bool {
	return false
}

func (db *txcacherdb) Set(id HashID, tx *TX) (*TX, error) {
	txs.Set(string(id[:]), tx, cache.DefaultExpiration)
	return tx, nil
}

func (db *txcacherdb) Del(id HashID) {
	txs.Delete(string(id[:]))
}

func (db *txcacherdb) Get(id HashID) (*TX, error) {
	v, ok := txs.Get(string(id[:]))
	if !ok {
		return nil, CacheNotFoundErr
	}
	return v.(*TX), nil
}

type blockcacherdb struct {
}

func (db *blockcacherdb) Only() bool {
	return false
}

func (db *blockcacherdb) Set(id HashID, b *MsgBlock) (*MsgBlock, error) {
	bxs.Set(string(id[:]), b, cache.DefaultExpiration)
	return b, nil
}

func (db *blockcacherdb) Del(id HashID) {
	bxs.Delete(string(id[:]))
}

func (db *blockcacherdb) Get(id HashID) (*MsgBlock, error) {
	v, ok := bxs.Get(string(id[:]))
	if !ok {
		return nil, CacheNotFoundErr
	}
	return v.(*MsgBlock), nil
}
