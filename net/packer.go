package net

import (
	"bitcoin/db"
	"bitcoin/script"
	"fmt"
)

type SigPacker interface {
	Pack(db db.DbImp) ([]byte, error)
}

type baseSigPacker struct {
	idx int    //current ints index
	in  *TxIn  //current in
	out *TxOut //in's out
	ctx *TX    //currenct tx'clone
	ht  uint32 //hash type script.SIGHASH_*
	typ TXType //tx type
}

func (sp *baseSigPacker) Pack(db db.DbImp) ([]byte, error) {
	if sp.ht != script.SIGHASH_ALL {
		return nil, fmt.Errorf("hash type %d not support imp", sp.ht)
	}
	w := NewMsgWriter()
	w.WriteInt32(sp.ctx.Ver)
	w.WriteVarInt(len(sp.ctx.Ins))
	for i, v := range sp.ctx.Ins {
		w.WriteBytes(v.OutHash[:])
		w.WriteUInt32(v.OutIndex)
		if i == sp.idx {
			w.WriteScript(sp.out.Script)
		} else {
			w.WriteScript(nil)
		}
		w.WriteUInt32(v.Sequence)
	}
	w.WriteVarInt(len(sp.ctx.Outs))
	for _, v := range sp.ctx.Outs {
		w.WriteUInt64(v.Value)
		w.WriteScript(v.Script)
	}
	w.WriteUInt32(sp.ctx.LockTime)
	w.WriteUInt32(sp.ht)
	return w.Bytes(), nil
}

type witnessPacker struct {
	idx int    //current ints index
	in  *TxIn  //current in
	out *TxOut //in's out
	ctx *TX    //currenct tx'clone
	ht  uint32 //hash type script.SIGHASH_*
	typ TXType //tx type
}

func (sp *witnessPacker) getOutputsHash() HashID {
	hash := HashID{}
	m := NewMsgWriter()
	for _, v := range sp.ctx.Outs {
		m.WriteUInt64(v.Value)
		m.WriteScript(v.Script)
	}
	HASH256To(m.Payload, &hash)
	return hash
}

func (sp *witnessPacker) getPrevoutHash() HashID {
	hash := HashID{}
	m := NewMsgWriter()
	for _, v := range sp.ctx.Ins {
		m.WriteBytes(v.OutHash[:])
		m.WriteUInt32(v.OutIndex)
	}
	HASH256To(m.Payload, &hash)
	return hash
}

func (sp *witnessPacker) getSequenceHash() HashID {
	hash := HashID{}
	m := NewMsgWriter()
	for _, v := range sp.ctx.Ins {
		m.WriteUInt32(v.Sequence)
	}
	HASH256To(m.Payload, &hash)
	return hash
}

func (sp *witnessPacker) getScriptCode() *script.Script {
	if sp.in.Script.IsP2WPKH() {
		hash := sp.in.Script.SubBytes(3, 23)
		ns := &script.Script{}
		ns = ns.PushOp(script.OP_DUP)
		ns = ns.PushOp(script.OP_HASH160)
		ns = ns.PushBytes(hash)
		ns = ns.PushOp(script.OP_EQUALVERIFY)
		ns = ns.PushOp(script.OP_CHECKSIG)
		return ns
	}
	return nil
}

func (sp *witnessPacker) Pack(db db.DbImp) ([]byte, error) {
	m := NewMsgWriter()
	m.WriteUInt32(uint32(sp.ctx.Ver))
	m.WriteHash(sp.getPrevoutHash())
	m.WriteHash(sp.getSequenceHash())
	m.WriteHash(sp.in.OutHash)
	m.WriteUInt32(sp.in.OutIndex)
	m.WriteScript(sp.getScriptCode())
	m.WriteUInt64(sp.out.Value)
	m.WriteUInt32(sp.in.Sequence)
	m.WriteHash(sp.getOutputsHash())
	m.WriteUInt32(sp.ctx.LockTime)
	m.WriteUInt32(sp.ht)
	return m.Bytes(), nil
}
