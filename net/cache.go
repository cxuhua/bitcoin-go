package net

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrNotFound   = errors.New("not found")
	ErrCacherFull = errors.New("cacher fill")
)

//save valid tx info
type TxCacher interface {
	Get(id HashID) (*TX, error)
	Set(id HashID, tx *TX) error
	Del(id HashID)
	Clean()
}

var (
	Txs = NewTxCacher(1024*2, time.Minute*30)
)

func NewTxCacher(max int, timeout time.Duration) TxCacher {
	return &memoryTxCacher{
		txs:     map[string]*memoryTxElement{},
		max:     max,
		timeout: timeout,
	}
}

type memoryTxElement struct {
	tx *TX
	tv time.Time
}

type memoryTxCacher struct {
	txs     map[string]*memoryTxElement
	max     int
	timeout time.Duration
	rw      sync.Mutex
}

func (c *memoryTxCacher) clean() {
	now := time.Now()
	ks := []string{}
	for k, v := range c.txs {
		if now.Sub(v.tv) > c.timeout {
			ks = append(ks, k)
		}
		if len(ks) > 128 {
			break
		}
	}
	for _, v := range ks {
		delete(c.txs, v)
	}
}

func (c *memoryTxCacher) Clean() {
	c.rw.Lock()
	defer c.rw.Unlock()
	c.clean()
}

func (c *memoryTxCacher) Del(id HashID) {
	c.rw.Lock()
	defer c.rw.Unlock()
	delete(c.txs, id.String())
}

func (c *memoryTxCacher) Get(id HashID) (*TX, error) {
	c.rw.Lock()
	defer c.rw.Unlock()
	ele, ok := c.txs[id.String()]
	if !ok {
		return nil, ErrNotFound
	}
	if time.Now().Sub(ele.tv) >= c.timeout {
		delete(c.txs, id.String())
		return nil, ErrNotFound
	}
	ele.tv = time.Now()
	return ele.tx, nil
}

func (c *memoryTxCacher) Set(id HashID, tx *TX) error {
	c.rw.Lock()
	defer c.rw.Unlock()
	if len(c.txs) >= c.max {
		c.clean()
	}
	if len(c.txs) >= c.max {
		return ErrCacherFull
	}
	ele, ok := c.txs[id.String()]
	if ok {
		ele.tv = time.Now()
	} else {
		c.txs[id.String()] = &memoryTxElement{tv: time.Now(), tx: tx}
	}
	return nil
}
