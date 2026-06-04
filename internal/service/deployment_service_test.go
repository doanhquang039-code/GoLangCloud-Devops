package service

import (
	"context"
	"testing"

	"hr-cloud-service/internal/model"
	"hr-cloud-service/internal/repository"
)

func TestGetDeploymentsFiltersByQueryEnvironmentAndStatus(t *testing.T) {
	ctx := context.Background()
	deploymentRepository := repository.NewInMemoryDeploymentRepository()
	deploymentService := NewDeploymentService(
		repository.NewInMemoryApplicationRepository(),
		repository.NewInMemoryClusterRepository(),
		deploymentRepository,
	)

	deployments := []model.Deployment{
		{
			ID:            "dep-payroll-100",
			ApplicationID: "app-payroll-api",
			ClusterID:     "cluster-staging",
			Namespace:     "hr-staging",
			Environment:   "Staging",
			Version:       "v1.2.3",
			Strategy:      "canary",
			Status:        "Running",
			RequestedBy:   "devops@example.com",
		},
		{
			ID:            "dep-people-100",
			ApplicationID: "app-people-web",
			ClusterID:     "cluster-prod",
			Namespace:     "hr-prod",
			Environment:   "production",
			Version:       "v2.0.0",
			Strategy:      "rolling",
			Status:        "succeeded",
			RequestedBy:   "people@example.com",
		},
	}
	for _, deployment := range deployments {
		if _, err := deploymentRepository.Save(ctx, deployment); err != nil {
			t.Fatal(err)
		}
	}

	filtered, err := deploymentService.GetDeployments(ctx, model.DeploymentFilter{
		Query:       "canary",
		Environment: "staging",
		Status:      "RUNNING",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(filtered) != 1 {
		t.Fatalf("expected 1 deployment, got %d", len(filtered))
	}
	if filtered[0].ID != "dep-payroll-100" {
		t.Fatalf("expected dep-payroll-100, got %q", filtered[0].ID)
	}
}
