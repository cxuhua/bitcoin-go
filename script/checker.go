package script

const (
	SEQUENCE_FINAL                 = uint32(0xffffffff)
	SEQUENCE_LOCKTIME_DISABLE_FLAG = uint32(1 << 31)
	SEQUENCE_LOCKTIME_TYPE_FLAG    = uint32(1 << 22)
	SEQUENCE_LOCKTIME_MASK         = uint32(0x0000ffff)
	SEQUENCE_LOCKTIME_GRANULARITY  = 9
)

type SigVersion uint

type SigChecker interface {
	CheckSig(sig []byte, pubkey []byte, script *Script, sigver SigVersion) bool
	CheckLockTime(num ScriptNum) bool
	CheckSequence(num ScriptNum) bool
}

type baseSigChecker struct {
}

func (sc *baseSigChecker) CheckSig(sig []byte, pubkey []byte, script *Script, sigver SigVersion) bool {
	panic("Not Imp")
}

func (sc *baseSigChecker) CheckLockTime(num ScriptNum) bool {
	panic("Not Imp")
}

func (sc *baseSigChecker) CheckSequence(num ScriptNum) bool {
	panic("Not Imp")
}

func NewBaseSigChecker() SigChecker {
	return &baseSigChecker{}
}
