package repository

import (
	"context"

	"hr-cloud-service/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoEmployeeRepository struct {
	collection *mongo.Collection
}

func NewMongoEmployeeRepository(db *mongo.Database) *MongoEmployeeRepository {
	return &MongoEmployeeRepository{
		collection: db.Collection("employees"),
	}
}

func (r *MongoEmployeeRepository) FindAll(ctx context.Context) ([]model.Employee, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var employees []model.Employee
	if err := cursor.All(ctx, &employees); err != nil {
		return nil, err
	}

	return employees, nil
}

func (r *MongoEmployeeRepository) FindByID(ctx context.Context, id string) (model.Employee, error) {
	var employee model.Employee
	err := r.collection.FindOne(ctx, bson.M{"id": id}).Decode(&employee)
	if err == mongo.ErrNoDocuments {
		return model.Employee{}, ErrEmployeeNotFound
	}
	if err != nil {
		return model.Employee{}, err
	}

	return employee, nil
}

func (r *MongoEmployeeRepository) Save(ctx context.Context, employee model.Employee) (model.Employee, error) {
	_, err := r.collection.ReplaceOne(
		ctx,
		bson.M{"id": employee.ID},
		employee,
		optionsReplaceUpsert(),
	)
	if err != nil {
		return model.Employee{}, err
	}

	return employee, nil
}

func (r *MongoEmployeeRepository) DeleteByID(ctx context.Context, id string) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return ErrEmployeeNotFound
	}

	return nil
}
