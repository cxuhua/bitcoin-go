package core

import (
	"bitcoin/script"
	"bitcoin/util"
	"bytes"
	"errors"
	"fmt"
)

type p2wshMSIGVerify struct {
	hsidx int
	baseVerify
}

func newP2WSHMSIGVerify(idx int, in *TxIn, out *TxOut, ctx *TX, typ TxType) *p2wshMSIGVerify {
	return &p2wshMSIGVerify{
		hsidx: -1,
		baseVerify: baseVerify{
			idx: idx,
			in:  in,
			out: out,
			ctx: ctx,
			typ: typ,
		},
	}
}

func (vfy *p2wshMSIGVerify) Packer(sig *script.SigValue) SigPacker {
	return &witnesSigPacker{
		idx: vfy.idx,
		in:  vfy.in,
		out: vfy.out,
		ctx: vfy.ctx,
		ht:  uint32(sig.HashType),
		typ: vfy.typ,
	}
}

func (vfy *p2wshMSIGVerify) SigScript() *script.Script {
	return vfy.in.Witness.Script[vfy.hsidx]
}

func (vfy *p2wshMSIGVerify) CheckSig(stack *script.Stack, sigv []byte, pubv []byte) error {
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

func (vfy *p2wshMSIGVerify) checkPublicHash() bool {
	sc := vfy.in.Witness.Script[vfy.hsidx]
	hv1 := util.SHA256(*sc)
	hv2 := (*vfy.out.Script)[2:]
	return bytes.Equal(hv1, hv2)
}

func (vfy *p2wshMSIGVerify) Verify(flags int) error {
	stack := script.NewStack()
	sv := script.NewScript([]byte{})
	vfy.hsidx = -1
	for i, v := range vfy.in.Witness.Script {
		if v.Len() == 0 {
			continue
		} else if v.HasMultiSig() {
			vfy.hsidx = i
			sv.Concat(v)
		} else {
			sv.PushBytes(*v)
		}
	}
	if vfy.hsidx < 0 || !vfy.checkPublicHash() {
		return errors.New("check public hash error")
	}
	if err := sv.Eval(stack, vfy, flags); err != nil {
		return err
	}
	if !script.StackTopBool(stack, -1) {
		return errors.New("verify error")
	}
	return nil
}
