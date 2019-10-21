package core

import (
	"bitcoin/script"
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
	typ TxType //tx type
}

func (sp *baseSigPacker) Pack(imp ISigScript) ([]byte, error) {
	anyone := (sp.ht & script.SIGHASH_ANYONECANPAY) != 0
	single := (sp.ht & 0x7F) == script.SIGHASH_SINGLE
	none := (sp.ht & 0x7F) == script.SIGHASH_NONE
	w := NewMsgWriter()
	w.WriteInt32(sp.ctx.Ver)
	w.WriteVarInt(len(sp.ctx.Ins))
	for i, v := range sp.ctx.Ins {
		if anyone {
			i = sp.idx
		}
		w.WriteBytes(v.OutHash[:])
		w.WriteUInt32(v.OutIndex)
		if i == sp.idx {
			w.WriteScript(imp.SigScript())
		} else {
			w.WriteScript(nil)
		}
		if i != sp.idx && (single || none) {
			w.WriteUInt32(0)
		} else {
			w.WriteUInt32(v.Sequence)
		}
	}
	outs := 0
	if none {
		outs = 0
	} else if single {
		outs = sp.idx + 1
	} else {
		outs = len(sp.ctx.Outs)
	}
	w.WriteVarInt(outs)
	for i, v := range sp.ctx.Outs {
		if i >= outs {
			break
		}
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
	typ TxType //tx type
}

func (sp *witnesSigPacker) getOutputsHash() HashID {
	single := (sp.ht & 0x7F) == script.SIGHASH_SINGLE
	none := (sp.ht & 0x7F) == script.SIGHASH_NONE
	hash := HashID{}
	if !single && !none {
		m := NewMsgWriter()
		for _, v := range sp.ctx.Outs {
			m.WriteUInt64(v.Value)
			m.WriteScript(v.Script)
		}
		return HASH256To(m.Bytes(), &hash)
	} else if single && sp.idx < len(sp.ctx.Outs) {
		ov := sp.ctx.Outs[sp.idx]
		m := NewMsgWriter()
		m.WriteUInt64(ov.Value)
		m.WriteScript(ov.Script)
		return HASH256To(m.Bytes(), &hash)
	}
	return hash
}

func (sp *witnesSigPacker) getPrevoutHash() HashID {
	anyone := (sp.ht & script.SIGHASH_ANYONECANPAY) != 0
	hash := HashID{}
	if !anyone {
		m := NewMsgWriter()
		for _, v := range sp.ctx.Ins {
			m.WriteBytes(v.OutHash[:])
			m.WriteUInt32(v.OutIndex)
		}
		return HASH256To(m.Bytes(), &hash)
	} else {
		return hash
	}
}

func (sp *witnesSigPacker) getSequenceHash() HashID {
	anyone := (sp.ht & script.SIGHASH_ANYONECANPAY) != 0
	single := (sp.ht & 0x7F) == script.SIGHASH_SINGLE
	none := (sp.ht & 0x7F) == script.SIGHASH_NONE
	hash := HashID{}
	if !anyone && !single && !none {
		m := NewMsgWriter()
		for _, v := range sp.ctx.Ins {
			m.WriteUInt32(v.Sequence)
		}
		return HASH256To(m.Bytes(), &hash)
	} else {
		return hash
	}
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
