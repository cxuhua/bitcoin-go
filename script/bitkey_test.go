package script

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"log"
	"testing"
)

func TestSelf(t *testing.T) {

	a, err := NewPrivateKey()
	if err != nil {
		panic(err)
	}
	log.Println(a.x, a.y)

	log.Println(DecompressY(a.x, uint(BigIntOdd(a.y))))

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
