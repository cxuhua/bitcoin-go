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
	CheckSig(sigdata []byte, pubdata []byte, sigver SigVersion) error
	CheckLockTime(ltime ScriptNum) error
	CheckSequence(seq ScriptNum) error
}
