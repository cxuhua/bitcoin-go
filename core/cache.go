package core

import (
	"container/list"
	"errors"
	"github.com/patrickmn/go-cache"
	"sync"
	"time"
)

type ICacher interface {
	Del(id HashID)
	Get(id HashID) (interface{}, error)
	Set(id HashID, v interface{}) (interface{}, error)
}

type memcacher struct {
	c *cache.Cache
}

func newMemCacher() ICacher {
	return &memcacher{
		c: cache.New(time.Minute*10, time.Minute*30),
	}
}

func (c *memcacher) Del(id HashID) {
	c.c.Delete(string(id[:]))
}

func (c *memcacher) Get(id HashID) (interface{}, error) {
	v, ok := c.c.Get(string(id[:]))
	if ok && v == nil {
		c.c.Delete(string(id[:]))
		return nil, CacheNotFoundErr
	}
	if !ok || v == nil {
		return nil, CacheNotFoundErr
	}
	return v, nil
}

func (c *memcacher) Set(id HashID, v interface{}) (interface{}, error) {
	c.c.Set(string(id[:]), v, cache.DefaultExpiration)
	return v, nil
}

type TxCacher interface {
	Del(id HashID)
	Get(id HashID) (*TX, error)
	Set(tx *TX) (*TX, error)
	Push(cv ICacher)
	Pop()
}

type BlockCacher interface {
	Del(id HashID)
	Get(id HashID) (*MsgBlock, error)
	Set(bl *MsgBlock) (*MsgBlock, error)
	Push(cv ICacher)
	Pop()
}

var (
	//tx cacher
	Txs = newTxs()
	//block cacher
	Bxs = newBxs()
)

var (
	CacheNotFoundErr = errors.New("cache not found")
	//block cache
	bxs = newBxsCache()
)

func newTxsCache() *cache.Cache {
	return cache.New(time.Minute*10, time.Minute*30)
}

func newBxsCache() *cache.Cache {
	return cache.New(time.Minute*5, time.Minute*20)
}

type txcacherdb struct {
	mu sync.Mutex
	xs ICacher
	lv *list.List
}

func newTxs() TxCacher {
	v := &txcacherdb{
		lv: list.New(),
	}
	v.xs = newMemCacher()
	v.lv.PushBack(v.xs)
	return v
}

func (db *txcacherdb) Push(cv ICacher) {
	db.mu.Lock()
	defer db.mu.Unlock()
	if cv == nil {
		db.xs = newMemCacher()
	} else {
		db.xs = cv
	}
	db.lv.PushBack(db.xs)
}

func (db *txcacherdb) Pop() {
	db.mu.Lock()
	defer db.mu.Unlock()
	//must keep one
	if db.lv.Len() <= 1 {
		return
	}
	db.lv.Remove(db.lv.Back())
	db.xs = db.lv.Back().Value.(ICacher)
}

func (db *txcacherdb) Set(tx *TX) (*TX, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	v, err := db.xs.Set(tx.Hash, tx)
	if err != nil {
		return nil, err
	}
	return v.(*TX), nil
}

func (db *txcacherdb) Del(id HashID) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.xs.Del(id)
}

func (db *txcacherdb) Get(id HashID) (*TX, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	v, err := db.xs.Get(id)
	if err != nil {
		return nil, err
	}
	return v.(*TX), nil
}

type blockcacherdb struct {
	mu sync.Mutex
	xs ICacher
	lv *list.List
}

func newBxs() BlockCacher {
	v := &blockcacherdb{
		lv: list.New(),
	}
	v.xs = newMemCacher()
	v.lv.PushBack(v.xs)
	return v
}

func (db *blockcacherdb) Push(cv ICacher) {
	db.mu.Lock()
	defer db.mu.Unlock()
	if cv == nil {
		db.xs = newMemCacher()
	} else {
		db.xs = cv
	}
	db.lv.PushBack(db.xs)
}

func (db *blockcacherdb) Pop() {
	db.mu.Lock()
	defer db.mu.Unlock()
	//must keep one
	if db.lv.Len() <= 1 {
		return
	}
	db.lv.Remove(db.lv.Back())
	db.xs = db.lv.Back().Value.(ICacher)
}

func (db *blockcacherdb) Set(b *MsgBlock) (*MsgBlock, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	v, err := db.xs.Set(b.Hash, b)
	if err != nil {
		return nil, err
	}
	return v.(*MsgBlock), nil
}

func (db *blockcacherdb) Del(id HashID) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.xs.Del(id)
}

func (db *blockcacherdb) Get(id HashID) (*MsgBlock, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	v, err := db.xs.Get(id)
	if err != nil {
		return nil, err
	}
	return v.(*MsgBlock), nil
}
