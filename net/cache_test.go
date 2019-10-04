package net

import (
	"testing"
	"time"
)

func TestTXCacher(t *testing.T) {
	c := NewMemoryCacher(10, time.Second*3)
	id1 := HashID{0, 1, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0}
	tx1 := &TX{}
	if err := c.Set(id1, tx1); err != nil {
		t.Errorf("set error %v", err)
	}
	if v, err := c.Get(id1); err != nil {
		t.Errorf("get error %v", err)
	} else if v != tx1 {
		t.Errorf("get value error %v", err)
	}
	c.Del(id1)
	if _, err := c.Get(id1); err == nil {
		t.Errorf("del get error %v", err)
	}
	if err := c.Set(id1, tx1); err != nil {
		t.Errorf("set error %v", err)
	}
	c.Clean()
	if _, err := c.Get(id1); err != nil {
		t.Errorf("Clean 1 get error %v", err)
	}
	time.Sleep(time.Second * 4)
	//c.Clean()
	if _, err := c.Get(id1); err == nil {
		t.Errorf("Clean 2 get error %v", err)
	}
}
