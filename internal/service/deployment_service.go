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

var ErrInvalidDeployment = errors.New("invalid deployment input")

type DeploymentService struct {
	applicationRepository repository.ApplicationRepository
	deploymentRepository  repository.DeploymentRepository
}

func NewDeploymentService(applicationRepository repository.ApplicationRepository, deploymentRepository repository.DeploymentRepository) *DeploymentService {
	return &DeploymentService{
		applicationRepository: applicationRepository,
		deploymentRepository:  deploymentRepository,
	}
}

func (s *DeploymentService) GetDeployments(ctx context.Context) ([]model.Deployment, error) {
	return s.deploymentRepository.FindAll(ctx)
}

func (s *DeploymentService) GetDeploymentByID(ctx context.Context, id string) (model.Deployment, error) {
	if strings.TrimSpace(id) == "" {
		return model.Deployment{}, ErrInvalidDeployment
	}

	return s.deploymentRepository.FindByID(ctx, id)
}

func (s *DeploymentService) CreateDeployment(ctx context.Context, request model.CreateDeploymentRequest) (model.Deployment, error) {
	request.ApplicationID = strings.TrimSpace(request.ApplicationID)
	request.Environment = strings.TrimSpace(request.Environment)
	request.Version = strings.TrimSpace(request.Version)
	request.RequestedBy = strings.TrimSpace(request.RequestedBy)

	if request.ApplicationID == "" || request.Environment == "" || request.Version == "" || request.RequestedBy == "" {
		return model.Deployment{}, ErrInvalidDeployment
	}

	if _, err := s.applicationRepository.FindByID(ctx, request.ApplicationID); err != nil {
		return model.Deployment{}, err
	}

	now := time.Now().UTC()
	deployment := model.Deployment{
		ID:            fmt.Sprintf("dep-%d", now.UnixNano()),
		ApplicationID: request.ApplicationID,
		Environment:   request.Environment,
		Version:       request.Version,
		Status:        "running",
		RequestedBy:   request.RequestedBy,
		StartedAt:     now,
	}

	return s.deploymentRepository.Save(ctx, deployment)
}

func (s *DeploymentService) UpdateDeploymentStatus(ctx context.Context, id string, request model.UpdateDeploymentStatusRequest) (model.Deployment, error) {
	status := strings.TrimSpace(request.Status)
	if strings.TrimSpace(id) == "" || status == "" {
		return model.Deployment{}, ErrInvalidDeployment
	}
	if status != "running" && status != "succeeded" && status != "failed" {
		return model.Deployment{}, ErrInvalidDeployment
	}

	deployment, err := s.deploymentRepository.FindByID(ctx, id)
	if err != nil {
		return model.Deployment{}, err
	}

	deployment.Status = status
	if status == "succeeded" || status == "failed" {
		deployment.FinishedAt = time.Now().UTC()
	}

	return s.deploymentRepository.Save(ctx, deployment)
}
