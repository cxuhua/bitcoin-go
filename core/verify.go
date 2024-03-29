package core

import (
	"bitcoin/script"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

var (
	ErrSigVerify = errors.New("sig veify error")
	ErrLockTime  = errors.New("locktime error")
	ErrSequence  = errors.New("sequence error")
)

type Verifyer interface {
	//check sig
	Verify(flags int) error
	//sig data packer
	Packer(sig *script.SigValue) SigPacker
}

type baseVerify struct {
	idx int    //current ints index
	in  *TxIn  //current in
	out *TxOut //in's out
	ctx *TX    //currenct tx'clone
	typ TxType //tx type
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

type TxType int

//tx type
const (
	TX_UNKNOW TxType = iota
	TX_NULL_DATA
	TX_NONSTANDARD
	TX_P2PK
	TX_P2PKH
	TX_P2SH_WPKH
	TX_P2SH_WSH
	TX_P2WSH_MSIG
	TX_P2SH_MSIG
	TX_P2WPKH
)

func (i *TxIn) OnlyHasWitness() bool {
	return len(*i.Script) == 0 && i.Witness != nil && len(i.Witness.Script) > 0
}

func (i *TxIn) HasWitnessMultiSig() bool {
	if i.Witness == nil || len(i.Witness.Script) == 0 {
		return false
	}
	for _, v := range i.Witness.Script {
		if v != nil && v.HasMultiSig() {
			return true
		}
	}
	return false
}

func (i *TxIn) HasScriptMultiSig() bool {
	if i.Script == nil || i.Script.Len() == 0 {
		return false
	}
	return i.Script.HasMultiSig()
}

//in input data,out=in's out
func CheckTXType(in *TxIn, out *TxOut) TxType {
	if in == nil || out == nil || out.Script == nil {
		return TX_UNKNOW
	}
	if out.Script.IsP2PK() {
		return TX_P2PK
	}
	if out.Script.IsP2PKH() {
		return TX_P2PKH
	}
	if out.Script.IsP2WPKH() && in.OnlyHasWitness() {
		return TX_P2WPKH
	}
	if out.Script.IsNull() {
		return TX_NULL_DATA
	}
	if out.Script.IsP2SH() && in.Script.IsP2WPKH() {
		return TX_P2SH_WPKH
	}
	if out.Script.IsP2SH() && in.Script.IsP2WSH() {
		return TX_P2SH_WSH
	}
	if out.Script.IsP2WSH() && in.HasWitnessMultiSig() {
		return TX_P2WSH_MSIG
	}
	if out.Script.IsP2SH() && in.HasScriptMultiSig() {
		return TX_P2SH_MSIG
	}
	return TX_NONSTANDARD
}

func VerifyTX(tx *TX, flags int) error {
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
		out, err := in.OutTx()
		if err != nil {
			return fmt.Errorf("load ref out error %w", err)
		}
		typ := CheckTXType(in, out)
		if typ == TX_UNKNOW {
			return fmt.Errorf("in %d checktype not support tx=%v", idx, tx.Hash)
		}
		var verifyer Verifyer
		switch typ {
		case TX_NULL_DATA:
			continue
		case TX_NONSTANDARD:
			h := NewNetHeader()
			tx.Write(h)
			ioutil.WriteFile(tx.Hash.String(), h.Bytes(), os.ModePerm)
			continue
		case TX_P2PKH, TX_P2PK:
			verifyer = newP2PKHVerify(idx, in, out, tx, typ)
		case TX_P2WPKH:
			verifyer = newP2WPKHVerify(idx, in, out, tx, typ)
		case TX_P2SH_WPKH:
			verifyer = newP2SHWPKHVerify(idx, in, out, tx, typ)
		case TX_P2WSH_MSIG:
			flags |= script.SCRIPT_WITNESS_V0_PUBKEYTYPE
			verifyer = newP2WSHMSIGVerify(idx, in, out, tx, typ)
		case TX_P2SH_MSIG:
			verifyer = newP2SHMSIGVerify(idx, in, out, tx, typ)
		case TX_P2SH_WSH:
			flags |= script.SCRIPT_WITNESS_V0_PUBKEYTYPE
			verifyer = newP2SHWSHVerify(idx, in, out, tx, typ)
		default:
			return fmt.Errorf("in %d checktype not support,miss Verifyer", idx)
		}
		if err := verifyer.Verify(flags); err != nil {
			return fmt.Errorf("Verifyer in %d error %v", idx, err)
		}
	}
	return nil
}
