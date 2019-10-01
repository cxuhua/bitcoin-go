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

var ta = [][3]string{
	// [0]-pubScr, [1]-sigScript, [2]-unsignedTx
	{
		"040eaebcd1df2df853d66ce0e1b0fda07f67d1cabefde98514aad795b86a6ea66dbeb26b67d7a00e2447baeccc8a4cef7cd3cad67376ac1c5785aeebb4f6441c16",
		"3045022100fe00e013c244062847045ae7eb73b03fca583e9aa5dbd030a8fd1c6dfcf11b1002207d0d04fed8fa1e93007468d5a9e134b0a7023b6d31db4e50942d43a250f4d07c01",
		"3382219555ddbb5b00e0090f469e590ba1eae03c7f28ab937de330aa60294ed6",
	},
	{
		"020eaebcd1df2df853d66ce0e1b0fda07f67d1cabefde98514aad795b86a6ea66d",
		"3045022100fe00e013c244062847045ae7eb73b03fca583e9aa5dbd030a8fd1c6dfcf11b1002207d0d04fed8fa1e93007468d5a9e134b0a7023b6d31db4e50942d43a250f4d07c01",
		"3382219555ddbb5b00e0090f469e590ba1eae03c7f28ab937de330aa60294ed6",
	},
	{
		"0411db93e1dcdb8a016b49840f8c53bc1eb68a382e97b1482ecad7b148a6909a5cb2e0eaddfb84ccf9744464f82e160bfa9b8b64f9d4c03f999b8643f656b412a3",
		"304402204e45e16932b8af514961a1d3a1a25fdf3f4f7732e9d624c6c61548ab5fb8cd410220181522ec8eca07de4860a4acdd12909d831cc56cbbac4622082221a8768d1d0901",
		"7a05c6145f10101e9d6325494245adf1297d80f8f38d4d576d57cdba220bcb19",
	},
	{
		"0311db93e1dcdb8a016b49840f8c53bc1eb68a382e97b1482ecad7b148a6909a5c",
		"304402204e45e16932b8af514961a1d3a1a25fdf3f4f7732e9d624c6c61548ab5fb8cd410220181522ec8eca07de4860a4acdd12909d831cc56cbbac4622082221a8768d1d0901",
		"7a05c6145f10101e9d6325494245adf1297d80f8f38d4d576d57cdba220bcb19",
	},
	{
		"0428f42723f81c70664e200088437282d0e11ae0d4ae139f88bdeef1550471271692970342db8e3f9c6f0123fab9414f7865d2db90c24824da775f00e228b791fd",
		"3045022100d557da5d9bf886e0c3f98fd6d5d337487cd01d5b887498679a57e3d32bd5d0af0220153217b63a75c3145b14f58c64901675fe28dba2352c2fa9f2a1579c74a2de1701",
		"c22de395adbb0720941e009e8a4e488791b2e428af775432ed94d2c7ec8e421a",
	},
	{
		"0328f42723f81c70664e200088437282d0e11ae0d4ae139f88bdeef15504712716",
		"3045022100d557da5d9bf886e0c3f98fd6d5d337487cd01d5b887498679a57e3d32bd5d0af0220153217b63a75c3145b14f58c64901675fe28dba2352c2fa9f2a1579c74a2de1701",
		"c22de395adbb0720941e009e8a4e488791b2e428af775432ed94d2c7ec8e421a",
	},
	{
		"041f2a00036b3cbd1abe71dca54d406a1e9dd5d376bf125bb109726ff8f2662edcd848bd2c44a86a7772442095c7003248cc619bfec3ddb65130b0937f8311c787",
		"3045022100ec6eb6b2aa0580c8e75e8e316a78942c70f46dd175b23b704c0330ab34a86a34022067a73509df89072095a16dbf350cc5f1ca5906404a9275ebed8a4ba219627d6701",
		"7c8e7c2cb887682ed04dc82c9121e16f6d669ea3d57a2756785c5863d05d2e6a",
	},
	{
		"031f2a00036b3cbd1abe71dca54d406a1e9dd5d376bf125bb109726ff8f2662edc",
		"3045022100ec6eb6b2aa0580c8e75e8e316a78942c70f46dd175b23b704c0330ab34a86a34022067a73509df89072095a16dbf350cc5f1ca5906404a9275ebed8a4ba219627d6701",
		"7c8e7c2cb887682ed04dc82c9121e16f6d669ea3d57a2756785c5863d05d2e6a",
	},
	{
		"04ee90bfdd4e07eb1cfe9c6342479ca26c0827f84bfe1ab39e32fc3e94a0fe00e6f7d8cd895704e974978766dd0f9fad3c97b1a0f23684e93b400cc9022b7ae532",
		"3045022100fe1f6e2c2c2cbc916f9f9d16497df2f66a4834e5582d6da0ee0474731c4a27580220682bad9359cd946dc97bb07ea8fad48a36f9b61186d47c6798ccce7ba20cc22701",
		"baff983e6dfb1052918f982090aa932f56d9301d1de9a726d2e85d5f6bb75464",
	},
}

func TestSecp256Data(t *testing.T) {
	for _, v := range ta {
		pkey, _ := hex.DecodeString(v[0])
		sign, _ := hex.DecodeString(v[1])
		hasz, _ := hex.DecodeString(v[2])

		pub, err := LoadPublicKey(pkey)
		if err != nil {
			panic(err)
		}
		sig := &SigValue{}
		err = sig.FromDER(sign)
		if err != nil {
			panic(err)
		}
		b := pub.Verify(hasz, sig)
		if !b {
			t.Errorf("test verify error")
		}
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
