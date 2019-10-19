package core

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"sync"

	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"

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

const (
	//blocks key prefix
	TPrefixBlock = byte(1)

	//txid -> block key prefox
	TPrefixTxId = byte(2)

	//address key prefix
	TPrefixAddress = byte(3)

	//Best block hash key -> blockid
	TBestBlockHashKey = "TBestBlockHashKey"
)

type TAddrKey []byte

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
	buf := bytes.NewReader(k)
	buf.ReadByte()
	al, _ := buf.ReadByte()
	ab := make([]byte, al)
	buf.Read(ab)
	txid := HashID{}
	buf.Read(txid[:])
	idx := uint32(0)
	binary.Read(buf, ByteOrder, &idx)
	return "A= " + string(ab) + " TX= " + txid.String() + " IDX= " + fmt.Sprintf("%d", idx)
}

type TAddrValue [8]byte

func NewTAddrValue(v uint64) TAddrValue {
	v8 := TAddrValue{}
	ByteOrder.PutUint64(v8[:], v)
	return v8
}

func (a TAddrValue) String() string {
	u64 := ByteOrder.Uint64(a[:])
	return fmt.Sprintf("%d", u64)
}

type TBlockKey [33]byte

//prefix + txid
type TTxKey [33]byte

//BlockHash[32] + txindex[4]
type TTxValue []byte

func NewTTxValue(hv HashID, idx uint32) TTxValue {
	buf := &bytes.Buffer{}
	buf.Write(hv[:])
	binary.Write(buf, ByteOrder, idx)
	return buf.Bytes()
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
	//cache
	key := NewTBlockKey(id)
	bb, err := DB().Get(key[:], nil)
	if err != nil {
		return nil, err
	}
	bv := TBlock(bb)
	m := bv.ToBlock()
	m.Height = bv.Height()
	//
	return m, nil
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
	v, err := LoadTxValue(tx)
	if err != nil {
		return nil, err
	}
	return v.GetTx()
}

func LoadTxValue(tx HashID) (TTxValue, error) {
	txkey := NewTxKey(tx)
	return DB().Get(txkey[:], nil)
}

//sb = save best
func (m *MsgBlock) Save(sb bool) error {
	batch := &leveldb.Batch{}
	//save block data
	bkey := NewTBlockKey(m.Hash)
	batch.Put(bkey[:], NewTBlock(m))
	//save tx index,addr index
	for idx, tx := range m.Txs {
		//txid  -> block txs[idx]
		txkey := NewTxKey(tx.Hash)
		batch.Put(txkey[:], NewTTxValue(m.Hash, uint32(idx)))
		//cost value
		for iidx, in := range tx.Ins {
			if iidx == 0 && tx.IsCoinBase() {
				continue
			}
			outtx, err := LoadTx(in.OutHash)
			if err != nil {
				return fmt.Errorf("load outtx failed: %v, tx=%v[%d] miss", err, in.OutHash, in.OutIndex)
			}
			if int(in.OutIndex) >= len(outtx.Outs) {
				return fmt.Errorf("outindex outbound outs block=%v tx=%v", m.Hash, tx.Hash)
			}
			outv := outtx.Outs[in.OutIndex]
			if outv.Script == nil {
				return fmt.Errorf("out script nil,error")
			}
			if outv.Value == 0 {
				continue
			}
			addr := outv.Script.GetAddress()
			if addr == "" {
				log.Println("warn, address parse failed 1")
				continue
			}
			//cost addr
			akey := NewTAddrKey(addr, in.OutHash, in.OutIndex)
			batch.Delete(akey)
		}
		//get value
		for oidx, out := range tx.Outs {
			if out.Value == 0 {
				continue
			}
			if out.Script == nil {
				return fmt.Errorf("out script nil,error")
			}
			addr := out.Script.GetAddress()
			if addr == "" {
				log.Println("warn, address parse failed 2")
				continue
			}
			akey := NewTAddrKey(addr, tx.Hash, uint32(oidx))
			aval := NewTAddrValue(out.Value)
			batch.Put(akey, aval[:])
		}
	}
	//update best block
	if sb {
		batch.Put([]byte(TBestBlockHashKey), m.Hash[:])
	}
	return DB().Write(batch, nil)
}

func NewTBlock(m *MsgBlock) TBlock {
	b := make(TBlock, 4+len(m.Body))
	ByteOrder.PutUint32(b[:4], m.Height)
	copy(b[4:], m.Body)
	return b
}
