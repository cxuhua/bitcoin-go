package util

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"net"
	"strconv"
	"strings"
)

func ParseAddr(addr string) (net.IP, uint16) {
	ip := net.IP{}
	port := uint16(0)
	vs := strings.Split(addr, ":")
	if len(vs) > 0 {
		ip = net.ParseIP(vs[0])
	}
	if len(vs) > 1 {
		iv, err := strconv.ParseInt(vs[1], 10, 32)
		if err != nil {
			panic(err)
		}
		port = uint16((iv))
	}
	return ip, port
}

func HexToBytes(s string) []byte {
	d, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return d
}

func HexDecode(s string) []byte {
	s = strings.Replace(s, "_", "", -1)
	d, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return d
}

//Pay-to-Script-Hash
func P2SHAddress(pk []byte) string {
	var a []byte = nil
	if len(pk) == 20 {
		a = pk
	} else {
		a = HASH160(pk)
	}
	b := []byte{5}
	b = append(b, a...)
	c := HASH256(b)
	b = append(b, c[:4]...)
	return B58Encode(b, BitcoinAlphabet)
}

//Pay-to-Public-Key-Hash
func P2PKHAddress(pk []byte) string {
	var a []byte = nil
	if len(pk) == 20 {
		a = pk
	} else {
		a = HASH160(pk)
	}
	b := []byte{0}
	b = append(b, a...)
	c := HASH256(b)
	b = append(b, c[:4]...)
	return B58Encode(b, BitcoinAlphabet)
}

func RandUInt64() uint64 {
	v := uint64(0)
	if err := binary.Read(rand.Reader, binary.LittleEndian, &v); err != nil {
		panic(err)
	}
	return v
}

func String(b []byte) string {
	for idx, c := range b {
		if c == 0 {
			return string(b[:idx])
		}
	}
	return string(b)
}
