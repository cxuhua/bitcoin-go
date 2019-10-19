package core

import (
	"bitcoin/script"
	"bitcoin/util"
	"errors"
	"fmt"
)

type p2shMSIGVerify struct {
	sigs     []*script.Script
	pkscript *script.Script
	baseVerify
}

func newP2SHMSIGVerify(idx int, in *TxIn, out *TxOut, ctx *TX, typ TXType) *p2shMSIGVerify {
	return &p2shMSIGVerify{
		sigs:     []*script.Script{},
		pkscript: nil,
		baseVerify: baseVerify{
			idx: idx,
			in:  in,
			out: out,
			ctx: ctx,
			typ: typ,
		},
	}
}

func (vfy *p2shMSIGVerify) Packer(sig *script.SigValue) SigPacker {
	return &baseSigPacker{
		idx: vfy.idx,
		in:  vfy.in,
		out: vfy.out,
		ctx: vfy.ctx,
		ht:  uint32(sig.HashType),
		typ: vfy.typ,
	}
}

func (vfy *p2shMSIGVerify) SigScript() *script.Script {
	return vfy.pkscript
}

func (vfy *p2shMSIGVerify) CheckSig(stack *script.Stack, sigv []byte, pubv []byte) error {
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

func (vfy *p2shMSIGVerify) Verify(flags int) error {
	stack := script.NewStack()
	if vfy.in.Script == nil {
		return errors.New("in script error nil")
	}
	//find pubkey and sigs
	for i := 0; i < vfy.in.Script.Len(); {
		b, p, op, ops := vfy.in.Script.GetOp(i)
		if !b {
			break
		}
		if pks := script.NewScript(ops); pks.HasMultiSig() {
			vfy.pkscript = pks
		} else if script.IsValidSignatureEncoding(ops) {
			vfy.sigs = append(vfy.sigs, script.NewScript(ops))
		} else if op == script.OP_CHECKMULTISIG {
			break
		}
		i = p
	}
	if vfy.pkscript == nil || len(vfy.sigs) == 0 {
		return errors.New("pubkey script or sigs miss")
	}
	//check hash equal
	sv := script.NewScript([]byte{})
	sv = sv.PushBytes(*vfy.pkscript)
	sv = sv.Concat(vfy.out.Script)
	if err := sv.Eval(stack, vfy, flags); err != nil {
		return err
	}
	if !script.StackTopBool(stack, -1) {
		return errors.New("has cmp failed")
	}
	sv.Clean()
	stack.Pop()
	//check multisigs
	for _, v := range vfy.sigs {
		sv.PushBytes(*v)
	}
	sv.Concat(vfy.pkscript)
	if err := sv.Eval(stack, vfy, flags); err != nil {
		return err
	}
	if !script.StackTopBool(stack, -1) {
		return errors.New("verify error")
	}
	return nil
}
