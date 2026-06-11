package repository

import (
	"context"

	"hr-cloud-service/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoActivityRepository struct {
	collection *mongo.Collection
}

func NewMongoActivityRepository(db *mongo.Database) *MongoActivityRepository {
	return &MongoActivityRepository{
		collection: db.Collection("activities"),
	}
}

func (r *MongoActivityRepository) FindAll(ctx context.Context) ([]model.Activity, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var activities []model.Activity
	if err := cursor.All(ctx, &activities); err != nil {
		return nil, err
	}

	return activities, nil
}

func (r *MongoActivityRepository) FindByID(ctx context.Context, id string) (model.Activity, error) {
	var activity model.Activity
	err := r.collection.FindOne(ctx, bson.M{"id": id}).Decode(&activity)
	if err == mongo.ErrNoDocuments {
		return model.Activity{}, ErrActivityNotFound
	}
	if err != nil {
		return model.Activity{}, err
	}

	return activity, nil
}

func (r *MongoActivityRepository) Save(ctx context.Context, activity model.Activity) (model.Activity, error) {
	_, err := r.collection.ReplaceOne(
		ctx,
		bson.M{"id": activity.ID},
		activity,
		optionsReplaceUpsert(),
	)
	if err != nil {
		return model.Activity{}, err
	}

	return activity, nil
}

func (r *MongoActivityRepository) DeleteByID(ctx context.Context, id string) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return ErrActivityNotFound
	}

	return nil
}
