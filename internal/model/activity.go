package model

import "time"

type Activity struct {
	ID            string            `json:"id" bson:"id"`
	Type          string            `json:"type" bson:"type"`
	Action        string            `json:"action" bson:"action"`
	Status        string            `json:"status" bson:"status"`
	Actor         string            `json:"actor" bson:"actor"`
	ResourceType  string            `json:"resource_type" bson:"resource_type"`
	ResourceID    string            `json:"resource_id" bson:"resource_id"`
	ApplicationID string            `json:"application_id,omitempty" bson:"application_id,omitempty"`
	OwnerTeam     string            `json:"owner_team,omitempty" bson:"owner_team,omitempty"`
	Summary       string            `json:"summary" bson:"summary"`
	Metadata      map[string]string `json:"metadata,omitempty" bson:"metadata,omitempty"`
	Tags          []string          `json:"tags,omitempty" bson:"tags,omitempty"`
	CreatedAt     time.Time         `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at" bson:"updated_at"`
}

type ActivityFilter struct {
	Query         string
	Type          string
	Action        string
	Status        string
	Actor         string
	ResourceType  string
	ResourceID    string
	ApplicationID string
	OwnerTeam     string
	Tag           string
}

type CreateActivityRequest struct {
	Type          string            `json:"type"`
	Action        string            `json:"action"`
	Status        string            `json:"status"`
	Actor         string            `json:"actor"`
	ResourceType  string            `json:"resource_type"`
	ResourceID    string            `json:"resource_id"`
	ApplicationID string            `json:"application_id"`
	OwnerTeam     string            `json:"owner_team"`
	Summary       string            `json:"summary"`
	Metadata      map[string]string `json:"metadata"`
	Tags          []string          `json:"tags"`
}

type UpdateActivityRequest CreateActivityRequest
