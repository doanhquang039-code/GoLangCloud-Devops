package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"hr-cloud-service/internal/model"
	"hr-cloud-service/internal/repository"
	"hr-cloud-service/internal/service"
)

func TestApplicationControllerCreateApplication(t *testing.T) {
	applicationRepository := repository.NewInMemoryApplicationRepository()
	applicationService := service.NewApplicationService(applicationRepository)
	applicationController := NewApplicationController(applicationService)

	body := bytes.NewBufferString(`{
		"name": "payroll-api",
		"repository": "git@example.com/payroll-api.git",
		"runtime": "go1.22",
		"owner_team": "platform",
		"criticality": "HIGH",
		"tags": ["backend", "payroll"]
	}`)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/applications", body)
	response := httptest.NewRecorder()

	applicationController.Index(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}

	var application model.Application
	if err := json.NewDecoder(response.Body).Decode(&application); err != nil {
		t.Fatal(err)
	}
	if application.ID == "" {
		t.Fatal("expected generated application id")
	}
	if application.Criticality != "high" {
		t.Fatalf("expected normalized criticality high, got %q", application.Criticality)
	}
	if application.Port != 8080 || application.Replicas != 1 || application.HealthEndpoint != "/healthz" {
		t.Fatalf("expected default runtime settings, got port=%d replicas=%d health=%q", application.Port, application.Replicas, application.HealthEndpoint)
	}
}

func TestApplicationControllerRejectsInvalidCreateApplication(t *testing.T) {
	applicationRepository := repository.NewInMemoryApplicationRepository()
	applicationService := service.NewApplicationService(applicationRepository)
	applicationController := NewApplicationController(applicationService)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/applications", bytes.NewBufferString(`{"name":"missing-fields"}`))
	response := httptest.NewRecorder()

	applicationController.Index(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

func TestApplicationControllerFiltersApplicationsByQuery(t *testing.T) {
	ctx := context.Background()
	applicationRepository := repository.NewInMemoryApplicationRepository()
	applicationService := service.NewApplicationService(applicationRepository)
	applicationController := NewApplicationController(applicationService)

	if _, err := applicationRepository.Save(ctx, model.Application{
		ID:          "app-payroll-api",
		Name:        "payroll-api",
		Repository:  "github.com/company/payroll-api",
		Runtime:     "go1.22",
		OwnerTeam:   "platform",
		Criticality: "high",
		Tags:        []string{"backend"},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := applicationRepository.Save(ctx, model.Application{
		ID:          "app-people-web",
		Name:        "people-web",
		Repository:  "github.com/company/people-web",
		Runtime:     "node20",
		OwnerTeam:   "people",
		Criticality: "medium",
	}); err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/applications?q=payroll&owner_team=platform", nil)
	response := httptest.NewRecorder()

	applicationController.Index(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var applications []model.Application
	if err := json.NewDecoder(response.Body).Decode(&applications); err != nil {
		t.Fatal(err)
	}
	if len(applications) != 1 {
		t.Fatalf("expected 1 application, got %d", len(applications))
	}
	if applications[0].ID != "app-payroll-api" {
		t.Fatalf("expected payroll application, got %q", applications[0].ID)
	}
}
