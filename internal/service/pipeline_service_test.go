package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"hr-cloud-service/internal/model"
	"hr-cloud-service/internal/repository"
)

func TestUpdatePipelineStageStatusDerivesRunStatus(t *testing.T) {
	ctx := context.Background()
	applicationRepository := repository.NewInMemoryApplicationRepository()
	pipelineRepository := repository.NewInMemoryPipelineRepository()
	pipelineService := NewPipelineService(applicationRepository, pipelineRepository)
	now := time.Date(2026, 5, 31, 10, 0, 0, 0, time.UTC)

	if _, err := pipelineRepository.Save(ctx, model.PipelineRun{
		ID:            "pipe-main-100",
		ApplicationID: "app-payroll-api",
		Branch:        "main",
		Status:        "running",
		Stages: []model.PipelineStage{
			{Name: "build", Status: "pending", StartedAt: now},
			{Name: "unit-test", Status: "pending", StartedAt: now},
		},
		StartedAt: now,
	}); err != nil {
		t.Fatal(err)
	}

	pipelineRun, err := pipelineService.UpdatePipelineStageStatus(ctx, "pipe-main-100", "BUILD", model.UpdatePipelineStageStatusRequest{
		Status: "succeeded",
	})
	if err != nil {
		t.Fatal(err)
	}
	if pipelineRun.Status != "running" {
		t.Fatalf("expected running pipeline after one stage succeeds, got %q", pipelineRun.Status)
	}
	if pipelineRun.Stages[0].EndedAt == nil {
		t.Fatal("expected completed stage to have ended_at")
	}

	pipelineRun, err = pipelineService.UpdatePipelineStageStatus(ctx, "pipe-main-100", "unit-test", model.UpdatePipelineStageStatusRequest{
		Status: "succeeded",
	})
	if err != nil {
		t.Fatal(err)
	}
	if pipelineRun.Status != "succeeded" {
		t.Fatalf("expected succeeded pipeline after all stages succeed, got %q", pipelineRun.Status)
	}
	if pipelineRun.FinishedAt == nil {
		t.Fatal("expected pipeline to have finished_at")
	}
}

func TestUpdatePipelineStageStatusReturnsNotFoundForUnknownStage(t *testing.T) {
	ctx := context.Background()
	pipelineRepository := repository.NewInMemoryPipelineRepository()
	pipelineService := NewPipelineService(repository.NewInMemoryApplicationRepository(), pipelineRepository)

	if _, err := pipelineRepository.Save(ctx, model.PipelineRun{
		ID:     "pipe-main-100",
		Status: "running",
		Stages: []model.PipelineStage{
			{Name: "build", Status: "pending"},
		},
	}); err != nil {
		t.Fatal(err)
	}

	_, err := pipelineService.UpdatePipelineStageStatus(ctx, "pipe-main-100", "deploy", model.UpdatePipelineStageStatusRequest{
		Status: "succeeded",
	})
	if !errors.Is(err, ErrPipelineStageNotFound) {
		t.Fatalf("expected stage not found error, got %v", err)
	}
}

func TestCreatePipelineRunRejectsBlankCustomStages(t *testing.T) {
	ctx := context.Background()
	applicationRepository := repository.NewInMemoryApplicationRepository()
	pipelineRepository := repository.NewInMemoryPipelineRepository()
	pipelineService := NewPipelineService(applicationRepository, pipelineRepository)

	if _, err := applicationRepository.Save(ctx, model.Application{ID: "app-payroll-api"}); err != nil {
		t.Fatal(err)
	}

	_, err := pipelineService.CreatePipelineRun(ctx, model.CreatePipelineRunRequest{
		ApplicationID: "app-payroll-api",
		Branch:        "main",
		CommitSHA:     "abc123",
		TriggeredBy:   "devops@example.com",
		Stages:        []string{" ", ""},
	})
	if !errors.Is(err, ErrInvalidPipelineRun) {
		t.Fatalf("expected invalid pipeline run for blank custom stages, got %v", err)
	}
}

func TestGetPipelineRunsFiltersStatusAndBranchCaseInsensitively(t *testing.T) {
	ctx := context.Background()
	pipelineRepository := repository.NewInMemoryPipelineRepository()
	pipelineService := NewPipelineService(repository.NewInMemoryApplicationRepository(), pipelineRepository)

	if _, err := pipelineRepository.Save(ctx, model.PipelineRun{
		ID:            "pipe-main-100",
		ApplicationID: "app-payroll-api",
		Branch:        "Main",
		CommitSHA:     "abc123",
		Status:        "Running",
		TriggeredBy:   "DevOps@Example.com",
		Stages: []model.PipelineStage{
			{Name: "security-scan", Status: "running"},
		},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := pipelineRepository.Save(ctx, model.PipelineRun{
		ID:          "pipe-dev-100",
		Branch:      "dev",
		Status:      "failed",
		TriggeredBy: "devops@example.com",
	}); err != nil {
		t.Fatal(err)
	}

	pipelineRuns, err := pipelineService.GetPipelineRuns(ctx, model.PipelineRunFilter{
		Query:       "security",
		Branch:      "main",
		Status:      "RUNNING",
		TriggeredBy: "devops@example.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(pipelineRuns) != 1 {
		t.Fatalf("expected 1 pipeline run, got %d", len(pipelineRuns))
	}
	if pipelineRuns[0].ID != "pipe-main-100" {
		t.Fatalf("expected main pipeline, got %q", pipelineRuns[0].ID)
	}
}

func TestUpdatePipelineStatusNormalizesUppercaseStatus(t *testing.T) {
	ctx := context.Background()
	pipelineRepository := repository.NewInMemoryPipelineRepository()
	pipelineService := NewPipelineService(repository.NewInMemoryApplicationRepository(), pipelineRepository)

	if _, err := pipelineRepository.Save(ctx, model.PipelineRun{
		ID:     "pipe-main-100",
		Status: "running",
		Stages: []model.PipelineStage{
			{Name: "build", Status: "running"},
		},
	}); err != nil {
		t.Fatal(err)
	}

	pipelineRun, err := pipelineService.UpdatePipelineRunStatus(ctx, "pipe-main-100", model.UpdatePipelineRunStatusRequest{
		Status: "SUCCEEDED",
	})
	if err != nil {
		t.Fatal(err)
	}
	if pipelineRun.Status != "succeeded" {
		t.Fatalf("expected normalized succeeded status, got %q", pipelineRun.Status)
	}
	if pipelineRun.Stages[0].Status != "succeeded" {
		t.Fatalf("expected stage to be marked succeeded, got %q", pipelineRun.Stages[0].Status)
	}
}
