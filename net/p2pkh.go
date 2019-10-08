package net

import (
	"bitcoin/db"
	"bitcoin/script"
	"bitcoin/util"
	"bytes"
	"errors"
	"fmt"
)

//out script format
//OP_DUP OP_HASH160 67327d111aa5bedde016f51a10acd4e850eb8c83 OP_EQUALVERIFY OP_CHECKSIG
//76a914 67327d111aa5bedde016f51a10acd4e850eb8c83 88 ac

//in script format
//sig:47 304402207f242d4a011d530231234bf30c8af14adc1e173c79be4a7e608ec2d1f6ea91cf0220239d11453406e486ade491a733395215cdc2c25343f99d0f5914489d02247496 01
//pub:21 03320ce8bccb8a30819860e72b326188f95f855a713bc88720cb0ca74efdbc2c87

type p2pkhVerify struct {
	idx int    //current ints index
	in  *TxIn  //current in
	out *TxOut //in's out
	ctx *TX    //currenct tx'clone
	typ TXType //tx type
}

func (vfy *p2pkhVerify) getSigInfo() (*script.SigValue, *script.PublicKey, error) {
	r := NewMsgReader(*vfy.in.Script)
	sl := r.ReadUint8()
	if sl > 0x80 {
		return nil, nil, errors.New("sig length error")
	}
	sd := make([]byte, sl)
	if _, err := r.Read(sd); err != nil {
		return nil, nil, err
	}
	sig, err := script.NewSigValue(sd)
	if err != nil {
		return nil, nil, err
	}
	pl := r.ReadUint8()
	if pl > 0x80 {
		return nil, nil, errors.New("pubkey length error")
	}
	pd := make([]byte, pl)
	if _, err := r.Read(pd); err != nil {
		return nil, nil, err
	}
	pub, err := script.NewPublicKey(pd)
	if err != nil {
		return nil, nil, err
	}
	return sig, pub, nil
}

func (vfy *p2pkhVerify) getPubHash() ([]byte, error) {
	r := NewMsgReader(*vfy.out.Script)
	h := []byte{0, 0}
	r.Read(h)
	if h[0] != script.OP_DUP && h[1] != script.OP_HASH160 {
		return nil, errors.New("head 2byte error")
	}
	l := r.ReadUint8()
	if l > 0x80 {
		return nil, errors.New("length byte error")
	}
	dat := make([]byte, l)
	if _, err := r.Read(dat); err != nil {
		return nil, err
	}
	e := []byte{0, 0}
	r.Read(e)
	if e[0] != script.OP_EQUALVERIFY && e[1] != script.OP_CHECKSIG {
		return nil, errors.New("end 2byte error")
	}
	return dat, nil
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

func (vfy *p2pkhVerify) Verify(db db.DbImp) error {
	//ctx need clone
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
