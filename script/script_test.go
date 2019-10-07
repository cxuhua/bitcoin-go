package script

import (
	"encoding/hex"
)

func hextobytes(s string) []byte {
	d, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return d
}
