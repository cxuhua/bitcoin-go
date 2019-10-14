package net

import (
	"bitcoin/config"
	"errors"
)

//
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
	limit := NewHexUHash(conf.PowLimit)
	if h.Cmp(limit) > 0 {
		return false
	}
	ch := hash.ToUHash()
	return ch.Cmp(h) <= 0
}

//ct = current height - 1 block time
//pt = current height - 2016 block time
//pw = current height - 1 bits
func CalculateWorkRequired(ct uint32, pt uint32, pw uint32, conf *config.Config) uint32 {
	span := uint32(conf.PowTargetTimespan)
	limit := NewHexUHash(conf.PowLimit)
	sub := ct - pt
	if sub <= 0 {
		panic(errors.New("ct pt error"))
	}
	if sub < span/4 {
		sub = span / 4
	}
	if sub > span*4 {
		sub = span * 4
	}
	n := UIHash{}
	n.SetCompact(pw)
	n = n.MulUInt32(sub)
	n = n.Div(NewU64Hash(uint64(span)))
	if n.Cmp(limit) > 0 {
		n = limit
	}
	return n.Compact(false)
}
