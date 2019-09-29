package script

import (
	"testing"
)

func TestBoolStack(t *testing.T) {
	bl := NewStack()
	bl.Push(NewValueBool(false))
	bl.Push(NewValueBool(false))
	bl.Push(NewValueBool(true))
	bl.Push(NewValueBool(true))
	bl.Push(NewValueBool(true))

	n := bl.Count(func(v Value) bool {
		return v.ToBool() == true
	})
	if n != 3 {
		t.Error("count true error")
	}
	n = bl.Count(func(v Value) bool {
		return v.ToBool() == false
	})
	if n != 2 {
		t.Error("count true error")
	}
	bl.EraseIndex(3)

	if bl.Len() != 4 {
		t.Error("len failed")
	}

	n = bl.Count(func(v Value) bool {
		return v.ToBool() == true
	})
	if n != 2 {
		t.Error("count true error")
	}
	n = bl.Count(func(v Value) bool {
		return v.ToBool() == false
	})
	if n != 2 {
		t.Error("count true error")
	}
}
