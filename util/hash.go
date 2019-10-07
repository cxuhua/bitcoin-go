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

func SHA1(b []byte) []byte {
	hv := sha1.Sum(b)
	return hv[:]
}

func RIPEMD160(b []byte) []byte {
	h160 := ripemd160.New()
	h160.Write(b)
	return h160.Sum(nil)
}

func SHA256(b []byte) []byte {
	hash := sha256.Sum256(b)
	return hash[:]
}

func HASH160(b []byte) []byte {
	v1 := SHA256(b)
	return RIPEMD160(v1)
}

func HASH256(b []byte) []byte {
	s2 := sha256.New()
	s2.Write(b)
	v1 := s2.Sum(nil)
	s2.Reset()
	s2.Write(v1)
	return s2.Sum(nil)
}
