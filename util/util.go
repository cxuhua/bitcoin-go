package util

import (
	"bitcoin/config"
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"net"
	"strconv"
	"strings"
)

func ParseAddr(addr string) (net.IP, uint16) {
	conf := config.GetConfig()
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		panic(err)
	}
	iport := conf.ListenPort
	ip := net.ParseIP(host)
	pv, err := strconv.ParseInt(port, 10, 32)
	if err == nil {
		iport = int(pv)
	}
	return ip, uint16(iport)
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

//3
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

//return ver prefix and public hash160
func DecodeAddr(a string) (byte, []byte, error) {
	conf := config.GetConfig()
	if len(a) < 10 {
		return 0, nil, errors.New("a length error")
	}
	if a[:2] == conf.Bech32HRP {
		b, err := SegWitAddressDecode(a)
		if err != nil {
			return 0, nil, err
		}
		l := b[1]
		return b[0], b[2 : l+2], nil
	} else {
		b, err := B58Decode(a, BitcoinAlphabet)
		if err != nil {
			return 0, nil, err
		}
		c := b[:21]
		d := HASH256(c)
		if !bytes.Equal(d[:4], b[21:]) {
			return 0, nil, errors.New("check num error")
		}
		return c[0], c[1:21], nil
	}
}

//1
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

//bc

func BECH32Address(pk []byte) string {
	conf := config.GetConfig()
	ver := byte(0)
	pl := byte(len(pk))
	var a []byte = nil
	if len(pk) == 20 {
		//P2WPKH
		a = pk
	} else if len(pk) == 32 {
		//P2WSH
		a = pk
	} else {
		a = HASH160(pk)
	}
	b := []byte{ver, pl}
	b = append(b, a...)
	addr, err := SegWitAddressEncode(conf.Bech32HRP, b)
	if err != nil {
		panic(err)
	}
	return addr
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
