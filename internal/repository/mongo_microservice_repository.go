package repository

import (
	"context"

	"hr-cloud-service/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoMicroserviceRepository struct {
	collection *mongo.Collection
}

func NewMongoMicroserviceRepository(db *mongo.Database) *MongoMicroserviceRepository {
	return &MongoMicroserviceRepository{
		collection: db.Collection("microservices"),
	}
}

func (r *MongoMicroserviceRepository) FindAll(ctx context.Context) ([]model.Microservice, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var microservices []model.Microservice
	if err := cursor.All(ctx, &microservices); err != nil {
		return nil, err
	}

	return microservices, nil
}

func (r *MongoMicroserviceRepository) FindByID(ctx context.Context, id string) (model.Microservice, error) {
	var microservice model.Microservice
	err := r.collection.FindOne(ctx, bson.M{"id": id}).Decode(&microservice)
	if err == mongo.ErrNoDocuments {
		return model.Microservice{}, ErrMicroserviceNotFound
	}
	if err != nil {
		return model.Microservice{}, err
	}

	return microservice, nil
}

func (r *MongoMicroserviceRepository) Save(ctx context.Context, microservice model.Microservice) (model.Microservice, error) {
	_, err := r.collection.ReplaceOne(
		ctx,
		bson.M{"id": microservice.ID},
		microservice,
		optionsReplaceUpsert(),
	)
	if err != nil {
		return model.Microservice{}, err
	}

	return microservice, nil
}

func (r *MongoMicroserviceRepository) DeleteByID(ctx context.Context, id string) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return ErrMicroserviceNotFound
	}

	return nil
}
