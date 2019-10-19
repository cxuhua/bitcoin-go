package core

import (
	"bitcoin/store"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

var (
	//tx memory cacher
	Txs = NewMemoryCacher(1024*8, time.Minute*30)
	//block memory cacher
	Bxs = NewMemoryCacher(1024*2, time.Minute*60)
	//test download tx data cacher
	Dxs = &networkCache{}
	//test file tx data cache
	Fxs = &filekCache{}
)

//from ../dat get tx
type filekCache struct {
}

func (c *filekCache) Clean() {
	//log.Println("test use")
}

func (c *filekCache) Del(id []byte) {
	//log.Println("test use")
}

func (c *filekCache) Get(id []byte) (interface{}, error) {
	hid := HashID{}
	copy(hid[:], id)
	path := fmt.Sprintf("../dat/tx%s.dat", hid)
	dat, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	h := NewNetHeader(dat)
	tx := &TX{}
	tx.Read(h)
	return tx, nil
}

func (c *filekCache) Set(id []byte, v interface{}) error {
	//log.Println("test use")
	return nil
}

//from btc.com get tx data
type networkCache struct {
}

func (c *networkCache) Clean() {
	//log.Println("test use")
}

func (c *networkCache) Del(id []byte) {
	//log.Println("test use")
}

func (c *networkCache) Get(id []byte) (interface{}, error) {
	hid := HashID{}
	copy(hid[:], id)
	url := fmt.Sprintf("https://btc.com/%s.rawhex", hid)
	log.Printf("Download TX %s raw data,from %s\n", hid, url)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Download %s error %v\n", hid, err)
		return nil, err
	}
	decoder := hex.NewDecoder(resp.Body)
	dat, err := ioutil.ReadAll(decoder)
	if err != nil {
		log.Printf("Download %s error %v\n", hid, err)
		return nil, err
	}
	log.Printf("Download %s OK\n", hid)
	h := NewNetHeader(dat)
	tx := &TX{}
	tx.Read(h)
	return tx, nil
}

func (c *networkCache) Set(id []byte, v interface{}) error {
	//log.Println("test use")
	return nil
}

func NewCacher() store.DbCacher {
	return &memoryCacher{
		txs:     map[string]*memoryElement{},
		max:     int(^uint16(0)),
		timeout: time.Hour * 24,
	}
}

func NewMemoryCacher(max int, timeout time.Duration) store.DbCacher {
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

func (c *memoryCacher) Del(id []byte) {
	c.rw.Lock()
	defer c.rw.Unlock()
	key := hex.EncodeToString(id)
	delete(c.txs, key)
}

func (c *memoryCacher) Get(id []byte) (interface{}, error) {
	c.rw.Lock()
	defer c.rw.Unlock()
	key := hex.EncodeToString(id)
	ele, ok := c.txs[key]
	if !ok {
		return nil, store.ErrNotFound
	}
	if time.Now().Sub(ele.tv) >= c.timeout {
		delete(c.txs, key)
		return nil, store.ErrNotFound
	}
	ele.tv = time.Now()
	return ele.value, nil
}

func (c *memoryCacher) Set(id []byte, v interface{}) error {
	c.rw.Lock()
	defer c.rw.Unlock()
	key := hex.EncodeToString(id)
	if len(c.txs) >= c.max {
		c.clean()
	}
	if len(c.txs) >= c.max {
		return store.ErrCacherFull
	}
	ele, ok := c.txs[key]
	if ok {
		ele.tv = time.Now()
	} else {
		c.txs[key] = &memoryElement{tv: time.Now(), value: v}
	}
	return nil
}
