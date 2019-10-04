package net

import (
	"bitcoin/script"
	"bitcoin/util"
	"errors"
)

var (
	ErrSigVerify = errors.New("check sig verify error")
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
	panic("Not Imp")
}

func (sc *baseTXSigChecker) CheckSequence(seq script.ScriptNum) error {
	panic("Not Imp")
}

func NewTxSigChecker(tx *TX, i int) script.SigChecker {
	return &baseTXSigChecker{tx: tx, idx: i}
}
