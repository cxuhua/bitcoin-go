package net

import (
	"bitcoin/db"
	"bitcoin/script"
	"bitcoin/util"
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func TestSaveTXToDat(t *testing.T) {
	data := util.HexDecode("0200000001c5cc15a54bbb6e963d21153fe9320ccef11a1405de54a2bb0a6ef1b08ae015a5010000006a473044022074e0e8a1c4c1dd70f55ebea3ef2472727839168143afb2cfe7d6494ecda4a9ef022048492f3865b6572c97d88aa334e19c6a0947a11c794706fc34a63968d8c76e93012103addffda7b0dc0856fcf248ab830c61e66d24b368393811fe58f955f0aa32e8b6fdffffff02e803000000000000160014751e76e8199196d454941c45d1b3a323f1433bd6004b0000000000001976a9145f2c746007a3171893ea09a21e0d4f4307be2e1a88acc6090900")
	h := NewNetHeader(data)
	tx := &TX{}
	tx.Read(h)
	log.Println(tx.Hash, "save")
	file := fmt.Sprintf("../dat/tx%s.dat", tx.Hash)
	err := ioutil.WriteFile(file, data, os.ModePerm)
	if err != nil {
		t.Errorf("save error %v", err)
	}
}

func TestInfo(t *testing.T) {
	data2 := util.HexDecode("020000000001010000000000000000000000000000000000000000000000000000000000000000ffffffff4e0338220904129f9d5d535a30322f4254432e434f4d2ffabe6d6db782a41cbf92a6bca1f0b0e98bee0d8fabc2b7ff8e5a91b3815b7f054abe54e2080000007296cd1092eba16aef5e010000000000ffffffff0321d18d4a0000000016001497cfc76442fe717f2a3f0cc9c175f7561b6619970000000000000000266a24aa21a9ed995e3f01e0ad89d8add62981129645ca5518da2d8a6e507d06b900cbf3b61acc0000000000000000266a24b9e11b6de0b00267354a66a640f4f5322efd45e50eef89cd32a3298fb10f6c58dbd267670120000000000000000000000000000000000000000000000000000000000000000000000000")
	h2 := NewNetHeader(data2)
	tx2 := &TX{}
	tx2.Read(h2)
	log.Println(hex.EncodeToString(*tx2.Ins[0].Script))
}

func TestCloneTX(t *testing.T) {
	data2 := util.HexDecode("010000000135fb3eb02753c97977ace0ae5234a8f81045b486c826f3cd46f93bb90f003517010000006b483045022100eb2d8fba743470bacf2c567962530a6d7a3e43b0552c040039122e01cf0983fc022070ba55ec13937ab51dad7c9daa2fc8b59f7190e2debee3a6075ff1994c47f68b012102693bd134d8c1e753603e5fee11b08fc206b46d626a3267947c8923239c33d64dffffffff02e8230d00000000001976a91460823c20ec9fa15f2f7ed0955545982b67d1043288ac76eb37000000000017a914a48a80d30a56b276fb698e7c172f4d4c694f828e8700000000")
	h2 := NewNetHeader(data2)
	tx2 := &TX{}
	tx2.Read(h2)

	tx1 := tx2.Clone()

	if !tx1.Hash.Equal(tx2.Hash) {
		t.Errorf("clone tx error")
	}
}

func TestSaveTX(t *testing.T) {
	data2 := util.HexDecode("0100000001b6da357dfa24917f1f32414a10e2fcdad971a52521ecf84901f63612b0a103090000000048473044021f01ea6d57bf0373f8242ff263893ed396ea3de374439dfd80d59c2c1e6ab50a022100ee013faa0138a5f014b8f308befe2995ae3cb2e9fcd0763d2c7ca7b91f74435901ffffffff0247e8846d00000000434104a39b9e4fbd213ef24bb9be69de4a118dd0644082e47c01fd9159d38637b83fbcdc115a5d6e970586a012d1cfe3e3a8b1a3d04e763bdc5a071c0e827c0bd834a5acc0ac0d0d000000001976a9146ced4fd6ab06f237d185aebb11a7b4d92d7f8c8088ac00000000")
	h := NewNetHeader(data2)
	tx := &TX{}
	tx.Read(h)
	log.Println(tx.Hash, "save")
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

func TestPSHScript(t *testing.T) {
	d := util.HexDecode("00483045022100c121721a09160127ee2f6073e83c84d709ae8ebe6810f3893cb158f0635c8a0a02204fb2cb9d5ee693cf3df44660d3de63152396641fbb48c15be891c4289e06c0c201473044022037a2b5d78ecde26a72b2bc95056d82c2ea56ae55d692768c42e80f5f9d56bde7022079dc481aa4ba7e11644e9579eadd46e63f1af04a8e19be214dc7bfcafcb09a3301473044022008906911805539eabe566ea9b0b2de6aefb1993ef9bf671d4dd3e2357a15c5a5022025ce9c2409dfd198a6ef8b7015be3c6eae1b2b5726acfa737d8aef0fbe9420d5014d0b01534104220936c3245597b1513a9a7fe96d96facf1a840ee21432a1b73c2cf42c1810284dd730f21ded9d818b84402863a2b5cd1afe3a3d13719d524482592fb23c88a3410472225d3abc8665cf01f703a270ee65be5421c6a495ce34830061eb0690ec27dfd1194e27b6b0b659418d9f91baec18923078aac18dc19699aae82583561fefe54104a24db5c0e8ed34da1fd3b6f9f797244981b928a8750c8f11f9252041daad7b2d95309074fed791af77dc85abdd8bb2774ed8d53379d28cd49f251b9c08cab7fc4104dd26300a280a4c64bb42608d8cebe0d76705eda9f598a7a9945845f080f34788e6711ed7d786d3cc714aee44201d69a770f1caaf1558b8076398cbb0fc48241a54ae")
	s := script.NewScript(d)
	log.Println(s.HasMultiSig())
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
	if tx.Hash.String() != "c09b7a4be56da07d0e27fdaa465d9fc60f420e9216dbac70b714125729ca63fb" {
		t.Errorf("coinbase tx hashid error %v", tx.Hash)
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
	if tx.Hash.String() != id {
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

func TestBlockData(t *testing.T) {
	blockId := NewHexBHash("0000000000000000002a2451180749294cd74058e0a0dd37cc19ad0ee66e77ff")

	data, err := ioutil.ReadFile("../dat/block.dat")
	if err != nil {
		panic(err)
	}
	h := NewNetHeader(data)
	m := NewMsgBlock()
	m.Read(h)
	if !m.Hash.Equal(blockId) {
		t.Errorf("block hashid error")
	}
	if !m.Merkle.Equal(m.MarkleId()) {
		t.Errorf("equal markle error")
	}
}
