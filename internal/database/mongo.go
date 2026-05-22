package database

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoConfig struct {
	URI      string
	Database string
}

func ConnectMongo(ctx context.Context, config MongoConfig) (*mongo.Database, func(context.Context) error, error) {
	connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(connectCtx, options.Client().ApplyURI(config.URI))
	if err != nil {
		return nil, nil, err
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx, readpref.Primary()); err != nil {
		_ = client.Disconnect(ctx)
		return nil, nil, err
	}

	return client.Database(config.Database), client.Disconnect, nil
}

func EnsureMongoIndexes(ctx context.Context, db *mongo.Database) error {
	collections := []string{"employees", "applications", "clusters", "environments", "deployments", "pipeline_runs", "incidents"}
	for _, collectionName := range collections {
		collection := db.Collection(collectionName)
		_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys:    bson.D{{Key: "id", Value: 1}},
			Options: options.Index().SetUnique(true),
		})
		if err != nil {
			return err
		}
	}

	if _, err := db.Collection("deployments").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "application_id", Value: 1},
			{Key: "cluster_id", Value: 1},
			{Key: "environment", Value: 1},
		},
	}); err != nil {
		return err
	}

	if _, err := db.Collection("pipeline_runs").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "application_id", Value: 1},
			{Key: "branch", Value: 1},
			{Key: "status", Value: 1},
		},
	}); err != nil {
		return err
	}

	if _, err := db.Collection("environments").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "application_id", Value: 1},
			{Key: "cluster_id", Value: 1},
			{Key: "type", Value: 1},
		},
	}); err != nil {
		return err
	}

	_, err := db.Collection("incidents").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "status", Value: 1},
			{Key: "severity", Value: 1},
			{Key: "application_id", Value: 1},
		},
	})
	return err
}
