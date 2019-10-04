package db

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

const (
	DATABASE = "bitcoin"
)

type KeyValue map[string]interface{}

type DbImp interface {
	context.Context
	//get trans raw data
	GetTX(id []byte, v interface{}) error
	//save or update tans data
	SetTX(id []byte, v interface{}) error
	//delete tx
	DelTX(id []byte) error
	//multiple opt id != nil will read blockid =id txs and order by index
	MulTX(v []interface{}, id ...[]byte) error
	//get block raw data
	GetBK(id []byte, v interface{}) error
	//save or update block data
	SetBK(id []byte, v interface{}) error
	//del block data
	DelBK(id []byte) error
}

type mongoDBImp struct {
	context.Context
}

func (m *mongoDBImp) blocks() *mongo.Collection {
	return m.database().Collection(BLOCK_TABLE)
}

func (m *mongoDBImp) txs() *mongo.Collection {
	return m.database().Collection(TX_TABLE)
}

func (m *mongoDBImp) database() *mongo.Database {
	return m.client().Database(DATABASE)
}

func (m *mongoDBImp) client() *mongo.Client {
	return m.Context.(mongo.SessionContext).Client()
}

func NewDBImp(ctx context.Context) DbImp {
	return &mongoDBImp{Context: ctx}
}
