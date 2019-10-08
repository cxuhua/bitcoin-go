package net

import (
	"bitcoin/db"
	"bitcoin/script"
	"errors"
	"log"
)

type p2wshOnlyVerify struct {
	baseVerify
}

func (vfy *p2wshOnlyVerify) Packer(sig *script.SigValue) SigPacker {
	return &witnessPacker{
		idx: vfy.idx,
		in:  vfy.in,
		out: vfy.out,
		ctx: vfy.ctx,
		ht:  uint32(sig.HashType),
		typ: vfy.typ,
	}
}

func (vfy *p2wshOnlyVerify) SigScript() *script.Script {
	return nil
}

func (vfy *p2wshOnlyVerify) CheckSig(stack *script.Stack, sigv []byte, pubv []byte) error {

	return nil
}

func (vfy *p2wshOnlyVerify) Verify(db db.DbImp) error {
	stack := script.NewStack()
	sv := script.NewScript([]byte{})
	for _, v := range vfy.in.Witness.Script {
		if v.Len() == 0 {
			continue
		}
		if n1, n2 := v.IsMultiSig(); n1 > 0 && n2 > 0 {
			log.Printf("MULTISUG %d of %d", n1, n2)
			sv.Concat(v)
		} else {
			sv.PushBytes(*v)
		}
	}
	if err := sv.Eval(stack, vfy); err != nil {
		return err
	}
	if !script.StackTopBool(stack, -1) {
		return errors.New("verify error")
	}
	return nil
}
