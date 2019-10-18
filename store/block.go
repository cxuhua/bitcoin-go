package store

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
	_, err = m.txs().DeleteMany(m, bson.M{"block": id})
	return err
}

//last block
func (m *mongoDBImp) lastBK(v interface{}) error {
	opts := options.FindOne()
	opts.SetSort(bson.M{"height": -1})
	ret := m.blocks().FindOne(m, bson.M{}, opts)
	if err := ret.Err(); err != nil {
		return err
	}
	return ret.Decode(v)
}

func (m *mongoDBImp) GetBK(id []byte, v interface{}) error {
	cond := bson.M{}
	if IsNewestBK(id) {
		return m.lastBK(v)
	} else if hv, ok := IsBKHeight(id); ok {
		cond["height"] = hv
	} else {
		cond["_id"] = id
	}
	ret := m.blocks().FindOne(m, cond)
	if err := ret.Err(); err != nil {
		return err
	}
	return ret.Decode(v)
}

func (m *mongoDBImp) ValidBK(id []byte) bool {
	opts := options.FindOne().SetProjection(bson.M{"_id": 1})
	cond := bson.M{"_id": id, "count": bson.M{"$gt": 0}}
	ret := m.blocks().FindOne(m, cond, opts)
	return ret.Err() == nil
}

//check tx exists
func (m *mongoDBImp) HasBK(id []byte) bool {
	opts := options.FindOne().SetProjection(bson.M{"_id": 1})
	cond := bson.M{"_id": id}
	ret := m.blocks().FindOne(m, cond, opts)
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
