package script

import (
	"bitcoin/config"
	"bitcoin/util"
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
)

const (
	PUBLIC_KEY_SIZE            = 65
	COMPRESSED_PUBLIC_KEY_SIZE = 33
	SIGNATURE_SIZE             = 72
	COMPACT_SIGNATURE_SIZE     = 65

	P256_PUBKEY_EVEN         = byte(0x02)
	P256_PUBKEY_ODD          = byte(0x03)
	P256_PUBKEY_UNCOMPRESSED = byte(0x04)
	P256_PUBKEY_HYBRID_EVEN  = byte(0x06)
	P256_PUBKEY_HYBRID_ODD   = byte(0x07)
)

var (
	curve = util.SECP256K1()
)

func GetPubKeyLen(header byte) int {
	if header == P256_PUBKEY_EVEN || header == P256_PUBKEY_ODD {
		return COMPRESSED_PUBLIC_KEY_SIZE
	}
	if header == P256_PUBKEY_UNCOMPRESSED || header == P256_PUBKEY_HYBRID_EVEN || header == P256_PUBKEY_HYBRID_ODD {
		return PUBLIC_KEY_SIZE
	}
	return 0
}

type PrivateKey struct {
	D          *big.Int
	compressed bool
}

func DecodePrivateKey(s string) (*PrivateKey, error) {
	key, err := NewPrivateKey()
	if err != nil {
		return nil, err
	}
	conf := config.GetConfig()
	data, err := util.B58Decode(s, util.BitcoinAlphabet)
	if err != nil {
		return nil, err
	}
	if len(data) < 4 {
		return nil, errors.New("size error")
	}
	prefix := conf.Base58Prefix(config.SECRET_KEY)
	dl := len(data)
	hv := util.HASH256(data[:dl-4])
	if bytes.Equal(hv[:4], data[dl-4:]) {
		data = data[:dl-4]
	}
	dl = len(data)
	pl := len(prefix)
	if (dl == pl+32 || (dl == pl+33 && data[dl-1] == 1)) && bytes.Equal(prefix, data[:pl]) {
		compressed := dl == 33+pl
		key.SetBytes(data[pl:pl+32], compressed)
	}
	return key, nil
}

func (pk *PrivateKey) IsCompressed() bool {
	return pk.compressed
}

func (pk *PrivateKey) IsValid() bool {
	return pk.PublicKey().IsValid()
}

func (pk *PrivateKey) SetBytes(b []byte, compressed bool) {
	pk.compressed = compressed
	pk.D = new(big.Int).SetBytes(b)
}

func (pk PrivateKey) String() string {
	return hex.EncodeToString(pk.D.Bytes())
}

func NewPrivateKey() (*PrivateKey, error) {
	d, err := util.GenPrivateKey()
	if err != nil {
		return nil, err
	}
	return &PrivateKey{
		D: d,
	}, nil
}

func LoadPrivateKey(d []byte) (*PrivateKey, error) {
	if len(d) != curve.Params().BitSize/8 {
		return nil, errors.New("private size error")
	}
	p := &PrivateKey{}
	p.D = new(big.Int).SetBytes(d)
	return p, nil
}

func (pk PrivateKey) Sign(hash []byte) (*SigValue, error) {
	sig := &SigValue{}
	priv := new(ecdsa.PrivateKey)
	priv.Curve = curve
	priv.D = pk.D
	pub := pk.PublicKey()
	priv.X, priv.Y = pub.X, pub.Y
	r, s, err := ecdsa.Sign(rand.Reader, priv, hash)
	if err != nil {
		return nil, err
	}
	sig.R, sig.S = r, s
	return sig, nil
}

func (pk PrivateKey) Marshal() []byte {
	return pk.D.Bytes()
}

func (pk PrivateKey) PublicKey() *PublicKey {
	pp := &PublicKey{}
	pp.X, pp.Y = curve.ScalarBaseMult(pk.Marshal())
	pp.Compressed(pk.compressed)
	return pp
}

type SigValue struct {
	R *big.Int
	S *big.Int
}

//secp256k1_ecdsa_sig_serialize
func (sig SigValue) Marshal() []byte {
	if sig.R == nil || sig.S == nil {
		panic(errors.New("null sig value"))
	}
	return nil
}

func (sig *SigValue) Unmarshal(b []byte) error {
	return nil
}

type PublicKey struct {
	X          *big.Int
	Y          *big.Int
	b0         byte
	compressed bool
}

func LoadPublicKey(data []byte) (*PublicKey, error) {
	pk := &PublicKey{}
	byteLen := (curve.Params().BitSize + 7) >> 3
	if len(data) == 0 {
		return nil, errors.New("data empty")
	}
	pk.b0 = data[0]
	bl := GetPubKeyLen(pk.b0)
	if len(data) != bl {
		return nil, errors.New("data size error")
	}
	if bl == PUBLIC_KEY_SIZE {
		if data[0] != P256_PUBKEY_UNCOMPRESSED && data[0] != P256_PUBKEY_HYBRID_EVEN && data[0] != P256_PUBKEY_HYBRID_ODD {
			return nil, errors.New("public head byte error")
		}
		p := curve.Params().P
		x := new(big.Int).SetBytes(data[1 : 1+byteLen])
		y := new(big.Int).SetBytes(data[1+byteLen:])
		d := byte(y.Bit(0))
		if data[0] != P256_PUBKEY_HYBRID_ODD && d != 1 {
			return nil, errors.New("public key odd error")
		}
		if data[0] != P256_PUBKEY_HYBRID_EVEN && d != 0 {
			return nil, errors.New("public key even error")
		}
		if x.Cmp(p) >= 0 || y.Cmp(p) >= 0 {
			return nil, errors.New(" x,y error")
		}
		if !curve.IsOnCurve(x, y) {
			return nil, errors.New(" x,y not at curve error")
		}
		pk.X, pk.Y = x, y
		pk.compressed = false
	} else if bl == COMPRESSED_PUBLIC_KEY_SIZE {
		if data[0] != P256_PUBKEY_EVEN && data[0] != P256_PUBKEY_ODD {
			return nil, errors.New(" compressed head byte error")
		}
		p := curve.Params().P
		x := new(big.Int).SetBytes(data[1 : 1+byteLen])
		var y *big.Int
		ybit := uint(0)
		if data[0] == P256_PUBKEY_ODD {
			ybit = 1
		}
		if v, err := util.DecompressY(x, ybit); err != nil {
			return nil, fmt.Errorf("decompress x -> y error %v", err)
		} else {
			y = v
		}
		d := byte(y.Bit(0))
		if data[0] != P256_PUBKEY_ODD && d != 1 {
			return nil, errors.New("cpmpressed public key odd error")
		}
		if data[0] != P256_PUBKEY_EVEN && d != 0 {
			return nil, errors.New("cpmpressed public key even error")
		}
		if x.Cmp(p) >= 0 || y.Cmp(p) >= 0 {
			return nil, errors.New("cpmpressed x,y error")
		}
		if !curve.IsOnCurve(x, y) {
			return nil, errors.New("cpmpressed x,y not at curve error")
		}
		pk.X, pk.Y = x, y
		pk.compressed = true
	}
	return pk, errors.New("data size error")
}

func (pb *PublicKey) IsValid() bool {
	return curve.IsOnCurve(pb.X, pb.Y)
}

func (pk *PublicKey) Compressed(v bool) {
	pk.compressed = v
}

//check marshal pubkey
func IsValidPublicKey(pk []byte) bool {
	return len(pk) > 0 && GetPubKeyLen(pk[0]) == len(pk)
}

func (pk *PublicKey) Verify(hash []byte, sig *SigValue) bool {
	pub := new(ecdsa.PublicKey)
	pub.Curve = curve
	pub.X, pub.Y = pk.X, pk.Y
	return ecdsa.Verify(pub, hash, sig.R, sig.S)
}

func (pk *PublicKey) Hybrid() []byte {
	ret := []byte{}
	d := byte(pk.Y.Bit(0))
	ret = append(ret, P256_PUBKEY_HYBRID_EVEN+d)
	ret = append(ret, pk.X.Bytes()...)
	ret = append(ret, pk.Y.Bytes()...)
	return ret
}

func (pk *PublicKey) Marshal() []byte {
	ret := []byte{}
	d := byte(pk.Y.Bit(0))
	if !pk.compressed {
		ret = append(ret, P256_PUBKEY_UNCOMPRESSED)
		ret = append(ret, pk.X.Bytes()...)
		ret = append(ret, pk.Y.Bytes()...)
	} else {
		ret = append(ret, P256_PUBKEY_EVEN+d)
		ret = append(ret, pk.X.Bytes()...)
	}
	return ret
}
