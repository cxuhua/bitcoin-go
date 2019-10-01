package script

import (
	"bitcoin/util"
	"bytes"
	"encoding/hex"
	"testing"
)

var (
	pk1even, _          = hex.DecodeString("e0a963079b610fc7711350da4430b7ac2de119f4046e2621593ae3395d54ff99")
	pk1evenpub, _       = hex.DecodeString("04c68645547106c18b2bfbd6badb6c01d589c59ae69c22adb0607e482938c735a709cdc78752430d662711333958660769905c381b8b76b8fe9af8bc777b2c61fa")
	pk1evencommpress, _ = hex.DecodeString("02c68645547106c18b2bfbd6badb6c01d589c59ae69c22adb0607e482938c735a7")
	pk1evenhyb, _       = hex.DecodeString("06c68645547106c18b2bfbd6badb6c01d589c59ae69c22adb0607e482938c735a709cdc78752430d662711333958660769905c381b8b76b8fe9af8bc777b2c61fa")
	pk1odd, _           = hex.DecodeString("966f721f5a49b9cece99f36b276e0fc0b7ce46c9abd6e3ee3ee03bf9618f286f")
	pk1oddpub, _        = hex.DecodeString("04402ae4cc64262e93f807fee4e79605d40142142e547707b26ad5dab97bb26bf1a5b6f01499ad935633dad4725fdbbe32247257b9ea4b24a7056b8357efa207d3")
	pk1oddcompress, _   = hex.DecodeString("03402ae4cc64262e93f807fee4e79605d40142142e547707b26ad5dab97bb26bf1")
	pk1oddhyb, _        = hex.DecodeString("07402ae4cc64262e93f807fee4e79605d40142142e547707b26ad5dab97bb26bf1a5b6f01499ad935633dad4725fdbbe32247257b9ea4b24a7056b8357efa207d3")
)

type pkv struct {
	pk         []byte
	pkpub      []byte
	pkcompress []byte
	pkhyb      []byte
}

func TestPrivatePublicKey(t *testing.T) {
	pkvs := []pkv{
		{
			pk:         pk1even,
			pkpub:      pk1evenpub,
			pkcompress: pk1evencommpress,
			pkhyb:      pk1evenhyb,
		},
		{
			pk:         pk1odd,
			pkpub:      pk1oddpub,
			pkcompress: pk1oddcompress,
			pkhyb:      pk1oddhyb,
		},
	}
	for _, v := range pkvs {
		pk, err := LoadPrivateKey(v.pk)
		if err != nil {
			t.Errorf("load private key error %v", err)
		}
		if !bytes.Equal(pk.Marshal(), v.pk) {
			t.Errorf("marsha private key error")
		}
		pb := pk.PublicKey()
		if pb.Compressed(true); !bytes.Equal(pb.Marshal(), v.pkcompress) {
			t.Errorf("compressed error")
		}
		if !IsValidPublicKey(pb.Marshal()) {
			t.Errorf("valid public key error")
		}
		if pb.Compressed(false); !bytes.Equal(pb.Marshal(), v.pkpub) {
			t.Errorf("not compressed error")
		}
		if !IsValidPublicKey(pb.Marshal()) {
			t.Errorf("valid public key error")
		}
		if !bytes.Equal(pb.Hybrid(), v.pkhyb) {
			t.Errorf("not compressed hyb error")
		}
		if !IsValidPublicKey(pb.Hybrid()) {
			t.Errorf("valid public key error")
		}
	}
}

const (
	strSecret1  = "5HxWvvfubhXpYYpS3tJkw6fq9jE9j18THftkZjHHfmFiWtmAbrj"
	strSecret2  = "5KC4ejrDjv152FGwP386VD1i2NYc5KkfSMyv1nGy1VGDxGHqVY3"
	strSecret1C = "Kwr371tjA9u2rFSMZjTNun2PXXP3WPZu2afRHTcta6KxEUdm1vEw"
	strSecret2C = "L3Hq7a8FEQwJkW1M2GNKDW28546Vp5miewcCzSqUD9kCAXrJdS3g"
	addr1       = "1QFqqMUD55ZV3PJEJZtaKCsQmjLT6JkjvJ"
	addr2       = "1F5y5E5FMc5YzdJtB9hLaUe43GDxEKXENJ"
	addr1C      = "1NoJrossxPBKfCHuJXT4HadJrXRE9Fxiqs"
	addr2C      = "1CRj2HyM1CXWzHAXLQtiGLyggNT9WQqsDs"
)

func TestBase58Key(t *testing.T) {
	pk1, err := DecodePrivateKey(strSecret1)
	if err != nil {
		t.Errorf("DecodePrivateKey error %v", err)
	}
	if !pk1.IsValid() {
		t.Errorf("pubkey IsValid = false ")
	}
	if pk1.IsCompressed() {
		t.Errorf("pubkey IsCompressed = true ")
	}
	pk2, err := DecodePrivateKey(strSecret1C)
	if err != nil {
		t.Errorf("DecodePrivateKey error %v", err)
	}
	if !pk2.IsValid() {
		t.Errorf("pubkey IsValid = false ")
	}
	if !pk2.IsCompressed() {
		t.Errorf("pubkey IsCompressed = false ")
	}
}

func TestSignVerify(t *testing.T) {
	msg := "Very deterministic message"
	hash := util.HASH256([]byte(msg))
	pk1, err := DecodePrivateKey(strSecret1)
	if err != nil {
		t.Errorf("DecodePrivateKey error %v", err)
	}
	pk2, err := DecodePrivateKey(strSecret1C)
	if err != nil {
		t.Errorf("DecodePrivateKey error %v", err)
	}
	pk3, err := DecodePrivateKey(strSecret2C)
	if err != nil {
		t.Errorf("DecodePrivateKey error %v", err)
	}
	sig1, err := pk1.Sign(hash)
	if err != nil {
		t.Errorf("sign 1 error %v", err)
	}
	if pub1 := pk1.PublicKey(); !pub1.Verify(hash, sig1) {
		t.Errorf("Verify 1 error")
	}
	sig2, err := pk2.Sign(hash)
	if err != nil {
		t.Errorf("sign 1 error %v", err)
	}
	if pub2 := pk2.PublicKey(); !pub2.Verify(hash, sig2) {
		t.Errorf("Verify 2 error")
	}
	if pub3 := pk3.PublicKey(); pub3.Verify(hash, sig2) {
		t.Errorf("Verify 3 should error")
	}
}
