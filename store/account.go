package store

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	ACCOUNT_TABLE = "accounts"
)

//get account info
func (m *mongoDBImp) GetAT(id []byte, v interface{}) error {
	cond := bson.M{"_id": id}
	ret := m.accounts().FindOne(m, cond)
	if err := ret.Err(); err != nil {
		return err
	}
	return ret.Decode(v)
}

//check account exists
func (m *mongoDBImp) HasAT(id []byte) bool {
	ret := m.accounts().FindOne(m, bson.M{"_id": id}, options.FindOne().SetProjection(bson.M{"_id": 1}))
	return ret.Err() == nil
}

//save account data
func (m *mongoDBImp) SetAT(id []byte, v interface{}) error {
	switch v.(type) {
	case IncValue:
		ds := bson.M{}
		for k, v := range v.(IncValue) {
			ds[k] = v
		}
		if len(ds) > 0 {
			_, err := m.accounts().UpdateOne(m, bson.M{"_id": id}, bson.M{"$inc": ds})
			return err
		}
	case SetValue:
		ds := bson.M{}
		for k, v := range v.(SetValue) {
			ds[k] = v
		}
		if len(ds) > 0 {
			_, err := m.accounts().UpdateOne(m, bson.M{"_id": id}, bson.M{"$set": ds})
			return err
		}
	default:
		opt := options.Update().SetUpsert(true)
		_, err := m.accounts().UpdateOne(m, bson.M{"_id": id}, bson.M{"$set": v}, opt)
		return err
	}
	return nil
}
