package internal

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Repository struct {
	client     *mongo.Client
	collection *mongo.Collection
}

func NewRepository(ctx context.Context, cfg *Config) (*Repository, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	db := client.Database(cfg.MongoDatabase)
	collection := db.Collection(cfg.MongoScrapeLogCollection)

	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    map[string]interface{}{"createdAt": -1},
		Options: options.Index().SetName("idx_created_at_desc"),
	})
	if err != nil {
		client.Disconnect(ctx)
		return nil, err
	}

	return &Repository{client: client, collection: collection}, nil
}

func (r *Repository) InsertLog(ctx context.Context, log *ScrapeLog) (string, error) {
	result, err := r.collection.InsertOne(ctx, log)
	if err != nil {
		return "", err
	}

	id, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return "", fmt.Errorf("unexpected inserted ID type: %T", result.InsertedID)
	}
	return id.Hex(), nil
}

func (r *Repository) Close(ctx context.Context) error {
	return r.client.Disconnect(ctx)
}
