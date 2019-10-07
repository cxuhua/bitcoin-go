package net

import (
	"bitcoin/db"
	"bitcoin/script"
	"bitcoin/util"
	"errors"
	"fmt"
)

type p2pkVerify struct {
	idx int    //current ints index
	in  *TxIn  //current in
	out *TxOut //in's out
	ctx *TX    //currenct tx'clone
	typ TXType //tx type
}

func (vfy *p2pkVerify) getSigInfo() (*script.SigValue, error) {
	r := NewMsgReader(*vfy.in.Script)
	sl := r.ReadUint8()
	if sl > 0x80 {
		return nil, errors.New("sig length error")
	}
	sd := make([]byte, sl)
	if _, err := r.Read(sd); err != nil {
		return nil, err
	}
	return script.NewSigValue(sd)
}

func (vfy *p2pkVerify) getPubKey() (*script.PublicKey, error) {
	r := NewMsgReader(*vfy.out.Script)
	l := r.ReadUint8()
	if l != 65 && l != 33 {
		return nil, errors.New("pubkey length error")
	}
	sd := make([]byte, l)
	if _, err := r.Read(sd); err != nil {
		return nil, err
	}
	if e := r.ReadUint8(); e != script.OP_CHECKSIG {
		return nil, errors.New("end op error")
	}
	return script.NewPublicKey(sd)
}

func (vfy *p2pkVerify) Packer(sig *script.SigValue) SigPacker {
	return &baseSigPacker{
		idx: vfy.idx,
		in:  vfy.in,
		out: vfy.out,
		ctx: vfy.ctx,
		ht:  sig.HashType,
	}
}

func (vfy *p2pkVerify) Verify(db db.DbImp) error {
	//ctx need clone
	sig, err := vfy.getSigInfo()
	if err != nil {
		return err
	}
	//check public hash
	pub, err := vfy.getPubKey()
	if err != nil {
		return err
	}
	//pack sig verify data
	data, err := vfy.Packer(sig).Pack(db)
	if err != nil {
		return fmt.Errorf("packer hash sig data error %v", err)
	}
	//sig checker
	hv := util.HASH256(data)
	if !pub.Verify(hv, sig) {
		return ErrSigVerify
	}
	return nil
}
