package core

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"sync"

	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"

	"github.com/syndtr/goleveldb/leveldb"
)

var (
	dbptr *leveldb.DB = nil
	once  sync.Once
)

func DB() *leveldb.DB {
	once.Do(func() {
		bf := filter.NewBloomFilter(5)
		opts := &opt.Options{
			Filter: bf,
		}
		sdb, err := leveldb.OpenFile("database", opts)
		if err != nil {
			panic(err)
		}
		dbptr = sdb
	})
	return dbptr
}

type Chain struct {
	mu   sync.Mutex
	Hash []HashID
	best string
}

func GetChain(b string, h int) *Chain {
	return NewChain(b, h)
}

func NewChain(b string, h int) *Chain {
	c := &Chain{}
	c.Hash = make([]HashID, h+1)
	c.best = b
	return c
}

func (c *Chain) WriteDB(path string, lh int) error {
	if lh == 0 {
		lh = -1
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	bhash := NewHashID(c.best)
	for idx := len(c.Hash) - 1; idx >= lh+1; idx-- {
		file := path + "\\" + bhash.String()
		data, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}
		h := NewNetHeader(data)
		m := &MsgBlock{}
		m.Read(h)
		m.Height = uint32(idx)
		if !bhash.Equal(m.Hash) {
			return fmt.Errorf("connect chain error %v -> %v", m.Hash, bhash)
		}
		c.Hash[idx] = m.Hash
		bhash = m.Prev
		if idx%100 == 0 {
			log.Println("load block ", idx, m.Hash, m.Height, "OK")
		}
		Bxs.Set(m)
	}
	for i := lh + 1; i < len(c.Hash); i++ {
		m, err := Bxs.Get(c.Hash[i])
		if err != nil {
			return fmt.Errorf("%v miss", c.Hash[i])
		}
		if !G.IsNextBlock(m) {
			return errors.New("connect to next block error")
		}
		if err := m.Check(); err != nil {
			return err
		}
		if err := m.Connect(true); err != nil {
			return err
		}
		G.SetBestBlock(m)
		log.Println(c.Hash[i], i, "WRITE OK")
	}
	return nil
}

const (
	//blocks key prefix
	TPrefixBlock = byte(1)

	//txid -> block key prefox
	TPrefixTxId = byte(2)

	//address key prefix
	TPrefixAddress = byte(3)

	//height index(4) -> blockid
	TPrefixHeight = byte(4)

	//Best block hash key -> blockid
	TBestBlockHashKey = "TBestBlockHashKey"
)

func LoadHeightBlock(h uint32) (*MsgBlock, error) {
	hkey := NewTHeightKey(h)
	hv, err := DB().Get(hkey[:], nil)
	if err != nil {
		return nil, fmt.Errorf("load height block %w h=%d", err, h)
	}
	return LoadBlock(NewHashID(hv))
}

type THeightKey [5]byte

func NewTHeightKey(h uint32) THeightKey {
	k := THeightKey{}
	k[0] = TPrefixHeight
	ByteOrder.PutUint32(k[1:], h)
	return k
}

type TAddrKey []byte

func (a TAddrKey) GetIndex() uint32 {
	off := a[1] + 2 + 32
	return ByteOrder.Uint32(a[off : off+4])
}

func (a TAddrKey) GetTx() HashID {
	off := a[1] + 2
	bb := a[off : off+32]
	return NewHashID([]byte(bb))
}

func (a TAddrKey) GetAddr() string {
	al := a[1]
	return string(a[2 : al+2])
}

//prefix[1] addrlen[1] addr(mstring) TxHashId[32]-index[4] -> int64(amount)
func NewTAddrKey(addr string, txid HashID, idx uint32) TAddrKey {
	buf := &bytes.Buffer{}
	buf.WriteByte(TPrefixAddress)
	buf.WriteByte(byte(len(addr)))
	buf.Write([]byte(addr))
	buf.Write(txid[:])
	binary.Write(buf, ByteOrder, idx)
	return buf.Bytes()
}

func (k TAddrKey) String() string {
	return "A= " + k.GetAddr() + " TX= " + k.GetTx().String() + " IDX= " + fmt.Sprintf("%d", k.GetIndex())
}

type TAddrValue [8]byte

type TAddrElement struct {
	TAddrKey
	TAddrValue
}

func ListAddrValues(addr string) []TAddrElement {
	eles := []TAddrElement{}
	prefix := util.BytesPrefix(append([]byte{TPrefixAddress, byte(len(addr))}, []byte(addr)...))
	iter := DB().NewIterator(prefix, nil)
	for iter.Next() {
		v := TAddrValue{}
		copy(v[:], iter.Value())
		ele := TAddrElement{TAddrKey: iter.Key(), TAddrValue: v}
		eles = append(eles, ele)
	}
	return eles
}

func (a TAddrValue) GetValue() Amount {
	return Amount(ByteOrder.Uint64(a[:]))
}

func NewTAddrValue(v uint64) TAddrValue {
	v8 := TAddrValue{}
	ByteOrder.PutUint64(v8[:], v)
	return v8
}

func (a TAddrValue) String() string {
	return fmt.Sprintf("%d", a.GetValue())
}

//prefix[1] hash[32]
type TBlockKey [33]byte

//prefix[1] + txid[32]
type TTxKey [33]byte

//BlockHash[32] + txindex[4]
type TTxValue []byte

func NewTTxValue(hv HashID, idx uint32) TTxValue {
	buf := &bytes.Buffer{}
	buf.Write(hv[:])
	binary.Write(buf, ByteOrder, idx)
	return buf.Bytes()
}

func HasTx(hv HashID) bool {
	tkey := NewTxKey(hv)
	ok, err := DB().Has(tkey[:], nil)
	return err == nil && ok
}

//value: height(4)+block data
type TBlock []byte

func NewTBlockKey(id HashID) TBlockKey {
	k := TBlockKey{}
	k[0] = TPrefixBlock
	copy(k[1:], id[:])
	return k
}

func (k TBlockKey) String() string {
	return NewHashID(k[1:]).String()
}

func (txk TTxKey) String() string {
	return NewHashID(txk[1:]).String()
}

func NewTxKey(tx HashID) TTxKey {
	k := TTxKey{}
	k[0] = TPrefixTxId
	copy(k[1:], tx[:])
	return k
}

func (b TBlock) Height() uint32 {
	return ByteOrder.Uint32(b[:4])
}

func (b TBlock) Body() []byte {
	return b[4:]
}

func (b TBlock) ToBlock() *MsgBlock {
	h := NewNetHeader(b.Body())
	m := &MsgBlock{}
	m.Read(h)
	return m
}

func LoadBestBlock() (*MsgBlock, error) {
	val, err := DB().Get([]byte(TBestBlockHashKey), nil)
	if err != nil {
		return nil, err
	}
	return LoadBlock(NewHashID(val))
}

func LoadBlock(id HashID) (*MsgBlock, error) {
	if bv, err := Bxs.Get(id); err == nil {
		return bv, nil
	}
	key := NewTBlockKey(id)
	bb, err := DB().Get(key[:], nil)
	if err != nil {
		return nil, err
	}
	bv := TBlock(bb)
	m := bv.ToBlock()
	m.Height = bv.Height()
	return Bxs.Set(m)
}

func (v TTxValue) TxIndex() uint32 {
	return ByteOrder.Uint32(v[32:])
}

func (v TTxValue) BlockHash() HashID {
	id := HashID{}
	copy(id[:], v[:32])
	return id
}

func (v TTxValue) GetTx() (*TX, error) {
	m, err := LoadBlock(v.BlockHash())
	if err != nil {
		return nil, err
	}
	return m.Txs[v.TxIndex()], nil
}

func LoadTx(tx HashID) (*TX, error) {
	//cache
	if tv, err := Txs.Get(tx); err == nil {
		return tv, nil
	}
	v, err := LoadTxValue(tx)
	if err != nil {
		return nil, err
	}
	tv, err := v.GetTx()
	if err != nil {
		return nil, err
	}
	return Txs.Set(tv)
}

func LoadTxValue(tx HashID) (TTxValue, error) {
	txkey := NewTxKey(tx)
	return DB().Get(txkey[:], nil)
}

func NewTBlock(m *MsgBlock) TBlock {
	h := NewNetHeader()
	m.Write(h)
	b := make(TBlock, 4+h.Len())
	ByteOrder.PutUint32(b[:4], m.Height)
	copy(b[4:], h.Bytes())
	return b
}
