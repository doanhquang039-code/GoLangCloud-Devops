package model

import "time"

type Environment struct {
	ID            string            `json:"id" bson:"id"`
	Name          string            `json:"name" bson:"name"`
	Type          string            `json:"type" bson:"type"`
	ApplicationID string            `json:"application_id" bson:"application_id"`
	ClusterID     string            `json:"cluster_id" bson:"cluster_id"`
	Namespace     string            `json:"namespace" bson:"namespace"`
	Status        string            `json:"status" bson:"status"`
	Variables     map[string]string `json:"variables,omitempty" bson:"variables,omitempty"`
	CreatedAt     time.Time         `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at" bson:"updated_at"`
}

type EnvironmentFilter struct {
	ApplicationID string
	ClusterID     string
	Type          string
	Status        string
}

type CreateEnvironmentRequest struct {
	Name          string            `json:"name"`
	Type          string            `json:"type"`
	ApplicationID string            `json:"application_id"`
	ClusterID     string            `json:"cluster_id"`
	Namespace     string            `json:"namespace"`
	Status        string            `json:"status"`
	Variables     map[string]string `json:"variables"`
}

type UpdateEnvironmentRequest struct {
	Name          string            `json:"name"`
	Type          string            `json:"type"`
	ApplicationID string            `json:"application_id"`
	ClusterID     string            `json:"cluster_id"`
	Namespace     string            `json:"namespace"`
	Status        string            `json:"status"`
	Variables     map[string]string `json:"variables"`
}

type EnvironmentVariableDrift struct {
	Key           string `json:"key"`
	ExpectedValue string `json:"expected_value,omitempty"`
	ActualValue   string `json:"actual_value,omitempty"`
}

type EnvironmentDriftReport struct {
	EnvironmentID    string                     `json:"environment_id"`
	EnvironmentName  string                     `json:"environment_name"`
	EnvironmentType  string                     `json:"environment_type"`
	ApplicationID    string                     `json:"application_id"`
	ApplicationName  string                     `json:"application_name"`
	ClusterID        string                     `json:"cluster_id"`
	Namespace        string                     `json:"namespace"`
	Status           string                     `json:"status"`
	MissingVariables []EnvironmentVariableDrift `json:"missing_variables,omitempty"`
	ChangedVariables []EnvironmentVariableDrift `json:"changed_variables,omitempty"`
	ExtraVariables   []EnvironmentVariableDrift `json:"extra_variables,omitempty"`
	DriftScore       int                        `json:"drift_score"`
	DriftLevel       string                     `json:"drift_level"`
	DriftReasons     []string                   `json:"drift_reasons,omitempty"`
}
