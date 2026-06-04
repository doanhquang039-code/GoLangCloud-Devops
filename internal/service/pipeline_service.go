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
var ErrPipelineStageNotFound = errors.New("pipeline stage not found")

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

func (s *PipelineService) GetPipelineRuns(ctx context.Context, filter model.PipelineRunFilter) ([]model.PipelineRun, error) {
	filter.Query = strings.TrimSpace(filter.Query)
	filter.ApplicationID = strings.TrimSpace(filter.ApplicationID)
	filter.Branch = strings.TrimSpace(filter.Branch)
	filter.Status = strings.ToLower(strings.TrimSpace(filter.Status))
	filter.TriggeredBy = strings.TrimSpace(filter.TriggeredBy)

	if filter.Status != "" && !isValidPipelineRunStatus(filter.Status) {
		return nil, ErrInvalidPipelineRun
	}

	pipelineRuns, err := s.pipelineRepository.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	if filter.Query == "" && filter.ApplicationID == "" && filter.Branch == "" && filter.Status == "" && filter.TriggeredBy == "" {
		return pipelineRuns, nil
	}

	filtered := make([]model.PipelineRun, 0, len(pipelineRuns))
	for _, pipelineRun := range pipelineRuns {
		if filter.Query != "" && !pipelineRunMatchesQuery(pipelineRun, filter.Query) {
			continue
		}
		if filter.ApplicationID != "" && pipelineRun.ApplicationID != filter.ApplicationID {
			continue
		}
		if filter.Branch != "" && !strings.EqualFold(pipelineRun.Branch, filter.Branch) {
			continue
		}
		if filter.Status != "" && !strings.EqualFold(pipelineRun.Status, filter.Status) {
			continue
		}
		if filter.TriggeredBy != "" && !strings.EqualFold(pipelineRun.TriggeredBy, filter.TriggeredBy) {
			continue
		}
		filtered = append(filtered, pipelineRun)
	}

	return filtered, nil
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
		StartedAt:     now,
	}
	pipelineRun.Stages = buildPipelineStages(request.Stages, now)
	if len(pipelineRun.Stages) == 0 {
		return model.PipelineRun{}, ErrInvalidPipelineRun
	}

	return s.pipelineRepository.Save(ctx, pipelineRun)
}

func (s *PipelineService) UpdatePipelineRunStatus(ctx context.Context, id string, request model.UpdatePipelineRunStatusRequest) (model.PipelineRun, error) {
	status := strings.ToLower(strings.TrimSpace(request.Status))
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
		pipelineRun.FinishedAt = &now
		pipelineRun.Stages = finishPipelineStages(pipelineRun.Stages, status, now)
	}

	return s.pipelineRepository.Save(ctx, pipelineRun)
}

func (s *PipelineService) UpdatePipelineStageStatus(ctx context.Context, id string, stageName string, request model.UpdatePipelineStageStatusRequest) (model.PipelineRun, error) {
	id = strings.TrimSpace(id)
	stageName = strings.TrimSpace(stageName)
	status := strings.ToLower(strings.TrimSpace(request.Status))
	if id == "" || stageName == "" || !isValidPipelineStageStatus(status) {
		return model.PipelineRun{}, ErrInvalidPipelineRun
	}

	pipelineRun, err := s.pipelineRepository.FindByID(ctx, id)
	if err != nil {
		return model.PipelineRun{}, err
	}

	now := time.Now().UTC()
	found := false
	for i := range pipelineRun.Stages {
		if !strings.EqualFold(pipelineRun.Stages[i].Name, stageName) {
			continue
		}
		pipelineRun.Stages[i].Status = status
		if isTerminalPipelineStageStatus(status) {
			pipelineRun.Stages[i].EndedAt = &now
		} else {
			pipelineRun.Stages[i].EndedAt = nil
		}
		found = true
		break
	}
	if !found {
		return model.PipelineRun{}, ErrPipelineStageNotFound
	}

	pipelineRun.Status = derivePipelineRunStatus(pipelineRun.Stages)
	if isTerminalPipelineRunStatus(pipelineRun.Status) {
		pipelineRun.FinishedAt = &now
	} else {
		pipelineRun.FinishedAt = nil
	}

	return s.pipelineRepository.Save(ctx, pipelineRun)
}

func (s *PipelineService) DeletePipelineRun(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrInvalidPipelineRun
	}

	return s.pipelineRepository.DeleteByID(ctx, id)
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
		stages[i].EndedAt = &now
	}

	return stages
}

func isValidPipelineRunStatus(status string) bool {
	return status == "running" || status == "succeeded" || status == "failed" || status == "cancelled"
}

func isTerminalPipelineRunStatus(status string) bool {
	return status == "succeeded" || status == "failed" || status == "cancelled"
}

func isValidPipelineStageStatus(status string) bool {
	return status == "pending" || status == "running" || status == "succeeded" || status == "failed" || status == "skipped"
}

func isTerminalPipelineStageStatus(status string) bool {
	return status == "succeeded" || status == "failed" || status == "skipped"
}

func derivePipelineRunStatus(stages []model.PipelineStage) string {
	if len(stages) == 0 {
		return "running"
	}

	allSucceeded := true
	allTerminal := true
	for _, stage := range stages {
		switch stage.Status {
		case "failed":
			return "failed"
		case "succeeded":
		case "skipped":
			allSucceeded = false
		default:
			allSucceeded = false
			allTerminal = false
		}
	}

	if allSucceeded {
		return "succeeded"
	}
	if allTerminal {
		return "cancelled"
	}

	return "running"
}

func pipelineRunMatchesQuery(pipelineRun model.PipelineRun, query string) bool {
	query = strings.ToLower(query)
	if strings.Contains(strings.ToLower(pipelineRun.ID), query) ||
		strings.Contains(strings.ToLower(pipelineRun.ApplicationID), query) ||
		strings.Contains(strings.ToLower(pipelineRun.Branch), query) ||
		strings.Contains(strings.ToLower(pipelineRun.CommitSHA), query) ||
		strings.Contains(strings.ToLower(pipelineRun.TriggeredBy), query) ||
		strings.Contains(strings.ToLower(pipelineRun.Status), query) {
		return true
	}
	for _, stage := range pipelineRun.Stages {
		if strings.Contains(strings.ToLower(stage.Name), query) ||
			strings.Contains(strings.ToLower(stage.Status), query) {
			return true
		}
	}

	return false
}
