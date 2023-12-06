package mongodb

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

type Transaction struct {
	client *mongo.Client
}

// NewTransaction create new transaction and session context
func NewTransaction(client *mongo.Client) *Transaction {
	return &Transaction{
		client: client,
	}
}

// CreateSession create new session for transaction
func (t *Transaction) CreateSession(sessionOpts ...*options.SessionOptions) (mongo.Session, error) {
	sess, err := t.client.StartSession(sessionOpts...)
	if err != nil {
		return nil, errors.Join(errors.New("mongodb: failed to start new session "), err)
	}

	return sess, nil
}

// RunTransaction run transaction base on transactionFunc
func (t *Transaction) RunTransaction(ctx context.Context, sessCtx mongo.SessionContext, transactionFunc func(sessCtx mongo.SessionContext) (interface{}, error)) (interface{}, error) {
	return sessCtx.WithTransaction(ctx, transactionFunc,
		options.Transaction().
			SetReadConcern(readconcern.Snapshot()).
			SetWriteConcern(writeconcern.Majority()))
}
