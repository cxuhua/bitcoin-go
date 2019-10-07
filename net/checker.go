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
	ctx *TX //cur verify ctx
	ptx *TX //cur in ref ctx
}

func (sc *baseTXSigChecker) GetTxIn() *TxIn {
	return sc.ctx.Ins[sc.idx]
}

func (sc *baseTXSigChecker) GetTxOut() *TxOut {
	cin := sc.GetTxIn()
	return sc.ptx.Outs[cin.OutIndex]
}

/*
	nVersion:     01000000
    hashPrevouts: b0287b4a252ac05af83d2dcef00ba313af78a3e9c329afa216eb3aa2a7b4613a
    hashSequence: 18606b350cd8bf565266bc352f0caddcf01e8fa789dd8a15386327cf8cabe198
    outpoint:     db6b1b20aa0fd7b23880be2ecbd4a98130974cf4748fb66092ac4d3ceb1a547701000000
    scriptCode:   1976a91479091972186c449eb1ded22b78e40d009bdf008988ac
    amount:       00ca9a3b00000000
    nSequence:    feffffff
    hashOutputs:  de984f44532e2173ca0d64314fcefe6d30da6f8cf27bafa706da61df8a226c83
    nLockTime:    92040000
    nHashType:    01000000
*/
func (sc *baseTXSigChecker) CheckSig(sigdata []byte, pubdata []byte, sigver script.SigVersion) error {
	sig, err := script.NewSigValue(sigdata)
	if err != nil {
		return err
	}
	pub, err := script.NewPublicKey(pubdata)
	if err != nil {
		return err
	}
	pto := sc.GetTxOut()
	h := NewNetHeader()
	if sigver == script.SIG_VER_BASE {
		tx := sc.ctx.Clone()
		for i, v := range tx.Ins {
			if i == sc.idx {
				v.Script = pto.Script
			} else {
				v.Script = nil
			}
		}
		tx.WriteSig(h, sig.HashType, sigver)
		h.WriteUInt32(uint32(sig.HashType))
	} else if sigver == script.SIG_VER_WITNESS_V0 {
		tx := sc.ctx
		cin := sc.GetTxIn()
		h.WriteUInt32(uint32(tx.Ver))
		h.WriteBytes(tx.GetPrevoutHash(sc.idx, sig.HashType))
		h.WriteBytes(tx.GetSequenceHash(sc.idx, sig.HashType))
		h.WriteBytes(cin.OutHash[:])
		h.WriteUInt32(cin.OutIndex)
		h.WriteScript(cin.Script.GetP2PKHScript())
		h.WriteUInt64(pto.Value)
		h.WriteUInt32(cin.Sequence)
		h.WriteBytes(tx.GetOutputsHash(sc.idx, sig.HashType))
		h.WriteUInt32(tx.LockTime)
		h.WriteUInt32(uint32(sig.HashType))
	}
	hash := util.HASH256(h.Bytes())
	if ok := pub.Verify(hash, sig); !ok {
		return ErrSigVerify
	}
	return nil
}

func (sc *baseTXSigChecker) CheckLockTime(ltime script.ScriptNum) error {
	if !((sc.ctx.LockTime < script.LOCKTIME_THRESHOLD && ltime < script.LOCKTIME_THRESHOLD) ||
		(sc.ctx.LockTime >= script.LOCKTIME_THRESHOLD && ltime >= script.LOCKTIME_THRESHOLD)) {
		return ErrLockTime
	}
	if ltime > script.ScriptNum(sc.ctx.LockTime) {
		return ErrLockTime
	}
	if sc.ctx.Ins[sc.idx].Sequence == script.SEQUENCE_FINAL {
		return ErrLockTime
	}
	return nil
}

func (sc *baseTXSigChecker) CheckSequence(seq script.ScriptNum) error {
	txseq := int64(sc.ctx.Ins[sc.idx].Sequence)
	if sc.ctx.Ver < 2 {
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

func NewTxSigChecker(ctx *TX, ptx *TX, i int) script.SigChecker {
	return &baseTXSigChecker{
		ctx: ctx,
		ptx: ptx,
		idx: i,
	}
}
