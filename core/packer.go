package core

import (
	"bitcoin/script"
	"fmt"
)

//get sig script interface
type ISigScript interface {
	SigScript() *script.Script
}

type SigPacker interface {
	//pack sig data
	//imp get sig script code
	Pack(imp ISigScript) ([]byte, error)
}

type baseSigPacker struct {
	idx int    //current ints index
	in  *TxIn  //current in
	out *TxOut //in's out
	ctx *TX    //currenct tx'clone
	ht  uint32 //hash type script.SIGHASH_*
	typ TXType //tx type
}

func (sp *baseSigPacker) Pack(imp ISigScript) ([]byte, error) {
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
			w.WriteScript(imp.SigScript())
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

type witnesSigPacker struct {
	idx int    //current ints index
	in  *TxIn  //current in
	out *TxOut //in's out
	ctx *TX    //currenct tx'clone
	ht  uint32 //hash type script.SIGHASH_*
	typ TXType //tx type
}

func (sp *witnesSigPacker) getOutputsHash() HashID {
	hash := HashID{}
	m := NewMsgWriter()
	for _, v := range sp.ctx.Outs {
		m.WriteUInt64(v.Value)
		m.WriteScript(v.Script)
	}
	HASH256To(m.Payload, &hash)
	return hash
}

func (sp *witnesSigPacker) getPrevoutHash() HashID {
	hash := HashID{}
	m := NewMsgWriter()
	for _, v := range sp.ctx.Ins {
		m.WriteBytes(v.OutHash[:])
		m.WriteUInt32(v.OutIndex)
	}
	HASH256To(m.Payload, &hash)
	return hash
}

func (sp *witnesSigPacker) getSequenceHash() HashID {
	hash := HashID{}
	m := NewMsgWriter()
	for _, v := range sp.ctx.Ins {
		m.WriteUInt32(v.Sequence)
	}
	HASH256To(m.Payload, &hash)
	return hash
}

func (sp *witnesSigPacker) Pack(imp ISigScript) ([]byte, error) {
	m := NewMsgWriter()
	m.WriteInt32(sp.ctx.Ver)
	m.WriteHash(sp.getPrevoutHash())
	m.WriteHash(sp.getSequenceHash())
	m.WriteHash(sp.in.OutHash)
	m.WriteUInt32(sp.in.OutIndex)
	m.WriteScript(imp.SigScript())
	m.WriteUInt64(sp.out.Value)
	m.WriteUInt32(sp.in.Sequence)
	m.WriteHash(sp.getOutputsHash())
	m.WriteUInt32(sp.ctx.LockTime)
	m.WriteUInt32(sp.ht)
	return m.Bytes(), nil
}
