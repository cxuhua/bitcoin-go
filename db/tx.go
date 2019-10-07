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

//get tx data
func (m *mongoDBImp) GetTX(id []byte, v interface{}) error {
	ret := m.txs().FindOne(m, bson.M{"_id": id})
	if err := ret.Err(); err != nil {
		return err
	}
	return ret.Decode(v)
}

//check tx exists
func (m *mongoDBImp) HasTX(id []byte) bool {
	ret := m.txs().FindOne(m, bson.M{"_id": id}, options.FindOne().SetProjection(bson.M{"_id": 1}))
	return ret.Err() == nil
}

//return blockid txs if id exists
func (m *mongoDBImp) MulTX(vs []interface{}, id ...[]byte) error {
	if len(id) == 1 {
		opts := options.Find().SetSort(bson.M{"index": 1})
		cur, err := m.txs().Find(m, bson.M{"block": id[0]}, opts)
		if err != nil {
			return err
		}
		for i := 0; i < len(vs) && cur.Next(m); i++ {
			err = cur.Decode(vs[i])
			if err != nil {
				return err
			}
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
	case IncValue:
		ds := bson.M{}
		for k, v := range v.(IncValue) {
			ds[k] = v
		}
		if len(ds) > 0 {
			_, err := m.blocks().UpdateOne(m, bson.M{"_id": id}, bson.M{"$inc": ds})
			return err
		}
	case SetValue:
		ds := bson.M{}
		for k, v := range v.(SetValue) {
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
