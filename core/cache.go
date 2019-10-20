package core

import (
	"errors"
	"github.com/patrickmn/go-cache"
	"sync"
	"time"
)

type TxCacher interface {
	Del(id HashID)
	Get(id HashID) (*TX, error)
	Set(id HashID, tx *TX) (*TX, error)
	Only() bool //only read from cacher return true
	Push(n TxCacher)
	SetTop(n TxCacher)
	Pop()
}

type BlockCacher interface {
	Del(id HashID)
	Get(id HashID) (*MsgBlock, error)
	Set(id HashID, bl *MsgBlock) (*MsgBlock, error)
	Only() bool //only read from cacher return true
	Push(n BlockCacher)
	SetTop(n BlockCacher)
	Pop()
}

var (
	//tx cacher
	Txs TxCacher = &txcacherdb{
		txstackcacher: &txstackcacher{},
	}
	//block cacher
	Bxs BlockCacher = &blockcacherdb{
		blockstackcacher: &blockstackcacher{},
	}
)

var (
	CacheNotFoundErr = errors.New("cache not found")
	//tx cache
	txs = cache.New(time.Minute*10, time.Minute*30)
	//block cache
	bxs = cache.New(time.Minute*5, time.Minute*20)
)

type txstackcacher struct {
	mu   sync.Mutex
	curr TxCacher
}

func (db *txstackcacher) Push(n TxCacher) {
	db.mu.Lock()
	defer db.mu.Unlock()
	if db.curr == nil {
		db.curr = Txs
	}
	n.SetTop(db.curr)
	Txs = n
}

func (db *txstackcacher) SetTop(curr TxCacher) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.curr = curr
}

func (db *txstackcacher) Pop() {
	db.mu.Lock()
	defer db.mu.Unlock()
	Txs = db.curr
}

type txcacherdb struct {
	*txstackcacher
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

type blockstackcacher struct {
	mu   sync.Mutex
	curr BlockCacher
}

func (db *blockstackcacher) Push(n BlockCacher) {
	db.mu.Lock()
	defer db.mu.Unlock()
	if db.curr == nil {
		db.curr = Bxs
	}
	n.SetTop(db.curr)
	Bxs = n
}

func (db *blockstackcacher) SetTop(curr BlockCacher) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.curr = curr
}

func (db *blockstackcacher) Pop() {
	db.mu.Lock()
	defer db.mu.Unlock()
	Bxs = db.curr
}

type blockcacherdb struct {
	*blockstackcacher
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
