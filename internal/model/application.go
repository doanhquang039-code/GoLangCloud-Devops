package model

import "time"

type Application struct {
	ID             string            `json:"id" bson:"id"`
	Name           string            `json:"name" bson:"name"`
	Repository     string            `json:"repository" bson:"repository"`
	Runtime        string            `json:"runtime" bson:"runtime"`
	OwnerTeam      string            `json:"owner_team" bson:"owner_team"`
	Criticality    string            `json:"criticality" bson:"criticality"`
	Port           int               `json:"port" bson:"port"`
	Replicas       int               `json:"replicas" bson:"replicas"`
	HealthEndpoint string            `json:"health_endpoint" bson:"health_endpoint"`
	Environment    map[string]string `json:"environment,omitempty" bson:"environment,omitempty"`
	Tags           []string          `json:"tags,omitempty" bson:"tags,omitempty"`
	CreatedAt      time.Time         `json:"created_at" bson:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at" bson:"updated_at"`
}

type ApplicationFilter struct {
	Query       string
	OwnerTeam   string
	Criticality string
	Runtime     string
	Tag         string
}

type CreateApplicationRequest struct {
	Name           string            `json:"name"`
	Repository     string            `json:"repository"`
	Runtime        string            `json:"runtime"`
	OwnerTeam      string            `json:"owner_team"`
	Criticality    string            `json:"criticality"`
	Port           int               `json:"port"`
	Replicas       int               `json:"replicas"`
	HealthEndpoint string            `json:"health_endpoint"`
	Environment    map[string]string `json:"environment"`
	Tags           []string          `json:"tags"`
}

type UpdateApplicationRequest struct {
	Name           string            `json:"name"`
	Repository     string            `json:"repository"`
	Runtime        string            `json:"runtime"`
	OwnerTeam      string            `json:"owner_team"`
	Criticality    string            `json:"criticality"`
	Port           int               `json:"port"`
	Replicas       int               `json:"replicas"`
	HealthEndpoint string            `json:"health_endpoint"`
	Environment    map[string]string `json:"environment"`
	Tags           []string          `json:"tags"`
}
