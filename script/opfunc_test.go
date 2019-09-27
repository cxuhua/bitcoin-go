package script

import (
	"log"
	"testing"
)

func TestScriptNum(t *testing.T) {
	b := ScriptNum(-0x80).Serialize()
	log.Println(b, GetScriptNum(b) == -0x80)

	b = ScriptNum(0x80).Serialize()
	log.Println(b, GetScriptNum(b) == 0x80)
}
