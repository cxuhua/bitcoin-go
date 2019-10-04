package net

import (
	"bitcoin/db"
	"bitcoin/util"
	"bytes"
	"context"
	"encoding/hex"
	"io/ioutil"
	"log"
	"testing"
)

func TestInfo(t *testing.T) {
	data2 := util.HexDecode("0100000000010144c061cb929936fec8d6a7f7b56a0f8356d19714c6918ef6a041c5a1e52515190200000017160014bb010ceb86c2d2ded56254d8616c11e9a6386d7affffffff022d0f0f00000000001976a9147044778addfbfeac72cd0bcd670aa34c570e55dd88ac9217d90100000000160014107f25d6ed604a6759925269fe812616a12c495d0247304402204f8cf1bce4b45aa0aad03da7b7c731265bd56acc9a5a427b036a724b1481eacc022014e5d97ec7ccf5bdc5646ca01e3d059763c608e478633217aa407dce2044c8110121022fbd03ad4867759315df475e2de5f7143a9630c54735cb5f9f9960009fded97400000000")
	h2 := NewNetHeader(data2)
	tx2 := &TX{}
	tx2.Read(h2)
	log.Println(tx2)
}

func TestCloneTX(t *testing.T) {
	data2 := util.HexDecode("010000000135fb3eb02753c97977ace0ae5234a8f81045b486c826f3cd46f93bb90f003517010000006b483045022100eb2d8fba743470bacf2c567962530a6d7a3e43b0552c040039122e01cf0983fc022070ba55ec13937ab51dad7c9daa2fc8b59f7190e2debee3a6075ff1994c47f68b012102693bd134d8c1e753603e5fee11b08fc206b46d626a3267947c8923239c33d64dffffffff02e8230d00000000001976a91460823c20ec9fa15f2f7ed0955545982b67d1043288ac76eb37000000000017a914a48a80d30a56b276fb698e7c172f4d4c694f828e8700000000")
	h2 := NewNetHeader(data2)
	tx2 := &TX{}
	tx2.Read(h2)

	tx1 := tx2.Clone()

	if tx1.HashID() != tx2.HashID() {
		t.Errorf("clone tx error")
	}
}

func TestSaveTX(t *testing.T) {
	data2 := util.HexDecode("0100000001b6da357dfa24917f1f32414a10e2fcdad971a52521ecf84901f63612b0a103090000000048473044021f01ea6d57bf0373f8242ff263893ed396ea3de374439dfd80d59c2c1e6ab50a022100ee013faa0138a5f014b8f308befe2995ae3cb2e9fcd0763d2c7ca7b91f74435901ffffffff0247e8846d00000000434104a39b9e4fbd213ef24bb9be69de4a118dd0644082e47c01fd9159d38637b83fbcdc115a5d6e970586a012d1cfe3e3a8b1a3d04e763bdc5a071c0e827c0bd834a5acc0ac0d0d000000001976a9146ced4fd6ab06f237d185aebb11a7b4d92d7f8c8088ac00000000")
	h := NewNetHeader(data2)
	tx := &TX{}
	tx.Read(h)
	err := db.UseSession(context.Background(), func(db db.DbImp) error {
		err := tx.Save(db)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		t.Errorf("test save tx error %v", err)
	}
}

func TestSign(t *testing.T) {
	err := db.UseSession(context.Background(), func(db db.DbImp) error {
		id := NewHexBHash("78470577b25f58e0b18fd21e57eb64c10eb66272a856208440362103de0f31da")
		tx2, err := LoadTX(id, db)
		if err != nil {
			return err
		}
		return tx2.Verify(db)
	})
	if err != nil {
		t.Errorf("Verify test failed  %v", err)
	}
}

func TestCoinBaseTX(t *testing.T) {
	data, err := hex.DecodeString("010000000001010000000000000000000000000000000000000000000000000000000000000000ffffffff60036872081d2f5669614254432f4d696e65642062792031383539313737393733372f2cfabe6d6d980e95c6480064939d0a9f0bc8fa2318fc30fb5373cf2e1418f4830d7c69ab850200000000000000107a6e3608bad2c4e31e92165d8a5b0200ffffffff02a9845f4b000000001976a914536ffa992491508dca0354e52f32a3a7a679a53a88ac0000000000000000266a24aa21a9ed8719f26a8751c9ce44f62ae9903804f82f4c80f73d63dbcbe4772092037729960120000000000000000000000000000000000000000000000000000000000000000000000000")
	if err != nil {
		panic(err)
	}
	h := NewNetHeader(data)
	tx := &TX{}
	tx.Read(h)
	if !tx.IsCoinBase() {
		t.Errorf("coinbase tx check error")
	}
	if tx.HashID().String() != "c09b7a4be56da07d0e27fdaa465d9fc60f420e9216dbac70b714125729ca63fb" {
		t.Errorf("coinbase tx hashid error")
	}
	oh := NewNetHeader()
	tx.Write(oh)
	if err := tx.Check(); err != nil {
		t.Errorf("check tx error %v", err)
	}
}

func TestTXID(t *testing.T) {
	id := "246ab84ed7e0187e395c796c0d05fe083b7379b6c6ab29edc19626a039004bcf"
	data, err := hex.DecodeString("02000000027cb7fef1b9192c534cc539f2a1e949779d49438d0c860132cfb19d61cc655bc2010000006a473044022025bb9bd14ac3bf2b627b7d2f762f6322e8bc503779a25f2d0200b0d16f91f920022052847e5abb8ffd57738de52b41430de0b06d3363ff19ade8fcb8adadc7ef316a0121020eb301f186eca040fe127fff26ae692894c07586fd94bd26db5f77ae08ce4818feffffff43957c51908895ea828b8077661f2c74e347c84424e5761a1a5da4c7e10a1348010000006a47304402200468722bd21014cb8bf67bd975e1bd1ac7efa2158b5feeb15f12ff945f40486302202c54469049b4f22e6fe30790f2336ebe783a3a4eaef05b168fb38341a6f4ce870121020eb301f186eca040fe127fff26ae692894c07586fd94bd26db5f77ae08ce4818feffffff02800808000000000017a9148f0ac79dfd0e097aef933f7b54b65d5eee72dc4087b0f91e000000000017a91435e9c35af6c5958e6f5c80575afb94b2cedec0fa877d1c0900")
	if err != nil {
		panic(err)
	}
	h := NewNetHeader(data)
	tx := &TX{}
	tx.Read(h)
	if tx.HashID().String() != id {
		t.Errorf("make tx hashid error")
	}

}

func TestTXWNoWitness(t *testing.T) {
	data, err := hex.DecodeString("0100000001ec8abe57241834aab404a43288ce8c57061dfc47fe68312e43e2baddb01dfcfe010000006b483045022100b0b297b340fc5f42640b1dd9c96d58a4c19616045b7c921e0a0177eabec1f06402201f5efe9c400b1b5005f82ba57dd653abfcda4974bbabe00dc88e1e5b7c30c74b01210346aa8f3d640d829c7d6d40243d5e052a316ed5186bebed8da300e3a2e676ce3ffdffffff011ca0bf06000000001976a914975aa8ed4bd63ed0f388f40b4a3533425e42909f88ac00000000")
	if err != nil {
		panic(err)
	}
	h := NewNetHeader(data)
	tx := &TX{}
	tx.Read(h)
	if !h.IsEOF() {
		t.Errorf("read bytes remain")
	}
	if len(tx.Ins) != 1 {
		t.Errorf("txin error")
	}
	if len(tx.Outs) != 1 {
		t.Errorf("txout error")
	}
	if tx.LockTime != 0x0 {
		t.Errorf("lock time error")
	}
	oh := NewNetHeader()
	tx.Write(oh)
	if !bytes.Equal(data, oh.Payload) {
		t.Errorf("write tx error")
	}
	if err := tx.Check(); err != nil {
		t.Errorf("check tx error %v", err)
	}
}

func TestTXWithWitness(t *testing.T) {
	data, err := hex.DecodeString("02000000027cb7fef1b9192c534cc539f2a1e949779d49438d0c860132cfb19d61cc655bc2010000006a473044022025bb9bd14ac3bf2b627b7d2f762f6322e8bc503779a25f2d0200b0d16f91f920022052847e5abb8ffd57738de52b41430de0b06d3363ff19ade8fcb8adadc7ef316a0121020eb301f186eca040fe127fff26ae692894c07586fd94bd26db5f77ae08ce4818feffffff43957c51908895ea828b8077661f2c74e347c84424e5761a1a5da4c7e10a1348010000006a47304402200468722bd21014cb8bf67bd975e1bd1ac7efa2158b5feeb15f12ff945f40486302202c54469049b4f22e6fe30790f2336ebe783a3a4eaef05b168fb38341a6f4ce870121020eb301f186eca040fe127fff26ae692894c07586fd94bd26db5f77ae08ce4818feffffff02800808000000000017a9148f0ac79dfd0e097aef933f7b54b65d5eee72dc4087b0f91e000000000017a91435e9c35af6c5958e6f5c80575afb94b2cedec0fa877d1c0900")
	if err != nil {
		panic(err)
	}
	h := NewNetHeader(data)
	tx := &TX{}
	tx.Read(h)
	if !h.IsEOF() {
		t.Errorf("read bytes remain")
	}
	if len(tx.Ins) != 2 {
		t.Errorf("txin error")
	}
	if len(tx.Outs) != 2 {
		t.Errorf("txout error")
	}
	if tx.LockTime != 0x91c7d {
		t.Errorf("lock time error")
	}
	oh := NewNetHeader()
	tx.Write(oh)
	if !bytes.Equal(data, oh.Payload) {
		t.Errorf("write tx error")
	}
	if err := tx.Check(); err != nil {
		t.Errorf("check tx error %v", err)
	}
}

//0595dcce885d55287483900d91dee54cd85f9325bfd053c948801c32f3e3edee
func TestGetOutput(t *testing.T) {
	s1 := "010000000001010000000000000000000000000000000000000000000000000000000000000000ffffffff5f03661c0904236d905d2f706f6f6c696e2e636f6d2ffabe6d6dfb14f40eab377f96913454bd1560dd9bed05d1343bc47e048648376d82029fc20100000000000000ccc3a91362229b793b2b18a48ad829190ed9bdf9f100aed13a00ffffffffffffffff035755dd4b0000000017a914b111f00eed1a8123dd0a1fed50b0793229ed47e7870000000000000000266a24b9e11b6ded607dba8ae11c44120a5aa1e209b1e95dbb35ef75d6f99f45ca7ecc0e4c6f230000000000000000266a24aa21a9ed8a7fa35398c81d6bafcfcf5013d96fe3c3b8b90efe8785709971529ee0588cb10120000000000000000000000000000000000000000000000000000000000000000000001e89"
	data, err := hex.DecodeString(s1)
	if err != nil {
		panic(err)
	}
	h1 := NewNetHeader(data)
	tx1 := &TX{}
	tx1.Read(h1)
	idx := tx1.Ins[0].OutIndex
	//log.Println(tx1.Ins[0].OutHash, idx)
	script1 := tx1.Ins[0].Script
	log.Println(hex.EncodeToString(*script1))

	//ins[0] output tx
	s2 := "0100000002227bf9487f0716e9bcf81d42343b9d31435f8cf24fbcc766ebb4b7cf7213d9aa01000000db0048304502201c719cb0c5030845ae4fcaf880ba8c9d9bc33a4048d4bc61f6c8768d34722235022100e97231f5fc703ffe22a1c3e375639804a452ccf0fcff8b153e12da66e82d79980148304502206dc0723ef31a7a24b130ccae2e0385dac01d30fec30ea99e70e280502213ef30022100aee1330040fb75b91ec58a8bf2d8191a195f610f5823c46411a4c34f03e4c938014752210293baf0397588acc1aba056e868fd188dc0eea7554b45370aae862f9d2493a4c121020ab7517cf22a46b503ee8dcae7f9f109ec4cd19f0ab9d77c89c607554f3d5aa952aeffffffff3040c42258489d633033906125dc9999a8b0fadebde325db13c7fa6a0126f1b3580000006a473044022022f21d6bbd933ff799586e482f91e682007eb4ed4d6f9d4cb3b3e9e1a82b3340022073f50be36c0a37541cabd84c3defb43c9a6e2cc0dfe3757d40ac4aeef3c02e2701210203635e5c184951e14fcfecc83b15960594f4fceec729e09a4a517b0a03a7f4b9ffffffff0260216000000000001976a914e46841d71f89d9a1b97dfe086bcd92615cda5ae688aca75cc4220000000017a914622854939d571b63df97f47e8302b700ab2932b68700000000"
	data, err = hex.DecodeString(s2)
	if err != nil {
		panic(err)
	}
	h2 := NewNetHeader(data)
	tx2 := &TX{}
	tx2.Read(h2)
	//log.Println(tx2.HashID())
	script2 := tx2.Outs[idx].Script
	log.Println(hex.EncodeToString(*script2))
}

func TestH1B(t *testing.T) {
	data, err := ioutil.ReadFile("h1b.dat")
	if err != nil {
		panic(err)
	}
	h := NewNetHeader(data)
	m := NewMsgBlock()
	m.Read(h)
}

func TestBlockData(t *testing.T) {
	blockId := "0000000000000000002a2451180749294cd74058e0a0dd37cc19ad0ee66e77ff"
	data, err := ioutil.ReadFile("../dat/block.dat")
	if err != nil {
		panic(err)
	}
	h := NewNetHeader(data)
	m := NewMsgBlock()
	m.Read(h)
	if m.Hash.String() != blockId {
		t.Errorf("block hashid error")
	}
	if !m.Merkle.Equal(m.NewMarkle()) {
		t.Errorf("equal markle error")
	}
}
