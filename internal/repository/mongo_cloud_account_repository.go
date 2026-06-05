package repository

import (
	"context"

	"hr-cloud-service/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoCloudAccountRepository struct {
	collection *mongo.Collection
}

func NewMongoCloudAccountRepository(db *mongo.Database) *MongoCloudAccountRepository {
	return &MongoCloudAccountRepository{
		collection: db.Collection("cloud_accounts"),
	}
}

func (r *MongoCloudAccountRepository) FindAll(ctx context.Context) ([]model.CloudAccount, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var accounts []model.CloudAccount
	if err := cursor.All(ctx, &accounts); err != nil {
		return nil, err
	}

	return accounts, nil
}

func (r *MongoCloudAccountRepository) FindByID(ctx context.Context, id string) (model.CloudAccount, error) {
	var account model.CloudAccount
	err := r.collection.FindOne(ctx, bson.M{"id": id}).Decode(&account)
	if err == mongo.ErrNoDocuments {
		return model.CloudAccount{}, ErrCloudAccountNotFound
	}
	if err != nil {
		return model.CloudAccount{}, err
	}

	return account, nil
}

func (r *MongoCloudAccountRepository) Save(ctx context.Context, account model.CloudAccount) (model.CloudAccount, error) {
	_, err := r.collection.ReplaceOne(ctx, bson.M{"id": account.ID}, account, optionsReplaceUpsert())
	if err != nil {
		return model.CloudAccount{}, err
	}

	return account, nil
}

func (r *MongoCloudAccountRepository) DeleteByID(ctx context.Context, id string) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return ErrCloudAccountNotFound
	}

	return nil
}
