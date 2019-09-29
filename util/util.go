package util

import (
	"crypto/rand"
	"encoding/binary"
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

func MakeAddress(pk []byte) string {
	a := HASH160(pk)
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
