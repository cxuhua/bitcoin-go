package script

import (
	"encoding/hex"
	"errors"
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

var (
	SCRIPT_ERR_BAD_OPCODE             = errors.New("SCRIPT_ERR_BAD_OPCODE")
	SCRIPT_ERR_OP_COUNT               = errors.New("SCRIPT_ERR_OP_COUNT")
	SCRIPT_ERR_DISABLED_OPCODE        = errors.New("SCRIPT_ERR_DISABLED_OPCODE")
	SCRIPT_ERR_MINIMALDATA            = errors.New("SCRIPT_ERR_MINIMALDATA")
	SCRIPT_ERR_STACK_SIZE             = errors.New("SCRIPT_ERR_STACK_SIZE")
	SCRIPT_ERR_UNBALANCED_CONDITIONAL = errors.New("SCRIPT_ERR_UNBALANCED_CONDITIONAL")
)

//stack []byte
func (s Script) Eval(stack *Stack, checker SigChecker, flags uint32, ver uint) (bool, error) {
	b, e := 0, s.Len()
	bl := NewStack() //bool list
	blf := func(v interface{}) bool {
		b, ok := v.(bool)
		return ok && !b
	}
	alts := NewStack()
	opc := 0
	fbMini := (flags & SCRIPT_VERIFY_MINIMALDATA) != 0
	for b < e {
		fexec := bl.Count(blf) > 0
		ok, idx, op, ops := s.GetOp(b)
		if !ok {
			return false, SCRIPT_ERR_BAD_OPCODE
		}
		opc++
		if op > OP_16 && opc > MAX_OPS_PER_SCRIPT {
			return false, SCRIPT_ERR_OP_COUNT
		}
		if OpIsDisabled(op) {
			return false, SCRIPT_ERR_DISABLED_OPCODE
		}
		if fexec && 0 <= op && op <= OP_PUSHDATA4 {
			if fbMini && !CheckMinimalPush(ops, op) {
				return false, SCRIPT_ERR_MINIMALDATA
			}
			stack.Push(ops)
		} else if fexec || (OP_IF <= op && op <= OP_ENDIF) {
			switch op {
			case OP_1NEGATE:
				fallthrough
			case OP_1:
				fallthrough
			case OP_2:
				fallthrough
			case OP_3:
				fallthrough
			case OP_4:
				fallthrough
			case OP_5:
				fallthrough
			case OP_6:
				fallthrough
			case OP_7:
				fallthrough
			case OP_8:
				fallthrough
			case OP_9:
				fallthrough
			case OP_10:
				fallthrough
			case OP_11:
				fallthrough
			case OP_12:
				fallthrough
			case OP_13:
				fallthrough
			case OP_14:
				fallthrough
			case OP_15:
				fallthrough
			case OP_16:
				{

				}
				break
			}
			if stack.Len()+alts.Len() > MAX_STACK_SIZE {
				return false, SCRIPT_ERR_STACK_SIZE
			}
		}
		b = idx
	}
	if !bl.Empty() {
		return false, SCRIPT_ERR_UNBALANCED_CONDITIONAL
	}
	return true, nil
}
