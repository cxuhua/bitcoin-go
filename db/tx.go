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
	_, err := m.txs().DeleteOne(m, bson.M{"_id": id})
	if err != nil {
		return err
	}
	return err
}

func (m *mongoDBImp) GetTX(id []byte, v interface{}) error {
	ret := m.txs().FindOne(m, bson.M{"_id": id})
	return ret.Decode(v)
}

//save tans data
func (m *mongoDBImp) SetTX(id []byte, v interface{}) error {
	var err error = nil
	switch v.(type) {
	case KeyValue:
		ds := bson.M{}
		for k, v := range v.(KeyValue) {
			ds[k] = v
		}
		if len(ds) > 0 {
			_, err = m.txs().UpdateOne(m, bson.M{"_id": id}, bson.M{"$set": ds})
		}
	default:
		opt := options.Update().SetUpsert(true)
		_, err = m.txs().UpdateOne(m, bson.M{"_id": id}, bson.M{"$set": v}, opt)
	}
	return err
}
