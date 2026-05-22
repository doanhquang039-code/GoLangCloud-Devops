package model

import "time"

type Environment struct {
	ID            string            `json:"id" bson:"id"`
	Name          string            `json:"name" bson:"name"`
	Type          string            `json:"type" bson:"type"`
	ApplicationID string            `json:"application_id" bson:"application_id"`
	ClusterID     string            `json:"cluster_id" bson:"cluster_id"`
	Namespace     string            `json:"namespace" bson:"namespace"`
	Status        string            `json:"status" bson:"status"`
	Variables     map[string]string `json:"variables,omitempty" bson:"variables,omitempty"`
	CreatedAt     time.Time         `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at" bson:"updated_at"`
}

type EnvironmentFilter struct {
	ApplicationID string
	ClusterID     string
	Type          string
	Status        string
}

type CreateEnvironmentRequest struct {
	Name          string            `json:"name"`
	Type          string            `json:"type"`
	ApplicationID string            `json:"application_id"`
	ClusterID     string            `json:"cluster_id"`
	Namespace     string            `json:"namespace"`
	Status        string            `json:"status"`
	Variables     map[string]string `json:"variables"`
}

type UpdateEnvironmentRequest struct {
	Name          string            `json:"name"`
	Type          string            `json:"type"`
	ApplicationID string            `json:"application_id"`
	ClusterID     string            `json:"cluster_id"`
	Namespace     string            `json:"namespace"`
	Status        string            `json:"status"`
	Variables     map[string]string `json:"variables"`
}
