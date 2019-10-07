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
	ht  byte   //hash type script.SIGHASH_*
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
	w.WriteUInt32(uint32(sp.ht))
	return w.Bytes(), nil
}
