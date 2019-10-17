package core

import "sync"

type TxMap struct {
	mu  sync.RWMutex
	txs map[HashID]*TX
}

func NewTxMap() *TxMap {
	return &TxMap{
		txs: map[HashID]*TX{},
	}
}

var (
	TxsMap = NewTxMap()
)

func (m *TxMap) Has(id HashID) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.txs[id]
	return ok
}

func (m *TxMap) Del(id HashID) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.txs, id)
}

func (m *TxMap) Set(tx *TX) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.txs[tx.Hash] = tx
	return len(m.txs)
}

func (m *TxMap) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.txs)
}
