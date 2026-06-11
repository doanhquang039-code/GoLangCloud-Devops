package repository

import (
	"context"

	"hr-cloud-service/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoTechnologyRepository struct {
	collection *mongo.Collection
}

func NewMongoTechnologyRepository(db *mongo.Database) *MongoTechnologyRepository {
	return &MongoTechnologyRepository{
		collection: db.Collection("technologies"),
	}
}

func (r *MongoTechnologyRepository) FindAll(ctx context.Context) ([]model.Technology, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var technologies []model.Technology
	if err := cursor.All(ctx, &technologies); err != nil {
		return nil, err
	}

	return technologies, nil
}

func (r *MongoTechnologyRepository) FindByID(ctx context.Context, id string) (model.Technology, error) {
	var technology model.Technology
	err := r.collection.FindOne(ctx, bson.M{"id": id}).Decode(&technology)
	if err == mongo.ErrNoDocuments {
		return model.Technology{}, ErrTechnologyNotFound
	}
	if err != nil {
		return model.Technology{}, err
	}

	return technology, nil
}

func (r *MongoTechnologyRepository) Save(ctx context.Context, technology model.Technology) (model.Technology, error) {
	_, err := r.collection.ReplaceOne(
		ctx,
		bson.M{"id": technology.ID},
		technology,
		optionsReplaceUpsert(),
	)
	if err != nil {
		return model.Technology{}, err
	}

	return technology, nil
}

func (r *MongoTechnologyRepository) DeleteByID(ctx context.Context, id string) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return ErrTechnologyNotFound
	}

	return nil
}
