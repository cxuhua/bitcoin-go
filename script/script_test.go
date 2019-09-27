package script

import (
	"log"
	"testing"
)

func TestScript(t *testing.T) {
	s := "483045022100fd2a947a1e8a399d32f391d3b4bb143397dea466673b373c1baf6ab35caef52f0220550dee70c0c767f5966763dd1a184bee39b8ff8cf55c864678af762d6455a92d0121024e2e8a57c1251868885aaa27da8aa16c040c3b9d7751b1b277f31208761afd06"
	script := NewScriptHex(s)
	log.Println(script.HasValidOps())
}
