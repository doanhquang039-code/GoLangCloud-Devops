package model

import "time"

type Cluster struct {
	ID        string    `json:"id" bson:"id"`
	Name      string    `json:"name" bson:"name"`
	Provider  string    `json:"provider" bson:"provider"`
	Region    string    `json:"region" bson:"region"`
	Endpoint  string    `json:"endpoint" bson:"endpoint"`
	Version   string    `json:"version" bson:"version"`
	Status    string    `json:"status" bson:"status"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

type ClusterFilter struct {
	Provider string
	Region   string
	Status   string
}

type CreateClusterRequest struct {
	Name     string `json:"name"`
	Provider string `json:"provider"`
	Region   string `json:"region"`
	Endpoint string `json:"endpoint"`
	Version  string `json:"version"`
	Status   string `json:"status"`
}

type UpdateClusterRequest struct {
	Name     string `json:"name"`
	Provider string `json:"provider"`
	Region   string `json:"region"`
	Endpoint string `json:"endpoint"`
	Version  string `json:"version"`
	Status   string `json:"status"`
}

type UpdateClusterStatusRequest struct {
	Status string `json:"status"`
}

type PlatformSummary struct {
	Applications int            `json:"applications"`
	Clusters     int            `json:"clusters"`
	Environments int            `json:"environments"`
	Deployments  int            `json:"deployments"`
	PipelineRuns int            `json:"pipeline_runs"`
	Incidents    int            `json:"incidents"`
	OpenIncidents int           `json:"open_incidents"`
	ByStatus     map[string]int `json:"deployments_by_status"`
}
