package core

import (
	"bitcoin/config"
	"errors"
)

// Check whether a block hash satisfies the proof-of-work requirement specified by nBits
func CheckProofOfWork(hash HashID, bits uint32, conf *config.Config) bool {
	h := UIHash{}
	n, o := h.SetCompact(bits)
	if n {
		return false
	}
	if h.IsZero() {
		return false
	}
	if o {
		return false
	}
	limit := NewUIHash(conf.PowLimit)
	if h.Cmp(limit) > 0 {
		return false
	}
	ch := hash.ToUHash()
	return ch.Cmp(h) <= 0
}

//ct = lastBlock blockTime
//pt = lastBlock - 2016 + 1 blockTime
//pw = lastBlock's bits
func CalculateWorkRequired(ct uint32, pt uint32, pw uint32, conf *config.Config) uint32 {
	span := uint32(conf.PowTargetTimespan)
	limit := NewUIHash(conf.PowLimit)
	sub := ct - pt
	if sub <= 0 {
		panic(errors.New("ct pt error"))
	}
	if sv := span / 4; sub < sv {
		sub = sv
	}
	if sv := span * 4; sub > sv {
		sub = sv
	}
	n := UIHash{}
	n.SetCompact(pw)
	n = n.MulUInt32(sub)
	n = n.Div(NewUIHash(span))
	if n.Cmp(limit) > 0 {
		n = limit
	}
	return n.Compact(false)
}
