package net

import (
	"bitcoin/db"
	"bitcoin/script"
	"bitcoin/util"
	"bytes"
	"errors"
	"fmt"
)

type p2wpkhVerify struct {
	idx int    //current ints index
	in  *TxIn  //current in
	out *TxOut //in's out
	ctx *TX    //currenct tx'clone
	typ TXType //tx type
}

func (vfy *p2wpkhVerify) Packer(sig *script.SigValue) SigPacker {
	return &witnessPacker{
		idx: vfy.idx,
		in:  vfy.in,
		out: vfy.out,
		ctx: vfy.ctx,
		ht:  uint32(sig.HashType),
		typ: vfy.typ,
	}
}

//from witeness script get
func (vfy *p2wpkhVerify) getSigInfo() (*script.SigValue, *script.PublicKey, error) {
	return nil, nil, errors.New("not imp")
}

func (vfy *p2wpkhVerify) getPubHash() ([]byte, error) {

	return nil, errors.New("not imp")
}

func (vfy *p2wpkhVerify) Verify(db db.DbImp) error {
	sig, pub, err := vfy.getSigInfo()
	if err != nil {
		return err
	}
	//check public hash
	phv, err := vfy.getPubHash()
	if err != nil {
		return err
	}
	chv := util.HASH160(pub.Marshal())
	if !bytes.Equal(chv, phv) {
		return errors.New("verify error,public hash error")
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
