package script

import (
	"container/list"
	"errors"
)

type Value []byte

func NewValueBool(v bool) Value {
	if v {
		return []byte{1, 1}
	} else {
		return []byte{0}
	}
}

func (v Value) ToBytes() []byte {
	return v
}

func (v Value) Len() int {
	return len(v)
}

func (v Value) ToBool() bool {
	return CastToBool(v)
}

func (v Value) ToInt(mini bool, siz int) int {
	return v.ToScriptNum(mini, siz).ToInt()
}

func (v Value) ToScriptNum(mini bool, siz int) ScriptNum {
	return GetScriptNum(v.ToBytes(), mini, siz)
}

type Stack struct {
	list *list.List
}

func NewStack() *Stack {
	return &Stack{
		list: list.New(),
	}
}

func (stack *Stack) InsertAfter(v interface{}, e *list.Element) {
	stack.list.InsertAfter(v, e)
}

func (stack *Stack) InsertBefore(v interface{}, e *list.Element) {
	stack.list.InsertBefore(v, e)
}

func (stack *Stack) Push(value Value) {
	stack.list.PushBack(value)
}

func (stack *Stack) Count(f func(v Value) bool) int {
	c := 0
	for e := stack.list.Front(); e != nil; e = e.Next() {
		if f(e.Value.(Value)) {
			c++
		}
	}
	return c
}

func (stack *Stack) TopElement(idx int) *list.Element {
	e := stack.list.Back()
	if e == nil {
		return nil
	}
	for idx < -1 && e != nil {
		e = e.Prev()
		idx++
	}
	if e == nil {
		return nil
	}
	return e
}

func (stack *Stack) Top(idx int) Value {
	e := stack.list.Back()
	if e == nil {
		return nil
	}
	for idx < -1 && e != nil {
		e = e.Prev()
		idx++
	}
	if e == nil {
		return nil
	}
	return e.Value.(Value)
}
func (stack *Stack) EraseIndex(idx int) {
	stack.EraseRange(idx, idx)

}

func (stack *Stack) EraseRange(from int, to int) {
	if from < 0 {
		from = stack.Len() + from
	}
	if to < 0 {
		to = stack.Len() + to
	}
	if from > to {
		panic(errors.New("from > to"))
	}
	if from < 0 || from >= stack.Len() {
		panic(errors.New("from outbound"))
	}
	if to < 0 || to >= stack.Len() {
		panic(errors.New("to outbound"))
	}
	ds := []*list.Element{}
	var e *list.Element
	num := to - from
	pos := from
	e = stack.list.Front()
	for pos > 0 {
		e = e.Next()
		pos--
	}
	ds = append(ds, e)
	for num > 0 {
		e = e.Next()
		ds = append(ds, e)
		num--
	}
	for _, v := range ds {
		stack.list.Remove(v)
	}
}

func (stack *Stack) Pop() Value {
	e := stack.list.Back()
	if e != nil {
		stack.list.Remove(e)
		return e.Value.(Value)
	}
	return nil
}

func (stack *Stack) Back() *list.Element {
	return stack.list.Back()
}

func (stack *Stack) Peak() interface{} {
	e := stack.list.Back()
	if e != nil {
		return e.Value
	}
	return nil
}

func (stack *Stack) Len() int {
	return stack.list.Len()
}

func (stack *Stack) Empty() bool {
	return stack.list.Len() == 0
}
