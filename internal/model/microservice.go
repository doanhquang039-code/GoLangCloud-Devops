package model

import "time"

type Microservice struct {
	ID                   string            `json:"id" bson:"id"`
	TenantID             string            `json:"tenant_id" bson:"tenant_id"`
	ApplicationID        string            `json:"application_id" bson:"application_id"`
	Name                 string            `json:"name" bson:"name"`
	OwnerTeam            string            `json:"owner_team" bson:"owner_team"`
	Protocol             string            `json:"protocol" bson:"protocol"`
	Endpoint             string            `json:"endpoint" bson:"endpoint"`
	Status               string            `json:"status" bson:"status"`
	CloudProvider        string            `json:"cloud_provider,omitempty" bson:"cloud_provider,omitempty"`
	Region               string            `json:"region,omitempty" bson:"region,omitempty"`
	ClusterID            string            `json:"cluster_id,omitempty" bson:"cluster_id,omitempty"`
	Namespace            string            `json:"namespace,omitempty" bson:"namespace,omitempty"`
	Environment          string            `json:"environment,omitempty" bson:"environment,omitempty"`
	Runtime              string            `json:"runtime,omitempty" bson:"runtime,omitempty"`
	Image                string            `json:"image,omitempty" bson:"image,omitempty"`
	Version              string            `json:"version,omitempty" bson:"version,omitempty"`
	Replicas             int               `json:"replicas" bson:"replicas"`
	CPURequest           string            `json:"cpu_request,omitempty" bson:"cpu_request,omitempty"`
	MemoryRequest        string            `json:"memory_request,omitempty" bson:"memory_request,omitempty"`
	HealthPath           string            `json:"health_path,omitempty" bson:"health_path,omitempty"`
	SLOTarget            float64           `json:"slo_target" bson:"slo_target"`
	ErrorBudgetRemaining float64           `json:"error_budget_remaining" bson:"error_budget_remaining"`
	Dependencies         []string          `json:"dependencies,omitempty" bson:"dependencies,omitempty"`
	Config               map[string]string `json:"config,omitempty" bson:"config,omitempty"`
	Tags                 []string          `json:"tags,omitempty" bson:"tags,omitempty"`
	CreatedAt            time.Time         `json:"created_at" bson:"created_at"`
	UpdatedAt            time.Time         `json:"updated_at" bson:"updated_at"`
}

type MicroserviceFilter struct {
	Query         string
	TenantID      string
	ApplicationID string
	OwnerTeam     string
	Protocol      string
	Status        string
	CloudProvider string
	Region        string
	ClusterID     string
	Namespace     string
	Environment   string
	Runtime       string
	Tag           string
	MinReplicas   int
	Limit         int
	Offset        int
	AfterID       string
	SortBy        string
	SortOrder     string
}

type CreateMicroserviceRequest struct {
	TenantID             string            `json:"tenant_id"`
	ApplicationID        string            `json:"application_id"`
	Name                 string            `json:"name"`
	OwnerTeam            string            `json:"owner_team"`
	Protocol             string            `json:"protocol"`
	Endpoint             string            `json:"endpoint"`
	Status               string            `json:"status"`
	CloudProvider        string            `json:"cloud_provider"`
	Region               string            `json:"region"`
	ClusterID            string            `json:"cluster_id"`
	Namespace            string            `json:"namespace"`
	Environment          string            `json:"environment"`
	Runtime              string            `json:"runtime"`
	Image                string            `json:"image"`
	Version              string            `json:"version"`
	Replicas             int               `json:"replicas"`
	CPURequest           string            `json:"cpu_request"`
	MemoryRequest        string            `json:"memory_request"`
	HealthPath           string            `json:"health_path"`
	SLOTarget            float64           `json:"slo_target"`
	ErrorBudgetRemaining float64           `json:"error_budget_remaining"`
	Dependencies         []string          `json:"dependencies"`
	Config               map[string]string `json:"config"`
	Tags                 []string          `json:"tags"`
}

type UpdateMicroserviceRequest struct {
	TenantID             string            `json:"tenant_id"`
	ApplicationID        string            `json:"application_id"`
	Name                 string            `json:"name"`
	OwnerTeam            string            `json:"owner_team"`
	Protocol             string            `json:"protocol"`
	Endpoint             string            `json:"endpoint"`
	Status               string            `json:"status"`
	CloudProvider        string            `json:"cloud_provider"`
	Region               string            `json:"region"`
	ClusterID            string            `json:"cluster_id"`
	Namespace            string            `json:"namespace"`
	Environment          string            `json:"environment"`
	Runtime              string            `json:"runtime"`
	Image                string            `json:"image"`
	Version              string            `json:"version"`
	Replicas             int               `json:"replicas"`
	CPURequest           string            `json:"cpu_request"`
	MemoryRequest        string            `json:"memory_request"`
	HealthPath           string            `json:"health_path"`
	SLOTarget            float64           `json:"slo_target"`
	ErrorBudgetRemaining float64           `json:"error_budget_remaining"`
	Dependencies         []string          `json:"dependencies"`
	Config               map[string]string `json:"config"`
	Tags                 []string          `json:"tags"`
}

type UpdateMicroserviceStatusRequest struct {
	Status string `json:"status"`
}
