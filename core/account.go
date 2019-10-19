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

//sync money record
func (b *MsgBlock) SyncMoneys(sdb store.DbImp) error {
	for _, v := range b.Txs {
		for iidx, in := range v.Ins {
			if in.OutHash.IsZero() {
				continue
			}
			out, err := in.OutTx(sdb)
			if err != nil {
				return err
			}
			sv := out.ToSubMoneys(v.Hash, uint32(iidx))
			if sv == nil {
				continue
			}
			if v.IsCoinBase() && sdb.HasMT(sv.Id) {
				sv = sv.LoseMoney()
			}
			if err := sdb.SetMT(sv.Id, sv); err != nil {
				return err
			}
		}
		for oidx, out := range v.Outs {
			if out.Value == 0 {
				continue
			}
			sv := out.ToAddMoneys(v.Hash, uint32(oidx))
			if sv == nil {
				continue
			}
			if v.IsCoinBase() && sdb.HasMT(sv.Id) {
				sv = sv.LoseMoney()
			}
			if err := sdb.SetMT(sv.Id, sv); err != nil {
				return err
			}
		}
	}
	return nil
}
