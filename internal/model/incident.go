package model

import "time"

type Incident struct {
	ID            string     `json:"id" bson:"id"`
	Title         string     `json:"title" bson:"title"`
	Summary       string     `json:"summary" bson:"summary"`
	Severity      string     `json:"severity" bson:"severity"`
	Status        string     `json:"status" bson:"status"`
	ApplicationID string     `json:"application_id,omitempty" bson:"application_id,omitempty"`
	ClusterID     string     `json:"cluster_id,omitempty" bson:"cluster_id,omitempty"`
	DeploymentID  string     `json:"deployment_id,omitempty" bson:"deployment_id,omitempty"`
	OwnerTeam     string     `json:"owner_team" bson:"owner_team"`
	CreatedAt     time.Time  `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" bson:"updated_at"`
	ResolvedAt    *time.Time `json:"resolved_at,omitempty" bson:"resolved_at,omitempty"`
}

type IncidentFilter struct {
	ApplicationID string
	ClusterID     string
	DeploymentID  string
	Severity      string
	Status        string
	OwnerTeam     string
}

type CreateIncidentRequest struct {
	Title         string `json:"title"`
	Summary       string `json:"summary"`
	Severity      string `json:"severity"`
	Status        string `json:"status"`
	ApplicationID string `json:"application_id"`
	ClusterID     string `json:"cluster_id"`
	DeploymentID  string `json:"deployment_id"`
	OwnerTeam     string `json:"owner_team"`
}

type UpdateIncidentRequest struct {
	Title         string `json:"title"`
	Summary       string `json:"summary"`
	Severity      string `json:"severity"`
	Status        string `json:"status"`
	ApplicationID string `json:"application_id"`
	ClusterID     string `json:"cluster_id"`
	DeploymentID  string `json:"deployment_id"`
	OwnerTeam     string `json:"owner_team"`
}

type UpdateIncidentStatusRequest struct {
	Status string `json:"status"`
}
