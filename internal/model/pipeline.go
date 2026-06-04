package model

import "time"

type PipelineRun struct {
	ID            string          `json:"id" bson:"id"`
	ApplicationID string          `json:"application_id" bson:"application_id"`
	Branch        string          `json:"branch" bson:"branch"`
	CommitSHA     string          `json:"commit_sha" bson:"commit_sha"`
	TriggeredBy   string          `json:"triggered_by" bson:"triggered_by"`
	Status        string          `json:"status" bson:"status"`
	Stages        []PipelineStage `json:"stages" bson:"stages"`
	StartedAt     time.Time       `json:"started_at" bson:"started_at"`
	FinishedAt    *time.Time      `json:"finished_at,omitempty" bson:"finished_at,omitempty"`
}

type PipelineStage struct {
	Name      string     `json:"name" bson:"name"`
	Status    string     `json:"status" bson:"status"`
	StartedAt time.Time  `json:"started_at" bson:"started_at"`
	EndedAt   *time.Time `json:"ended_at,omitempty" bson:"ended_at,omitempty"`
}

type PipelineRunFilter struct {
	Query         string
	ApplicationID string
	Branch        string
	Status        string
	TriggeredBy   string
}

type CreatePipelineRunRequest struct {
	ApplicationID string   `json:"application_id"`
	Branch        string   `json:"branch"`
	CommitSHA     string   `json:"commit_sha"`
	TriggeredBy   string   `json:"triggered_by"`
	Stages        []string `json:"stages"`
}

type UpdatePipelineRunStatusRequest struct {
	Status string `json:"status"`
}

type UpdatePipelineStageStatusRequest struct {
	Status string `json:"status"`
}
