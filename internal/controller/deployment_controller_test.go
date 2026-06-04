package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"hr-cloud-service/internal/model"
	"hr-cloud-service/internal/repository"
	"hr-cloud-service/internal/service"
)

func TestDeploymentControllerFiltersDeploymentsByQuery(t *testing.T) {
	ctx := t.Context()
	deploymentRepository := repository.NewInMemoryDeploymentRepository()
	deploymentService := service.NewDeploymentService(
		repository.NewInMemoryApplicationRepository(),
		repository.NewInMemoryClusterRepository(),
		deploymentRepository,
	)
	deploymentController := NewDeploymentController(deploymentService)

	if _, err := deploymentRepository.Save(ctx, model.Deployment{
		ID:            "dep-payroll-100",
		ApplicationID: "app-payroll-api",
		ClusterID:     "cluster-staging",
		Namespace:     "HR-Staging",
		Environment:   "Staging",
		Version:       "v1.2.3",
		Strategy:      "canary",
		Status:        "Running",
		RequestedBy:   "devops@example.com",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := deploymentRepository.Save(ctx, model.Deployment{
		ID:            "dep-people-100",
		ApplicationID: "app-people-web",
		ClusterID:     "cluster-prod",
		Namespace:     "hr-prod",
		Environment:   "production",
		Version:       "v2.0.0",
		Strategy:      "rolling",
		Status:        "succeeded",
		RequestedBy:   "people@example.com",
	}); err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/deployments?q=canary&environment=staging&status=running", nil)
	response := httptest.NewRecorder()

	deploymentController.Index(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var deployments []model.Deployment
	if err := json.NewDecoder(response.Body).Decode(&deployments); err != nil {
		t.Fatal(err)
	}
	if len(deployments) != 1 || deployments[0].ID != "dep-payroll-100" {
		t.Fatalf("expected filtered deployment dep-payroll-100, got %#v", deployments)
	}
}
