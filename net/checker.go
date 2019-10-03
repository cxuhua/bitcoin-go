package net

import "bitcoin/script"

type baseSigChecker struct {
	tx *TX
}

func (sc *baseSigChecker) CheckSig(sig []byte, pubkey []byte, script *script.Script, sigver script.SigVersion) bool {
	panic("Not Imp")
}

func (sc *baseSigChecker) CheckLockTime(num script.ScriptNum) bool {
	panic("Not Imp")
}

func (sc *baseSigChecker) CheckSequence(num script.ScriptNum) bool {
	panic("Not Imp")
}

func NewBaseSigChecker(tx *TX) script.SigChecker {
	return &baseSigChecker{tx: tx}
}
