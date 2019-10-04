package net

import (
	"bitcoin/script"
	"bitcoin/util"
	"errors"
)

var (
	ErrSigVerify = errors.New("check sig verify error")
	ErrLockTime  = errors.New("check locktime error")
	ErrSequence  = errors.New("check sequence error")
)

type baseTXSigChecker struct {
	idx int //cur index ins index
	tx  *TX
}

func (sc *baseTXSigChecker) CheckSig(sigdata []byte, pubdata []byte, ls *script.Script, sigver script.SigVersion) error {
	sig, err := script.NewSigValue(sigdata)
	if err != nil {
		return err
	}
	pub, err := script.NewPublicKey(pubdata)
	if err != nil {
		return err
	}
	tx := sc.tx.Clone()
	for i, v := range tx.Ins {
		if i == sc.idx {
			v.Script = ls
		} else {
			v.Script = nil
		}
	}
	h := NewNetHeader()
	tx.Write(h)
	h.WriteUInt32(uint32(sig.HashType))
	hash := util.HASH256(h.Bytes())
	if ok := pub.Verify(hash, sig); !ok {
		return ErrSigVerify
	}
	return nil
}

func (sc *baseTXSigChecker) CheckLockTime(ltime script.ScriptNum) error {
	if !((sc.tx.LockTime < script.LOCKTIME_THRESHOLD && ltime < script.LOCKTIME_THRESHOLD) ||
		(sc.tx.LockTime >= script.LOCKTIME_THRESHOLD && ltime >= script.LOCKTIME_THRESHOLD)) {
		return ErrLockTime
	}
	if ltime > script.ScriptNum(sc.tx.LockTime) {
		return ErrLockTime
	}
	if sc.tx.Ins[sc.idx].Sequence == script.SEQUENCE_FINAL {
		return ErrLockTime
	}
	return nil
}

func (sc *baseTXSigChecker) CheckSequence(seq script.ScriptNum) error {
	txseq := int64(sc.tx.Ins[sc.idx].Sequence)
	if sc.tx.Ver < 2 {
		return ErrSequence
	}
	if txseq&int64(script.SEQUENCE_LOCKTIME_DISABLE_FLAG) != 0 {
		return ErrSequence
	}
	timeMask := script.SEQUENCE_LOCKTIME_TYPE_FLAG | script.SEQUENCE_LOCKTIME_MASK
	txseqMask := txseq & int64(timeMask)
	seqMask := int64(seq) & int64(timeMask)
	if !((txseqMask < int64(script.SEQUENCE_LOCKTIME_TYPE_FLAG) && seqMask < int64(script.SEQUENCE_LOCKTIME_TYPE_FLAG)) ||
		(txseqMask >= int64(script.SEQUENCE_LOCKTIME_TYPE_FLAG) && seqMask >= int64(script.SEQUENCE_LOCKTIME_TYPE_FLAG))) {
		return ErrSequence
	}
	if seqMask > txseqMask {
		return ErrSequence
	}
	return nil
}

func NewTxSigChecker(tx *TX, i int) script.SigChecker {
	return &baseTXSigChecker{tx: tx, idx: i}
}
