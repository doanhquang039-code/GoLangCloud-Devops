package service

import (
	"context"
	"testing"

	"hr-cloud-service/internal/model"
	"hr-cloud-service/internal/repository"
)

func TestGetApplicationsFiltersByOwnerCriticalityRuntimeAndTag(t *testing.T) {
	ctx := context.Background()
	applicationRepository := repository.NewInMemoryApplicationRepository()
	applicationService := NewApplicationService(applicationRepository)

	applications := []model.Application{
		{
			ID:          "app-payroll-api",
			Name:        "payroll-api",
			Repository:  "github.com/company/payroll-api",
			OwnerTeam:   "platform",
			Criticality: "high",
			Runtime:     "go1.22",
			Environment: map[string]string{"FEATURE": "payroll-v2"},
			Tags:        []string{"payroll", "backend"},
		},
		{
			ID:          "app-people-web",
			Name:        "people-web",
			OwnerTeam:   "hr",
			Criticality: "medium",
			Runtime:     "node20",
			Tags:        []string{"frontend"},
		},
	}

	for _, application := range applications {
		if _, err := applicationRepository.Save(ctx, application); err != nil {
			t.Fatal(err)
		}
	}

	filtered, err := applicationService.GetApplications(ctx, model.ApplicationFilter{
		Query:       "payroll-v2",
		OwnerTeam:   "PLATFORM",
		Criticality: "HIGH",
		Runtime:     "GO1.22",
		Tag:         "BACKEND",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(filtered) != 1 {
		t.Fatalf("expected 1 application, got %d", len(filtered))
	}
	if filtered[0].ID != "app-payroll-api" {
		t.Fatalf("expected payroll application, got %q", filtered[0].ID)
	}
}

func TestCreateApplicationAppliesDefaultsAndNormalizesTags(t *testing.T) {
	ctx := context.Background()
	applicationRepository := repository.NewInMemoryApplicationRepository()
	applicationService := NewApplicationService(applicationRepository)

	application, err := applicationService.CreateApplication(ctx, model.CreateApplicationRequest{
		Name:        "  payroll-api  ",
		Repository:  " git@example.com/payroll-api.git ",
		Runtime:     " go1.22 ",
		OwnerTeam:   " platform ",
		Criticality: " HIGH ",
		Tags:        []string{" backend ", "", "backend", "payroll"},
	})
	if err != nil {
		t.Fatal(err)
	}

	if application.Port != 8080 {
		t.Fatalf("expected default port 8080, got %d", application.Port)
	}
	if application.Replicas != 1 {
		t.Fatalf("expected default replicas 1, got %d", application.Replicas)
	}
	if application.HealthEndpoint != "/healthz" {
		t.Fatalf("expected default health endpoint, got %q", application.HealthEndpoint)
	}
	if application.Criticality != "high" {
		t.Fatalf("expected normalized criticality high, got %q", application.Criticality)
	}
	if len(application.Tags) != 2 || application.Tags[0] != "backend" || application.Tags[1] != "payroll" {
		t.Fatalf("expected normalized unique tags, got %#v", application.Tags)
	}
}
