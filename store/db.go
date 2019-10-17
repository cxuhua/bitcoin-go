package store

import (
	"context"
	"encoding/binary"
	"sync"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	client      *mongo.Client = nil
	dbonce                    = sync.Once{}
	NewestBK                  = []byte{0} //use GetBK method
	UseBKHeight               = []byte{1} //1-8bytes height=9 byte LittleEndian
)

func IsNewestBK(id []byte) bool {
	return len(id) == 1 && id[0] == NewestBK[0]
}

func IsBKHeight(id []byte) (uint64, bool) {
	if len(id) == 9 && id[0] == UseBKHeight[0] {
		v := binary.LittleEndian.Uint64(id[1:])
		return v, true
	}
	return 0, false
}

func BKHeight(h uint64) []byte {
	b := make([]byte, 9)
	b[0] = 1
	binary.LittleEndian.PutUint64(b[1:], h)
	return b
}

type mongoDBImp struct {
	context.Context
	cache DbCacher
}

//get dbcacher
func (m *mongoDBImp) TXCacher() DbCacher {
	if m.cache != nil {
		return m.cache
	} else {
		return &cacherInvoker{DbCacher: m.cache}
	}
}

//set txcacher
func (m *mongoDBImp) SetTXCacher(c DbCacher) {
	m.cache = c
}

func (m *mongoDBImp) blocks() *mongo.Collection {
	return m.database().Collection(BLOCK_TABLE)
}

func (m *mongoDBImp) txs() *mongo.Collection {
	return m.database().Collection(TX_TABLE)
}

func (m *mongoDBImp) database() *mongo.Database {
	return m.client().Database(DATABASE)
}

func (m *mongoDBImp) client() *mongo.Client {
	return m.Context.(mongo.SessionContext).Client()
}

func NewDBImp(ctx context.Context) DbImp {
	return &mongoDBImp{Context: ctx}
}

func InitDB(ctx context.Context) *mongo.Client {
	dbonce.Do(func() {
		c := options.Client().ApplyURI("mongodb://127.0.0.1:27017/")
		cptr, err := mongo.NewClient(c)
		if err != nil {
			panic(err)
		}
		err = cptr.Connect(ctx)
		if err != nil {
			panic(err)
		}
		client = cptr
	})
	return client
}

func UseSession(ctx context.Context, fn func(db DbImp) error) error {
	client = InitDB(ctx)
	return client.UseSession(ctx, func(sess mongo.SessionContext) error {
		return fn(NewDBImp(sess))
	})
}
