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

func TestIncidentControllerFiltersIncidentsByQuery(t *testing.T) {
	ctx := context.Background()
	incidentRepository := repository.NewInMemoryIncidentRepository()
	incidentService := service.NewIncidentService(
		repository.NewInMemoryApplicationRepository(),
		repository.NewInMemoryClusterRepository(),
		repository.NewInMemoryDeploymentRepository(),
		incidentRepository,
	)
	incidentController := NewIncidentController(incidentService)

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

	request := httptest.NewRequest(http.MethodGet, "/api/v1/incidents?q=canary&severity=sev2&status=investigating&owner_team=platform", nil)
	response := httptest.NewRecorder()

	incidentController.Index(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var incidents []model.Incident
	if err := json.NewDecoder(response.Body).Decode(&incidents); err != nil {
		t.Fatal(err)
	}
	if len(incidents) != 1 || incidents[0].ID != "inc-payroll-100" {
		t.Fatalf("expected filtered incident inc-payroll-100, got %#v", incidents)
	}
}
