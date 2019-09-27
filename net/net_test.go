package net

import (
	"testing"
	"time"
)

func TestVersionMessage(t *testing.T) {
	c := NewClient(ClientTypeOut, "172.81.183.236:8333")
	c.Run()
	time.Sleep(time.Second * 3)
	m := NewMsgGetData()
	m.Add(&Inventory{
		Type: MSG_TX,
		Hash: NewBHashWithString("7681580A6611D1FE9287184B0D20A9108E94E456CA56F4C03C722D164EAAA05A"),
	})
	c.WC <- m

	time.Sleep(time.Hour)
}
