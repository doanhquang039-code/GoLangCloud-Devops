package repository

import (
	"context"

	"hr-cloud-service/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoEnvironmentRepository struct {
	collection *mongo.Collection
}

func NewMongoEnvironmentRepository(db *mongo.Database) *MongoEnvironmentRepository {
	return &MongoEnvironmentRepository{
		collection: db.Collection("environments"),
	}
}

func (r *MongoEnvironmentRepository) FindAll(ctx context.Context) ([]model.Environment, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var environments []model.Environment
	if err := cursor.All(ctx, &environments); err != nil {
		return nil, err
	}

	return environments, nil
}

func (r *MongoEnvironmentRepository) FindByID(ctx context.Context, id string) (model.Environment, error) {
	var environment model.Environment
	err := r.collection.FindOne(ctx, bson.M{"id": id}).Decode(&environment)
	if err == mongo.ErrNoDocuments {
		return model.Environment{}, ErrEnvironmentNotFound
	}
	if err != nil {
		return model.Environment{}, err
	}

	return environment, nil
}

func (r *MongoEnvironmentRepository) Save(ctx context.Context, environment model.Environment) (model.Environment, error) {
	_, err := r.collection.ReplaceOne(
		ctx,
		bson.M{"id": environment.ID},
		environment,
		optionsReplaceUpsert(),
	)
	if err != nil {
		return model.Environment{}, err
	}

	return environment, nil
}
