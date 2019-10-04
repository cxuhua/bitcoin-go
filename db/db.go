package db

import (
	"context"
	"sync"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	client     *mongo.Client = nil
	dbinitonce               = sync.Once{}
)

func InitDB(ctx context.Context) *mongo.Client {
	dbinitonce.Do(func() {
		c := options.Client().ApplyURI("mongodb://127.0.0.1:27017/")
		cptr, err := mongo.NewClient(c)
		if err != nil {
			panic(err)
		}
		err = cptr.Connect(ctx)
		if err != nil {
			panic(err)
		}
		client = cptr
	})
	return client
}

func UseTransaction(ctx context.Context, fn func(db DbImp) error) error {
	client = InitDB(ctx)
	return client.UseSession(ctx, func(sess mongo.SessionContext) error {
		_, err := sess.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (i interface{}, e error) {
			return nil, fn(NewDBImp(sessCtx))
		})
		return err
	})
}

func UseSession(ctx context.Context, fn func(db DbImp) error) error {
	client = InitDB(ctx)
	return client.UseSession(ctx, func(sess mongo.SessionContext) error {
		return fn(NewDBImp(sess))
	})
}
