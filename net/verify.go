package net

import (
	"bitcoin/db"
	"bitcoin/script"
	"errors"
	"fmt"
)

var (
	ErrSigVerify = errors.New("sig veify error")
	ErrLockTime  = errors.New("locktime error")
	ErrSequence  = errors.New("sequence error")
)

type Verifyer interface {
	//check sig
	Verify(db db.DbImp) error
	//sig data packer
	Packer(sig *script.SigValue) SigPacker
}

type baseVerify struct {
	idx int    //current ints index
	in  *TxIn  //current in
	out *TxOut //in's out
	ctx *TX    //currenct tx'clone
	typ TXType //tx type
}

func (vfy *baseVerify) CheckLockTime(ltime script.ScriptNum) error {
	if !((vfy.ctx.LockTime < script.LOCKTIME_THRESHOLD && ltime < script.LOCKTIME_THRESHOLD) ||
		(vfy.ctx.LockTime >= script.LOCKTIME_THRESHOLD && ltime >= script.LOCKTIME_THRESHOLD)) {
		return ErrLockTime
	}
	if ltime > script.ScriptNum(vfy.ctx.LockTime) {
		return ErrLockTime
	}
	if vfy.ctx.Ins[vfy.idx].Sequence == script.SEQUENCE_FINAL {
		return ErrLockTime
	}
	return nil
}

func (vfy *baseVerify) CheckSequence(seq script.ScriptNum) error {
	txseq := int64(vfy.ctx.Ins[vfy.idx].Sequence)
	if vfy.ctx.Ver < 2 {
		return ErrSequence
	}
	if txseq&int64(script.SEQUENCE_LOCKTIME_DISABLE_FLAG) != 0 {
		return ErrSequence
	}
	timeMask := script.SEQUENCE_LOCKTIME_TYPE_FLAG | script.SEQUENCE_LOCKTIME_MASK
	txseqMask := txseq & int64(timeMask)
	seqMask := int64(seq) & int64(timeMask)
	if !((txseqMask < int64(script.SEQUENCE_LOCKTIME_TYPE_FLAG) && seqMask < int64(script.SEQUENCE_LOCKTIME_TYPE_FLAG)) ||
		(txseqMask >= int64(script.SEQUENCE_LOCKTIME_TYPE_FLAG) && seqMask >= int64(script.SEQUENCE_LOCKTIME_TYPE_FLAG))) {
		return ErrSequence
	}
	if seqMask > txseqMask {
		return ErrSequence
	}
	return nil
}

type TXType int

//tx type
const (
	TX_NONE TXType = iota
	TX_P2PK
	TX_P2PKH
	TX_P2SH_WPKH
	TX_P2SH_WSH
	TX_ONLY_P2WSH
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
		return TX_P2SH_WPKH
	}
	if out.Script.IsP2SH() && in.Script.IsP2WSH() {
		return TX_P2SH_WSH
	}
	if out.Script.IsP2WSH() && in.Script.Len() == 0 {
		return TX_ONLY_P2WSH
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
		case TX_P2PKH, TX_P2PK:
			verifyer = &p2pkhVerify{
				baseVerify: baseVerify{
					idx: idx,
					in:  in,
					out: out,
					ctx: tx,
					typ: typ,
				},
			}
		case TX_P2SH_WPKH:
			verifyer = &p2wpkhVerify{
				baseVerify: baseVerify{
					idx: idx,
					in:  in,
					out: out,
					ctx: tx,
					typ: typ,
				},
			}
		case TX_ONLY_P2WSH:
			verifyer = &p2wshOnlyVerify{
				hsidx: -1,
				less:  -1,
				size:  -1,
				baseVerify: baseVerify{
					idx: idx,
					in:  in,
					out: out,
					ctx: tx,
					typ: typ,
				},
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
