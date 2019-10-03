package db

import (
	"context"
	"log"
	"testing"
)

func TestGetSetTX(t *testing.T) {
	type dt struct {
		Id []byte  `bson:"_id"`
		A  int     `bson:"a"`
		B  string  `bson:"b"`
		C  int64   `bson:"c"`
		D  float64 `bson:"d"`
		X  []byte  `bson:"x"`
	}
	data := dt{
		Id: make([]byte, 32),
		A:  11,
		B:  "22",
		C:  33,
		D:  0.44,
		X:  make([]byte, 1024),
	}
	err := UseSession(context.Background(), func(db DbImp) error {
		err := db.SetTX(data.Id, data)
		if err != nil {
			return err
		}
		v := dt{}
		err = db.GetTX(data.Id, &v)
		if err != nil {
			return err
		}
		err = db.DelTX(data.Id)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}
