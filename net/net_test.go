package net

import (
	"bitcoin/script"
	"bitcoin/util"
	"bytes"
	"encoding/hex"
	"io/ioutil"
	"log"
	"testing"
)

func TestSign(t *testing.T) {
	data2 := util.HexDecode("010000000135fb3eb02753c97977ace0ae5234a8f81045b486c826f3cd46f93bb90f003517010000006b483045022100eb2d8fba743470bacf2c567962530a6d7a3e43b0552c040039122e01cf0983fc022070ba55ec13937ab51dad7c9daa2fc8b59f7190e2debee3a6075ff1994c47f68b012102693bd134d8c1e753603e5fee11b08fc206b46d626a3267947c8923239c33d64dffffffff02e8230d00000000001976a91460823c20ec9fa15f2f7ed0955545982b67d1043288ac76eb37000000000017a914a48a80d30a56b276fb698e7c172f4d4c694f828e8700000000")
	h2 := NewNetHeader(data2)
	tx2 := &TX{}
	tx2.Read(h2)
	log.Println(tx2.HashID())

	pub := new(script.PublicKey)
	err := pub.FromHEX("02693bd134d8c1e753603e5fee11b08fc206b46d626a3267947c8923239c33d64d")
	log.Println(pub.P2PKHAddress(), err)

	sig := &script.SigValue{}

	err = sig.FromHEX("3045022100eb2d8fba743470bacf2c567962530a6d7a3e43b0552c040039122e01cf0983fc022070ba55ec13937ab51dad7c9daa2fc8b59f7190e2debee3a6075ff1994c47f68b01")

	log.Println(hex.EncodeToString(sig.Encode()))

	pre := util.HexDecode("0100000001b5379f9fd6aba4b7bf5689622c84197b78f7c075d1157a98e4eaf891c59beef4000000006b483045022100f90a2cf0ef23261f8f36ccb75ff914c35d069aa2b633c0bb33d2f2165ee4ec060220071c2ac90deb8ee6eccb6f748d6496e0b0b0fee901dc0c4db551ef6b820a7df0012103cea3c970ee8d84a74454aab2121aaa6ba41d2e35aee9db6734f6f901c84e57d6ffffffff1a704607010000000017a9140ae0de681dd9ec84f4074b71d4109a98415f88fb8752124500000000001976a9145d29b8b0cb9ee7986de2cb1ec626e7399776d80e88ac986907010000000017a9148daa7788691d5d64df2661370df8000242b7b81e87102016000000000017a914095dc8927aeb1bbdcac56c226c49acc50c6b31e68720202b080000000017a91403c4d9440da5067cee631a5b5b4f447de9a30b7687becc2c00000000001976a914e4bbc1ce479c29052f3ee06bed654c79d27bf9a888ac50a06e000000000017a914d2f1b3e811c9c0358ecfb9db64b77196ddaf517587a0f52805000000001976a9146589b1f750bc1e3a2508b8b342543c6cb6a1756588acb0e460000000000017a914aaed3586bc83b6ab9cd96307a184411145327218873e564d040000000017a9140958fb1439b69d2fee0b714926408745e798d18f879871a3020000000017a914db038df59d86b1b04d72b300e9c285fec50c330b87a05e3c010000000017a914b15c324f727ebf19b2665078814c3a070dde10fc87e0ed15030000000017a914178b5e49f3a41c8699c8f8a8a051bc5f9ba5c86987f8e8d2000000000017a91409c2d384c7f88f0397de251fd810206f91db2f9887d06083000000000017a9147f246347e1cc17c98521325abf8cd34cfdf233a98728173400000000001976a914dd1ed7d8f6c18cdba1ca7a94416af435b8ed47d488ac60c50f020000000017a914b8dbc02941eeee9c19784be570df86e93d4de78d87740676000000000017a914c33c783bb0625dd9704595757c7ba06f6c3c537587251a7f040000000017a91420f6475f53a52094650f77521dca65690b00312087f0787300000000001976a914d68b9f805d09cb9bfb0fc93787b4402285bdf7c288acf8ed1f040000000017a91449bd0d011662b25f81fa9aa874cc8597e1d785da87138766000000000017a9148a2899e317005e7df1b62a16be43f1e181b29ebc8747cd3001000000001976a914e5e0d9fa2bea7808f7a5cc5ea6da7b6a491948fe88ac782314000000000017a914e3a384f186d478940c7d1c69ebd370f6a6a1dae387700c6a000000000017a914be961e3f6c0ca98f2a18a068a8e1d96692b1b5b9875f766410000000001976a9144136d4cacdb685eb7f33f9d4a77d52bc2fe9119088ac00000000")
	h1 := NewNetHeader(pre)
	pretx := &TX{}
	pretx.Read(h1)

	log.Println(tx2.Ins[0].OutHash, tx2.Ins[0].OutIndex)
	log.Println(pretx.HashID())

	stack := script.NewStack()
	tx2.Ins[0].Script.Eval(stack, nil, 0, script.SIG_VER_BASE)
	pretx.Outs[tx2.Ins[0].OutIndex].Script.Eval(stack, nil, 0, script.SIG_VER_BASE)

	tx2.Ins[0].Script = pretx.Outs[tx2.Ins[0].OutIndex].Script
	bv := tx2.SigBytes(sig.HashType)
	//bv = append(bv, 1)
	//prebytes := pretx.Outs[1].Script
	//bv = append(bv, *prebytes...)
	log.Println(hex.EncodeToString(bv))
	hash := util.HASH256(bv)
	log.Println(hex.EncodeToString(hash))
	log.Println("Verify=", pub.Verify(hash, sig))
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
	if err := tx.Check(oh, true); err != nil {
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
	if err := tx.Check(oh, true); err != nil {
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
	if err := tx.Check(oh, true); err != nil {
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
	if m.HashID().String() != blockId {
		t.Errorf("block hashid error")
	}
	for _, v := range m.Txs {
		if v.HashID().String() == "5b0d4175e73cc0e3c705f93269eb1714c269ad1c2ad11eff39f7a9fad180ed2e" {
			log.Println(v.Ins[0].OutIndex, v.Ins[0].OutHash)
			stack := script.NewStack()
			v.Ins[0].Script.Eval(stack, nil, 0, script.SIG_VER_BASE)
		}
	}
}
