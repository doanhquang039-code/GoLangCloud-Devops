package model

import "time"

type Technology struct {
	ID               string    `json:"id" bson:"id"`
	Name             string    `json:"name" bson:"name"`
	Category         string    `json:"category" bson:"category"`
	Version          string    `json:"version" bson:"version"`
	OwnerTeam        string    `json:"owner_team" bson:"owner_team"`
	Status           string    `json:"status" bson:"status"`
	RiskLevel        string    `json:"risk_level" bson:"risk_level"`
	AdoptionStage    string    `json:"adoption_stage" bson:"adoption_stage"`
	License          string    `json:"license,omitempty" bson:"license,omitempty"`
	DocumentationURL string    `json:"documentation_url,omitempty" bson:"documentation_url,omitempty"`
	Tags             []string  `json:"tags,omitempty" bson:"tags,omitempty"`
	CreatedAt        time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" bson:"updated_at"`
}

type TechnologyFilter struct {
	Query         string
	Category      string
	OwnerTeam     string
	Status        string
	RiskLevel     string
	AdoptionStage string
	Tag           string
}

type CreateTechnologyRequest struct {
	Name             string   `json:"name"`
	Category         string   `json:"category"`
	Version          string   `json:"version"`
	OwnerTeam        string   `json:"owner_team"`
	Status           string   `json:"status"`
	RiskLevel        string   `json:"risk_level"`
	AdoptionStage    string   `json:"adoption_stage"`
	License          string   `json:"license"`
	DocumentationURL string   `json:"documentation_url"`
	Tags             []string `json:"tags"`
}

type UpdateTechnologyRequest CreateTechnologyRequest
