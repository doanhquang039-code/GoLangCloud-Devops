package repository

import (
	"context"
	"strings"

	"hr-cloud-service/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

func (r *MongoMicroserviceRepository) FindByFilter(ctx context.Context, filter model.MicroserviceFilter) ([]model.Microservice, error) {
	query := bson.M{}
	if filter.TenantID != "" {
		query["tenant_id"] = filter.TenantID
	}
	if filter.ApplicationID != "" {
		query["application_id"] = filter.ApplicationID
	}
	if filter.OwnerTeam != "" {
		query["owner_team"] = filter.OwnerTeam
	}
	if filter.Protocol != "" {
		query["protocol"] = filter.Protocol
	}
	if filter.Status != "" {
		query["status"] = filter.Status
	}
	if filter.CloudProvider != "" {
		query["cloud_provider"] = filter.CloudProvider
	}
	if filter.Region != "" {
		query["region"] = filter.Region
	}
	if filter.ClusterID != "" {
		query["cluster_id"] = filter.ClusterID
	}
	if filter.Namespace != "" {
		query["namespace"] = filter.Namespace
	}
	if filter.Environment != "" {
		query["environment"] = filter.Environment
	}
	if filter.Runtime != "" {
		query["runtime"] = filter.Runtime
	}
	if filter.Tag != "" {
		query["tags"] = filter.Tag
	}
	if filter.MinReplicas > 0 {
		query["replicas"] = bson.M{"$gte": filter.MinReplicas}
	}
	if filter.AfterID != "" {
		operator := "$gt"
		if strings.EqualFold(filter.SortOrder, "desc") {
			operator = "$lt"
		}
		query["id"] = bson.M{operator: filter.AfterID}
	}
	if filter.Query != "" {
		regex := bson.M{"$regex": filter.Query, "$options": "i"}
		query["$or"] = []bson.M{
			{"id": regex},
			{"tenant_id": regex},
			{"application_id": regex},
			{"name": regex},
			{"owner_team": regex},
			{"protocol": regex},
			{"endpoint": regex},
			{"status": regex},
			{"cloud_provider": regex},
			{"region": regex},
			{"cluster_id": regex},
			{"namespace": regex},
			{"environment": regex},
			{"runtime": regex},
			{"image": regex},
			{"version": regex},
			{"dependencies": regex},
			{"tags": regex},
		}
	}

	findOptions := options.Find().
		SetLimit(int64(filter.Limit)).
		SetSkip(int64(filter.Offset)).
		SetSort(microserviceMongoSort(filter.SortBy, filter.SortOrder))

	cursor, err := r.collection.Find(ctx, query, findOptions)
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

func microserviceMongoSort(sortBy string, sortOrder string) bson.D {
	direction := 1
	if strings.EqualFold(sortOrder, "desc") {
		direction = -1
	}
	switch sortBy {
	case "name":
		return bson.D{{Key: "name", Value: direction}, {Key: "id", Value: 1}}
	case "updated_at":
		return bson.D{{Key: "updated_at", Value: direction}, {Key: "id", Value: 1}}
	case "replicas":
		return bson.D{{Key: "replicas", Value: direction}, {Key: "id", Value: 1}}
	case "id":
		return bson.D{{Key: "id", Value: direction}}
	default:
		return bson.D{{Key: "id", Value: direction}}
	}
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
		bson.M{"tenant_id": microservice.TenantID, "id": microservice.ID},
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
