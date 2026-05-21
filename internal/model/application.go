package model

import "time"

type Application struct {
	ID          string    `json:"id" bson:"id"`
	Name        string    `json:"name" bson:"name"`
	Repository  string    `json:"repository" bson:"repository"`
	Runtime     string    `json:"runtime" bson:"runtime"`
	OwnerTeam   string    `json:"owner_team" bson:"owner_team"`
	Criticality string    `json:"criticality" bson:"criticality"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" bson:"updated_at"`
}

type CreateApplicationRequest struct {
	Name        string `json:"name"`
	Repository  string `json:"repository"`
	Runtime     string `json:"runtime"`
	OwnerTeam   string `json:"owner_team"`
	Criticality string `json:"criticality"`
}
