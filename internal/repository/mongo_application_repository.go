package repository

import (
	"context"

	"hr-cloud-service/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoApplicationRepository struct {
	collection *mongo.Collection
}

func NewMongoApplicationRepository(db *mongo.Database) *MongoApplicationRepository {
	return &MongoApplicationRepository{
		collection: db.Collection("applications"),
	}
}

func (r *MongoApplicationRepository) FindAll(ctx context.Context) ([]model.Application, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var applications []model.Application
	if err := cursor.All(ctx, &applications); err != nil {
		return nil, err
	}

	return applications, nil
}

func (r *MongoApplicationRepository) FindByID(ctx context.Context, id string) (model.Application, error) {
	var application model.Application
	err := r.collection.FindOne(ctx, bson.M{"id": id}).Decode(&application)
	if err == mongo.ErrNoDocuments {
		return model.Application{}, ErrApplicationNotFound
	}
	if err != nil {
		return model.Application{}, err
	}

	return application, nil
}

func (r *MongoApplicationRepository) Save(ctx context.Context, application model.Application) (model.Application, error) {
	_, err := r.collection.ReplaceOne(
		ctx,
		bson.M{"id": application.ID},
		application,
		optionsReplaceUpsert(),
	)
	if err != nil {
		return model.Application{}, err
	}

	return application, nil
}

func (r *MongoApplicationRepository) DeleteByID(ctx context.Context, id string) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return ErrApplicationNotFound
	}

	return nil
}
