package script

import (
	"bitcoin/util"
	"encoding/binary"
	"errors"
)

const (
	INT_MAX           = int(^uint(0) >> 1)
	INT_MIN           = ^INT_MAX
	DEFAULT_MINI_SIZE = 4
)

type ScriptNum int64

func (v ScriptNum) ToInt() int {
	if int64(v) > int64(INT_MAX) {
		return INT_MAX
	} else if int64(v) < int64(INT_MIN) {
		return INT_MIN
	}
	return int(v)
}

func CastToBool(vch []byte) bool {
	for i := 0; i < len(vch); i++ {
		if vch[i] != 0 {
			if i == len(vch)-1 && vch[i] == 0x80 {
				return false
			}
			return true
		}
	}
	return false
}

func GetScriptNum(b []byte, mini bool, max int) ScriptNum {
	bl := len(b)
	if bl > max {
		panic(errors.New("script number overflow"))
	}
	if mini && bl > 0 {
		lv := b[bl-1]
		if lv&0x7f == 0 {
			if bl <= 1 || (b[bl-2]&0x80) == 0 {
				panic(errors.New("non-minimally encoded script number"))
			}
		}
	}
	if bl == 0 {
		return ScriptNum(0)
	}
	result := int64(0)
	for i := 0; i != bl; i++ {
		result |= int64(b[i]) << int64(8*i)
	}
	if b[bl-1]&0x80 != 0 {
		tmp := int64(0x80) << (8 * (bl - 1))
		return ScriptNum(-(result & (^tmp)))
	}
	return ScriptNum(result)
}

func IsValidSignatureEncoding(sig []byte) bool {
	if len(sig) < 9 {
		return false
	}
	if len(sig) > 73 {
		return false
	}
	if sig[0] != 0x30 {
		return false
	}
	if sig[1] != byte(len(sig)-3) {
		return false
	}
	lenR := sig[3]

	if 5+lenR >= byte(len(sig)) {
		return false
	}
	lenS := sig[5+lenR]
	if int(lenR)+int(lenS)+7 != len(sig) {
		return false
	}
	if sig[2] != 0x02 {
		return false
	}
	if lenR == 0 {
		return false
	}
	if sig[4]&0x80 != 0 {
		return false
	}
	if lenR > 1 && (sig[4] == 0x00) && (sig[5]&0x80 != 0) {
		return false
	}
	if sig[lenR+4] != 0x02 {
		return false
	}
	if lenS == 0 {
		return false
	}
	if sig[lenR+6]&0x80 != 0 {
		return false
	}
	if lenS > 1 && (sig[lenR+6] == 0x00) && (sig[lenR+7]&0x80) != 0 {
		return false
	}
	return true
}

func PubKeyCheckLowS(pk []byte) bool {
	return true
}

func IsLowDERSignature(sig []byte) error {
	if !IsValidSignatureEncoding(sig) {
		return SCRIPT_ERR_SIG_DER
	}
	cpy := make([]byte, len(sig))
	copy(cpy, sig)
	if !PubKeyCheckLowS(cpy) {
		return SCRIPT_ERR_SIG_HIGH_S
	}
	return nil
}

func IsDefinedHashtypeSignature(sig []byte) bool {
	if len(sig) == 0 {
		return false
	}
	htype := sig[len(sig)-1] & (^byte(SIGHASH_ANYONECANPAY))
	if htype < SIGHASH_ALL || htype > SIGHASH_SINGLE {
		return false
	}
	return true
}

func IsCompressedOrUncompressedPubKey(pb []byte) bool {
	if len(pb) < COMPRESSED_PUBLIC_KEY_SIZE {
		return false
	}
	if pb[0] == 0x04 {
		if len(pb) != PUBLIC_KEY_SIZE {
			return false
		}
	} else if pb[0] == 0x02 || pb[0] == 0x03 {
		if len(pb) != COMPRESSED_PUBLIC_KEY_SIZE {
			return false
		}
	} else {
		return false
	}
	return true
}

func IsCompressedPubKey(pb []byte) bool {
	if len(pb) != COMPRESSED_PUBLIC_KEY_SIZE {
		return false
	}
	if pb[0] != 0x02 && pb[0] != 0x03 {
		return false
	}
	return true
}

func CheckPubKeyEncoding(pb []byte, flags uint32, sigver SigVersion) error {
	if (flags&SCRIPT_VERIFY_STRICTENC) != 0 && !IsCompressedOrUncompressedPubKey(pb) {
		return SCRIPT_ERR_PUBKEYTYPE
	}
	if (flags&SCRIPT_VERIFY_WITNESS_PUBKEYTYPE) != 0 && sigver == SIG_VER_WITNESS_V0 && !IsCompressedPubKey(pb) {
		return SCRIPT_ERR_WITNESS_PUBKEYTYPE
	}
	return nil
}

func CheckSignatureEncoding(sig []byte, flags uint32) error {
	if len(sig) == 0 {
		return nil
	}
	if (flags&(SCRIPT_VERIFY_DERSIG|SCRIPT_VERIFY_LOW_S|SCRIPT_VERIFY_STRICTENC)) != 0 && !IsValidSignatureEncoding(sig) {
		return SCRIPT_ERR_SIG_DER
	}
	if (flags & SCRIPT_VERIFY_LOW_S) != 0 {
		if err := IsLowDERSignature(sig); err != nil {
			return err
		}
	}
	if (flags&SCRIPT_VERIFY_STRICTENC) != 0 && !IsDefinedHashtypeSignature(sig) {
		return SCRIPT_ERR_SIG_HASHTYPE
	}
	return nil
}

func (v ScriptNum) Serialize() []byte {
	ret := []byte{}
	if v == 0 {
		return ret
	}
	absv := uint64(0)
	neg := v < 0
	if neg {
		absv = uint64(-v)
	} else {
		absv = uint64(v)
	}
	for absv != 0 {
		ret = append(ret, byte(absv&0xFF))
		absv >>= 8
	}
	bi := len(ret) - 1
	if ret[bi]&0x80 != 0 {
		if neg {
			ret = append(ret, 0x80)
		} else {
			ret = append(ret, 0)
		}
	} else if neg {
		ret[bi] |= 0x80
	}
	return ret
}

func (s Script) HasValidOps() bool {
	b := 0
	for b < len(s) {
		ok, i, op, item := s.GetOp(b)
		if !ok || op > MAX_OPCODE || len(item) > MAX_SCRIPT_ELEMENT_SIZE {
			return false
		}
		b = i
	}
	return true
}

/*
// payToPubKeyHashScript creates a new script to pay a transaction
// output to a 20-byte pubkey hash. It is expected that the input is a valid
// hash
//P2PKH
func payToPubKeyHashScript(pubKeyHash []byte) ([]byte, error) {
   return NewScriptBuilder().AddOp(OP_DUP).AddOp(OP_HASH160).
      AddData(pubKeyHash).AddOp(OP_EQUALVERIFY).AddOp(OP_CHECKSIG).
      Script()
}

//P2WPKH
// payToWitnessPubKeyHashScript creates a new script to pay to a version 0
// pubkey hash witness program. The passed hash is expected to be valid.
func payToWitnessPubKeyHashScript(pubKeyHash []byte) ([]byte, error) {
   return NewScriptBuilder().AddOp(OP_0).AddData(pubKeyHash).Script()
}

//P2SH
// payToScriptHashScript creates a new script to pay a transaction output to a
// script hash. It is expected that the input is a valid hash.
func payToScriptHashScript(scriptHash []byte) ([]byte, error) {
   return NewScriptBuilder().AddOp(OP_HASH160).AddData(scriptHash).
      AddOp(OP_EQUAL).Script()
}

//P2WSH
func payToWitnessScriptHashScript(scriptHash []byte) ([]byte, error) {
   return NewScriptBuilder().AddOp(OP_0).AddData(scriptHash).Script()
}

//P2PK
// payToPubkeyScript creates a new script to pay a transaction output to a
// public key. It is expected that the input is a valid pubkey.
func payToPubKeyScript(serializedPubKey []byte) ([]byte, error) {
   return NewScriptBuilder().AddData(serializedPubKey).
      AddOp(OP_CHECKSIG).Script()
}

*/

func NewP2PKScript(pub *PublicKey) *Script {
	b := pub.Marshal()
	s := &Script{}
	return s.PushBytes(b).PushOp(OP_CHECKSIG)
}

//2103c9f4836b9a4f77fc0d81f7bcb01b7f1b35916864b9476c241ce9fc198bd25432ac
//or
//4104a39b9e4fbd213ef24bb9be69de4a118dd0644082e47c01fd9159d38637b83fbcdc115a5d6e970586a012d1cfe3e3a8b1a3d04e763bdc5a071c0e827c0bd834a5ac
func (s Script) IsP2PK() bool {
	return (s.Len() == 35 && s[0] == 0x21 && s[34] == OP_CHECKSIG) || (s.Len() == 67 && s[0] == 0x42 && s[66] == OP_CHECKSIG)
}

//00141d0f172a0ecb48aee1be1f2687d2963ae33f71a1
func (s Script) IsP2WPKH() bool {
	return s.Len() == 23 && s[0] == 22 && s[1] == 0 && s[2] == byte(s.Len()-3)
}

//IsP2WPKH
//public key hash to p2pkh script
func (s Script) GetP2PKHScript() *Script {
	if !s.IsP2WPKH() {
		panic(errors.New("s not IsP2WPKH"))
	}
	ns := &Script{}
	return ns.PushOp(OP_DUP).PushOp(OP_HASH160).PushBytes(s[3:]).PushOp(OP_EQUALVERIFY).PushOp(OP_CHECKSIG)
}

func NewP2PKHScript(pub *PublicKey) *Script {
	s := &Script{}
	b := pub.Marshal()
	hv := util.HASH160(b)
	return s.PushOp(OP_DUP).PushOp(OP_HASH160).PushBytes(hv).PushOp(OP_EQUALVERIFY).PushOp(OP_CHECKSIG)
}

//
func (s Script) IsP2PKH() bool {
	return s.Len() == 25 && s[0] == OP_DUP && s[1] == OP_HASH160 && s[2] == 20 && s[23] == OP_EQUALVERIFY && s[24] == OP_CHECKSIG
}

func NewP2SHScript(pub *PublicKey) *Script {
	s := &Script{}
	b := pub.Marshal()
	hv := util.HASH160(b)
	return s.PushOp(OP_HASH160).PushBytes(hv).PushOp(OP_EQUAL)
}

//a9144733f37cf4db86fbc2efed2500b4f4e49f31202387
func (s Script) IsP2SH() bool {
	return s.Len() == 23 && s[0] == OP_HASH160 && s[1] == 0x14 && s[22] == OP_EQUAL
}

//return ver programe ok
func (s Script) IsWitnessProgram() (int, []byte, bool) {
	if s.Len() < 4 || s.Len() > 42 {
		return 0, nil, false
	}
	if s[0] != OP_0 && (s[0] < OP_1 || s[0] > OP_16) {
		return 0, nil, false
	}
	if int(s[1]+2) == s.Len() {
		ver := DecodeOP(s[0])
		return ver, s[2:], true
	}
	return 0, nil, false
}

func (s Script) GetScriptSigOpCount(script *Script) int {
	if !s.IsP2SH() {
		return s.GetSigOpCount(true)
	}
	var subd []byte = nil
	pc := 0
	for pc < script.Len() {
		ok, idx, op, item := script.GetOp(pc)
		if !ok {
			return 0
		}
		if op > OP_16 {
			return 0
		}
		pc = idx
		subd = item
	}
	sub := NewScript(subd)
	return sub.GetSigOpCount(true)
}

func (s Script) GetSigOpCount(accurate bool) int {
	pc := 0
	n := 0
	lastop := byte(OP_INVALIDOPCODE)
	for pc < s.Len() {
		ok, idx, op, _ := s.GetOp(pc)
		if !ok {
			break
		}
		if op == OP_CHECKSIG || op == OP_CHECKSIGVERIFY {
			n++
		} else if op == OP_CHECKMULTISIG || op == OP_CHECKMULTISIGVERIFY {
			if accurate && lastop >= OP_1 && lastop <= OP_16 {
				n += DecodeOP(lastop)
			} else {
				n += MAX_PUBKEYS_PER_MULTISIG
			}
		}
		pc = idx
		lastop = op
	}
	return n
}

func (s Script) IsPushOnly(idx int) (int, bool) {
	for idx < s.Len() {
		ok, pos, op, _ := s.GetOp(idx)
		if !ok {
			return idx, false
		}
		if op > OP_16 {
			return idx, false
		}
		idx = pos
	}
	return idx, true
}

func (s Script) IsPayToWitnessScriptHash() bool {
	return (s.Len() == 34 && s[0] == OP_0 && s[1] == 0x20)
}

func (s Script) IsUnspendable() bool {
	return (s.Len() > 0 && s[0] == OP_RETURN) || (s.Len() > MAX_SCRIPT_SIZE)
}

func (s *Script) PushInt64(v int64) *Script {
	if v == -1 || (v >= 1 && v <= 16) {
		*s = append(*s, byte(v+(OP_1-1)))
	} else if v == 0 {
		*s = append(*s, OP_0)
	} else {
		*s = append(*s, ScriptNum(v).Serialize()...)
	}
	return s
}

func (s *Script) PushOp(v byte) *Script {
	*s = append(*s, v)
	return s
}

func (s *Script) PushBytes(b []byte) *Script {
	l := len(b)
	if l < OP_PUSHDATA1 {
		s.PushOp(byte(len(b)))
	} else if l <= 0xff {
		s.PushOp(OP_PUSHDATA1)
		s.PushOp(byte(len(b)))
	} else if l <= 0xffff {
		s.PushOp(OP_PUSHDATA2)
		b2 := []byte{0, 0}
		binary.LittleEndian.PutUint16(b2, uint16(l))
		*s = append(*s, b2...)
	} else {
		s.PushOp(OP_PUSHDATA4)
		b4 := []byte{0, 0, 0, 0}
		binary.LittleEndian.PutUint32(b4, uint32(l))
		*s = append(*s, b4...)
	}
	*s = append(*s, b...)
	return s
}

func CheckMinimalPush(ops []byte, op byte) bool {
	if op > OP_PUSHDATA4 {
		panic(OpCodeErr)
	}
	l := len(ops)
	if l == 0 {
		return op == OP_0
	} else if l == 1 && ops[0] >= 1 && ops[0] <= 16 {
		return false
	} else if l == 1 && ops[0] == 0x81 {
		return false
	} else if l <= 75 {
		return op == byte(l)
	} else if l <= 255 {
		return op == OP_PUSHDATA1
	} else if l <= 65535 {
		return op == OP_PUSHDATA2
	}
	return true
}

//return ok,idx,op,ops
func (s *Script) GetOp(b int) (bool, int, byte, []byte) {
	e := s.Len()
	ret := []byte{}
	if b >= e {
		return false, b, OP_INVALIDOPCODE, ret
	}
	if e-b < 1 {
		return false, b, OP_INVALIDOPCODE, ret
	}
	op := (*s)[b]
	b++
	if op > OP_PUSHDATA4 {
		return true, b, op, ret
	}
	size := uint(0)
	if op < OP_PUSHDATA1 {
		size = uint(op)
	} else if op == OP_PUSHDATA1 {
		if e-b < 1 {
			return false, b, op, ret
		}
		op = (*s)[b]
		b++
		size = uint(op)
	} else if op == OP_PUSHDATA2 {
		if e-b < 2 {
			return false, b, op, ret
		}
		size = uint(binary.LittleEndian.Uint16((*s)[b:]))
		b += 2
	} else if op == OP_PUSHDATA4 {
		if e-b < 4 {
			return false, b, op, ret
		}
		size = uint(binary.LittleEndian.Uint32((*s)[b:]))
		b += 4
	}
	if e-b < 0 || uint(e-b) < size {
		return false, b, op, ret
	}
	ret = (*s)[b : b+int(size)]
	b += int(size)
	return true, b, op, ret
}
