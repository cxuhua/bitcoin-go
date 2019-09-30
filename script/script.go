package script

import (
	"bitcoin/util"
	"bytes"
	"encoding/hex"
	"errors"
	"log"
)

const (
	MAX_SCRIPT_ELEMENT_SIZE  = 520
	MAX_OPS_PER_SCRIPT       = 201
	MAX_PUBKEYS_PER_MULTISIG = 20
	MAX_SCRIPT_SIZE          = 10000
	MAX_STACK_SIZE           = 1000
	LOCKTIME_THRESHOLD       = 500000000
	LOCKTIME_MAX             = uint32(0xFFFFFFFF)
)

type OpCodeType byte

const (
	OP_0         = 0x00
	OP_FALSE     = OP_0
	OP_PUSHDATA1 = 0x4c
	OP_PUSHDATA2 = 0x4d
	OP_PUSHDATA4 = 0x4e
	OP_1NEGATE   = 0x4f
	OP_RESERVED  = 0x50
	OP_1         = 0x51
	OP_TRUE      = OP_1
	OP_2         = 0x52
	OP_3         = 0x53
	OP_4         = 0x54
	OP_5         = 0x55
	OP_6         = 0x56
	OP_7         = 0x57
	OP_8         = 0x58
	OP_9         = 0x59
	OP_10        = 0x5a
	OP_11        = 0x5b
	OP_12        = 0x5c
	OP_13        = 0x5d
	OP_14        = 0x5e
	OP_15        = 0x5f
	OP_16        = 0x60

	// control
	OP_NOP      = 0x61
	OP_VER      = 0x62
	OP_IF       = 0x63
	OP_NOTIF    = 0x64
	OP_VERIF    = 0x65
	OP_VERNOTIF = 0x66
	OP_ELSE     = 0x67
	OP_ENDIF    = 0x68
	OP_VERIFY   = 0x69
	OP_RETURN   = 0x6a

	// stack ops
	OP_TOALTSTACK   = 0x6b
	OP_FROMALTSTACK = 0x6c
	OP_2DROP        = 0x6d
	OP_2DUP         = 0x6e
	OP_3DUP         = 0x6f
	OP_2OVER        = 0x70
	OP_2ROT         = 0x71
	OP_2SWAP        = 0x72
	OP_IFDUP        = 0x73
	OP_DEPTH        = 0x74
	OP_DROP         = 0x75
	OP_DUP          = 0x76
	OP_NIP          = 0x77
	OP_OVER         = 0x78
	OP_PICK         = 0x79
	OP_ROLL         = 0x7a
	OP_ROT          = 0x7b
	OP_SWAP         = 0x7c
	OP_TUCK         = 0x7d

	// splice ops
	OP_CAT    = 0x7e
	OP_SUBSTR = 0x7f
	OP_LEFT   = 0x80
	OP_RIGHT  = 0x81
	OP_SIZE   = 0x82

	// bit logic
	OP_INVERT      = 0x83
	OP_AND         = 0x84
	OP_OR          = 0x85
	OP_XOR         = 0x86
	OP_EQUAL       = 0x87
	OP_EQUALVERIFY = 0x88
	OP_RESERVED1   = 0x89
	OP_RESERVED2   = 0x8a

	// numeric
	OP_1ADD      = 0x8b
	OP_1SUB      = 0x8c
	OP_2MUL      = 0x8d
	OP_2DIV      = 0x8e
	OP_NEGATE    = 0x8f
	OP_ABS       = 0x90
	OP_NOT       = 0x91
	OP_0NOTEQUAL = 0x92

	OP_ADD    = 0x93
	OP_SUB    = 0x94
	OP_MUL    = 0x95
	OP_DIV    = 0x96
	OP_MOD    = 0x97
	OP_LSHIFT = 0x98
	OP_RSHIFT = 0x99

	OP_BOOLAND            = 0x9a
	OP_BOOLOR             = 0x9b
	OP_NUMEQUAL           = 0x9c
	OP_NUMEQUALVERIFY     = 0x9d
	OP_NUMNOTEQUAL        = 0x9e
	OP_LESSTHAN           = 0x9f
	OP_GREATERTHAN        = 0xa0
	OP_LESSTHANOREQUAL    = 0xa1
	OP_GREATERTHANOREQUAL = 0xa2
	OP_MIN                = 0xa3
	OP_MAX                = 0xa4

	OP_WITHIN = 0xa5

	// crypto
	OP_RIPEMD160           = 0xa6
	OP_SHA1                = 0xa7
	OP_SHA256              = 0xa8
	OP_HASH160             = 0xa9
	OP_HASH256             = 0xaa
	OP_CODESEPARATOR       = 0xab
	OP_CHECKSIG            = 0xac
	OP_CHECKSIGVERIFY      = 0xad
	OP_CHECKMULTISIG       = 0xae
	OP_CHECKMULTISIGVERIFY = 0xaf

	// expansion
	OP_NOP1                = 0xb0
	OP_CHECKLOCKTIMEVERIFY = 0xb1
	OP_NOP2                = OP_CHECKLOCKTIMEVERIFY
	OP_CHECKSEQUENCEVERIFY = 0xb2
	OP_NOP3                = OP_CHECKSEQUENCEVERIFY
	OP_NOP4                = 0xb3
	OP_NOP5                = 0xb4
	OP_NOP6                = 0xb5
	OP_NOP7                = 0xb6
	OP_NOP8                = 0xb7
	OP_NOP9                = 0xb8
	OP_NOP10               = 0xb9

	OP_INVALIDOPCODE = 0xff
)

const (
	MAX_OPCODE = OP_NOP10
)

const (
	SIG_VER_BASE                       = 0
	SIG_VER_WITNESS_V0                 = 1
	SIG_VER_WITNESS_V0_SCRIPTHASH_SIZE = 32
	SIG_VER_WITNESS_V0_KEYHASH_SIZE    = 20
)

var (
	OpCodeErr = errors.New("op code error")
)

func DecodeOP(opcode byte) int {
	if opcode == OP_0 {
		return 0
	}
	if opcode < OP_1 || opcode > OP_16 {
		panic(OpCodeErr)
	}
	return int(opcode) - int(OP_1-1)
}

func EncodeOP(n int) byte {
	if n < 0 || n > 16 {
		panic(OpCodeErr)
	}
	if n == 0 {
		return OP_0
	}
	return byte(OP_1 + n - 1)
}

//disable op
func OpIsDisabled(op byte) bool {
	return op == OP_CAT ||
		op == OP_SUBSTR ||
		op == OP_LEFT ||
		op == OP_RIGHT ||
		op == OP_INVERT ||
		op == OP_AND ||
		op == OP_OR ||
		op == OP_XOR ||
		op == OP_2MUL ||
		op == OP_2DIV ||
		op == OP_MUL ||
		op == OP_DIV ||
		op == OP_MOD ||
		op == OP_LSHIFT ||
		op == OP_RSHIFT
}

type Script []byte

func NewScript(b []byte) *Script {
	ret := Script(b)
	return &ret
}

func NewScriptHex(s string) *Script {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return NewScript(b)
}

func (s Script) Len() int {
	return len(s)
}

func (s Script) Bytes() []byte {
	return s
}

const (
	N_ZERO = ScriptNum(0)
	N_ONE  = ScriptNum(1)
)

const (
	SIGHASH_ALL          = 1
	SIGHASH_NONE         = 2
	SIGHASH_SINGLE       = 3
	SIGHASH_ANYONECANPAY = 0x80
)

const (
	SCRIPT_VERIFY_NONE                                  = 0
	SCRIPT_VERIFY_P2SH                                  = uint32(1) << 0
	SCRIPT_VERIFY_STRICTENC                             = uint32(1) << 1
	SCRIPT_VERIFY_DERSIG                                = uint32(1) << 2
	SCRIPT_VERIFY_LOW_S                                 = uint32(1) << 3
	SCRIPT_VERIFY_NULLDUMMY                             = uint32(1) << 4
	SCRIPT_VERIFY_SIGPUSHONLY                           = uint32(1) << 5
	SCRIPT_VERIFY_MINIMALDATA                           = uint32(1) << 6
	SCRIPT_VERIFY_DISCOURAGE_UPGRADABLE_NOPS            = uint32(1) << 7
	SCRIPT_VERIFY_CLEANSTACK                            = uint32(1) << 8
	SCRIPT_VERIFY_CHECKLOCKTIMEVERIFY                   = uint32(1) << 9
	SCRIPT_VERIFY_CHECKSEQUENCEVERIFY                   = uint32(1) << 10
	SCRIPT_VERIFY_WITNESS                               = uint32(1) << 11
	SCRIPT_VERIFY_DISCOURAGE_UPGRADABLE_WITNESS_PROGRAM = uint32(1) << 12
	SCRIPT_VERIFY_MINIMALIF                             = uint32(1) << 13
	SCRIPT_VERIFY_NULLFAIL                              = uint32(1) << 14
	SCRIPT_VERIFY_WITNESS_PUBKEYTYPE                    = uint32(1) << 15
	SCRIPT_VERIFY_CONST_SCRIPTCODE                      = uint32(1) << 16
)

func (s Script) SubScript(b, e int) *Script {
	sub := s[b:e]
	return NewScript(sub)
}

//stack []byte
func (s Script) Eval(stack *Stack, checker SigChecker, flags uint32, sigver SigVersion) (bool, error) {
	vchFalse := []byte{0}
	vchTrue := []byte{1, 1}
	pc, pe := 0, s.Len()
	pb := pc
	vfExec := NewStack() //bool list
	blf := func(v Value) bool {
		return v.ToBool() == false
	}
	alts := NewStack()
	opc := 0
	fbMini := (flags & SCRIPT_VERIFY_MINIMALDATA) != 0
	for pc < pe {
		fexec := vfExec.Count(blf) == 0
		ok, idx, op, ops := s.GetOp(pc)
		if !ok {
			return false, SCRIPT_ERR_BAD_OPCODE
		}
		opc++
		pc = idx
		if op > OP_16 && opc > MAX_OPS_PER_SCRIPT {
			return false, SCRIPT_ERR_OP_COUNT
		}
		if OpIsDisabled(op) {
			return false, SCRIPT_ERR_DISABLED_OPCODE
		}
		if op == OP_CODESEPARATOR && sigver == SIG_VER_BASE && flags&SCRIPT_VERIFY_CONST_SCRIPTCODE != 0 {
			return false, SCRIPT_ERR_OP_CODESEPARATOR
		}
		if fexec && 0 <= op && op <= OP_PUSHDATA4 {
			if fbMini && !CheckMinimalPush(ops, op) {
				return false, SCRIPT_ERR_MINIMALDATA
			}
			log.Println(string(ops), op, hex.EncodeToString(ops))
			stack.Push(ops)
		} else if fexec || (OP_IF <= op && op <= OP_ENDIF) {
			switch op {
			case OP_1NEGATE, OP_1, OP_2, OP_3, OP_4, OP_5, OP_6, OP_7, OP_8, OP_9, OP_10, OP_11, OP_12, OP_13, OP_14, OP_15, OP_16:
				num := ScriptNum(int(op) - int(OP_1-1))
				stack.Push(num.Serialize())
			case OP_NOP:
				break
			case OP_CHECKLOCKTIMEVERIFY:
				if flags&SCRIPT_VERIFY_CHECKLOCKTIMEVERIFY == 0 {
					break
				}
				t1 := stack.Top(-1)
				if t1 == nil {
					return false, SCRIPT_ERR_INVALID_STACK_OPERATION
				}
				locktime := t1.ToScriptNum(fbMini, 5)
				if locktime < 0 {
					return false, SCRIPT_ERR_NEGATIVE_LOCKTIME
				}
				if !checker.CheckLockTime(locktime) {
					return false, SCRIPT_ERR_UNSATISFIED_LOCKTIME
				}
			case OP_CHECKSEQUENCEVERIFY:
				if flags&SCRIPT_VERIFY_CHECKSEQUENCEVERIFY == 0 {
					break
				}
				t1 := stack.Top(-1)
				if t1 == nil {
					return false, SCRIPT_ERR_INVALID_STACK_OPERATION
				}
				seq := t1.ToScriptNum(fbMini, 5)
				if seq < 0 {
					return false, SCRIPT_ERR_NEGATIVE_LOCKTIME
				}
				if uint32(seq)&SEQUENCE_LOCKTIME_DISABLE_FLAG != 0 {
					break
				}
				if !checker.CheckSequence(seq) {
					return false, SCRIPT_ERR_UNSATISFIED_LOCKTIME
				}
			case OP_NOP1, OP_NOP4, OP_NOP5, OP_NOP6, OP_NOP7, OP_NOP8, OP_NOP9, OP_NOP10:
				if flags&SCRIPT_VERIFY_DISCOURAGE_UPGRADABLE_NOPS != 0 {
					return false, SCRIPT_ERR_DISCOURAGE_UPGRADABLE_NOPS
				}
			case OP_IF, OP_NOTIF:
				fValue := false
				if fexec {
					if stack.Len() < 1 {
						return false, SCRIPT_ERR_UNBALANCED_CONDITIONAL
					}
					vch := stack.Top(-1)
					if sigver == SIG_VER_WITNESS_V0 && flags&SCRIPT_VERIFY_MINIMALIF != 0 {
						if vch.Len() > 1 {
							return false, SCRIPT_ERR_MINIMALIF
						}
						if vch.Len() == 1 && vch[0] != 1 {
							return false, SCRIPT_ERR_MINIMALIF
						}
					}
					fValue = vch.ToBool()
					if op == OP_NOTIF {
						fValue = !fValue
					}
					stack.Pop()
				}
				vfExec.Push(NewValueBool(fValue))
			case OP_ELSE:
				if vfExec.Empty() {
					return false, SCRIPT_ERR_UNBALANCED_CONDITIONAL
				}
				e := vfExec.Back()
				e.Value = !e.Value.(bool)
			case OP_ENDIF:
				if vfExec.Empty() {
					return false, SCRIPT_ERR_UNBALANCED_CONDITIONAL
				}
				vfExec.Pop()
			case OP_VERIFY:
				if stack.Len() < 1 {
					return false, SCRIPT_ERR_INVALID_STACK_OPERATION
				}
				fValue := stack.Top(-1).ToBool()
				if fValue {
					stack.Pop()
				} else {
					return false, SCRIPT_ERR_VERIFY
				}
			case OP_RETURN:
				return false, SCRIPT_ERR_OP_RETURN
			case OP_TOALTSTACK:
				if stack.Len() < 1 {
					return false, SCRIPT_ERR_INVALID_STACK_OPERATION
				}
				alts.Push(stack.Pop())
			case OP_FROMALTSTACK:
				if alts.Len() < 1 {
					return false, SCRIPT_ERR_INVALID_STACK_OPERATION
				}
				stack.Push(alts.Pop())
			case OP_2DROP:
				if stack.Len() < 2 {
					return false, SCRIPT_ERR_INVALID_STACK_OPERATION
				}
				stack.Pop()
				stack.Pop()
			case OP_2DUP:
				if stack.Len() < 2 {
					return false, SCRIPT_ERR_INVALID_STACK_OPERATION
				}
				v1 := stack.Top(-2)
				v2 := stack.Top(-1)
				stack.Push(v1)
				stack.Push(v2)
			case OP_3DUP:
				if stack.Len() < 3 {
					return false, SCRIPT_ERR_INVALID_STACK_OPERATION
				}
				v1 := stack.Top(-3)
				v2 := stack.Top(-2)
				v3 := stack.Top(-1)
				stack.Push(v1)
				stack.Push(v2)
				stack.Push(v3)
			case OP_2OVER:
				if stack.Len() < 4 {
					return false, SCRIPT_ERR_INVALID_STACK_OPERATION
				}
				v1 := stack.Top(-4)
				v2 := stack.Top(-3)
				stack.Push(v1)
				stack.Push(v2)
			case OP_2ROT:
				if stack.Len() < 6 {
					return false, SCRIPT_ERR_INVALID_STACK_OPERATION
				}
				v1 := stack.Top(-6)
				v2 := stack.Top(-5)
				stack.EraseRange(-6, -4)
				stack.Push(v1)
				stack.Push(v2)
			case OP_2SWAP:
				if stack.Len() < 4 {
					return false, SCRIPT_ERR_INVALID_STACK_OPERATION
				}
				v4 := stack.TopElement(-4)
				v3 := stack.TopElement(-3)
				v2 := stack.TopElement(-2)
				v1 := stack.TopElement(-1)
				v4.Value, v2.Value = v2.Value, v4.Value
				v3.Value, v1.Value = v1.Value, v3.Value
			case OP_IFDUP:
				if stack.Len() < 1 {
					return false, SCRIPT_ERR_INVALID_STACK_OPERATION
				}
				v1 := stack.Top(-1).ToBytes()
				if CastToBool(v1) {
					stack.Push(v1)
				}
			case OP_DEPTH:
				num := ScriptNum(stack.Len())
				stack.Push(num.Serialize())
			case OP_DROP:
				if stack.Len() < 1 {
					return false, SCRIPT_ERR_INVALID_STACK_OPERATION
				}
				stack.Pop()
			case OP_DUP:
				if stack.Len() < 1 {
					return false, SCRIPT_ERR_INVALID_STACK_OPERATION
				}
				v1 := stack.Top(-1)
				stack.Push(v1)
			case OP_NIP:
				if stack.Len() < 2 {
					return false, SCRIPT_ERR_INVALID_STACK_OPERATION
				}
				stack.EraseIndex(-2)
			case OP_OVER:
				if stack.Len() < 2 {
					return false, SCRIPT_ERR_INVALID_STACK_OPERATION
				}
				v1 := stack.Top(-2)
				stack.Push(v1)
			case OP_PICK, OP_ROLL:
				if stack.Len() < 2 {
					return false, SCRIPT_ERR_INVALID_STACK_OPERATION
				}
				v1 := stack.Top(-1).ToBytes()
				n := GetScriptNum(v1, fbMini, DEFAULT_MINI_SIZE).ToInt()
				stack.Pop()
				if n < 0 || n >= stack.Len() {
					return false, SCRIPT_ERR_INVALID_STACK_OPERATION
				}
				v2 := stack.Top(-n - 1)
				if op == OP_ROLL {
					stack.EraseIndex(-n - 1)
				}
				stack.Push(v2)
			case OP_ROT:
				if stack.Len() < 3 {
					return false, SCRIPT_ERR_INVALID_STACK_OPERATION
				}
				v3 := stack.TopElement(-3)
				v2 := stack.TopElement(-2)
				v1 := stack.TopElement(-1)
				v3.Value, v2.Value = v2.Value, v3.Value
				v2.Value, v1.Value = v1.Value, v2.Value
			case OP_SWAP:
				if stack.Len() < 2 {
					return false, SCRIPT_ERR_INVALID_STACK_OPERATION
				}
				v2 := stack.TopElement(-2)
				v1 := stack.TopElement(-1)
				v2.Value, v1.Value = v1.Value, v2.Value
			case OP_TUCK:
				if stack.Len() < 2 {
					return false, SCRIPT_ERR_INVALID_STACK_OPERATION
				}
				v1 := stack.Top(-1)
				v2 := stack.TopElement(-2)
				stack.InsertBefore(v1, v2)
			case OP_SIZE:
				if stack.Len() < 1 {
					return false, SCRIPT_ERR_INVALID_STACK_OPERATION
				}
				siz := ScriptNum(stack.Top(-1).Len())
				stack.Push(siz.Serialize())
			case OP_EQUAL, OP_EQUALVERIFY:
				if stack.Len() < 2 {
					return false, SCRIPT_ERR_INVALID_STACK_OPERATION
				}
				v1 := stack.Top(-2).ToBytes()
				v2 := stack.Top(-1).ToBytes()
				stack.Pop()
				stack.Pop()
				fEqual := bytes.Equal(v1, v2)
				if fEqual {
					stack.Push(vchTrue)
				} else {
					stack.Push(vchFalse)
				}
				if op == OP_EQUALVERIFY {
					if fEqual {
						stack.Pop()
					} else {
						return false, SCRIPT_ERR_EQUALVERIFY
					}
				}
			case OP_1ADD, OP_1SUB, OP_NEGATE, OP_ABS, OP_NOT, OP_0NOTEQUAL:
				if stack.Len() < 1 {
					return false, SCRIPT_ERR_INVALID_STACK_OPERATION
				}
				bn := stack.Top(-1).ToScriptNum(fbMini, DEFAULT_MINI_SIZE)
				switch op {
				case OP_1ADD:
					bn++
				case OP_1SUB:
					bn--
				case OP_NEGATE:
					bn = -bn
				case OP_ABS:
					if bn < 0 {
						bn = -bn
					}
				case OP_NOT:
					if bn == 0 {
						bn = 1
					} else {
						bn = 0
					}
				case OP_0NOTEQUAL:
					if bn != 0 {
						bn = 1
					} else {
						bn = 0
					}
				}
				stack.Pop()
				stack.Push(bn.Serialize())
			case OP_ADD, OP_SUB, OP_BOOLAND, OP_BOOLOR, OP_NUMEQUAL, OP_NUMEQUALVERIFY, OP_NUMNOTEQUAL, OP_LESSTHAN, OP_GREATERTHAN, OP_LESSTHANOREQUAL, OP_GREATERTHANOREQUAL, OP_MIN, OP_MAX:
				if stack.Len() < 2 {
					return false, SCRIPT_ERR_INVALID_STACK_OPERATION
				}
				bn1 := stack.Top(-2).ToScriptNum(fbMini, DEFAULT_MINI_SIZE)
				bn2 := stack.Top(-1).ToScriptNum(fbMini, DEFAULT_MINI_SIZE)
				bn := ScriptNum(0)
				switch op {
				case OP_ADD:
					bn = bn1 + bn2
				case OP_SUB:
					bn = bn1 - bn2
				case OP_BOOLAND:
					if bn1 != 0 && bn2 != 0 {
						bn = 1
					} else {
						bn = 0
					}
				case OP_BOOLOR:
					if bn1 != 0 || bn2 != 0 {
						bn = 1
					} else {
						bn = 0
					}
				case OP_NUMEQUAL:
					if bn1 == bn2 {
						bn = 1
					} else {
						bn = 0
					}
				case OP_NUMEQUALVERIFY:
					if bn1 == bn2 {
						bn = 1
					} else {
						bn = 0
					}
				case OP_NUMNOTEQUAL:
					if bn1 != bn2 {
						bn = 1
					} else {
						bn = 0
					}
				case OP_LESSTHAN:
					if bn1 < bn2 {
						bn = 1
					} else {
						bn = 0
					}
				case OP_GREATERTHAN:
					if bn1 > bn2 {
						bn = 1
					} else {
						bn = 0
					}
				case OP_LESSTHANOREQUAL:
					if bn1 <= bn2 {
						bn = 1
					} else {
						bn = 0
					}
				case OP_GREATERTHANOREQUAL:
					if bn1 >= bn2 {
						bn = 1
					} else {
						bn = 0
					}
				case OP_MIN:
					if bn1 < bn2 {
						bn = bn1
					} else {
						bn = bn2
					}
				case OP_MAX:
					if bn1 > bn2 {
						bn = bn1
					} else {
						bn = bn2
					}
				}
				stack.Pop()
				stack.Pop()
				stack.Push(bn.Serialize())
				if op == OP_NUMEQUALVERIFY {
					v1 := stack.Top(-1).ToBool()
					if v1 {
						stack.Pop()
					} else {
						return false, SCRIPT_ERR_NUMEQUALVERIFY
					}
				}
			case OP_WITHIN:
				if stack.Len() < 3 {
					return false, SCRIPT_ERR_INVALID_STACK_OPERATION
				}
				bn1 := stack.Top(-3).ToScriptNum(fbMini, DEFAULT_MINI_SIZE)
				bn2 := stack.Top(-2).ToScriptNum(fbMini, DEFAULT_MINI_SIZE)
				bn3 := stack.Top(-1).ToScriptNum(fbMini, DEFAULT_MINI_SIZE)
				fvalue := (bn2 <= bn1 && bn1 < bn3)
				stack.Pop()
				stack.Pop()
				stack.Pop()
				if fvalue {
					stack.Push(vchTrue)
				} else {
					stack.Push(vchFalse)
				}
			case OP_RIPEMD160, OP_SHA1, OP_SHA256, OP_HASH160, OP_HASH256:
				if stack.Len() < 1 {
					return false, SCRIPT_ERR_INVALID_STACK_OPERATION
				}
				vch1 := stack.Top(-1).ToBytes()
				var hv []byte = nil
				if op == OP_RIPEMD160 {
					hv = util.RIPEMD160(vch1)
				} else if op == OP_SHA256 {
					hv = util.SHA256(vch1)
				} else if op == OP_HASH160 {
					hv = util.HASH160(vch1)
				} else if op == OP_HASH256 {
					hv = util.HASH256(vch1)
				} else if op == OP_SHA1 {
					hv = util.SHA1(vch1)
				} else {
					return false, SCRIPT_ERR_BAD_OPCODE
				}
				stack.Pop()
				stack.Push(hv)
			case OP_CODESEPARATOR:
				pb = pc
				log.Println("OP_CODESEPARATOR", pb)
			case OP_CHECKSIG, OP_CHECKSIGVERIFY:
				if stack.Len() < 2 {
					return false, SCRIPT_ERR_INVALID_STACK_OPERATION
				}
				sig := stack.Top(-2).ToBytes()
				pk := stack.Top(-1).ToBytes()
				sub := s.SubScript(pb, pe)
				if sigver == SIG_VER_BASE {
					found := FindAndDelete(sub, NewScript(sig))
					if found > 0 && flags&SCRIPT_VERIFY_CONST_SCRIPTCODE != 0 {
						return false, SCRIPT_ERR_SIG_FINDANDDELETE
					}
				}
				if err := CheckSignatureEncoding(sig, flags); err != nil {
					return false, err
				}
				if err := CheckPubKeyEncoding(pk, flags, sigver); err != nil {
					return false, err
				}
				success := checker.CheckSig(sig, pk, sub, sigver)
				if !success && flags&SCRIPT_VERIFY_NULLFAIL != 0 && len(sig) > 0 {
					return false, SCRIPT_ERR_SIG_NULLFAIL
				}
				stack.Pop()
				stack.Pop()
				if op == OP_CHECKSIGVERIFY {
					if success {
						stack.Pop()
					} else {
						return false, SCRIPT_ERR_CHECKSIGVERIFY
					}
				}
			case OP_CHECKMULTISIG, OP_CHECKMULTISIGVERIFY:
				i := 1
				if stack.Len() < i {
					return false, SCRIPT_ERR_INVALID_STACK_OPERATION
				}
				keyscount := stack.Top(-1).ToInt(fbMini, DEFAULT_MINI_SIZE)
				if keyscount < 0 || keyscount > MAX_PUBKEYS_PER_MULTISIG {
					return false, SCRIPT_ERR_PUBKEY_COUNT
				}
				opc += keyscount
				if opc > MAX_OPS_PER_SCRIPT {
					return false, SCRIPT_ERR_OP_COUNT
				}
				i++
				ikey1 := i
				ikey2 := keyscount + 2
				i += keyscount
				if stack.Len() < i {
					return false, SCRIPT_ERR_INVALID_STACK_OPERATION
				}
				sigcount := stack.Top(-1).ToInt(fbMini, DEFAULT_MINI_SIZE)
				if sigcount < 0 || sigcount > keyscount {
					return false, SCRIPT_ERR_SIG_COUNT
				}
				i++
				isig := i
				i += sigcount
				if stack.Len() < i {
					return false, SCRIPT_ERR_INVALID_STACK_OPERATION
				}
				sub := s.SubScript(pb, pe)
				for k := 0; k < sigcount; k++ {
					sig := stack.Top(-isig - k).ToBytes()
					if sigver == SIG_VER_BASE {
						found := FindAndDelete(sub, NewScript(sig))
						if found > 0 && flags&SCRIPT_VERIFY_CONST_SCRIPTCODE != 0 {
							return false, SCRIPT_ERR_SIG_FINDANDDELETE
						}
					}
				}
				success := true
				for success && sigcount > 0 {
					sig := stack.Top(-isig).ToBytes()
					pk := stack.Top(-ikey1).ToBytes()
					if err := CheckSignatureEncoding(sig, flags); err != nil {
						return false, err
					}
					if err := CheckPubKeyEncoding(pk, flags, sigver); err != nil {
						return false, err
					}
					fok := checker.CheckSig(sig, pk, sub, sigver)
					if fok {
						isig++
						sigcount++
					}
					ikey1++
					keyscount++
					if sigcount > keyscount {
						success = false
					}
				}
				for i > 1 {
					v1 := stack.Top(-1).ToBytes()
					if !success && (flags&SCRIPT_VERIFY_NULLFAIL) != 0 && ikey2 != 0 && len(v1) > 0 {
						return false, SCRIPT_ERR_SIG_NULLFAIL
					}
					if ikey2 > 0 {
						ikey2--
					}
					stack.Pop()
					i--
				}
				if stack.Len() < 1 {
					return false, SCRIPT_ERR_INVALID_STACK_OPERATION
				}
				v1 := stack.Top(-1).ToBytes()
				if (flags&SCRIPT_VERIFY_NULLDUMMY) != 0 && len(v1) > 0 {
					return false, SCRIPT_ERR_SIG_NULLDUMMY
				}
				stack.Pop()
				if success {
					stack.Push(vchTrue)
				} else {
					stack.Push(vchFalse)
				}
				if op == OP_CHECKMULTISIGVERIFY {
					if success {
						stack.Pop()
					} else {
						return false, SCRIPT_ERR_CHECKMULTISIGVERIFY
					}
				}
			default:
				return false, SCRIPT_ERR_BAD_OPCODE
			}
			if stack.Len()+alts.Len() > MAX_STACK_SIZE {
				return false, SCRIPT_ERR_STACK_SIZE
			}
		}
	}
	if !vfExec.Empty() {
		return false, SCRIPT_ERR_UNBALANCED_CONDITIONAL
	}
	return true, nil
}
