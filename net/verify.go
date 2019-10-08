package net

import (
	"bitcoin/db"
	"bitcoin/script"
	"errors"
	"fmt"
)

type Verifyer interface {
	Verify(db db.DbImp) error
	Packer(sig *script.SigValue) SigPacker
}

type TXType int

//tx type
const (
	TX_NONE TXType = iota
	TX_P2PK
	TX_P2PKH
	TX_P2WPKH
)

//in input data,out=in's out
func CheckTXType(in *TxIn, out *TxOut) TXType {
	if in == nil || out == nil || out.Script == nil {
		return TX_NONE
	}
	if out.Script.IsP2PK() {
		return TX_P2PK
	}
	if out.Script.IsP2PKH() {
		return TX_P2PKH
	}
	if out.Script.IsP2SH() && in.Script.IsP2WPKH() {
		return TX_P2WPKH
	}
	return TX_NONE
}

func VerifyTX(tx *TX, db db.DbImp) error {
	if tx == nil {
		return errors.New("args nil")
	}
	if err := tx.Check(); err != nil {
		return err
	}
	if tx.IsCoinBase() {
		return nil
	}
	for idx, in := range tx.Ins {
		ptx, err := LoadTX(in.OutHash, db)
		if err != nil {
			return fmt.Errorf("load prev tx error %v", err)
		}
		if int(in.OutIndex) >= len(ptx.Outs) {
			return errors.New("out index out bound")
		}
		out := ptx.Outs[in.OutIndex]
		typ := CheckTXType(in, out)
		if typ == TX_NONE {
			return fmt.Errorf("in %d checktype not support", idx)
		}
		var verifyer Verifyer
		switch typ {
		case TX_P2PKH:
			verifyer = &p2pkhVerify{
				idx: idx,
				in:  in,
				out: out,
				ctx: tx,
				typ: typ,
			}
		case TX_P2PK:
			verifyer = &p2pkVerify{
				idx: idx,
				in:  in,
				out: out,
				ctx: tx,
				typ: typ,
			}
		case TX_P2WPKH:
			verifyer = &p2wpkhVerify{
				idx: idx,
				in:  in,
				out: out,
				ctx: tx,
				typ: typ,
			}
		default:
			return fmt.Errorf("in %d checktype not support,miss Verifyer", idx)
		}
		if err := verifyer.Verify(db); err != nil {
			return fmt.Errorf("Verifyer in %d error %v", idx, err)
		}
	}
	return nil
}
