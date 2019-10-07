package db

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//use gridfs save raw block data
const (
	BLOCK_TABLE = "blocks"
)

//delete data
func (m *mongoDBImp) DelBK(id []byte) error {
	_, err := m.blocks().DeleteOne(m, bson.M{"_id": id})
	if err != nil {
		return err
	}
	return err
}

func (m *mongoDBImp) GetBK(id []byte, v interface{}) error {
	ret := m.blocks().FindOne(m, bson.M{"_id": id})
	return ret.Decode(v)
}

//check tx exists
func (m *mongoDBImp) HasBK(id []byte) bool {
	ret := m.blocks().FindOne(m, bson.M{"_id": id}, options.FindOne().SetProjection(bson.M{"_id": 1}))
	return ret.Err() == nil
}

//save tans data
func (m *mongoDBImp) SetBK(id []byte, v interface{}) error {
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
			_, err := m.blocks().UpdateOne(m, bson.M{"_id": id}, bson.M{"$set": ds})
			return err
		}
	default:
		opt := options.Update().SetUpsert(true)
		_, err := m.blocks().UpdateOne(m, bson.M{"_id": id}, bson.M{"$set": v}, opt)
		return err
	}
	return nil
}
