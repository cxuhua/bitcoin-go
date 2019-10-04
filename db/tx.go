package db

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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

func (m *mongoDBImp) MulTX(vs []interface{}, id ...[]byte) error {
	if len(id) == 1 {
		opts := options.Find().SetSort(bson.M{"index": 1})
		cur, err := m.txs().Find(m, bson.M{"block": id[0]}, opts)
		if err != nil {
			return err
		}
		i := 0
		for i < len(vs) && cur.Next(m) {
			cur.Decode(vs[i])
			i++
		}
		return nil
	} else {
		opts := options.BulkWrite()
		if len(vs) == 0 {
			return nil
		}
		mvs := []mongo.WriteModel{}
		for _, v := range vs {
			vv := mongo.NewInsertOneModel().SetDocument(v)
			mvs = append(mvs, vv)
		}
		_, err := m.txs().BulkWrite(m, mvs, opts)
		return err
	}
}

//save tans data
func (m *mongoDBImp) SetTX(id []byte, v interface{}) error {
	switch v.(type) {
	case KeyValue:
		ds := bson.M{}
		for k, v := range v.(KeyValue) {
			ds[k] = v
		}
		if len(ds) > 0 {
			_, err := m.txs().UpdateOne(m, bson.M{"_id": id}, bson.M{"$set": ds})
			return err
		}
	default:
		opt := options.Update().SetUpsert(true)
		_, err := m.txs().UpdateOne(m, bson.M{"_id": id}, bson.M{"$set": v}, opt)
		return err
	}
	return nil
}
