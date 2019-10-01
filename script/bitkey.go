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

//prefix[1] key[32] checknum[hash256-prefix-4]
func DecodePrivateKey(s string) (*PrivateKey, error) {
	key, err := NewPrivateKey()
	if err != nil {
		return nil, err
	}
	if err := key.Decode(s); err != nil {
		return nil, err
	}
	return key, nil
}

func (pk *PrivateKey) Encode() string {
	conf := config.GetConfig()
	prefix := conf.Base58Prefix(config.SECRET_KEY)
	pb := pk.D.Bytes()
	buf := &bytes.Buffer{}
	buf.Write(prefix)
	buf.Write(pb)
	if pk.compressed {
		buf.WriteByte(1)
	}
	hv := util.HASH256(buf.Bytes())
	buf.Write(hv[:4])
	return util.B58Encode(buf.Bytes(), util.BitcoinAlphabet)
}

func (pk *PrivateKey) Decode(s string) error {
	conf := config.GetConfig()
	data, err := util.B58Decode(s, util.BitcoinAlphabet)
	if err != nil {
		return err
	}
	if len(data) < 4 {
		return errors.New("size error")
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
		pk.SetBytes(data[pl:pl+32], compressed)
	}
	return nil
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
		D:          d,
		compressed: true,
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

func (sig SigValue) ToBytes() []byte {
	r := []byte{}
	r = append(r, sig.R.Bytes()...)
	r = append(r, sig.S.Bytes()...)
	return r
}

func (sig SigValue) ToDER() []byte {
	r := sig.R.Bytes()
	if r[0] >= 0x80 {
		r = append([]byte{0}, r...)
	}
	s := sig.S.Bytes()
	if s[0] >= 0x80 {
		s = append([]byte{0}, s...)
	}
	res := new(bytes.Buffer)
	res.WriteByte(0x30)
	res.WriteByte(byte(4 + len(r) + len(s)))
	res.WriteByte(0x02)
	res.WriteByte(byte(len(r)))
	res.Write(r)
	res.WriteByte(0x02)
	res.WriteByte(byte(len(s)))
	res.Write(s)
	return res.Bytes()
}
func (sig *SigValue) FromBytes(b []byte) error {
	if len(b) != 64 {
		return errors.New("b size error")
	}
	sig.R = new(big.Int).SetBytes(b[:32])
	sig.S = new(big.Int).SetBytes(b[32:])
	return nil
}

func (sig *SigValue) FromDER(b []byte) error {
	if b[0] != 0x30 || len(b) < 5 {
		return errors.New("der format error")
	}
	lenr := int(b[3])
	if lenr == 0 || 5+lenr >= len(b) || b[lenr+4] != 0x02 {
		return errors.New("der length error")
	}
	lens := int(b[lenr+5])
	if lens == 0 || int(b[1]) != lenr+lens+4 || lenr+lens+6 > len(b) || b[2] != 0x02 {
		return errors.New("der length error")
	}
	sig.R = new(big.Int).SetBytes(b[4 : 4+lenr])
	sig.S = new(big.Int).SetBytes(b[6+lenr : 6+lenr+lens])
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
	err := pk.From(data)
	return pk, err
}

//2-2
//2-3
//1-2
//多重pubkey签名脚本
func GetP2SHPublicScript(pubs []*PublicKey, opc int) ([]byte, error) {

	if len(pubs) == 0 {
		return nil, errors.New("pub error")
	}
	if opc < 1 || opc > len(pubs) {
		opc = len(pubs)
	}
	buf := &bytes.Buffer{}
	op1 := byte(OP_1 + opc - 1)
	buf.WriteByte(op1)
	for _, v := range pubs {
		kv := v.Marshal()
		buf.WriteByte(byte(len(kv)))
		buf.Write(kv)
	}
	op2 := byte(OP_1 + len(pubs) - 1) //total
	buf.WriteByte(op2)
	buf.WriteByte(OP_CHECKMULTISIG)
	return buf.Bytes(), nil
}

//pay to script address
func (pk PublicKey) P2SHAddress() string {
	return util.P2SHAddress(pk.Marshal())
}

//bitcoin address
func (pk PublicKey) P2PKHAddress() string {
	return util.P2PKHAddress(pk.Marshal())
}

func (pk *PublicKey) From(data []byte) error {
	byteLen := (curve.Params().BitSize + 7) >> 3
	if len(data) == 0 {
		return errors.New("data empty")
	}
	pk.b0 = data[0]
	bl := GetPubKeyLen(pk.b0)
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
		d := byte(y.Bit(0))
		if data[0] == P256_PUBKEY_HYBRID_ODD && d != 1 {
			return errors.New("public key odd error")
		}
		if data[0] == P256_PUBKEY_HYBRID_EVEN && d != 0 {
			return errors.New("public key even error")
		}
		if x.Cmp(p) >= 0 || y.Cmp(p) >= 0 {
			return errors.New(" x,y error")
		}
		if !curve.IsOnCurve(x, y) {
			return errors.New(" x,y not at curve error")
		}
		pk.X, pk.Y = x, y
		pk.compressed = false
		return nil
	}
	if bl == COMPRESSED_PUBLIC_KEY_SIZE {
		if data[0] != P256_PUBKEY_EVEN && data[0] != P256_PUBKEY_ODD {
			return errors.New(" compressed head byte error")
		}
		p := curve.Params().P
		x := new(big.Int).SetBytes(data[1 : 1+byteLen])
		var y *big.Int
		ybit := uint(0)
		if data[0] == P256_PUBKEY_ODD {
			ybit = 1
		}
		if v, err := util.DecompressY(x, ybit); err != nil {
			return fmt.Errorf("decompress x -> y error %v", err)
		} else {
			y = v
		}
		d := byte(y.Bit(0))
		if data[0] == P256_PUBKEY_ODD && d != 1 {
			return errors.New("decompress public key odd error")
		}
		if data[0] == P256_PUBKEY_EVEN && d != 0 {
			return errors.New("decompress public key even error")
		}
		if x.Cmp(p) >= 0 || y.Cmp(p) >= 0 {
			return errors.New("decompress x,y error")
		}
		if !curve.IsOnCurve(x, y) {
			return errors.New("cpmpressed x,y not at curve error")
		}
		pk.X, pk.Y = x, y
		pk.compressed = true
		return nil
	}
	return errors.New("data size error")
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
