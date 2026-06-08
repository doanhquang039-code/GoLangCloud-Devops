package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"hr-cloud-service/internal/model"
	"hr-cloud-service/internal/repository"
	"hr-cloud-service/internal/service"
)

func TestEnvironmentControllerFiltersEnvironmentsByQuery(t *testing.T) {
	ctx := context.Background()
	environmentRepository := repository.NewInMemoryEnvironmentRepository()
	environmentService := service.NewEnvironmentService(
		repository.NewInMemoryApplicationRepository(),
		repository.NewInMemoryClusterRepository(),
		environmentRepository,
	)
	environmentController := NewEnvironmentController(environmentService)

	if _, err := environmentRepository.Save(ctx, model.Environment{
		ID:            "env-payroll-staging",
		Name:          "payroll-staging",
		Type:          "Staging",
		ApplicationID: "app-payroll-api",
		ClusterID:     "cls-staging-sg",
		Namespace:     "hr-staging",
		Status:        "Active",
		Variables: map[string]string{
			"FEATURE_FLAGS": "payroll-v2",
		},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := environmentRepository.Save(ctx, model.Environment{
		ID:            "env-people-prod",
		Name:          "people-production",
		Type:          "production",
		ApplicationID: "app-people-web",
		ClusterID:     "cls-prod-us",
		Namespace:     "people-prod",
		Status:        "inactive",
	}); err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/environments?q=payroll-v2&type=staging&status=active", nil)
	response := httptest.NewRecorder()

	environmentController.Index(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var environments []model.Environment
	if err := json.NewDecoder(response.Body).Decode(&environments); err != nil {
		t.Fatal(err)
	}
	if len(environments) != 1 || environments[0].ID != "env-payroll-staging" {
		t.Fatalf("expected filtered environment env-payroll-staging, got %#v", environments)
	}
}
