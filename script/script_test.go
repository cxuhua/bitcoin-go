package script

import (
	"encoding/hex"
	"log"
	"testing"
)

func hextobytes(s string) []byte {
	d, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return d
}

func TestScript(t *testing.T) {
	data := hextobytes("47304402204c086d889102a00f695e41cf179cbb56f5c7387867c400e5b67c2de13f5ca09902204d0584dcff68fdd51310dc53e05b949d6ef57de968a514af0087d86ed68147360121034b548001bb4f241648e42178e53342eb9d7256f31bcc4364b1d5cb90aa85dc46")
	stack := NewStack()
	script := NewScript(data)
	ok, err := script.Eval(stack, nil, 0, SIG_VER_BASE)
	log.Println(ok, err)
}
