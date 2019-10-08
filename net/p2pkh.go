package net

import (
	"bitcoin/db"
	"bitcoin/script"
	"bitcoin/util"
	"errors"
	"fmt"
)

type p2pkhVerify struct {
	baseVerify
}

func (vfy *p2pkhVerify) Packer(sig *script.SigValue) SigPacker {
	return &baseSigPacker{
		idx: vfy.idx,
		in:  vfy.in,
		out: vfy.out,
		ctx: vfy.ctx,
		ht:  uint32(sig.HashType),
		typ: vfy.typ,
	}
}

func (vfy *p2pkhVerify) SigScript() *script.Script {
	return nil
}

func (vfy *p2pkhVerify) CheckSig(stack *script.Stack, sigv []byte, pubv []byte) error {
	sig, err := script.NewSigValue(sigv)
	if err != nil {
		return err
	}
	pub, err := script.NewPublicKey(pubv)
	if err != nil {
		return err
	}
	data, err := vfy.Packer(sig).Pack(vfy)
	if err != nil {
		return fmt.Errorf("packer hash sig data error %v", err)
	}
	hash := util.HASH256(data)
	if !pub.Verify(hash, sig) {
		return ErrSigVerify
	}
	return nil
}

func (vfy *p2pkhVerify) Verify(db db.DbImp) error {
	stack := script.NewStack()
	sv := script.NewScript([]byte{})
	sv = sv.Concat(vfy.in.Script)
	sv = sv.Concat(vfy.out.Script)
	if err := sv.Eval(stack, vfy); err != nil {
		return err
	}
	if !script.StackTopBool(stack, -1) {
		return errors.New("verify error,stack top false")
	}
	return nil
}
