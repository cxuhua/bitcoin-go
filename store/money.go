package store

import (
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	MONEYS_TABLE = "moneys"
)

func IsMoneyId(id []byte) bool {
	return len(id) == 37 && id[0] == 0
}

func NewOptMoneyId(id []byte) []byte {
	b := make([]byte, 37)
	b[0] = 0
	copy(b[1:], id)
	return b
}

func NewAddrId(addr string) []byte {
	b := make([]byte, len(addr)+1)
	b[0] = 1
	copy(b[1:], addr)
	return b
}

func IsAddrId(id []byte) bool {
	return len(id) > 0 && id[0] == 1
}

//total all money value
func (m *mongoDBImp) TotalMT() int64 {
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
	}
	return ret
}

//set money record
func (m *mongoDBImp) SetMT(id []byte, v interface{}) error {
	switch v.(type) {
	case IncValue:
		ds := bson.M{}
		for k, v := range v.(IncValue) {
			ds[k] = v
		}
		if len(ds) > 0 {
			_, err := m.moneys().UpdateOne(m, bson.M{"_id": id}, bson.M{"$inc": ds})
			return err
		}
	case SetValue:
		ds := bson.M{}
		for k, v := range v.(SetValue) {
			ds[k] = v
		}
		if len(ds) > 0 {
			_, err := m.moneys().UpdateOne(m, bson.M{"_id": id}, bson.M{"$set": ds})
			return err
		}
	default:
		opt := options.Update().SetUpsert(true)
		_, err := m.moneys().UpdateOne(m, bson.M{"_id": id}, bson.M{"$set": v}, opt)
		return err
	}
	return nil
}

//get money record
func (m *mongoDBImp) GetMT(id []byte, v interface{}) error {
	if IsMoneyId(id) {
		ret := m.moneys().FindOne(m, bson.M{"_id": id[1:]})
		if err := ret.Err(); err != nil {
			return err
		}
		return ret.Decode(v)
	}
	if !IsAddrId(id) {
		return errors.New("id error")
	}
	fn, ok := v.(IterFunc)
	if !ok {
		return errors.New("v args type error,IterFunc")
	}
	sid := string(id[1:])
	opts := options.Find()
	opts.SetSort(bson.M{"index": 1})
	iter, err := m.moneys().Find(m, bson.M{"addr": sid}, opts)
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

//delete money record
func (m *mongoDBImp) DelMT(id []byte) error {
	_, err := m.moneys().DeleteOne(m, bson.M{"_id": id})
	return err
}
