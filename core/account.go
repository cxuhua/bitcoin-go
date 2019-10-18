package core

import (
	"bitcoin/store"
)

type AccountKey struct {
	Private []byte `bson:"private"`
	Public  []byte `bson:"public"`
}

type Account struct {
	Addr string       `bson:"_id"`
	KOpt int          `bson:"kopt"` //mulsig use 2-3sig opt=2
	Keys []AccountKey `bson:"keys"`
}

func AvailableBlockComing(sdb store.DbImp, b *MsgBlock) error {
	for _, v := range b.Txs {
		for _, in := range v.Ins {
			if in.OutHash.IsZero() {
				continue
			}
			oid := NewMoneyId(in.OutHash, in.OutIndex)
			if err := sdb.DelMT(oid); err != nil {
				return err
			}
		}
		for oidx, out := range v.Outs {
			if out.Value == 0 {
				continue
			}
			sv := out.ToMoneys(v.Hash, uint32(oidx))
			if err := sdb.SetMT(sv.Id, sv); err != nil {
				return err
			}
		}
	}
	return nil
}
