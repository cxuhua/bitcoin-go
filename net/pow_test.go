package net

import (
	"bitcoin/config"
	"testing"
	"time"
)

//201600 bits compute
//(201600-2016) -> 2012-09-20 06:14:11
//(201600-1) -> 2012-10-03 09:17:01
//(201600-1) Bits = 0x1a05db8b
//result:1a057e08
func TestCalculateWorkRequired(t *testing.T) {
	conf := config.GetConfig()
	t1, _ := time.Parse("2006-01-02 15:04:05", "2012-09-20 06:14:11")
	t2, _ := time.Parse("2006-01-02 15:04:05", "2012-10-03 09:17:01")
	x := CalculateWorkRequired(uint32(t2.Unix()), uint32(t1.Unix()), 0x1a05db8b, conf)
	if x != 0x1a057e08 {
		t.Errorf("failed")
	}
}

//Check whether a block hash satisfies the proof-of-work requirement specified by nBits
func TestCheckProofOfWork(t *testing.T) {
	h := NewHexBHash("00000000000003010530e33a849b27ded874202911e9e63263cb49245744fb9e")
	b := CheckProofOfWork(h, 0x1a0575ef, config.GetConfig())
	if !b {
		t.Errorf("test failed")
	}
}
