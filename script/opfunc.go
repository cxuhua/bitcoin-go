package script

import (
	"encoding/binary"
)

type ScriptNum int64

func GetScriptNum(b []byte) ScriptNum {
	bl := len(b)
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
