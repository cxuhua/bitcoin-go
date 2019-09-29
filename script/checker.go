package script

const (
	SEQUENCE_FINAL                 = uint32(0xffffffff)
	SEQUENCE_LOCKTIME_DISABLE_FLAG = uint32(1 << 31)
	SEQUENCE_LOCKTIME_TYPE_FLAG    = uint32(1 << 22)
	SEQUENCE_LOCKTIME_MASK         = uint32(0x0000ffff)
	SEQUENCE_LOCKTIME_GRANULARITY  = 9
)

const (
	PUBLIC_KEY_SIZE            = 65
	COMPRESSED_PUBLIC_KEY_SIZE = 33
	SIGNATURE_SIZE             = 72
	COMPACT_SIGNATURE_SIZE     = 65
)

type SigVersion uint

type SigChecker interface {
	CheckSig(sig []byte, pubkey []byte, script *Script, sigver SigVersion) bool
	CheckLockTime(num ScriptNum) bool
	CheckSequence(num ScriptNum) bool
}
