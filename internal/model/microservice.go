package model

import "time"

type Microservice struct {
	ID            string            `json:"id" bson:"id"`
	ApplicationID string            `json:"application_id" bson:"application_id"`
	Name          string            `json:"name" bson:"name"`
	OwnerTeam     string            `json:"owner_team" bson:"owner_team"`
	Protocol      string            `json:"protocol" bson:"protocol"`
	Endpoint      string            `json:"endpoint" bson:"endpoint"`
	Status        string            `json:"status" bson:"status"`
	Dependencies  []string          `json:"dependencies,omitempty" bson:"dependencies,omitempty"`
	Config        map[string]string `json:"config,omitempty" bson:"config,omitempty"`
	Tags          []string          `json:"tags,omitempty" bson:"tags,omitempty"`
	CreatedAt     time.Time         `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at" bson:"updated_at"`
}

type MicroserviceFilter struct {
	Query         string
	ApplicationID string
	OwnerTeam     string
	Protocol      string
	Status        string
	Tag           string
}

type CreateMicroserviceRequest struct {
	ApplicationID string            `json:"application_id"`
	Name          string            `json:"name"`
	OwnerTeam     string            `json:"owner_team"`
	Protocol      string            `json:"protocol"`
	Endpoint      string            `json:"endpoint"`
	Status        string            `json:"status"`
	Dependencies  []string          `json:"dependencies"`
	Config        map[string]string `json:"config"`
	Tags          []string          `json:"tags"`
}

type UpdateMicroserviceRequest struct {
	ApplicationID string            `json:"application_id"`
	Name          string            `json:"name"`
	OwnerTeam     string            `json:"owner_team"`
	Protocol      string            `json:"protocol"`
	Endpoint      string            `json:"endpoint"`
	Status        string            `json:"status"`
	Dependencies  []string          `json:"dependencies"`
	Config        map[string]string `json:"config"`
	Tags          []string          `json:"tags"`
}

type UpdateMicroserviceStatusRequest struct {
	Status string `json:"status"`
}
