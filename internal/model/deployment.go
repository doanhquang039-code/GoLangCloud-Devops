package model

import "time"

type Deployment struct {
	ID            string    `json:"id" bson:"id"`
	ApplicationID string    `json:"application_id" bson:"application_id"`
	Environment   string    `json:"environment" bson:"environment"`
	Version       string    `json:"version" bson:"version"`
	Status        string    `json:"status" bson:"status"`
	RequestedBy   string    `json:"requested_by" bson:"requested_by"`
	StartedAt     time.Time `json:"started_at" bson:"started_at"`
	FinishedAt    time.Time `json:"finished_at,omitempty" bson:"finished_at,omitempty"`
}

type CreateDeploymentRequest struct {
	ApplicationID string `json:"application_id"`
	Environment   string `json:"environment"`
	Version       string `json:"version"`
	RequestedBy   string `json:"requested_by"`
}

type UpdateDeploymentStatusRequest struct {
	Status string `json:"status"`
}
