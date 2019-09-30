package net

import (
	"bitcoin/script"
	"bitcoin/util"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

const (
	COIN      = Amount(100000000)
	MAX_MONEY = Amount(21000000 * COIN)
)

type Amount int64

func (a Amount) IsRange() bool {
	return a >= 0 && a < MAX_MONEY
}

type HashID [32]byte

func (h HashID) String() string {
	sv := h.Swap()
	return hex.EncodeToString(sv[:])
}

func (b HashID) IsZero() bool {
	bz := make([]byte, len(b))
	return bytes.Equal(b[:], bz)
}

func (b HashID) Swap() HashID {
	v := HashID{}
	j := 0
	for i := len(b) - 1; i >= 0; i-- {
		v[j] = b[i]
		j++
	}
	return v
}

func NewHexBHash(s string) HashID {
	b := HashID{}
	if len(s) != len(b)*2 {
		panic(SizeError)
	}
	v, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	copy(b[:], v)
	return b
}

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
}

func (m *BHeader) Read(h *NetHeader) {
	m.Ver = h.ReadUInt32()
	h.ReadBytes(m.Prev[:])
	h.ReadBytes(m.Merkle[:])
	m.Timestamp = h.ReadUInt32()
	m.Bits = h.ReadUInt32()
	m.Nonce = h.ReadUInt32()
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

func (m *TxIn) OutBytes() []byte {
	buf := &bytes.Buffer{}
	binary.Write(buf, ByteOrder, m.OutHash[:])
	binary.Write(buf, ByteOrder, m.OutIndex)
	return buf.Bytes()
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

func (m *TxWitnesses) Read(h *NetHeader) {
	wl, _ := h.ReadVarInt()
	m.Script = make([]*script.Script, wl)
	for i, _ := range m.Script {
		v := h.ReadScript()
		m.Script[i] = v
	}
}

func (m *TxWitnesses) Write(h *NetHeader) {
	h.WriteVarInt(uint64(len(m.Script)))
	for _, v := range m.Script {
		h.WriteScript(v)
	}
}

const (
	INVALID_REASON_NONE = iota
	INVALID_REASON_CONSENSUS
	INVALID_REASON_RECENT_CONSENSUS_CHANGE
	INVALID_REASON_CACHED_INVALID
	INVALID_REASON_BLOCK_INVALID_HEADER
	INVALID_REASON_BLOCK_MUTATED
	INVALID_REASON_BLOCK_MISSING_PREV
	INVALID_REASON_BLOCK_INVALID_PREV
	INVALID_REASON_BLOCK_TIME_FUTURE
	INVALID_REASON_BLOCK_CHECKPOINT
	INVALID_REASON_TX_NOT_STANDARD
	INVALID_REASON_TX_MISSING_INPUTS
	INVALID_REASON_TX_PREMATURE_SPEND
	INVALID_REASON_TX_WITNESS_MUTATED
	INVALID_REASON_TX_CONFLICT
	INVALID_REASON_TX_MEMPOOL_POLICY
)

func IsTransactionReason(r int) bool {
	return r == INVALID_REASON_NONE ||
		r == INVALID_REASON_CONSENSUS ||
		r == INVALID_REASON_RECENT_CONSENSUS_CHANGE ||
		r == INVALID_REASON_TX_NOT_STANDARD ||
		r == INVALID_REASON_TX_PREMATURE_SPEND ||
		r == INVALID_REASON_TX_MISSING_INPUTS ||
		r == INVALID_REASON_TX_WITNESS_MUTATED ||
		r == INVALID_REASON_TX_CONFLICT ||
		r == INVALID_REASON_TX_MEMPOOL_POLICY
}

func IsBlockReason(r int) bool {
	return r == INVALID_REASON_NONE ||
		r == INVALID_REASON_CONSENSUS ||
		r == INVALID_REASON_RECENT_CONSENSUS_CHANGE ||
		r == INVALID_REASON_CACHED_INVALID ||
		r == INVALID_REASON_BLOCK_INVALID_HEADER ||
		r == INVALID_REASON_BLOCK_MUTATED ||
		r == INVALID_REASON_BLOCK_MISSING_PREV ||
		r == INVALID_REASON_BLOCK_INVALID_PREV ||
		r == INVALID_REASON_BLOCK_TIME_FUTURE ||
		r == INVALID_REASON_BLOCK_CHECKPOINT
}

type ValidationError struct {
	reason int
	reject byte
	err    string
}

func (v *ValidationError) Reject() byte {
	return v.reject
}

func (v *ValidationError) Error() string {
	return fmt.Sprintf("reason=%d reject=%d err=%s", v.reason, v.reject, v.err)
}

func NewValidationError(reason int, reject byte, err string) *ValidationError {
	return &ValidationError{
		reason: reason,
		reject: reject,
		err:    err,
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
	Ver      int32
	Flag     []byte //If present, always 0001
	Ins      []*TxIn
	Outs     []*TxOut
	LockTime uint32
	wbpos    int //witness wpos begin
	wepos    int //witness wpos end
}

func (m *TX) Check(h *NetHeader, checkDupIn bool) error {
	if len(m.Ins) == 0 {
		return NewValidationError(INVALID_REASON_CONSENSUS, REJECT_INVALID, "bad-txns-vin-empty")
	}
	if len(m.Outs) == 0 {
		return NewValidationError(INVALID_REASON_CONSENSUS, REJECT_INVALID, "bad-txns-outs-empty")
	}
	if uint(h.Pos()-m.WitnessesLen())*WITNESS_SCALE_FACTOR > MAX_BLOCK_WEIGHT {
		return NewValidationError(INVALID_REASON_CONSENSUS, REJECT_INVALID, "bad-txns-oversize")
	}
	vout := Amount(0)
	for _, v := range m.Outs {
		if v.Value < 0 {
			return NewValidationError(INVALID_REASON_CONSENSUS, REJECT_INVALID, "bad-txns-vout-negative")
		}
		if Amount(v.Value) > MAX_MONEY {
			return NewValidationError(INVALID_REASON_CONSENSUS, REJECT_INVALID, "bad-txns-vout-toolarge")
		}
		vout += Amount(v.Value)
		if !vout.IsRange() {
			return NewValidationError(INVALID_REASON_CONSENSUS, REJECT_INVALID, "bad-txns-txouttotal-toolarge")
		}
	}
	if checkDupIn {
		inos := util.NewBytesArray()
		for _, v := range m.Ins {
			vbs := v.OutBytes()
			if inos.Has(vbs) {
				return NewValidationError(INVALID_REASON_CONSENSUS, REJECT_INVALID, "bad-txns-inputs-duplicate")
			}
			inos.Append(vbs)
		}
		inos = nil
	}
	if m.IsCoinBase() {
		if m.Ins[0].Script.Len() < 2 || m.Ins[0].Script.Len() > 100 {
			return NewValidationError(INVALID_REASON_CONSENSUS, REJECT_INVALID, "bad-cb-length")
		}
	} else {
		for _, v := range m.Ins {
			if v.OutHash.IsZero() {
				return NewValidationError(INVALID_REASON_CONSENSUS, REJECT_INVALID, "bad-txns-prevout-null")
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
func (m *TX) HashID() HashID {
	//not include witnesses
	h := NewNetHeader()
	h.WriteUInt32(uint32(m.Ver))
	h.WriteVarInt(uint64(len(m.Ins)))
	for _, v := range m.Ins {
		v.Write(h)
	}
	h.WriteVarInt(uint64(len(m.Outs)))
	for _, v := range m.Outs {
		v.Write(h)
	}
	h.WriteUInt32(m.LockTime)
	hid := util.HASH256(h.Bytes())
	id := HashID{}
	copy(id[:], hid)
	return id
}

func (m *TX) IsCoinBase() bool {
	return len(m.Ins) > 0 && m.Ins[0].OutHash.IsZero()
}

func (m *TX) HasFlag() bool {
	return len(m.Flag) == 2 && m.Flag[0] == 0 && m.Flag[1] == 1
}

func (m *TX) ReadWitnesses(h *NetHeader) {
	m.wbpos = h.Pos()
	for i, _ := range m.Ins {
		v := &TxWitnesses{}
		v.Read(h)
		m.Ins[i].Witness = v
	}
	m.wepos = h.Pos()
}

func (m *TX) Read(h *NetHeader) {
	m.Ver = int32(h.ReadUInt32())
	//check flag for witnesses
	m.Flag = h.Peek(2)
	if m.HasFlag() {
		h.Skip(2)
	}
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
	//if has witnesses
	if m.HasFlag() {
		m.ReadWitnesses(h)
	}
	m.LockTime = h.ReadUInt32()
}

func (m *TX) HasWitness() bool {
	for _, v := range m.Ins {
		if v.Witness == nil {
			return false
		}
	}
	return true
}

func (m *TX) WitnessesLen() int {
	return m.wepos - m.wbpos
}

func (m *TX) WriteWitnesses(h *NetHeader) {
	m.wbpos = h.Pos()
	for _, v := range m.Ins {
		if v.Witness == nil {
			continue
		}
		v.Witness.Write(h)
	}
	m.wepos = h.Pos()
}

func (m *TX) Write(h *NetHeader) {
	h.WriteUInt32(uint32(m.Ver))
	if m.HasWitness() {
		h.WriteBytes([]byte{0, 1})
	}
	h.WriteVarInt(uint64(len(m.Ins)))
	for _, v := range m.Ins {
		v.Write(h)
	}
	h.WriteVarInt(uint64(len(m.Outs)))
	for _, v := range m.Outs {
		v.Write(h)
	}
	if m.HasWitness() {
		m.WriteWitnesses(h)
	}
	h.WriteUInt32(m.LockTime)
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
	h.WriteVarInt(uint64(len(m.Headers)))
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
	Blocks []*HashID
	Stop   *HashID
}

func (m *MsgGetBlocks) Command() string {
	return NMT_GETBLOCKS
}

func (m *MsgGetBlocks) Read(h *NetHeader) {
	m.Ver = h.ReadUInt32()
	num, _ := h.ReadVarInt()
	m.Blocks = make([]*HashID, num)
	for i, _ := range m.Blocks {
		v := &HashID{}
		h.ReadBytes(v[:])
		m.Blocks[i] = v
	}
	m.Stop = &HashID{}
	h.ReadBytes(m.Stop[:])
}

func (m *MsgGetBlocks) Write(h *NetHeader) {
	h.WriteUInt32(m.Ver)
	h.WriteVarInt(uint64(len(m.Blocks)))
	for _, v := range m.Blocks {
		h.WriteBytes(v[:])
	}
	h.WriteBytes(m.Stop[:])
}

func NewMsgGetBlocks() *MsgGetBlocks {
	return &MsgGetBlocks{
		Ver:  PROTOCOL_VERSION,
		Stop: &HashID{},
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
	h.WriteVarInt(uint64(len(m.Invs)))
	for _, v := range m.Invs {
		v.Write(h)
	}
}

func NewMsgNotFound() *MsgNotFound {
	return &MsgNotFound{}
}

//

type MsgBlock struct {
	Ver       uint32
	Prev      HashID
	Merkle    HashID //Merkle tree root
	Timestamp uint32
	Bits      uint32
	Nonce     uint32
	Txs       []*TX
}

func (m *MsgBlock) Command() string {
	return NMT_BLOCK
}

func (m *MsgBlock) HashID() HashID {
	buf := NewMsgBuffer([]byte{}, MSG_BUFFER_WRITE)
	buf.WriteUInt32(m.Ver)
	buf.WriteBytes(m.Prev[:])
	buf.WriteBytes(m.Merkle[:])
	buf.WriteUInt32(m.Timestamp)
	buf.WriteUInt32(m.Bits)
	buf.WriteUInt32(m.Nonce)
	hid := util.HASH256(buf.Bytes())
	id := HashID{}
	copy(id[:], hid)
	return id
}

func (m *MsgBlock) Read(h *NetHeader) {
	m.Ver = h.ReadUInt32()
	h.ReadBytes(m.Prev[:])
	h.ReadBytes(m.Merkle[:])
	m.Timestamp = h.ReadUInt32()
	m.Bits = h.ReadUInt32()
	m.Nonce = h.ReadUInt32()
	l, _ := h.ReadVarInt()
	m.Txs = make([]*TX, l)
	for i, _ := range m.Txs {
		v := &TX{}
		v.Read(h)
		m.Txs[i] = v
	}
}

func (m *MsgBlock) Write(h *NetHeader) {
	h.WriteUInt32(m.Ver)
	h.WriteBytes(m.Prev[:])
	h.WriteBytes(m.Merkle[:])
	h.WriteUInt32(m.Timestamp)
	h.WriteUInt32(m.Bits)
	h.WriteUInt32(m.Nonce)
	h.WriteVarInt(uint64(len(m.Txs)))
	for _, v := range m.Txs {
		v.Write(h)
	}
}

func NewMsgBlock() *MsgBlock {
	return &MsgBlock{}
}

//
type MsgGetData struct {
	Invs []*Inventory
}

func (m *MsgGetData) Command() string {
	return NMT_GETDATA
}

func (m *MsgGetData) Read(h *NetHeader) {
	num, _ := h.ReadVarInt()
	m.Invs = make([]*Inventory, num)
	for i, _ := range m.Invs {
		v := &Inventory{}
		v.Read(h)
		m.Invs[i] = v
	}
}
func (m *MsgGetData) Add(inv *Inventory) {
	m.Invs = append(m.Invs, inv)
}

func (m *MsgGetData) Write(h *NetHeader) {
	h.WriteVarInt(uint64(len(m.Invs)))
	for _, v := range m.Invs {
		v.Write(h)
	}
}

func NewMsgGetData() *MsgGetData {
	return &MsgGetData{
		Invs: []*Inventory{},
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
	h.WriteVarInt(uint64(len(m.Invs)))
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
	Blocks []*HashID
	Stop   *HashID
}

func (m *MsgGetHeaders) Command() string {
	return NMT_GETHEADERS
}

func (m *MsgGetHeaders) Read(h *NetHeader) {
	m.Ver = h.ReadUInt32()
	num, _ := h.ReadVarInt()
	m.Blocks = make([]*HashID, num)
	for i, _ := range m.Blocks {
		m.Blocks[i] = &HashID{}
		h.ReadBytes(m.Blocks[i][:])
	}
	m.Stop = &HashID{}
	h.ReadBytes(m.Stop[:])
}

func (m *MsgGetHeaders) Write(h *NetHeader) {
	h.WriteUInt32(m.Ver)
	h.WriteVarInt(uint64(len(m.Blocks)))
	for _, v := range m.Blocks {
		h.WriteBytes(v[:])
	}
	h.WriteBytes(m.Stop[:])
}

func NewMsgGetHeaders() *MsgGetHeaders {
	return &MsgGetHeaders{}
}
