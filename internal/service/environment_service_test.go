package service

import (
	"context"
	"testing"

	"hr-cloud-service/internal/model"
	"hr-cloud-service/internal/repository"
)

func TestGetEnvironmentsFiltersByQueryTypeAndStatus(t *testing.T) {
	ctx := context.Background()
	environmentRepository := repository.NewInMemoryEnvironmentRepository()
	environmentService := NewEnvironmentService(
		repository.NewInMemoryApplicationRepository(),
		repository.NewInMemoryClusterRepository(),
		environmentRepository,
	)

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

	filtered, err := environmentService.GetEnvironments(ctx, model.EnvironmentFilter{
		Query:  "payroll-v2",
		Type:   "STAGING",
		Status: "ACTIVE",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(filtered) != 1 || filtered[0].ID != "env-payroll-staging" {
		t.Fatalf("expected filtered environment env-payroll-staging, got %#v", filtered)
	}
}
