package store

import (
	"container/list"
	"context"
	"encoding/binary"
	"sync"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	client       *mongo.Client = nil
	dbonce                     = sync.Once{}
	NewestBK                   = []byte{0} //use GetBK method
	UseBKHeight                = []byte{1} //5bytes height=4 byte LittleEndian
	ListSyncBK                 = []byte{2}
	ListBlockTxs               = []byte{0} //33bytes [1:] = block id
)

type IterFunc func(cursor *mongo.Cursor) error

func IsListBlockTxs(id []byte) ([]byte, bool) {
	b := len(id) == 33 && id[0] == ListBlockTxs[0]
	return id[1:], b
}

func NewListBlockTxs(bid []byte) []byte {
	b := make([]byte, 33)
	b[0] = ListBlockTxs[0]
	copy(b[1:], bid)
	return b
}

func IsListSyncBK(id []byte) bool {
	return len(id) == 1 && id[0] == ListSyncBK[0]
}

func IsNewestBK(id []byte) bool {
	return len(id) == 1 && id[0] == NewestBK[0]
}

func IsBKHeight(id []byte) (uint32, bool) {
	if len(id) == 5 && id[0] == UseBKHeight[0] {
		v := binary.LittleEndian.Uint32(id[1:])
		return v, true
	}
	return 0, false
}

func BKHeight(h uint32) []byte {
	b := make([]byte, 5)
	b[0] = 1
	binary.LittleEndian.PutUint32(b[1:], h)
	return b
}

type mongoDBImp struct {
	context.Context
	clist *list.List
}

//get dbcacher
func (m *mongoDBImp) TopTxCacher() DbCacher {
	if m.clist.Len() == 0 {
		return &cacherInvoker{DbCacher: nil}
	}
	cacher := m.clist.Front().Value.(DbCacher)
	return &cacherInvoker{DbCacher: cacher}
}

func (m *mongoDBImp) PopTxCacher() {
	if m.clist.Len() == 0 {
		return
	}
	m.clist.Remove(m.clist.Front())
}

//set txcacher
func (m *mongoDBImp) PushTxCacher(c DbCacher) {
	m.clist.PushFront(c)
}

func (m *mongoDBImp) accounts() *mongo.Collection {
	return m.database().Collection(ACCOUNT_TABLE)
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

func (m *mongoDBImp) Transaction(fn func(db DbImp) error) error {
	ctx := m.Context.(mongo.SessionContext)
	_, err := ctx.WithTransaction(m, func(sess mongo.SessionContext) (interface{}, error) {
		err := fn(NewDBImp(sess))
		return nil, err
	})
	return err
}

func NewDBImp(ctx context.Context) DbImp {
	return &mongoDBImp{
		Context: ctx,
		clist:   list.New(),
	}
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
