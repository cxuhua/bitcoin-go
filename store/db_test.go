package store

import (
	"context"
	"errors"
	"log"
	"testing"

	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"

	"github.com/syndtr/goleveldb/leveldb"
)

func TestLevelDB(t *testing.T) {
	bb := filter.NewBloomFilter(10)
	opts := &opt.Options{
		Filter: bb,
	}
	db, err := leveldb.OpenFile("./leveldb", opts)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	//wg := sync.WaitGroup{}
	//wg.Add(3)
	//go func() {
	//	for i := 0; i < 100000; i++ {
	//		key := []byte{0, 0, 0, 0}
	//		binary.BigEndian.PutUint32(key, uint32(i))
	//		if err := db.Put(key, key, nil); err != nil {
	//			panic(err)
	//		}
	//	}
	//	wg.Done()
	//}()
	//go func() {
	//	for i := 0; i < 100000; i++ {
	//		key := []byte{0, 0, 0, 0}
	//		binary.BigEndian.PutUint32(key, uint32(i))
	//		if err := db.Put(key, key, nil); err != nil {
	//			panic(err)
	//		}
	//	}
	//	wg.Done()
	//}()
	//go func() {
	//	for i := 0; i < 100000; i++ {
	//		key := []byte{0, 0, 0, 0}
	//		binary.BigEndian.PutUint32(key, uint32(i))
	//		if err := db.Put(key, key, nil); err != nil {
	//			panic(err)
	//		}
	//	}
	//	wg.Done()
	//}()
	//wg.Wait()
	s := &leveldb.DBStats{}
	err = db.Stats(s)
	log.Println(s, err)

	//
	//start := []byte{0, 0, 0, 0}
	//binary.BigEndian.PutUint32(start, uint32(500))
	//
	//limit := []byte{0, 0, 0, 0}
	//binary.BigEndian.PutUint32(limit, uint32(600))
	//
	//r := util.BytesPrefix([]byte{0, 0, 1})
	//iter := db.NewIterator(r, nil)
	//for iter.Next() {
	//	log.Println(binary.BigEndian.Uint32(iter.Key()), iter.Value())
	//}

	//mdb := memdb.New(comparer.DefaultComparer, 1024)
	//defer mdb.Free()
	//mdb.Put([]byte{1}, []byte{2})
	//iter := mdb.NewIterator(nil)
	//for iter.Next() {
	//	log.Println(iter.Value())
	//}
}

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
		if !db.HasTX(data.Id) {
			return errors.New("test hastx error")
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}
