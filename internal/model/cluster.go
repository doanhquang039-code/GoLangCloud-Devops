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
	Query    string
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
	Applications  int            `json:"applications"`
	Clusters      int            `json:"clusters"`
	Environments  int            `json:"environments"`
	Deployments   int            `json:"deployments"`
	PipelineRuns  int            `json:"pipeline_runs"`
	Incidents     int            `json:"incidents"`
	OpenIncidents int            `json:"open_incidents"`
	ByStatus      map[string]int `json:"deployments_by_status"`
}

type PlatformScorecardFilter struct {
	Query       string
	OwnerTeam   string
	Criticality string
	RiskLevel   string
	MinScore    int
	SortBy      string
	SortOrder   string
}

type PlatformScorecard struct {
	ApplicationID                     string   `json:"application_id"`
	ApplicationName                   string   `json:"application_name"`
	OwnerTeam                         string   `json:"owner_team"`
	Criticality                       string   `json:"criticality"`
	EnvironmentCount                  int      `json:"environment_count"`
	ActiveEnvironmentCount            int      `json:"active_environment_count"`
	ClusterCount                      int      `json:"cluster_count"`
	DeploymentCount                   int      `json:"deployment_count"`
	SuccessfulDeploymentCount         int      `json:"successful_deployment_count"`
	FailedDeploymentCount             int      `json:"failed_deployment_count"`
	RunningDeploymentCount            int      `json:"running_deployment_count"`
	DeploymentSuccessRate             float64  `json:"deployment_success_rate"`
	PipelineRunCount                  int      `json:"pipeline_run_count"`
	SuccessfulPipelineRunCount        int      `json:"successful_pipeline_run_count"`
	FailedPipelineRunCount            int      `json:"failed_pipeline_run_count"`
	RunningPipelineRunCount           int      `json:"running_pipeline_run_count"`
	PipelineSuccessRate               float64  `json:"pipeline_success_rate"`
	IncidentCount                     int      `json:"incident_count"`
	OpenIncidentCount                 int      `json:"open_incident_count"`
	Sev1IncidentCount                 int      `json:"sev1_incident_count"`
	Sev2IncidentCount                 int      `json:"sev2_incident_count"`
	MeanTimeToResolveMinutes          float64  `json:"mean_time_to_resolve_minutes"`
	AverageDeploymentDurationMinutes  float64  `json:"average_deployment_duration_minutes"`
	AveragePipelineRunDurationMinutes float64  `json:"average_pipeline_run_duration_minutes"`
	OperationalReadinessScore         int      `json:"operational_readiness_score"`
	RiskLevel                         string   `json:"risk_level"`
	RiskReasons                       []string `json:"risk_reasons,omitempty"`
}

type EnvironmentDriftReportFilter struct {
	Query           string
	ApplicationID   string
	EnvironmentType string
	Status          string
	DriftLevel      string
	MaxDriftScore   int
	SortBy          string
	SortOrder       string
}
