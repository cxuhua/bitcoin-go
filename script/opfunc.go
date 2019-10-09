package script

import (
	"encoding/binary"
	"errors"
)

const (
	INT_MAX           = int(^uint(0) >> 1)
	INT_MIN           = ^INT_MAX
	DEFAULT_MINI_SIZE = 5
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

func GetScriptNum(b []byte) ScriptNum {
	bl := len(b)
	if bl > DEFAULT_MINI_SIZE {
		panic(errors.New("script number overflow"))
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
	if lenR > 1 && (sig[4] == 0x00) && (sig[5]&0x80 == 0) {
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
	if lenS > 1 && (sig[lenR+6] == 0x00) && (sig[lenR+7]&0x80) == 0 {
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

//2103c9f4836b9a4f77fc0d81f7bcb01b7f1b35916864b9476c241ce9fc198bd25432ac
//or
//4104a39b9e4fbd213ef24bb9be69de4a118dd0644082e47c01fd9159d38637b83fbcdc115a5d6e970586a012d1cfe3e3a8b1a3d04e763bdc5a071c0e827c0bd834a5ac
//for out
func (s Script) IsP2PK() bool {
	return (s.Len() == 35 && s[0] == 0x21 && s[34] == OP_CHECKSIG) || (s.Len() == 67 && s[0] == 0x41 && s[66] == OP_CHECKSIG)
}

//1600146bdeebc5218e565401db3e1c4510eebd2570cc07
//for in
func (s Script) IsP2WPKH() bool {
	return s.Len() == 23 && s[0] == 0x16 && s[1] == 0 && s[2] == byte(s.Len()-3)
}

//for out
func (s Script) IsP2PKH() bool {
	return s.Len() == 25 && s[0] == OP_DUP && s[1] == OP_HASH160 && s[2] == 20 && s[23] == OP_EQUALVERIFY && s[24] == OP_CHECKSIG
}

//a9144733f37cf4db86fbc2efed2500b4f4e49f31202387
//for out
func (s Script) IsP2SH() bool {
	return s.Len() == 23 && s[0] == OP_HASH160 && s[1] == 0x14 && s[22] == OP_EQUAL
}

//for out/in
func (s Script) IsP2WSH() bool {
	return (s.Len() == 34 && s[0] == OP_0 && s[1] == 0x20) || (s.Len() == 35 && s[0] == 34 && s[1] == OP_0 && s[2] == 0x20)
}

//return lessnum,pubnum
func (s Script) HasMultiSig() bool {
	if s.Len() == 0 || s[s.Len()-1] != OP_CHECKMULTISIG {
		return false
	}
	lnum, pnum, knum := 0, 0, 0
	for i := 0; ; {
		b, p, op, ops := s.GetOp(i)
		if !b {
			break
		}
		if op >= OP_PUSHDATA1 && op <= OP_PUSHDATA4 && NewScript(ops).HasMultiSig() {
			return true
		} else if lnum == 0 && op >= OP_1 && op <= OP_16 {
			lnum = int(op-OP_1) + 1
		} else if pnum == 0 && op >= OP_1 && op <= OP_16 {
			pnum = int(op-OP_1) + 1
		} else if IsCompressedOrUncompressedPubKey(ops) {
			knum++
		} else if op == OP_CHECKMULTISIG {
			break
		}
		i = p
	}
	if knum != pnum || lnum > pnum {
		return false
	}
	return lnum > 0 && pnum >= 3 && lnum <= pnum
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

func (s *Script) Clean() *Script {
	*s = (*s)[0:0]
	return s
}

func (s *Script) Concat(v *Script) *Script {
	*s = append(*s, *v...)
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
