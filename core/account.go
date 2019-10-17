package core

type AccountKey struct {
	Private []byte `bson:"private"`
	Public  []byte `bson:"public"`
}

type Account struct {
	Addr string       `bson:"_id"`
	KOpt int          `bson:"kopt"` //mulsig use 2-3sig opt=2
	Keys []AccountKey `bson:"keys"`
}
