package core

import (
	"bitcoin/script"
	"bitcoin/util"
	"errors"
	"fmt"
)

type p2shwpkhVerify struct {
	baseVerify
}

func newP2SHWPKHVerify(idx int, in *TxIn, out *TxOut, ctx *TX, typ TxType) *p2shwpkhVerify {
	return &p2shwpkhVerify{
		baseVerify: baseVerify{
			idx: idx,
			in:  in,
			out: out,
			ctx: ctx,
			typ: typ,
		},
	}
}

func (vfy *p2shwpkhVerify) Packer(sig *script.SigValue) SigPacker {
	return &witnesSigPacker{
		idx: vfy.idx,
		in:  vfy.in,
		out: vfy.out,
		ctx: vfy.ctx,
		ht:  uint32(sig.HashType),
		typ: vfy.typ,
	}
}

//get sigcode
func (vfy *p2shwpkhVerify) SigScript() *script.Script {
	hash := (*vfy.in.Script)[3:]
	ns := &script.Script{}
	ns = ns.PushOp(script.OP_DUP)
	ns = ns.PushOp(script.OP_HASH160)
	ns = ns.PushBytes(hash)
	ns = ns.PushOp(script.OP_EQUALVERIFY)
	ns = ns.PushOp(script.OP_CHECKSIG)
	return ns
}

//-1 pubkey
//-2 sigvalue
//-3 hash equal
func (vfy *p2shwpkhVerify) CheckSig(stack *script.Stack, sigv []byte, pubv []byte) error {
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

func (vfy *p2shwpkhVerify) Verify(flags int) error {
	stack := script.NewStack()
	sv := script.NewScript([]byte{})
	//concat hash equal script
	sv = sv.Concat(vfy.in.Script)
	sv = sv.Concat(vfy.out.Script)
	if err := sv.Eval(stack, vfy, flags); err != nil {
		return err
	}
	if !script.StackTopBool(stack, -1) {
		return errors.New("hash equal error")
	}
	stack.Pop()
	sv.Clean()
	//push sig pub data
	for _, v := range vfy.in.Witness.Script {
		sv = sv.PushBytes(*v)
	}
	//add sig check op code
	sv = sv.PushOp(script.OP_CHECKSIG)
	//run script checksig
	if err := sv.Eval(stack, vfy, flags); err != nil {
		return err
	}
	if !script.StackTopBool(stack, -1) {
		return errors.New("verify error")
	}
	return nil
}
