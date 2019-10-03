package db

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	TX_TABLE = "txs"
)

//delete data
func (m *mongoDBImp) DelTX(id []byte) error {
	_, err := m.database().Collection(TX_TABLE).DeleteOne(m, bson.M{"_id": id})
	if err != nil {
		return err
	}
	return err
}

func (m *mongoDBImp) GetTX(id []byte, v interface{}) error {
	ret := m.database().Collection(TX_TABLE).FindOne(m, bson.M{"_id": id})
	return ret.Decode(v)
}

//save tans data
func (m *mongoDBImp) SetTX(id []byte, v interface{}) error {
	opt := options.Update().SetUpsert(true)
	_, err := m.database().Collection(TX_TABLE).UpdateOne(m, bson.M{"_id": id}, bson.M{"$set": v}, opt)
	if err != nil {
		return err
	}
	return err
}
