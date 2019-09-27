package script

import "container/list"

type Stack struct {
	list *list.List
}

func NewStack() *Stack {
	return &Stack{
		list: list.New(),
	}
}

func (stack *Stack) Push(value interface{}) {
	stack.list.PushBack(value)
}

func (stack *Stack) Count(f func(v interface{}) bool) int {
	c := 0
	for e := stack.list.Front(); e != nil; e = e.Next() {
		if f(e.Value) {
			c++
		}
	}
	return c
}

func (stack *Stack) Pop() interface{} {
	e := stack.list.Back()
	if e != nil {
		stack.list.Remove(e)
		return e.Value
	}
	return nil
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
