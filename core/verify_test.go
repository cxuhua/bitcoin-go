package core

import (
	"testing"
)

//8d5bc6ff636d9cfb3a3b37cc2ad7681e5ba8078d8c7eb4a47531d75c18c8487f
func TestP2WPKHSign(t *testing.T) {
	id := NewHashID("8d5bc6ff636d9cfb3a3b37cc2ad7681e5ba8078d8c7eb4a47531d75c18c8487f")
	tx2, err := LoadTx(id)
	if err != nil {
		t.Errorf("load tx error %v", err)
	}
	for i, v := range tx2.Outs {
		if i == 0 && v.Script.GetAddress() != "395MUYNnnhaUDhm4VDSKn7jtafQbU5kXRB" {
			t.Errorf("get out %d address", i)
		}
	}
	if err := VerifyTX(tx2, 0); err != nil {
		t.Errorf("verify tx error %v", err)
	}
}

//
//func TestP2SHMSIGSign(t *testing.T) {
//	err := store.UseSession(context.Background(), func(db store.DbImp) error {
//		db.PushTxCacher(Fxs)
//		id := NewHashID("c7f04832fc99b87a0140da2377ec81d1e1a062ed72f507f84533e572db1f6d15")
//		tx2, err := LoadTX(id, db)
//		if err != nil {
//			return err
//		}
//		for i, v := range tx2.Outs {
//			if i == 0 && v.Script.GetAddress() != "3AAhq47sBv78RWNTWF5vsAeDdWmA2EqV88" {
//				return fmt.Errorf("get out %d address", i)
//			}
//			if i == 1 && v.Script.GetAddress() != "3BMEXQxztwFkN3E6FSf3VuGNTeUQzG41Vf" {
//				return fmt.Errorf("get out %d address", i)
//			}
//		}
//		return VerifyTX(tx2, db, 0)
//	})
//	if err != nil {
//		t.Errorf("Verify test failed  err=%v", err)
//	}
//}
//
//func TestP2WSHMSIGSign(t *testing.T) {
//	err := store.UseSession(context.Background(), func(db store.DbImp) error {
//		db.PushTxCacher(Fxs)
//		id := NewHashID("2cc59f3c646b3917ed9b5224f71b335a2eab70ca4610a01dee90c2536d35d940")
//		tx2, err := LoadTX(id, db)
//		if err != nil {
//			return err
//		}
//		for i, v := range tx2.Outs {
//			if i == 0 && v.Script.GetAddress() != "3EMvHQQrqHuX8vDBtW6SATSdVYPX2Yc529" {
//				return fmt.Errorf("get out %d address", i)
//			}
//			if i == 1 && v.Script.GetAddress() != "bc1qwqdg6squsna38e46795at95yu9atm8azzmyvckulcc7kytlcckxswvvzej" {
//				return fmt.Errorf("get out %d address", i)
//			}
//		}
//		return VerifyTX(tx2, db, 0)
//	})
//	if err != nil {
//		t.Errorf("Verify test failed  err=%v", err)
//	}
//}
//
//func TestP2SHWPKHSign(t *testing.T) {
//	err := store.UseSession(context.Background(), func(db store.DbImp) error {
//		db.PushTxCacher(Fxs)
//		id := NewHashID("0ae88f93be14b77994da8ebb948e817e6fbb98d66c0091366e46df0663ea3813")
//		tx2, err := LoadTX(id, db)
//		if err != nil {
//			return err
//		}
//		for i, v := range tx2.Outs {
//			if i == 0 && v.Script.GetAddress() != "3GDiJ4gRqnzAws1bFvkBwimh8Pykx5cUPi" {
//				return fmt.Errorf("get out %d address", i)
//			}
//			if i == 1 && v.Script.GetAddress() != "3FAX1sAtk1NTVpjLuNYJEp9D489ZvrRsvW" {
//				return fmt.Errorf("get out %d address", i)
//			}
//		}
//		return VerifyTX(tx2, db, 0)
//	})
//	if err != nil {
//		t.Errorf("Verify test failed  err=%v", err)
//	}
//}
//
//func TestP2PKSign(t *testing.T) {
//	err := store.UseSession(context.Background(), func(db store.DbImp) error {
//		db.PushTxCacher(Fxs)
//		id := NewHashID("80d417567b5a032465474052cca4dc38c57f6d5dc10dc7519b2ca20ac7d5512b")
//		tx2, err := LoadTX(id, db)
//		if err != nil {
//			return err
//		}
//		for i, v := range tx2.Outs {
//			if i == 0 && v.Script.GetAddress() != "1VayNert3x1KzbpzMGt2qdqrAThiRovi8" {
//				return fmt.Errorf("get out %d address", i)
//			}
//			if i == 1 && v.Script.GetAddress() != "1AvxGSFo8sVJKkfwHhtt6stHyuKUyLaKZp" {
//				return fmt.Errorf("get out %d address", i)
//			}
//		}
//		return VerifyTX(tx2, db, 0)
//	})
//	if err != nil {
//		t.Errorf("Verify test failed  err=%v", err)
//	}
//}
//
//func TestP2PKHSign(t *testing.T) {
//	err := store.UseSession(context.Background(), func(db store.DbImp) error {
//		db.PushTxCacher(Fxs)
//		id := NewHashID("78470577b25f58e0b18fd21e57eb64c10eb66272a856208440362103de0f31da")
//		tx2, err := LoadTX(id, db)
//		if err != nil {
//			return err
//		}
//		for i, v := range tx2.Outs {
//			if i == 0 && v.Script.GetAddress() != "1MX1S4dniXHPJdySEszvM42nYryJx6NPgG" {
//				return fmt.Errorf("get out %d address", i)
//			}
//			if i == 1 && v.Script.GetAddress() != "1k2saXX9kkxcSpx5W9yDMbuzRLsGfE71W" {
//				return fmt.Errorf("get out %d address", i)
//			}
//		}
//		return VerifyTX(tx2, db, 0)
//	})
//	if err != nil {
//		t.Errorf("Verify test failed  err=%v", err)
//	}
//}
