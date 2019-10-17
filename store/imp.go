package store

import (
	"context"
	"errors"
)

const (
	DATABASE = "bitcoin"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrCacherFull   = errors.New("cacher fill")
	ErrNotSetCacher = errors.New("not set cacher")
)

//save valid tx info
type DbCacher interface {
	Get(id []byte) (interface{}, error)
	Set(id []byte, v interface{}) error
	Del(id []byte)
	Clean()
}

type cacherInvoker struct {
	DbCacher
}

func (c *cacherInvoker) Get(id []byte) (interface{}, error) {
	if c.DbCacher == nil {
		return nil, ErrNotSetCacher
	}
	return c.DbCacher.Get(id)
}

func (c *cacherInvoker) Set(id []byte, v interface{}) error {
	if c.DbCacher == nil {
		return ErrNotSetCacher
	}
	return c.DbCacher.Set(id, v)
}

func (c *cacherInvoker) Del(id []byte) {
	if c.DbCacher == nil {
		return
	}
	c.DbCacher.Del(id)
}

func (c *cacherInvoker) Clean() {
	if c.DbCacher == nil {
		return
	}
	c.DbCacher.Clean()
}

type SetValue map[string]interface{}
type IncValue map[string]int

type DbImp interface {
	context.Context
	//get account data
	GetAT(id []byte, v interface{}) error
	//exists account
	HasAT(id []byte) bool
	//save or update account
	SetAT(id []byte, v interface{}) error
	//use txcacher
	TopTxCacher() DbCacher
	//push txcacher
	PushTxCacher(c DbCacher)
	//pop cacher
	PopTxCacher()
	//get trans raw data
	GetTX(id []byte, v interface{}) error
	//save or update tans data
	SetTX(id []byte, v interface{}) error
	//exists tx
	HasTX(id []byte) bool
	//delete tx
	DelTX(id []byte) error
	//multiple tx save
	MulTX(v []interface{}) error
	//get block raw data
	GetBK(id []byte, v interface{}) error
	//exists bk
	HasBK(id []byte) bool
	//valid bk: exists and check pass,count > 0
	ValidBK(id []byte) bool
	//save or update block data
	SetBK(id []byte, v interface{}) error
	//del block data
	DelBK(id []byte) error
	//transaction
	Transaction(fn func(db DbImp) error) error
}
