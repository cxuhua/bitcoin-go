package core

import (
	"bitcoin/config"
	"bitcoin/script"
	"bitcoin/store"
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"log"

	"github.com/willf/bitset"

	"go.mongodb.org/mongo-driver/mongo"
)

type Inventory struct {
	Type uint32
	ID   HashID
}

func (m *Inventory) Read(h *NetHeader) {
	m.Type = h.ReadUInt32()
	h.ReadBytes(m.ID[:])
}

func (m *Inventory) Write(h *NetHeader) {
	h.WriteUInt32(m.Type)
	h.WriteBytes(m.ID[:])
}

type BHeader struct {
	Ver       uint32
	Prev      HashID
	Merkle    HashID //Merkle tree root
	Timestamp uint32
	Bits      uint32
	Nonce     uint32
	Count     uint64
	Hash      HashID
}

/*
	Hash      []byte `bson:"_id"`
	Ver       uint32 `bson:"ver"`
	Prev      []byte `bson:"prev"`
	Merkle    []byte `bson:"merkel"`
	Timestamp uint32 `bson:"time"`
	Bits      uint32 `bson:"bits"`
	Nonce     uint32 `bson:"nonce"`
	Height    uint64 `bson:"height"`
	Count     int    `bson:"count"` //tx count
*/
func (m *BHeader) ToBlockHeader() *BlockHeader {
	bh := NewBlockHeader()
	copy(bh.Hash, m.Hash[:])
	bh.Ver = m.Ver
	copy(bh.Prev, m.Prev[:])
	copy(bh.Merkle, m.Merkle[:])
	bh.Timestamp = m.Timestamp
	bh.Nonce = m.Nonce
	bh.Bits = m.Bits
	bh.Count = int(m.Count)
	return bh
}

func (m *BHeader) Read(h *NetHeader) {
	sp := h.Pos()
	m.Ver = h.ReadUInt32()
	h.ReadBytes(m.Prev[:])
	h.ReadBytes(m.Merkle[:])
	m.Timestamp = h.ReadUInt32()
	m.Bits = h.ReadUInt32()
	m.Nonce = h.ReadUInt32()
	ep := h.Pos()
	HASH256To(h.SubBytes(sp, ep), &m.Hash)
	//always 0
	m.Count, _ = h.ReadVarInt()
}

func (m *BHeader) Write(h *NetHeader) {
	h.WriteUInt32(m.Ver)
	h.WriteBytes(m.Prev[:])
	h.WriteBytes(m.Merkle[:])
	h.WriteUInt32(m.Timestamp)
	h.WriteUInt32(m.Bits)
	h.WriteUInt32(m.Nonce)
	h.WriteVarInt(m.Count)
}

type TxOut struct {
	Value  uint64
	Script *script.Script
}

func (m *TxOut) ToSubMoneys(txid HashID, idx uint32) *Moneys {
	mv := NewMoneys()
	mv.Id = NewMoneyId(txid, idx, store.MT_IO_IN)
	if addr := m.Script.GetAddress(); addr == "" {
		log.Println("address parse failed", hex.EncodeToString(*m.Script))
		return nil
	} else {
		mv.Addr = addr
	}
	mv.Value = Amount(-m.Value)
	return mv
}

func (m *TxOut) ToAddMoneys(txid HashID, idx uint32) *Moneys {
	mv := NewMoneys()
	mv.Id = NewMoneyId(txid, idx, store.MT_IO_OUT)
	if addr := m.Script.GetAddress(); addr == "" {
		log.Println("address parse failed", hex.EncodeToString(*m.Script))
		return nil
	} else {
		mv.Addr = addr
	}
	mv.Value = Amount(m.Value)
	return mv
}

func (m *TxOut) Clone() *TxOut {
	v := &TxOut{}
	v.Value = m.Value
	v.Script = m.Script.Clone()
	return v
}

func (m *TxOut) Read(h *NetHeader) {
	m.Value = h.ReadUInt64()
	m.Script = h.ReadScript()
}

func (m *TxOut) Write(h *NetHeader) {
	h.WriteUInt64(m.Value)
	h.WriteScript(m.Script)
}

type TxIn struct {
	OutHash  HashID
	OutIndex uint32
	Script   *script.Script
	Sequence uint32
	Witness  *TxWitnesses
}

func (m *TxIn) OutTx(db store.DbImp) (*TxOut, error) {
	if m.OutHash.IsZero() {
		return nil, errors.New("coinbase first txin,No previous out")
	}
	tx, err := LoadTX(m.OutHash, db)
	if err != nil {
		return nil, err
	}
	if int(m.OutIndex) >= len(tx.Outs) {
		return nil, errors.New("out index outbound pre tx outs")
	}
	return tx.Outs[m.OutIndex], nil
}

func (m *TxIn) Clone() *TxIn {
	v := &TxIn{}
	copy(v.OutHash[:], m.OutHash[:])
	v.OutIndex = m.OutIndex
	v.Script = m.Script.Clone()
	v.Sequence = m.Sequence
	if m.Witness != nil {
		v.Witness = m.Witness.Clone()
	} else {
		v.Witness = nil
	}
	return v
}

func (m *TxIn) Read(h *NetHeader) {
	h.ReadBytes(m.OutHash[:])
	m.OutIndex = h.ReadUInt32()
	m.Script = h.ReadScript()
	m.Sequence = h.ReadUInt32()
}

func (m *TxIn) Write(h *NetHeader) {
	h.WriteBytes(m.OutHash[:])
	h.WriteUInt32(m.OutIndex)
	h.WriteScript(m.Script)
	h.WriteUInt32(m.Sequence)
}

type TxWitnesses struct {
	Script []*script.Script
}

func (m *TxWitnesses) Clone() *TxWitnesses {
	if len(m.Script) == 0 {
		return nil
	}
	ss := make([]*script.Script, len(m.Script))
	for i, v := range m.Script {
		ss[i] = v.Clone()
	}
	return &TxWitnesses{Script: ss}
}

func (m *TxWitnesses) Read(h *NetHeader) {
	wl, _ := h.ReadVarInt()
	m.Script = make([]*script.Script, wl)
	for i, _ := range m.Script {
		m.Script[i] = h.ReadScript()
	}
}

func (m *TxWitnesses) Write(h *NetHeader) {
	h.WriteVarInt(len(m.Script))
	for _, v := range m.Script {
		h.WriteScript(v)
	}
}

/**
 * Basic TX serialization format:
 * - int32_t nVersion
 * - std::vector<CTxIn> vin
 * - std::vector<CTxOut> vout
 * - uint32_t nLockTime
 *
 * Extended TX serialization format:
 * - int32_t nVersion
 * - unsigned char dummy = 0x00
 * - unsigned char flags (!= 0)
 * - std::vector<CTxIn> vin
 * - std::vector<CTxOut> vout
 * - if (flags & 1):
 *   - CTxWitness wit;
 * - uint32_t nLockTime
 */
type TX struct {
	Hash     HashID //
	Block    HashID //block
	Index    uint32 //block index
	Ver      int32
	Flag     []byte //If present, always 0001
	Ins      []*TxIn
	Outs     []*TxOut
	LockTime uint32
	wbpos    int    //witness wpos begin
	wepos    int    //witness wpos end
	Witness  []byte //Witnesses data
	bbpos    int    //body begin pos
	bepos    int    //body end pos
	Body     []byte //body data PS: not include witness data and flag
	rbpos    int    //raw data begin
	repos    int    //raw data end
	Raw      []byte //raw data
}

func NewTX(bid HashID, idx uint32) *TX {
	return &TX{
		Block: bid,
		Index: uint32(idx),
	}
}

type TXHeader struct {
	Id    []byte `bson:"_id"`
	Block []byte `bson:"block"` //block id
	Index uint32 `bson:"index"` //block tx index
	Ver   int32  `bson:"ver"`
	State int8   `bson:"state"`
	Raw   []byte `bson:"raw"`
}

func (txh *TXHeader) ToTX() *TX {
	tx := &TX{}
	h := NewNetHeader(txh.Raw)
	tx.Read(h)
	copy(tx.Block[:], txh.Block)
	tx.Index = txh.Index
	return tx
}

func LoadTX(id HashID, d store.DbImp) (*TX, error) {
	//from cache get
	if v, err := d.TopTxCacher().Get(id[:]); err == nil {
		return v.(*TX), nil
	}
	//from database
	txh := &TXHeader{}
	err := d.GetTX(id.Bytes(), txh)
	if err != nil {
		return nil, err
	}
	tx := txh.ToTX()
	//set to cache
	d.TopTxCacher().Set(id[:], tx)
	return tx, nil
}

func NewTXFrom(tx *TX) *TXHeader {
	txh := &TXHeader{}
	txh.Id = tx.Hash[:]
	txh.Ver = tx.Ver
	txh.Raw = tx.Raw
	txh.Block = tx.Block[:]
	txh.Index = tx.Index
	return txh
}

//check tx ,return trans fee
func (m *TX) GetFee(db store.DbImp) (Amount, error) {
	iv := Amount(0)
	ov := Amount(0)
	if m.IsCoinBase() {
		for _, v := range m.Outs {
			ov += Amount(v.Value)
		}
		if !ov.IsRange() {
			return 0, errors.New("out amount sum error")
		}
		return ov, nil
	}
	for _, v := range m.Ins {
		out, err := v.OutTx(db)
		if err != nil {
			return 0, err
		}
		iv += Amount(out.Value)
	}
	if !iv.IsRange() {
		return 0, errors.New("in amount sum error")
	}
	for _, v := range m.Outs {
		ov += Amount(v.Value)
	}
	if !ov.IsRange() {
		return 0, errors.New("out amount sum error")
	}
	if iv < ov {
		return 0, errors.New("in out amount sub error")
	}
	return iv - ov, nil
}

func (m *TX) Save(d store.DbImp) error {
	h := NewTXFrom(m)
	return d.SetTX(m.Hash[:], h)
}

func (m *TX) Clone() *TX {
	v := &TX{}
	v.Ver = m.Ver
	v.Block = m.Block
	v.Index = m.Index
	v.Hash = m.Hash
	if len(m.Flag) > 0 {
		v.Flag = make([]byte, len(m.Flag))
		copy(v.Flag, m.Flag)
	}
	v.Ins = make([]*TxIn, len(m.Ins))
	for i, iv := range m.Ins {
		v.Ins[i] = iv.Clone()
	}
	v.Outs = make([]*TxOut, len(m.Outs))
	for i, ov := range m.Outs {
		v.Outs[i] = ov.Clone()
	}
	v.LockTime = m.LockTime
	v.Body = m.Body
	v.Witness = m.Witness
	return v
}

/*
Check syntactic correctness
Make sure neither in or out lists are empty
Size in bytes <= MAX_BLOCK_SIZE
Each output value, as well as the total, must be in legal money range
Make sure none of the inputs have hash=0, n=-1 (coinbase transactions)
Check that nLockTime <= INT_MAX[1], size in bytes >= 100[2], and sig opcount <= 2[3]
Reject "nonstandard" transactions: scriptSig doing anything other than pushing numbers on the stack, or scriptPubkey not matching the two usual forms[4]
Reject if we already have matching tx in the pool, or in a block in the main branch
For each input, if the referenced output exists in any other tx in the pool, reject this transaction.[5]
For each input, look in the main branch and the transaction pool to find the referenced output transaction. If the output transaction is missing for any input, this will be an orphan transaction. Add to the orphan transactions, if a matching transaction is not in there already.
For each input, if the referenced output transaction is coinbase (i.e. only 1 input, with hash=0, n=-1), it must have at least COINBASE_MATURITY (100) confirmations; else reject this transaction
For each input, if the referenced output does not exist (e.g. never existed or has already been spent), reject this transaction[6]
Using the referenced output transactions to get input values, check that each input value, as well as the sum, are in legal money range
Reject if the sum of input values < sum of output values
Reject if transaction fee (defined as sum of input values minus sum of output values) would be too low to get into an empty block
Verify the scriptPubKey accepts for each input; reject if any are bad
Add to transaction pool[7]
"Add to wallet if mine"
Relay transaction to peers*/
func (m *TX) Check() error {
	if len(m.Ins) == 0 {
		return errors.New("bad-txns-vin-empty")
	}
	if len(m.Outs) == 0 {
		return errors.New("bad-txns-outs-empty")
	}
	vout := Amount(0)
	for _, v := range m.Outs {
		if v.Script == nil {
			return errors.New("script miss")
		}
		if len(*v.Script) > script.MAX_SCRIPT_SIZE {
			return errors.New("script too long")
		}
		if int64(v.Value) < 0 {
			return errors.New("bad-txns-vout-negative")
		}
		if !Amount(v.Value).IsRange() {
			return errors.New("bad-txns-vout-toolarge")
		}
		vout += Amount(v.Value)
		if !vout.IsRange() {
			return errors.New("bad-txns-txouttotal-toolarge")
		}
	}
	if m.IsCoinBase() {
		if m.Ins[0].Script.Len() < 2 || m.Ins[0].Script.Len() > 100 {
			return errors.New("bad-cb-length")
		}
		if len(m.Ins) != 1 {
			return errors.New("bad-cb-ins-count")
		}
	} else {
		for _, v := range m.Ins {
			if len(*v.Script) > script.MAX_SCRIPT_SIZE {
				return errors.New("script too long")
			}
			if v.OutHash.IsZero() {
				return errors.New("bad-txns-prevout-null")
			}
		}
	}
	return nil
}

func (m *TX) IsFinal(blockHeight, blockTime int64) bool {
	if m.LockTime == 0 {
		return false
	}
	lt := int64(0)
	if m.LockTime < script.LOCKTIME_THRESHOLD {
		lt = blockHeight
	} else {
		lt = blockTime
	}
	if int64(m.LockTime) < lt {
		return true
	}
	for _, v := range m.Ins {
		if v.Sequence != script.SEQUENCE_FINAL {
			return false
		}
	}
	return true
}

func (m *TX) GetValueOut() Amount {
	tv := Amount(0)
	for _, v := range m.Outs {
		tv += Amount(v.Value)
	}
	return tv
}

func (m *TX) IsCoinBase() bool {
	if len(m.Ins) <= 0 {
		return false
	}
	if !m.Ins[0].OutHash.IsZero() {
		return false
	}
	if m.Ins[0].Script == nil {
		return false
	}
	sl := m.Ins[0].Script.Len()
	return sl >= 2 && sl <= 100
}

func (m *TX) ReadWitnesses(h *NetHeader) {
	m.wbpos = h.Pos()
	for i, _ := range m.Ins {
		v := &TxWitnesses{}
		v.Read(h)
		m.Ins[i].Witness = v
	}
	m.wepos = h.Pos()
	m.Witness = h.SubBytes(m.wbpos, m.wepos)
}

func (m *TX) Read(h *NetHeader) {
	m.rbpos = h.Pos()
	buf := bytes.Buffer{}
	//+ver
	m.bbpos = h.Pos()
	m.Ver = int32(h.ReadUInt32())
	m.bepos = h.Pos()
	buf.Write(h.SubBytes(m.bbpos, m.bepos))
	//check flag for witnesses
	m.Flag = h.Peek(2)
	if m.HasWitness() {
		h.Skip(2)
	}
	//+ins outs
	m.bbpos = h.Pos()
	il, _ := h.ReadVarInt()
	m.Ins = make([]*TxIn, il)
	for i, _ := range m.Ins {
		v := &TxIn{}
		v.Read(h)
		m.Ins[i] = v
	}
	ol, _ := h.ReadVarInt()
	m.Outs = make([]*TxOut, ol)
	for i, _ := range m.Outs {
		v := &TxOut{}
		v.Read(h)
		m.Outs[i] = v
	}
	m.bepos = h.Pos()
	buf.Write(h.SubBytes(m.bbpos, m.bepos))
	//if has witnesses
	if m.HasWitness() {
		m.ReadWitnesses(h)
	}
	//lock time
	m.bbpos = h.Pos()
	m.LockTime = h.ReadUInt32()
	m.repos = h.Pos()
	m.bepos = h.Pos()
	buf.Write(h.SubBytes(m.bbpos, m.bepos))
	//hash get tx id
	m.Body = buf.Bytes()
	HASH256To(m.Body, &m.Hash)
	m.Raw = h.SubBytes(m.rbpos, m.repos)
}

func (m *TX) HasWitness() bool {
	return len(m.Flag) == 2 && m.Flag[0] == 0 && m.Flag[1] == 1
}

func (m *TX) WriteWitnesses(h *NetHeader) {
	m.wbpos = h.Pos()
	for _, v := range m.Ins {
		v.Witness.Write(h)
	}
	m.wepos = h.Pos()
	m.Witness = h.SubBytes(m.wbpos, m.wepos)
}

func (m *TX) SetHasWitness(v bool) {
	if v {
		m.Flag = []byte{0, 1}
	} else {
		m.Flag = []byte{}
	}
}

func (m *TX) Write(h *NetHeader) {
	buf := bytes.Buffer{}
	m.bbpos = h.Pos()
	h.WriteUInt32(uint32(m.Ver))
	m.bepos = h.Pos()
	buf.Write(h.SubBytes(m.bbpos, m.bepos))

	if m.HasWitness() {
		h.WriteBytes(m.Flag)
	}

	m.bbpos = h.Pos()
	h.WriteVarInt(len(m.Ins))
	for _, v := range m.Ins {
		v.Write(h)
	}
	h.WriteVarInt(len(m.Outs))
	for _, v := range m.Outs {
		v.Write(h)
	}
	m.bepos = h.Pos()
	buf.Write(h.SubBytes(m.bbpos, m.bepos))

	if m.HasWitness() {
		m.WriteWitnesses(h)
	}
	m.bbpos = h.Pos()
	h.WriteUInt32(m.LockTime)
	m.bepos = h.Pos()
	buf.Write(h.SubBytes(m.bbpos, m.bepos))

	m.Body = buf.Bytes()
	HASH256To(m.Body, &m.Hash)
}

//
type MsgHeaders struct {
	Headers []*BHeader
}

func (m *MsgHeaders) Command() string {
	return NMT_HEADERS
}

func (m *MsgHeaders) Read(h *NetHeader) {
	num, _ := h.ReadVarInt()
	m.Headers = make([]*BHeader, num)
	for i, _ := range m.Headers {
		v := &BHeader{}
		v.Read(h)
		m.Headers[i] = v
	}
}

func (m *MsgHeaders) Write(h *NetHeader) {
	h.WriteVarInt(len(m.Headers))
	for _, v := range m.Headers {
		v.Write(h)
	}
}

func NewMsgHeaders() *MsgHeaders {
	return &MsgHeaders{}
}

//
type MsgGetBlocks struct {
	Ver    uint32
	Blocks []HashID
	Stop   HashID
}

func (m *MsgGetBlocks) Command() string {
	return NMT_GETBLOCKS
}

func (m *MsgGetBlocks) AddHash(hv []byte) {
	m.Blocks = append(m.Blocks, NewHashID(hv))
}

func (m *MsgGetBlocks) AddHashID(hv HashID) {
	m.Blocks = append(m.Blocks, hv)
}

func (m *MsgGetBlocks) Read(h *NetHeader) {
	m.Ver = h.ReadUInt32()
	num, _ := h.ReadVarInt()
	m.Blocks = make([]HashID, num)
	for i, _ := range m.Blocks {
		h.ReadBytes(m.Blocks[i][:])
	}
	h.ReadBytes(m.Stop[:])
}

func (m *MsgGetBlocks) Write(h *NetHeader) {
	h.WriteUInt32(m.Ver)
	h.WriteVarInt(len(m.Blocks))
	for _, v := range m.Blocks {
		h.WriteBytes(v[:])
	}
	h.WriteBytes(m.Stop[:])
}

func NewMsgGetBlocks() *MsgGetBlocks {
	return &MsgGetBlocks{
		Ver:  PROTOCOL_VERSION,
		Stop: HashID{},
	}
}

//

type MsgNotFound struct {
	Invs []*Inventory
}

func (m *MsgNotFound) Command() string {
	return NMT_NOTFOUND
}

func (m *MsgNotFound) Read(h *NetHeader) {
	size, _ := h.ReadVarInt()
	m.Invs = make([]*Inventory, size)
	for i, _ := range m.Invs {
		v := &Inventory{}
		v.Read(h)
		m.Invs[i] = v
	}
}

func (m *MsgNotFound) Write(h *NetHeader) {
	h.WriteVarInt(len(m.Invs))
	for _, v := range m.Invs {
		v.Write(h)
	}
}

func NewMsgNotFound() *MsgNotFound {
	return &MsgNotFound{}
}

//

type MsgBlock struct {
	Hash      HashID //compute get
	Ver       uint32
	Prev      HashID
	Merkle    HashID //Merkle tree root
	Timestamp uint32
	Bits      uint32
	Nonce     uint32
	Txs       []*TX
	Body      []byte //don't include header 80 bytes
	bbpos     int
	bepos     int
	Count     int
}

//block header
type BlockHeader struct {
	Hash      []byte `bson:"_id"`
	Ver       uint32 `bson:"ver"`
	Prev      []byte `bson:"prev"`
	Merkle    []byte `bson:"merkel"`
	Timestamp uint32 `bson:"time"`
	Bits      uint32 `bson:"bits"`
	Nonce     uint32 `bson:"nonce"`
	Height    uint32 `bson:"height"`
	Count     int    `bson:"count"` //tx count ,=0 not dowload
}

func (b *BlockHeader) Equal(v *BlockHeader) bool {
	if !bytes.Equal(b.Hash, v.Hash) {
		return false
	}
	if b.Ver != v.Ver {
		return false
	}
	if !bytes.Equal(b.Prev, v.Prev) {
		return false
	}
	if !bytes.Equal(b.Merkle, v.Merkle) {
		return false
	}
	if b.Timestamp != v.Timestamp {
		return false
	}
	if b.Bits != v.Bits {
		return false
	}
	if b.Nonce != v.Nonce {
		return false
	}
	return true
}

func NewBlockHeader() *BlockHeader {
	bh := &BlockHeader{}
	bh.Hash = make([]byte, len(HashID{}))
	bh.Prev = make([]byte, len(HashID{}))
	bh.Merkle = make([]byte, len(HashID{}))
	return bh
}

func (b *BlockHeader) IsGenesis() bool {
	conf := config.GetConfig()
	gid := NewHashID(conf.GenesisBlock)
	return bytes.Equal(ZeroHashID[:], b.Prev[:]) && bytes.Equal(gid[:], b.Hash[:])
}

func (b *BlockHeader) ToBlock() *MsgBlock {
	m := &MsgBlock{}
	copy(m.Hash[:], b.Hash)
	m.Ver = b.Ver
	copy(m.Prev[:], b.Prev)
	copy(m.Merkle[:], b.Merkle)
	m.Timestamp = b.Timestamp
	m.Bits = b.Bits
	m.Nonce = b.Nonce
	m.Count = b.Count
	return m
}

func (m *MsgBlock) ToBlockHeader() *BlockHeader {
	h := NewBlockHeader()
	copy(h.Hash, m.Hash[:])
	h.Ver = m.Ver
	copy(h.Prev, m.Prev[:])
	copy(h.Merkle, m.Merkle[:])
	h.Timestamp = m.Timestamp
	h.Bits = m.Bits
	h.Nonce = m.Nonce
	h.Height = 0
	h.Count = len(m.Txs)
	return h
}

func (b *MsgBlock) IsGenesis() bool {
	conf := config.GetConfig()
	gid := NewHashID(conf.GenesisBlock)
	return bytes.Equal(ZeroHashID[:], b.Prev[:]) && bytes.Equal(gid[:], b.Hash[:])
}

func (m *MsgBlock) GetAncestor(h int, db store.DbImp) (*BlockHeader, error) {
	bh := &BlockHeader{}
	err := db.GetBK(store.BKHeight(uint32(h)), bh)
	return bh, err
}

//get prev bits
func (m *MsgBlock) prevBits(lb *BlockHeader, db store.DbImp) uint32 {
	conf := config.GetConfig()
	hfv := int(lb.Height) - conf.DiffAdjusInterval() + 1
	if hfv < 0 {
		log.Panicf("first height %d error", hfv)
	}
	bh, err := m.GetAncestor(hfv, db)
	if err != nil {
		log.Panicf("get block for height %d error", hfv)
	}
	return CalculateWorkRequired(lb.Timestamp, bh.Timestamp, lb.Bits)
}

//check proof of work return height
func (m *MsgBlock) checkBits(lb *BlockHeader, db store.DbImp) (int, error) {
	conf := config.GetConfig()
	if !CheckProofOfWork(m.Hash, m.Bits) {
		return 0, errors.New("block proof of work check error")
	}
	height := 0
	if lb != nil {
		dav := conf.DiffAdjusInterval()
		bits := uint32(0)
		if (int(lb.Height)+1)%dav != 0 {
			bits = lb.Bits
		} else {
			bits = m.prevBits(lb, db)
		}
		if m.Bits != bits {
			return 0, errors.New("block bits error")
		}
		height = int(lb.Height + 1)
	} else if !m.IsGenesis() {
		return 0, errors.New("last block get error")
	}
	return height, nil
}

//check recv block
func (m *MsgBlock) CheckBlock(lb *BlockHeader, db store.DbImp) error {
	txids := []HashID{}
	if len(m.Txs) == 0 {
		return errors.New("miss tx data")
	}
	height, err := m.checkBits(lb, db)
	if err != nil {
		return err
	}
	bfee, vfee, cfee := Amount(0), GetCoinbaseReward(height), Amount(0)
	if !vfee.IsRange() {
		return errors.New("get coinbase reward error")
	}
	db.PushTxCacher(NewCacher())
	defer db.PopTxCacher()
	for i, v := range m.Txs {
		if i == 0 && !v.IsCoinBase() {
			return errors.New("0 tx not coinbase")
		}
		//coinbase txid,There may be the same
		if !v.IsCoinBase() && db.HasTX(v.Hash[:]) {
			return fmt.Errorf("block tx exists txid=%v", v.Hash)
		}
		if i == 4 {
			log.Println("a")
		}
		if err := VerifyTX(v, db); err != nil {
			return fmt.Errorf("verify tx error %v", err)
		}
		txids = append(txids, v.Hash)
		av, err := v.GetFee(db)
		if err != nil {
			return fmt.Errorf("check tx amount error %v", err)
		}
		if v.IsCoinBase() {
			cfee = av
		} else {
			bfee += av
		}
		db.TopTxCacher().Set(v.Hash[:], v)
	}
	if !cfee.IsRange() || !bfee.IsRange() {
		return errors.New("check block fee error")
	}
	if vfee != (cfee - bfee) {
		return errors.New("block amount error")
	}
	root, _, _ := BuildMerkleTree(txids).Extract()
	if root.IsZero() {
		return errors.New("merkle tree root error")
	}
	if !root.Equal(m.Merkle) {
		return errors.New("Calc merkle id error")
	}
	return nil
}

//load txs from database
func (m *MsgBlock) LoadTXS(db store.DbImp) error {
	m.Txs = []*TX{}
	err := db.GetTX(store.NewListBlockTxs(m.Hash[:]), store.IterFunc(func(cur *mongo.Cursor) (bool, error) {
		txh := TXHeader{}
		if err := cur.Decode(&txh); err != nil {
			return true, err
		}
		m.Txs = append(m.Txs, txh.ToTX())
		return true, nil
	}))
	m.Count = len(m.Txs)
	return err
}
func (m *MsgBlock) PrevBlock(db store.DbImp) (*MsgBlock, error) {
	return LoadBlock(m.Prev, db)
}

func (m *MsgBlock) SaveTXS(db store.DbImp) error {
	vs := make([]interface{}, 0)
	for _, v := range m.Txs {
		//allow coinbase txid same
		if v.IsCoinBase() && db.HasTX(v.Hash[:]) {
			continue
		}
		vs = append(vs, NewTXFrom(v))
	}
	if len(vs) == 0 {
		return nil
	}
	return db.MulTX(vs)
}

func (m *MsgBlock) Save(db store.DbImp) error {
	b := &BlockHeader{}
	b.Hash = m.Hash[:]
	b.Ver = m.Ver
	b.Prev = m.Prev[:]
	b.Merkle = m.Merkle[:]
	b.Timestamp = m.Timestamp
	b.Bits = m.Bits
	b.Nonce = m.Nonce
	b.Count = m.Count
	return db.SetBK(b.Hash[:], b)
}

func LoadBlock(id HashID, db store.DbImp) (*MsgBlock, error) {
	h := &BlockHeader{}
	if err := db.GetBK(id[:], h); err != nil {
		return nil, err
	}
	return h.ToBlock(), nil
}

func (m *MsgBlock) Command() string {
	return NMT_BLOCK
}

func (m *MsgBlock) Read(h *NetHeader) {
	hs := h.Pos()
	m.Ver = h.ReadUInt32()
	h.ReadBytes(m.Prev[:])
	h.ReadBytes(m.Merkle[:])
	m.Timestamp = h.ReadUInt32()
	m.Bits = h.ReadUInt32()
	m.Nonce = h.ReadUInt32()
	HASH256To(h.Payload[hs:h.Pos()], &m.Hash)
	m.bbpos = h.Pos()
	l, _ := h.ReadVarInt()
	m.Txs = make([]*TX, l)
	for i, _ := range m.Txs {
		v := NewTX(m.Hash, uint32(i))
		v.Read(h)
		m.Txs[i] = v
	}
	m.bepos = h.Pos()
	m.Body = h.SubBytes(m.bbpos, m.bepos)
	m.Count = len(m.Txs)
}

func (m *MsgBlock) Write(h *NetHeader) {
	hp := h.Pos()
	h.WriteUInt32(m.Ver)
	h.WriteBytes(m.Prev[:])
	h.WriteBytes(m.Merkle[:])
	h.WriteUInt32(m.Timestamp)
	h.WriteUInt32(m.Bits)
	h.WriteUInt32(m.Nonce)
	h.WriteVarInt(len(m.Txs))
	HASH256To(h.Payload[hp:h.Pos()], &m.Hash)
	m.bbpos = h.Pos()
	for _, v := range m.Txs {
		v.Write(h)
	}
	m.bepos = h.Pos()
	m.Body = h.SubBytes(m.bbpos, m.bepos)
	m.Count = len(m.Txs)
}

func (m *MsgBlock) BuildMarkleTree() *MerkleTree {
	hvs := []HashID{}
	for _, v := range m.Txs {
		hvs = append(hvs, v.Hash)
	}
	num := len(hvs)
	bits := bitset.New(uint(num))
	tree := NewMerkleTree(num)
	return tree.Build(hvs, bits)
}

func NewMsgBlock() *MsgBlock {
	return &MsgBlock{}
}

//
type MsgGetData struct {
	Invs []Inventory
}

func (m *MsgGetData) Command() string {
	return NMT_GETDATA
}

func (m *MsgGetData) Read(h *NetHeader) {
	num, _ := h.ReadVarInt()
	m.Invs = make([]Inventory, num)
	for i, _ := range m.Invs {
		m.Invs[i].Read(h)
	}
}

func (m *MsgGetData) AddHash(typ uint32, hv []byte) {
	m.Invs = append(m.Invs, Inventory{
		Type: typ,
		ID:   NewHashID(hv),
	})
}
func (m *MsgGetData) Add(inv Inventory) {
	m.Invs = append(m.Invs, inv)
}

func (m *MsgGetData) Write(h *NetHeader) {
	h.WriteVarInt(len(m.Invs))
	for _, v := range m.Invs {
		v.Write(h)
	}
}

func NewMsgGetData() *MsgGetData {
	return &MsgGetData{
		Invs: []Inventory{},
	}
}

//
type MsgTX struct {
	Tx TX
}

func (m *MsgTX) Command() string {
	return NMT_TX
}

func (m *MsgTX) Read(h *NetHeader) {
	m.Tx.Read(h)
}

func (m *MsgTX) Write(h *NetHeader) {
	m.Tx.Write(h)
}

func NewMsgTX() *MsgTX {
	return &MsgTX{}
}

type MsgINV struct {
	Invs []*Inventory
}

func (m *MsgINV) Command() string {
	return NMT_INV
}

func (m *MsgINV) Read(h *NetHeader) {
	num, _ := h.ReadVarInt()
	m.Invs = make([]*Inventory, num)
	for i, _ := range m.Invs {
		v := &Inventory{}
		v.Read(h)
		m.Invs[i] = v
	}
}

func (m *MsgINV) Write(h *NetHeader) {
	h.WriteVarInt(len(m.Invs))
	for _, v := range m.Invs {
		v.Write(h)
	}
}

func NewMsgINV() *MsgINV {
	return &MsgINV{}
}

//

type MsgGetHeaders struct {
	Ver    uint32
	Blocks []HashID
	Stop   HashID
}

func (m *MsgGetHeaders) Command() string {
	return NMT_GETHEADERS
}

func (m *MsgGetHeaders) AddHashID(hv HashID) {
	m.Blocks = append(m.Blocks, hv)
}

func (m *MsgGetHeaders) AddHash(hv []byte) {
	m.Blocks = append(m.Blocks, NewHashID(hv))
}

func (m *MsgGetHeaders) Read(h *NetHeader) {
	m.Ver = h.ReadUInt32()
	num, _ := h.ReadVarInt()
	m.Blocks = make([]HashID, num)
	for i, _ := range m.Blocks {
		h.ReadBytes(m.Blocks[i][:])
	}
	m.Stop = HashID{}
	h.ReadBytes(m.Stop[:])
}

func (m *MsgGetHeaders) Write(h *NetHeader) {
	h.WriteUInt32(m.Ver)
	h.WriteVarInt(len(m.Blocks))
	for _, v := range m.Blocks {
		h.WriteBytes(v[:])
	}
	h.WriteBytes(m.Stop[:])
}

func NewMsgGetHeaders() *MsgGetHeaders {
	return &MsgGetHeaders{Ver: PROTOCOL_VERSION}
}
