package net

import (
	"testing"
	"time"
)

func TestVersionMessage(t *testing.T) {
	c := NewClient(ClientTypeOut, "24.28.64.159:8333")
	c.Run()
	time.Sleep(time.Second * 3)
	m := NewMsgGetData()
	m.Add(&Inventory{
		Type: MSG_TX,
		Hash: NewBHashWithString("09b267dcba09d1904222abb2018f07ebcb9e7a05ede920276cc9f229ddb910d2"),
	})
	c.WC <- m

	time.Sleep(time.Hour)
}
