package core

import (
	"sync"
	"testing"
)

func TestLock(t *testing.T) {
	l := sync.Mutex{}
	l.Lock()
	l.Unlock()
}
