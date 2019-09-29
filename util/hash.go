package util

import (
	"crypto/sha1"
	"crypto/sha256"

	"golang.org/x/crypto/ripemd160"
)

//prefix 4 bytes
func HashP4(b []byte) []byte {
	v := HASH256(b)
	return v[:4]
}

var (
	hash160 = ripemd160.New()
	hash256 = sha256.New()
	hash1   = sha1.New()
)

func SHA1(b []byte) []byte {
	hash1.Reset()
	hash1.Write(b)
	return hash1.Sum(nil)
}

func RIPEMD160(b []byte) []byte {
	hash160.Reset()
	hash160.Write(b)
	return hash160.Sum(nil)
}

func SHA256(b []byte) []byte {
	hash256.Reset()
	hash256.Write(b)
	return hash256.Sum(nil)
}

func HASH160(b []byte) []byte {
	v1 := SHA256(b)
	return RIPEMD160(v1)
}

func HASH256(b []byte) []byte {
	v1 := SHA256(b)
	return SHA256(v1)
}
