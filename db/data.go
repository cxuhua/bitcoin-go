package db

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

const (
	DATABASE = "bitcoin"
)

type DbImp interface {
	context.Context
	//get trans raw data
	GetTX(id []byte, v interface{}) error
	//save tans data
	SetTX(id []byte, v interface{}) error
	//delete tx
	DelTX(id []byte) error
}

type mongoDBImp struct {
	context.Context
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
