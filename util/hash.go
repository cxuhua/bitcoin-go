package util

import "crypto/sha256"

//data checksum
func Hash(b []byte) []byte {
	v1 := sha256.Sum256(b)
	v2 := sha256.Sum256(v1[:])
	return v2[:]
}

//prefix 4 bytes
func HashP4(b []byte) []byte {
	v := Hash(b)
	return v[:4]
}
