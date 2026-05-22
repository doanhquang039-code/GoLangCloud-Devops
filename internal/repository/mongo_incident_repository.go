package repository

import (
	"context"

	"hr-cloud-service/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoIncidentRepository struct {
	collection *mongo.Collection
}

func NewMongoIncidentRepository(db *mongo.Database) *MongoIncidentRepository {
	return &MongoIncidentRepository{
		collection: db.Collection("incidents"),
	}
}

func (r *MongoIncidentRepository) FindAll(ctx context.Context) ([]model.Incident, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var incidents []model.Incident
	if err := cursor.All(ctx, &incidents); err != nil {
		return nil, err
	}

	return incidents, nil
}

func (r *MongoIncidentRepository) FindByID(ctx context.Context, id string) (model.Incident, error) {
	var incident model.Incident
	err := r.collection.FindOne(ctx, bson.M{"id": id}).Decode(&incident)
	if err == mongo.ErrNoDocuments {
		return model.Incident{}, ErrIncidentNotFound
	}
	if err != nil {
		return model.Incident{}, err
	}

	return incident, nil
}

func (r *MongoIncidentRepository) Save(ctx context.Context, incident model.Incident) (model.Incident, error) {
	_, err := r.collection.ReplaceOne(
		ctx,
		bson.M{"id": incident.ID},
		incident,
		optionsReplaceUpsert(),
	)
	if err != nil {
		return model.Incident{}, err
	}

	return incident, nil
}
