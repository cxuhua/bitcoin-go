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
	data := hextobytes("0048304502210080075aa29c42f8062f75cf6ab32004944417af974775581719008052c78719710220409fee54c6ddf2ca83e090077e443f95b427a63cc1ad87fca2625951b789d1c201493046022100b61d8f206d17efd6db32dad106f754f231ee8a16882929b1eb39a58bfd36b39e022100c62cff92dd6fb22b373025fc9b87044cf1b33502acc9de707e5f54d1c8a042a7014752210293baf0397588acc1aba056e868fd188dc0eea7554b45370aae862f9d2493a4c121020ab7517cf22a46b503ee8dcae7f9f109ec4cd19f0ab9d77c89c607554f3d5aa952ae")
	stack := NewStack()
	script := NewScript(data)
	ok, err := script.Eval(stack, nil, 0, SIG_VER_BASE)
	log.Println(ok, err)
}
