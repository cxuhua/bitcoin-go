package core

import (
	"errors"
	"github.com/patrickmn/go-cache"
	"time"
)

var (
	CacheNotFoundErr = errors.New("cache not found")
	//tx cache
	txs = cache.New(time.Minute*10, time.Minute*30)
	//block cache
	bxs = cache.New(time.Minute*5, time.Minute*20)
)

func TxCacheSet(id HashID, tx *TX) (*TX, error) {
	txs.Set(string(id[:]), tx, cache.DefaultExpiration)
	return tx, nil
}

func TxCacheDel(id HashID) {
	txs.Delete(string(id[:]))
}

func TxCacheGet(id HashID) (*TX, error) {
	v, ok := txs.Get(string(id[:]))
	if !ok {
		return nil, CacheNotFoundErr
	}
	return v.(*TX), nil
}

func BlockCacheSet(id HashID, b *MsgBlock) (*MsgBlock, error) {
	bxs.Set(string(id[:]), b, cache.DefaultExpiration)
	return b, nil
}

func BlockCacheDel(id HashID) {
	bxs.Delete(string(id[:]))
}

func BlockCacheGet(id HashID) (*MsgBlock, error) {
	v, ok := bxs.Get(string(id[:]))
	if !ok {
		return nil, CacheNotFoundErr
	}
	return v.(*MsgBlock), nil
}
