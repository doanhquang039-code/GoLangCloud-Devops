package service

import (
	"context"
	"testing"

	"hr-cloud-service/internal/model"
	"hr-cloud-service/internal/repository"
)

func TestGetIncidentsFiltersByQuerySeverityStatusAndOwner(t *testing.T) {
	ctx := context.Background()
	incidentRepository := repository.NewInMemoryIncidentRepository()
	incidentService := NewIncidentService(
		repository.NewInMemoryApplicationRepository(),
		repository.NewInMemoryClusterRepository(),
		repository.NewInMemoryDeploymentRepository(),
		incidentRepository,
	)

	if _, err := incidentRepository.Save(ctx, model.Incident{
		ID:            "inc-payroll-100",
		Title:         "Payroll API error rate elevated",
		Summary:       "Canary rollout caused 5xx spikes.",
		Severity:      "Sev2",
		Status:        "Investigating",
		ApplicationID: "app-payroll-api",
		OwnerTeam:     "Platform",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := incidentRepository.Save(ctx, model.Incident{
		ID:        "inc-people-100",
		Title:     "People web cache warming",
		Summary:   "Background maintenance.",
		Severity:  "sev4",
		Status:    "resolved",
		OwnerTeam: "people",
	}); err != nil {
		t.Fatal(err)
	}

	filtered, err := incidentService.GetIncidents(ctx, model.IncidentFilter{
		Query:     "canary",
		Severity:  "SEV2",
		Status:    "INVESTIGATING",
		OwnerTeam: "platform",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(filtered) != 1 || filtered[0].ID != "inc-payroll-100" {
		t.Fatalf("expected filtered incident inc-payroll-100, got %#v", filtered)
	}
}
