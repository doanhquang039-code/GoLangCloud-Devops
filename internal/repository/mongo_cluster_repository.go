package repository

import (
	"context"

	"hr-cloud-service/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoClusterRepository struct {
	collection *mongo.Collection
}

func NewMongoClusterRepository(db *mongo.Database) *MongoClusterRepository {
	return &MongoClusterRepository{
		collection: db.Collection("clusters"),
	}
}

func (r *MongoClusterRepository) FindAll(ctx context.Context) ([]model.Cluster, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var clusters []model.Cluster
	if err := cursor.All(ctx, &clusters); err != nil {
		return nil, err
	}

	return clusters, nil
}

func (r *MongoClusterRepository) FindByID(ctx context.Context, id string) (model.Cluster, error) {
	var cluster model.Cluster
	err := r.collection.FindOne(ctx, bson.M{"id": id}).Decode(&cluster)
	if err == mongo.ErrNoDocuments {
		return model.Cluster{}, ErrClusterNotFound
	}
	if err != nil {
		return model.Cluster{}, err
	}

	return cluster, nil
}

func (r *MongoClusterRepository) Save(ctx context.Context, cluster model.Cluster) (model.Cluster, error) {
	_, err := r.collection.ReplaceOne(
		ctx,
		bson.M{"id": cluster.ID},
		cluster,
		optionsReplaceUpsert(),
	)
	if err != nil {
		return model.Cluster{}, err
	}

	return cluster, nil
}

func (r *MongoClusterRepository) DeleteByID(ctx context.Context, id string) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return ErrClusterNotFound
	}

	return nil
}
