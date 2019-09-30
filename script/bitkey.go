package script

import (
	"bitcoin/util"
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"errors"
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
	one   = big.NewInt(1)
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
	curve elliptic.Curve
	d     *big.Int
	x     *big.Int
	y     *big.Int
	data  []byte
}

func (pk PrivateKey) HasPublic() bool {
	cl := (curve.Params().BitSize + 7) >> 3
	dl := len(pk.data)
	return dl == cl+COMPRESSED_PUBLIC_KEY_SIZE || dl == cl+PUBLIC_KEY_SIZE
}

func (pk *PrivateKey) Load(b []byte) error {
	cl := (curve.Params().BitSize + 7) >> 3
	dl := len(b)
	if dl < cl {
		return errors.New("data size error")
	}
	pk.d = &big.Int{}
	pk.d.SetBytes(b[:cl])
	pk.curve = curve
	pk.data = b
	//not load public key
	if pk.HasPublic() {
		pb := b[cl:]
		pub := &PublicKey{}
		err := pub.Unmarshal(pb)
		if err != nil {
			return err
		}
		pk.x, pk.y = pub.x, pub.y
	}
	return nil
}

//include public
func (pk PrivateKey) Dump(haspub bool, compressed bool) []byte {
	b := pk.Marshal()
	if !haspub {
		return b
	}
	pb := pk.PublicKey()
	if compressed {
		pb = pb.Compress()
	}
	b = append(b, pb.Marshal()...)
	return b
}

func DecompressY(x *big.Int, ybit uint) (*big.Int, error) {
	c := curve.Params()

	// y^2 = x^3 + b
	// y   = sqrt(x^3 + b)
	var y, x3b big.Int
	x3b.Mul(x, x)
	x3b.Mul(&x3b, x)
	x3b.Add(&x3b, c.B)
	x3b.Mod(&x3b, c.P)
	y.ModSqrt(&x3b, c.P)

	if y.Bit(0) != ybit {
		y.Sub(c.P, &y)
	}
	if y.Bit(0) != ybit {
		return nil, errors.New("incorrectly encoded X and Y bit")
	}
	return &y, nil
}

func (pk PrivateKey) String() string {
	return pk.d.String()
}

func NewPrivateKey() (*PrivateKey, error) {
	pk, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, err
	}
	kv := &PrivateKey{
		curve: curve,
		d:     pk.D,
		x:     pk.X,
		y:     pk.Y,
	}
	kv.data = kv.Dump(true, false)
	return kv, nil
}

func (pk *PrivateKey) Marshal() []byte {
	if len(pk.data) == 0 {
		pk.data = pk.d.Bytes()
	}
	return pk.data
}

func (pk *PrivateKey) PublicKey() *PublicKey {
	return &PublicKey{
		curve: pk.curve,
		x:     pk.x,
		y:     pk.y,
	}
}

type PublicKey struct {
	curve      elliptic.Curve
	x          *big.Int
	y          *big.Int
	data       []byte
	compressed bool
}

func (pk *PublicKey) String() string {
	if len(pk.data) == 0 {
		pk.data = elliptic.Marshal(pk.curve, pk.x, pk.y)
	}
	return hex.EncodeToString(pk.data)
}

func (pk *PublicKey) IsCompressed() bool {
	return pk.compressed
}

func BigIntOdd(v *big.Int) byte {
	z := big.NewInt(0).And(v, one)
	return byte(z.Int64())
}

func (pk *PublicKey) Compress() *PublicKey {
	pk.compressed = true
	return pk
}

func (pk *PublicKey) Unmarshal(data []byte) error {
	byteLen := (curve.Params().BitSize + 7) >> 3
	if len(data) == 0 {
		return errors.New("data empty")
	}
	bl := GetPubKeyLen(data[0])
	if len(data) != bl {
		return errors.New("data size error")
	}
	if bl == PUBLIC_KEY_SIZE {
		if data[0] != P256_PUBKEY_UNCOMPRESSED && data[0] != P256_PUBKEY_HYBRID_EVEN && data[0] != P256_PUBKEY_HYBRID_ODD {
			return errors.New("public head byte error")
		}
		p := curve.Params().P
		x := new(big.Int).SetBytes(data[1 : 1+byteLen])
		y := new(big.Int).SetBytes(data[1+byteLen:])
		d := BigIntOdd(y)
		if data[0] != P256_PUBKEY_HYBRID_ODD && d != 1 {
			return errors.New("public key odd error")
		}
		if data[0] != P256_PUBKEY_HYBRID_EVEN && d != 0 {
			return errors.New("public key even error")
		}
		if x.Cmp(p) >= 0 || y.Cmp(p) >= 0 {
			return errors.New(" x,y error")
		}
		if !curve.IsOnCurve(x, y) {
			return errors.New(" x,y not at curve error")
		}
		pk.x, pk.y = x, y
	} else if bl == COMPRESSED_PUBLIC_KEY_SIZE {
		if data[0] != P256_PUBKEY_EVEN && data[0] != P256_PUBKEY_ODD {
			return errors.New(" cpmpressed head byte error")
		}
		p := curve.Params().P
		x := new(big.Int).SetBytes(data[1 : 1+byteLen])
		y := new(big.Int)
		//
		d := BigIntOdd(y)
		if data[0] != P256_PUBKEY_ODD && d != 1 {
			return errors.New("cpmpressed public key odd error")
		}
		if data[0] != P256_PUBKEY_EVEN && d != 0 {
			return errors.New("cpmpressed public key even error")
		}
		if x.Cmp(p) >= 0 || y.Cmp(p) >= 0 {
			return errors.New("cpmpressed x,y error")
		}
		if !curve.IsOnCurve(x, y) {
			return errors.New("cpmpressed x,y not at curve error")
		}
		pk.x, pk.y = x, y
	}
	return errors.New("data size error")
}

func (pk *PublicKey) Marshal() []byte {
	if len(pk.data) == 0 {
		pk.data = elliptic.Marshal(pk.curve, pk.x, pk.y)
	}
	return pk.data
}

type PubKey [PUBLIC_KEY_SIZE]byte

func NewPubKey() *PubKey {
	k := &PubKey{}
	k.Invalidate()
	return k
}

func (p *PubKey) Equal(o *PubKey) bool {
	return p[0] == o[0] && bytes.Equal(p[:p.Size()], o[:o.Size()])
}

func (p *PubKey) IsValid() bool {
	return p.Size() > 0
}

func (p *PubKey) Verify(hv []byte, sig []byte) bool {
	x, y := elliptic.Unmarshal(elliptic.P256(), p.Bytes())
	if x == nil || y == nil {
		return false
	}
	return false
}

func (p *PubKey) IsFullyValid() bool {
	if !p.IsValid() {
		return false
	}
	x, y := elliptic.Unmarshal(elliptic.P256(), p.Bytes())
	if x == nil || y == nil {
		return false
	}
	return true
}

func (p *PubKey) IsCompressed() bool {
	return p.Size() == COMPRESSED_PUBLIC_KEY_SIZE
}

func (p *PubKey) Bytes() []byte {
	return p[:p.Size()]
}

func (p *PubKey) HashID() []byte {
	return util.HASH160(p[:p.Size()])
}

func (p *PubKey) Set(buf []byte) {
	if len(buf) != GetPubKeyLen(buf[0]) {
		p.Invalidate()
	} else {
		copy(p[:], buf)
	}
}

func (p *PubKey) Invalidate() {
	p[0] = 0xFF
}

func (p *PubKey) Size() int {
	return GetPubKeyLen(p[0])
}

func (p *PubKey) ValidSize(key []byte) bool {
	return len(key) > 0 && GetPubKeyLen(key[0]) == p.Size()
}

func (p *PubKey) SetRange(buf []byte, b, e int) {
	if (e - b) != GetPubKeyLen(buf[b]) {
		p.Invalidate()
	} else {
		copy(p[:], buf[b:e])
	}
}
