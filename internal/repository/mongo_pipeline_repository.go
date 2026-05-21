package repository

import (
	"context"

	"hr-cloud-service/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoPipelineRepository struct {
	collection *mongo.Collection
}

func NewMongoPipelineRepository(db *mongo.Database) *MongoPipelineRepository {
	return &MongoPipelineRepository{
		collection: db.Collection("pipeline_runs"),
	}
}

func (r *MongoPipelineRepository) FindAll(ctx context.Context) ([]model.PipelineRun, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var pipelineRuns []model.PipelineRun
	if err := cursor.All(ctx, &pipelineRuns); err != nil {
		return nil, err
	}

	return pipelineRuns, nil
}

func (r *MongoPipelineRepository) FindByID(ctx context.Context, id string) (model.PipelineRun, error) {
	var pipelineRun model.PipelineRun
	err := r.collection.FindOne(ctx, bson.M{"id": id}).Decode(&pipelineRun)
	if err == mongo.ErrNoDocuments {
		return model.PipelineRun{}, ErrPipelineRunNotFound
	}
	if err != nil {
		return model.PipelineRun{}, err
	}

	return pipelineRun, nil
}

func (r *MongoPipelineRepository) Save(ctx context.Context, pipelineRun model.PipelineRun) (model.PipelineRun, error) {
	_, err := r.collection.ReplaceOne(
		ctx,
		bson.M{"id": pipelineRun.ID},
		pipelineRun,
		optionsReplaceUpsert(),
	)
	if err != nil {
		return model.PipelineRun{}, err
	}

	return pipelineRun, nil
}
