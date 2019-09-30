package script

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"log"
	"math/big"
	"testing"
)

func TestSelf(t *testing.T) {

	a, err := NewPrivateKey()
	if err != nil {
		panic(err)
	}
	log.Println(a.x, a.y)

	//x*x*x
	x2 := big.NewInt(0).Mul(a.x, a.x)
	x3 := big.NewInt(0).Mul(x2, a.x)
	c := big.NewInt(7)
	c.Add(c, x3)

	y := big.NewInt(0).Sqrt(c)
	if BigIntOdd(y) == 0 {
		log.Println(y.Neg(y))
	} else {
		log.Println(y)
	}
	//dump := pri.Dump(true, true)
	//np := &PrivateKey{}
	//err = np.Load(dump, true)
	//if err != nil {
	//	t.Errorf("Load private error %v", err)
	//}
}

func TestMakePubKey(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}
	v := elliptic.Marshal(elliptic.P256(), privateKey.X, privateKey.Y)
	x, y := elliptic.Unmarshal(elliptic.P256(), v)
	log.Println(len(v), len(x.Bytes()), len(y.Bytes()), len(privateKey.D.Bytes()))
	msg := "hello, world"
	hash := sha256.Sum256([]byte(msg))

	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash[:])
	if err != nil {
		panic(err)
	}
	fmt.Printf("signature: (0x%x, 0x%x)\n", r, s)

	valid := ecdsa.Verify(&privateKey.PublicKey, hash[:], r, s)
	fmt.Println("signature verified:", valid)
}
