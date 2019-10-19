package script

import (
	"bitcoin/util"
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

func IsDefinedHashtypeSignature(sig []byte) bool {
	if len(sig) == 0 {
		return false
	}
	ht := sig[len(sig)-1] &^ SIGHASH_ANYONECANPAY
	if ht < SIGHASH_ALL || ht > SIGHASH_SINGLE {
		return false
	}
	return true
}

func IsLowDERSignature(sig []byte) error {
	if !IsValidSignatureEncoding(sig) {
		return SCRIPT_ERR_SIG_DER
	}
	nsig := sig[:len(sig)-1]
	_, _, err := CheckLowS(nsig)
	return err
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

func CheckSignatureEncoding(sig []byte, flags int) error {
	if len(sig) == 0 {
		return nil
	}
	if flags&(SCRIPT_VERIFY_DERSIG|SCRIPT_VERIFY_LOW_S|SCRIPT_VERIFY_STRICTENC) != 0 && !IsValidSignatureEncoding(sig) {
		return SCRIPT_ERR_SIG_DER
	} else if err := IsLowDERSignature(sig); flags&SCRIPT_VERIFY_LOW_S != 0 && err != nil {
		return err
	} else if flags&SCRIPT_VERIFY_STRICTENC != 0 && !IsDefinedHashtypeSignature(sig) {
		return SCRIPT_ERR_SIG_HASHTYPE
	}
	return nil
}

func IsCompressedPubKey(pub []byte) bool {
	if len(pub) != COMPRESSED_PUBLIC_KEY_SIZE {
		return false
	}
	if pub[0] != 0x02 && pub[0] != 0x03 {
		return false
	}
	return true
}

func CheckPubKeyEncoding(pub []byte, flags int) error {
	if flags&SCRIPT_VERIFY_STRICTENC != 0 && !IsCompressedOrUncompressedPubKey(pub) {
		return SCRIPT_ERR_PUBKEYTYPE
	}
	if flags&SCRIPT_VERIFY_WITNESS_PUBKEYTYPE != 0 && flags&SCRIPT_WITNESS_V0_PUBKEYTYPE != 0 && !IsCompressedPubKey(pub) {
		return SCRIPT_ERR_WITNESS_PUBKEYTYPE
	}
	return nil
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

func (s Script) HasValidOps() bool {
	for i := 0; i < s.Len(); {
		b, p, op, ops := s.GetOp(i)
		if !b {
			return false
		}
		if op > MAX_OPCODE {
			return false
		}
		if len(ops) > MAX_SCRIPT_ELEMENT_SIZE {
			return false
		}
		i = p
	}
	return true
}

func (s Script) IsPushOnly() bool {
	for i := 0; i < s.Len(); {
		b, p, op, _ := s.GetOp(i)
		if !b {
			return false
		}
		if op > OP_16 {
			return false
		}
		i = p
	}
	return true
}

func (s Script) GetSigOpCount() int {
	n := 0
	for i := 0; i < s.Len(); {
		b, p, op, _ := s.GetOp(i)
		if !b {
			break
		}
		if op >= OP_CHECKSIG || op == OP_CHECKSIGVERIFY {
			n++
		} else if op == OP_CHECKMULTISIG || op == OP_CHECKMULTISIGVERIFY {
			n += MAX_PUBKEYS_PER_MULTISIG
		}
		i = p
	}
	return n
}

func (s Script) GetAddress() string {
	var ab []byte
	if s.IsP2PK(&ab) || s.IsP2PKH(&ab) {
		return util.P2PKHAddress(ab)
	}
	if s.IsP2SH(&ab) {
		return util.P2SHAddress(ab)
	}
	if s.IsP2WSH(&ab) {
		return util.BECH32Address(ab)
	}
	return ""
}

func (s Script) IsNull() bool {
	return s.Len() >= 1 && s[0] == OP_RETURN && NewScript(s[1:]).IsPushOnly()
}

func (s Script) IsP2PK(v ...*[]byte) bool {
	b := s.Len() == COMPRESSED_PUBLIC_KEY_SIZE+2 && s[0] == COMPRESSED_PUBLIC_KEY_SIZE && s[34] == OP_CHECKSIG
	if b {
		if len(v) > 0 {
			*v[0] = s.SubBytes(1, 33)
		}
		return b
	}
	b = s.Len() == PUBLIC_KEY_SIZE+2 && s[0] == PUBLIC_KEY_SIZE && s[66] == OP_CHECKSIG
	if b {
		if len(v) > 0 {
			*v[0] = s.SubBytes(1, 66)
		}
		return b
	}
	return b
}

//out || in
func (s Script) IsP2WPKH(v ...*[]byte) bool {
	b := s.Len() == 22 && s[0] == OP_0 && s[1] == byte(s.Len()-2)
	if b {
		if len(v) > 0 {
			*v[0] = s.SubBytes(2, 22)
		}
		return b
	}
	b = s.Len() == 23 && s[0] == 0x16 && s[1] == OP_0 && s[2] == byte(s.Len()-3)
	if b {
		if len(v) > 0 {
			*v[0] = s.SubBytes(2, 22)
		}
		return b
	}
	return b
}

//for out
func (s Script) IsP2PKH(v ...*[]byte) bool {
	b := s.Len() == 25 && s[0] == OP_DUP && s[1] == OP_HASH160 && s[2] == 20 && s[23] == OP_EQUALVERIFY && s[24] == OP_CHECKSIG
	if b {
		if len(v) > 0 {
			*v[0] = s.SubBytes(3, 23)
		}
		return b
	}
	//parse error,try
	b = s.Len() >= 25 && s.Len() < MAX_SCRIPT_SIZE && s[0] == OP_DUP && s[1] == OP_HASH160 && s[2] == 20 && s[23] == OP_EQUALVERIFY && s[24] == OP_CHECKSIG
	if b {
		if len(v) > 0 {
			*v[0] = s.SubBytes(3, 23)
		}
		return b
	}
	return b
}

func (s Script) IsP2SH(v ...*[]byte) bool {
	b := s.Len() == 23 && s[0] == OP_HASH160 && s[1] == 0x14 && s[22] == OP_EQUAL
	if b {
		if len(v) > 0 {
			*v[0] = s.SubBytes(2, 22)
		}
	}
	return b
}

//out || in
func (s Script) IsP2WSH(v ...*[]byte) bool {
	b := s.Len() == 34 && s[0] == OP_0 && s[1] == 0x20
	if b {
		if len(v) > 0 {
			*v[0] = s.SubBytes(2, 34)
		}
		return b
	}
	b = s.Len() == 35 && s[0] == 34 && s[1] == OP_0 && s[2] == 0x20
	if b {
		if len(v) > 0 {
			*v[0] = s.SubBytes(3, 35)
		}
		return b
	}
	return b
}

//return lessnum,pubnum
func (s Script) HasMultiSig() bool {
	if s.Len() == 0 || s[s.Len()-1] != OP_CHECKMULTISIG {
		return false
	}
	lnum, pnum, knum := 0, 0, 0
	for i := 0; i < s.Len(); {
		b, p, op, ops := s.GetOp(i)
		if !b {
			break
		}
		if op >= OP_PUSHDATA1 && op <= OP_PUSHDATA4 && Script(ops).HasMultiSig() {
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

var (
	border = binary.LittleEndian
)

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
		border.PutUint16(b2, uint16(l))
		*s = append(*s, b2...)
	} else {
		s.PushOp(OP_PUSHDATA4)
		b4 := []byte{0, 0, 0, 0}
		border.PutUint32(b4, uint32(l))
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
		size = uint(border.Uint16((*s)[b:]))
		b += 2
	} else if op == OP_PUSHDATA4 {
		if e-b < 4 {
			return false, b, op, ret
		}
		size = uint(border.Uint32((*s)[b:]))
		b += 4
	}
	if e-b < 0 || uint(e-b) < size {
		return false, b, op, ret
	}
	ret = (*s)[b : b+int(size)]
	b += int(size)
	return true, b, op, ret
}
