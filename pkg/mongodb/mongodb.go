package mongodb

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"sync"
	"time"
)

var (
	_ Connector = (*Client)(nil)
)

type Connector interface {
	GetClient() *Client
	SetDatabase(string)
	GetMongoClient() *mongo.Client
	GetDatabase() *mongo.Database
	GetCollection(string) *mongo.Collection
}

type Client struct {
	mongoClient *mongo.Client
	database    *mongo.Database
	mu          *sync.Mutex
}

// New create new object from mongo mongoClient
func New(ctx context.Context, uri string, mu *sync.Mutex, opts ...*options.ClientOptions) (Connector, error) {
	mu.Lock()
	defer mu.Unlock()
	opts = append(opts, options.Client().ApplyURI(uri))
	opts = append(opts, options.Client().SetConnectTimeout(2*time.Second))
	opts = append(opts, options.Client().SetMaxConnecting(500))
	cli, err := mongo.Connect(ctx, opts...)
	if err != nil {
		return nil, err
	}
	if err := cli.Ping(ctx, nil); err != nil {
		return nil, errors.Join(errors.New("mongodb: can't verify client connection "), err)
	}
	return &Client{mongoClient: cli, mu: mu}, nil
}

// GetClient return client
func (c *Client) GetClient() *Client {
	return c
}

// SetDatabase set database for client
func (c *Client) SetDatabase(dbName string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.database = c.mongoClient.Database(dbName)
}

// GetMongoClient return mongo client
func (c *Client) GetMongoClient() *mongo.Client {
	return c.mongoClient
}

// GetDatabase return database object
func (c *Client) GetDatabase() *mongo.Database {
	return c.database
}

// GetCollection return collection object
func (c *Client) GetCollection(collectionName string) *mongo.Collection {
	return c.database.Collection(collectionName)
}
