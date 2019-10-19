package store

import (
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	MONEYS_TABLE = "moneys"
	MT_IO_IN     = byte('I')
	MT_IO_OUT    = byte('O')
	MT_IO_LOSE   = byte('L')
)

//total all money value
func (m *mongoDBImp) SumMT(lose bool) int64 {
	ret := int64(0)
	vv := bson.M{}
	cur, err := m.moneys().Find(m, bson.M{})
	if err != nil {
		return -1
	}
	defer cur.Close(m)
	for cur.Next(m) {
		err := cur.Decode(&vv)
		if err != nil {
			return -1
		}
		ret += vv["value"].(int64)
		if lose {
			ret += vv["lose"].(int64)
		}
	}
	return ret
}

//check tx exists
func (m *mongoDBImp) HasMT(id []byte) bool {
	ret := m.moneys().FindOne(m, bson.M{"_id": id}, options.FindOne().SetProjection(bson.M{"_id": 1}))
	return ret.Err() == nil
}

//set money record
func (m *mongoDBImp) SetMT(id []byte, v interface{}) error {
	_, err := m.moneys().InsertOne(m, v)
	return err
}

//get money record
//id == string(address)
func (m *mongoDBImp) GetMT(id []byte, v interface{}) error {
	fn, ok := v.(IterFunc)
	if !ok {
		return errors.New("v args type error,IterFunc")
	}
	sid := string(id)
	iter, err := m.moneys().Find(m, bson.M{"addr": sid})
	if err != nil {
		return err
	}
	defer iter.Close(m)
	for iter.Next(m) {
		if b, err := fn(iter); err != nil {
			return err
		} else if !b {
			break
		}
	}
	return nil
}
