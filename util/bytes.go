package util

import "bytes"

type BytesArray [][]byte

func (b *BytesArray) Append(v []byte) {
	*b = append(*b, v)
}

func (b *BytesArray) Has(v []byte) bool {
	for _, vv := range *b {
		if bytes.Equal(vv, v) {
			return true
		}
	}
	return false
}

func NewBytesArray() *BytesArray {
	return &BytesArray{}
}
