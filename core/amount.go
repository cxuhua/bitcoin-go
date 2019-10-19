package core

import (
	"bitcoin/config"
)

const (
	COIN      = Amount(100000000)
	MAX_MONEY = Amount(21000000 * COIN)
)

func GetCoinbaseReward(h int) Amount {
	conf := config.GetConfig()
	halvings := h / conf.SubHalving
	if halvings >= 64 {
		return 0
	}
	n := 50 * COIN
	n >>= halvings
	return n
}

type Amount int64

func (a Amount) IsRange() bool {
	return a >= 0 && a < MAX_MONEY
}
