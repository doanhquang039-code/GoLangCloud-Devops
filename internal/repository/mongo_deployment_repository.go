package repository

import (
	"context"

	"hr-cloud-service/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoDeploymentRepository struct {
	collection *mongo.Collection
}

func NewMongoDeploymentRepository(db *mongo.Database) *MongoDeploymentRepository {
	return &MongoDeploymentRepository{
		collection: db.Collection("deployments"),
	}
}

func (r *MongoDeploymentRepository) FindAll(ctx context.Context) ([]model.Deployment, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var deployments []model.Deployment
	if err := cursor.All(ctx, &deployments); err != nil {
		return nil, err
	}

	return deployments, nil
}

func (r *MongoDeploymentRepository) FindByID(ctx context.Context, id string) (model.Deployment, error) {
	var deployment model.Deployment
	err := r.collection.FindOne(ctx, bson.M{"id": id}).Decode(&deployment)
	if err == mongo.ErrNoDocuments {
		return model.Deployment{}, ErrDeploymentNotFound
	}
	if err != nil {
		return model.Deployment{}, err
	}

	return deployment, nil
}

func (r *MongoDeploymentRepository) Save(ctx context.Context, deployment model.Deployment) (model.Deployment, error) {
	_, err := r.collection.ReplaceOne(
		ctx,
		bson.M{"id": deployment.ID},
		deployment,
		optionsReplaceUpsert(),
	)
	if err != nil {
		return model.Deployment{}, err
	}

	return deployment, nil
}
