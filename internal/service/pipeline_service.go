package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"hr-cloud-service/internal/model"
	"hr-cloud-service/internal/repository"
)

var ErrInvalidPipelineRun = errors.New("invalid pipeline run input")

type PipelineService struct {
	applicationRepository repository.ApplicationRepository
	pipelineRepository    repository.PipelineRepository
}

func NewPipelineService(
	applicationRepository repository.ApplicationRepository,
	pipelineRepository repository.PipelineRepository,
) *PipelineService {
	return &PipelineService{
		applicationRepository: applicationRepository,
		pipelineRepository:    pipelineRepository,
	}
}

func (s *PipelineService) GetPipelineRuns(ctx context.Context) ([]model.PipelineRun, error) {
	return s.pipelineRepository.FindAll(ctx)
}

func (s *PipelineService) GetPipelineRunByID(ctx context.Context, id string) (model.PipelineRun, error) {
	if strings.TrimSpace(id) == "" {
		return model.PipelineRun{}, ErrInvalidPipelineRun
	}

	return s.pipelineRepository.FindByID(ctx, id)
}

func (s *PipelineService) CreatePipelineRun(ctx context.Context, request model.CreatePipelineRunRequest) (model.PipelineRun, error) {
	request.ApplicationID = strings.TrimSpace(request.ApplicationID)
	request.Branch = strings.TrimSpace(request.Branch)
	request.CommitSHA = strings.TrimSpace(request.CommitSHA)
	request.TriggeredBy = strings.TrimSpace(request.TriggeredBy)

	if request.ApplicationID == "" || request.Branch == "" || request.CommitSHA == "" || request.TriggeredBy == "" {
		return model.PipelineRun{}, ErrInvalidPipelineRun
	}

	if _, err := s.applicationRepository.FindByID(ctx, request.ApplicationID); err != nil {
		return model.PipelineRun{}, err
	}

	now := time.Now().UTC()
	pipelineRun := model.PipelineRun{
		ID:            fmt.Sprintf("pipe-%d", now.UnixNano()),
		ApplicationID: request.ApplicationID,
		Branch:        request.Branch,
		CommitSHA:     request.CommitSHA,
		TriggeredBy:   request.TriggeredBy,
		Status:        "running",
		Stages:        buildPipelineStages(request.Stages, now),
		StartedAt:     now,
	}

	return s.pipelineRepository.Save(ctx, pipelineRun)
}

func (s *PipelineService) UpdatePipelineRunStatus(ctx context.Context, id string, request model.UpdatePipelineRunStatusRequest) (model.PipelineRun, error) {
	status := strings.TrimSpace(request.Status)
	if strings.TrimSpace(id) == "" || !isValidPipelineRunStatus(status) {
		return model.PipelineRun{}, ErrInvalidPipelineRun
	}

	pipelineRun, err := s.pipelineRepository.FindByID(ctx, id)
	if err != nil {
		return model.PipelineRun{}, err
	}

	now := time.Now().UTC()
	pipelineRun.Status = status
	if status != "running" {
		pipelineRun.FinishedAt = now
		pipelineRun.Stages = finishPipelineStages(pipelineRun.Stages, status, now)
	}

	return s.pipelineRepository.Save(ctx, pipelineRun)
}

func buildPipelineStages(stageNames []string, now time.Time) []model.PipelineStage {
	if len(stageNames) == 0 {
		stageNames = []string{"build", "test", "security-scan", "package"}
	}

	stages := make([]model.PipelineStage, 0, len(stageNames))
	seen := map[string]struct{}{}
	for _, stageName := range stageNames {
		stageName = strings.TrimSpace(stageName)
		if stageName == "" {
			continue
		}
		if _, ok := seen[stageName]; ok {
			continue
		}
		seen[stageName] = struct{}{}
		stages = append(stages, model.PipelineStage{
			Name:      stageName,
			Status:    "pending",
			StartedAt: now,
		})
	}

	return stages
}

func finishPipelineStages(stages []model.PipelineStage, status string, now time.Time) []model.PipelineStage {
	stageStatus := status
	if status == "cancelled" {
		stageStatus = "skipped"
	}

	for i := range stages {
		stages[i].Status = stageStatus
		stages[i].EndedAt = now
	}

	return stages
}

func isValidPipelineRunStatus(status string) bool {
	return status == "running" || status == "succeeded" || status == "failed" || status == "cancelled"
}
