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
type MemoryCacher interface {
	Get(id HashID) (interface{}, error)
	Set(id HashID, v interface{}) error
	Del(id HashID)
	Clean()
}

var (
	//tx memory cacher
	Txs = NewMemoryCacher(1024*2, time.Minute*30)
	//block memory cacher
	Bxs = NewMemoryCacher(256, time.Minute*60)
)

func NewMemoryCacher(max int, timeout time.Duration) MemoryCacher {
	return &memoryCacher{
		txs:     map[string]*memoryElement{},
		max:     max,
		timeout: timeout,
	}
}

type memoryElement struct {
	value interface{}
	tv    time.Time
}

type memoryCacher struct {
	txs     map[string]*memoryElement
	max     int
	timeout time.Duration
	rw      sync.Mutex
}

func (c *memoryCacher) clean() {
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

func (c *memoryCacher) Clean() {
	c.rw.Lock()
	defer c.rw.Unlock()
	c.clean()
}

func (c *memoryCacher) Del(id HashID) {
	c.rw.Lock()
	defer c.rw.Unlock()
	delete(c.txs, id.String())
}

func (c *memoryCacher) Get(id HashID) (interface{}, error) {
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
	return ele.value, nil
}

func (c *memoryCacher) Set(id HashID, v interface{}) error {
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
		c.txs[id.String()] = &memoryElement{tv: time.Now(), value: v}
	}
	return nil
}
