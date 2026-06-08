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

func TestPipelineControllerUpdateStageStatus(t *testing.T) {
	ctx := context.Background()
	pipelineRepository := repository.NewInMemoryPipelineRepository()
	pipelineService := service.NewPipelineService(repository.NewInMemoryApplicationRepository(), pipelineRepository)
	pipelineController := NewPipelineController(pipelineService)

	if _, err := pipelineRepository.Save(ctx, model.PipelineRun{
		ID:     "pipe-main-100",
		Status: "running",
		Stages: []model.PipelineStage{
			{Name: "build", Status: "pending"},
			{Name: "unit-test", Status: "pending"},
		},
	}); err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodPatch, "/api/v1/pipelines/pipe-main-100/stages/build", bytes.NewBufferString(`{"status":"SUCCEEDED"}`))
	response := httptest.NewRecorder()

	pipelineController.ShowOrUpdateStatus(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var pipelineRun model.PipelineRun
	if err := json.NewDecoder(response.Body).Decode(&pipelineRun); err != nil {
		t.Fatal(err)
	}
	if pipelineRun.Stages[0].Status != "succeeded" {
		t.Fatalf("expected normalized succeeded stage, got %q", pipelineRun.Stages[0].Status)
	}
	if pipelineRun.Status != "running" {
		t.Fatalf("expected pipeline to remain running, got %q", pipelineRun.Status)
	}
}

func TestPipelineControllerReturnsNotFoundForUnknownStage(t *testing.T) {
	ctx := context.Background()
	pipelineRepository := repository.NewInMemoryPipelineRepository()
	pipelineService := service.NewPipelineService(repository.NewInMemoryApplicationRepository(), pipelineRepository)
	pipelineController := NewPipelineController(pipelineService)

	if _, err := pipelineRepository.Save(ctx, model.PipelineRun{
		ID:     "pipe-main-100",
		Status: "running",
		Stages: []model.PipelineStage{
			{Name: "build", Status: "pending"},
		},
	}); err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodPatch, "/api/v1/pipelines/pipe-main-100/stages/deploy", bytes.NewBufferString(`{"status":"succeeded"}`))
	response := httptest.NewRecorder()

	pipelineController.ShowOrUpdateStatus(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d: %s", http.StatusNotFound, response.Code, response.Body.String())
	}
}

func TestPipelineControllerFiltersPipelineRunsByQuery(t *testing.T) {
	ctx := context.Background()
	pipelineRepository := repository.NewInMemoryPipelineRepository()
	pipelineService := service.NewPipelineService(repository.NewInMemoryApplicationRepository(), pipelineRepository)
	pipelineController := NewPipelineController(pipelineService)

	if _, err := pipelineRepository.Save(ctx, model.PipelineRun{
		ID:            "pipe-main-100",
		ApplicationID: "app-payroll-api",
		Branch:        "main",
		CommitSHA:     "abc123",
		TriggeredBy:   "devops@example.com",
		Status:        "running",
		Stages: []model.PipelineStage{
			{Name: "security-scan", Status: "running"},
		},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := pipelineRepository.Save(ctx, model.PipelineRun{
		ID:          "pipe-dev-100",
		Branch:      "dev",
		CommitSHA:   "def456",
		TriggeredBy: "devops@example.com",
		Status:      "failed",
	}); err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/pipelines?q=security&status=running", nil)
	response := httptest.NewRecorder()

	pipelineController.Index(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var pipelineRuns []model.PipelineRun
	if err := json.NewDecoder(response.Body).Decode(&pipelineRuns); err != nil {
		t.Fatal(err)
	}
	if len(pipelineRuns) != 1 || pipelineRuns[0].ID != "pipe-main-100" {
		t.Fatalf("expected filtered pipeline run pipe-main-100, got %#v", pipelineRuns)
	}
}
