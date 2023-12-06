package migration

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// AddNormalIndex create normal index for migration
func AddNormalIndex(ctx context.Context, collection *mongo.Collection, field string) error {
	opt := options.Index().SetName(fmt.Sprintf("%s_%s_normal", collection.Name(), field)).SetUnique(false)
	keys := bson.D{{field, 1}}
	model := mongo.IndexModel{Keys: keys, Options: opt}
	_, err := collection.Indexes().CreateOne(ctx, model)
	return err
}

// AddUniqueIndex create unique index, if sparse true the index will only reference documents that contain the fields specified in the index
func AddUniqueIndex(ctx context.Context, collection *mongo.Collection, field string, sparse bool) error {
	opt := options.Index().SetName(fmt.Sprintf("%s_%s_unique", collection.Name(), field)).SetUnique(true)
	if sparse {
		opt = opt.SetSparse(true)
	}
	keys := bson.D{{field, 1}}
	model := mongo.IndexModel{Keys: keys, Options: opt}
	_, err := collection.Indexes().CreateOne(ctx, model)
	return err
}
