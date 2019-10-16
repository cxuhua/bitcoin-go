package core

import (
	"bitcoin/store"
	"context"
	"testing"
)

//8d5bc6ff636d9cfb3a3b37cc2ad7681e5ba8078d8c7eb4a47531d75c18c8487f
func TestP2WPKHSign(t *testing.T) {
	err := store.UseSession(context.Background(), func(db store.DbImp) error {
		db.SetTXCacher(Fxs)
		id := NewHashID("8d5bc6ff636d9cfb3a3b37cc2ad7681e5ba8078d8c7eb4a47531d75c18c8487f")
		tx2, err := LoadTX(id, db)
		if err != nil {
			return err
		}
		return VerifyTX(tx2, db)
	})
	if err != nil {
		t.Errorf("Verify test failed  err=%v", err)
	}
}

func TestP2SHMSIGSign(t *testing.T) {
	err := store.UseSession(context.Background(), func(db store.DbImp) error {
		db.SetTXCacher(Fxs)
		id := NewHashID("c7f04832fc99b87a0140da2377ec81d1e1a062ed72f507f84533e572db1f6d15")
		tx2, err := LoadTX(id, db)
		if err != nil {
			return err
		}
		return VerifyTX(tx2, db)
	})
	if err != nil {
		t.Errorf("Verify test failed  err=%v", err)
	}
}

func TestP2WSHMSIGSign(t *testing.T) {
	err := store.UseSession(context.Background(), func(db store.DbImp) error {
		db.SetTXCacher(Fxs)
		id := NewHashID("2cc59f3c646b3917ed9b5224f71b335a2eab70ca4610a01dee90c2536d35d940")
		tx2, err := LoadTX(id, db)
		if err != nil {
			return err
		}
		return VerifyTX(tx2, db)
	})
	if err != nil {
		t.Errorf("Verify test failed  err=%v", err)
	}
}

func TestP2SHWPKHSign(t *testing.T) {
	err := store.UseSession(context.Background(), func(db store.DbImp) error {
		db.SetTXCacher(Fxs)
		id := NewHashID("0ae88f93be14b77994da8ebb948e817e6fbb98d66c0091366e46df0663ea3813")
		tx2, err := LoadTX(id, db)
		if err != nil {
			return err
		}
		return VerifyTX(tx2, db)
	})
	if err != nil {
		t.Errorf("Verify test failed  err=%v", err)
	}
}

func TestP2PKSign(t *testing.T) {
	err := store.UseSession(context.Background(), func(db store.DbImp) error {
		db.SetTXCacher(Fxs)
		id := NewHashID("80d417567b5a032465474052cca4dc38c57f6d5dc10dc7519b2ca20ac7d5512b")
		tx2, err := LoadTX(id, db)
		if err != nil {
			return err
		}
		return VerifyTX(tx2, db)
	})
	if err != nil {
		t.Errorf("Verify test failed  err=%v", err)
	}
}

func TestP2PKHSign(t *testing.T) {
	err := store.UseSession(context.Background(), func(db store.DbImp) error {
		db.SetTXCacher(Fxs)
		id := NewHashID("78470577b25f58e0b18fd21e57eb64c10eb66272a856208440362103de0f31da")
		tx2, err := LoadTX(id, db)
		if err != nil {
			return err
		}
		return VerifyTX(tx2, db)
	})
	if err != nil {
		t.Errorf("Verify test failed  err=%v", err)
	}
}
