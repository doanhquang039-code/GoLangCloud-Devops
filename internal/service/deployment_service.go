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
	clusterRepository     repository.ClusterRepository
	deploymentRepository  repository.DeploymentRepository
}

func NewDeploymentService(
	applicationRepository repository.ApplicationRepository,
	clusterRepository repository.ClusterRepository,
	deploymentRepository repository.DeploymentRepository,
) *DeploymentService {
	return &DeploymentService{
		applicationRepository: applicationRepository,
		clusterRepository:     clusterRepository,
		deploymentRepository:  deploymentRepository,
	}
}

func (s *DeploymentService) GetDeployments(ctx context.Context, filter model.DeploymentFilter) ([]model.Deployment, error) {
	filter.Query = strings.TrimSpace(filter.Query)
	filter.ApplicationID = strings.TrimSpace(filter.ApplicationID)
	filter.ClusterID = strings.TrimSpace(filter.ClusterID)
	filter.Namespace = strings.TrimSpace(filter.Namespace)
	filter.Environment = strings.TrimSpace(filter.Environment)
	filter.Status = strings.ToLower(strings.TrimSpace(filter.Status))

	if filter.Status != "" && !isValidDeploymentStatus(filter.Status) {
		return nil, ErrInvalidDeployment
	}

	deployments, err := s.deploymentRepository.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	if filter.Query == "" && filter.ApplicationID == "" && filter.ClusterID == "" && filter.Namespace == "" && filter.Environment == "" && filter.Status == "" {
		return deployments, nil
	}

	filtered := make([]model.Deployment, 0, len(deployments))
	for _, deployment := range deployments {
		if filter.Query != "" && !deploymentMatchesQuery(deployment, filter.Query) {
			continue
		}
		if filter.ApplicationID != "" && deployment.ApplicationID != filter.ApplicationID {
			continue
		}
		if filter.ClusterID != "" && deployment.ClusterID != filter.ClusterID {
			continue
		}
		if filter.Namespace != "" && !strings.EqualFold(deployment.Namespace, filter.Namespace) {
			continue
		}
		if filter.Environment != "" && !strings.EqualFold(deployment.Environment, filter.Environment) {
			continue
		}
		if filter.Status != "" && !strings.EqualFold(deployment.Status, filter.Status) {
			continue
		}
		filtered = append(filtered, deployment)
	}

	return filtered, nil
}

func (s *DeploymentService) GetDeploymentByID(ctx context.Context, id string) (model.Deployment, error) {
	if strings.TrimSpace(id) == "" {
		return model.Deployment{}, ErrInvalidDeployment
	}

	return s.deploymentRepository.FindByID(ctx, id)
}

func (s *DeploymentService) CreateDeployment(ctx context.Context, request model.CreateDeploymentRequest) (model.Deployment, error) {
	request.ApplicationID = strings.TrimSpace(request.ApplicationID)
	request.ClusterID = strings.TrimSpace(request.ClusterID)
	request.Namespace = strings.TrimSpace(request.Namespace)
	request.Environment = strings.TrimSpace(request.Environment)
	request.Version = strings.TrimSpace(request.Version)
	request.Strategy = strings.TrimSpace(request.Strategy)
	request.RequestedBy = strings.TrimSpace(request.RequestedBy)

	if request.ApplicationID == "" || request.ClusterID == "" || request.Environment == "" || request.Version == "" || request.RequestedBy == "" {
		return model.Deployment{}, ErrInvalidDeployment
	}
	if request.Namespace == "" {
		request.Namespace = "default"
	}
	if request.Strategy == "" {
		request.Strategy = "rolling"
	}
	if !isValidDeploymentStrategy(request.Strategy) {
		return model.Deployment{}, ErrInvalidDeployment
	}

	if _, err := s.applicationRepository.FindByID(ctx, request.ApplicationID); err != nil {
		return model.Deployment{}, err
	}
	if _, err := s.clusterRepository.FindByID(ctx, request.ClusterID); err != nil {
		return model.Deployment{}, err
	}

	now := time.Now().UTC()
	deployment := model.Deployment{
		ID:            fmt.Sprintf("dep-%d", now.UnixNano()),
		ApplicationID: request.ApplicationID,
		ClusterID:     request.ClusterID,
		Namespace:     request.Namespace,
		Environment:   request.Environment,
		Version:       request.Version,
		Strategy:      request.Strategy,
		Status:        "running",
		RequestedBy:   request.RequestedBy,
		StartedAt:     now,
	}

	return s.deploymentRepository.Save(ctx, deployment)
}

func isValidDeploymentStrategy(strategy string) bool {
	return strategy == "rolling" || strategy == "blue-green" || strategy == "canary"
}

func (s *DeploymentService) UpdateDeployment(ctx context.Context, id string, request model.UpdateDeploymentRequest) (model.Deployment, error) {
	id = strings.TrimSpace(id)
	request.ApplicationID = strings.TrimSpace(request.ApplicationID)
	request.ClusterID = strings.TrimSpace(request.ClusterID)
	request.Namespace = strings.TrimSpace(request.Namespace)
	request.Environment = strings.TrimSpace(request.Environment)
	request.Version = strings.TrimSpace(request.Version)
	request.Strategy = strings.TrimSpace(request.Strategy)
	request.RequestedBy = strings.TrimSpace(request.RequestedBy)

	if id == "" || request.ApplicationID == "" || request.ClusterID == "" || request.Environment == "" || request.Version == "" || request.RequestedBy == "" {
		return model.Deployment{}, ErrInvalidDeployment
	}
	if request.Namespace == "" {
		request.Namespace = "default"
	}
	if request.Strategy == "" {
		request.Strategy = "rolling"
	}
	if !isValidDeploymentStrategy(request.Strategy) {
		return model.Deployment{}, ErrInvalidDeployment
	}

	if _, err := s.applicationRepository.FindByID(ctx, request.ApplicationID); err != nil {
		return model.Deployment{}, err
	}
	if _, err := s.clusterRepository.FindByID(ctx, request.ClusterID); err != nil {
		return model.Deployment{}, err
	}

	deployment, err := s.deploymentRepository.FindByID(ctx, id)
	if err != nil {
		return model.Deployment{}, err
	}

	deployment.ApplicationID = request.ApplicationID
	deployment.ClusterID = request.ClusterID
	deployment.Namespace = request.Namespace
	deployment.Environment = request.Environment
	deployment.Version = request.Version
	deployment.Strategy = request.Strategy
	deployment.RequestedBy = request.RequestedBy

	return s.deploymentRepository.Save(ctx, deployment)
}

func (s *DeploymentService) UpdateDeploymentStatus(ctx context.Context, id string, request model.UpdateDeploymentStatusRequest) (model.Deployment, error) {
	status := strings.TrimSpace(request.Status)
	if strings.TrimSpace(id) == "" || !isValidDeploymentStatus(status) {
		return model.Deployment{}, ErrInvalidDeployment
	}

	deployment, err := s.deploymentRepository.FindByID(ctx, id)
	if err != nil {
		return model.Deployment{}, err
	}

	deployment.Status = status
	if status == "succeeded" || status == "failed" {
		finishedAt := time.Now().UTC()
		deployment.FinishedAt = &finishedAt
	}

	return s.deploymentRepository.Save(ctx, deployment)
}

func (s *DeploymentService) DeleteDeployment(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrInvalidDeployment
	}

	return s.deploymentRepository.DeleteByID(ctx, id)
}

func isValidDeploymentStatus(status string) bool {
	return status == "running" || status == "succeeded" || status == "failed"
}

func deploymentMatchesQuery(deployment model.Deployment, query string) bool {
	query = strings.ToLower(query)
	return strings.Contains(strings.ToLower(deployment.ID), query) ||
		strings.Contains(strings.ToLower(deployment.ApplicationID), query) ||
		strings.Contains(strings.ToLower(deployment.ClusterID), query) ||
		strings.Contains(strings.ToLower(deployment.Namespace), query) ||
		strings.Contains(strings.ToLower(deployment.Environment), query) ||
		strings.Contains(strings.ToLower(deployment.Version), query) ||
		strings.Contains(strings.ToLower(deployment.Strategy), query) ||
		strings.Contains(strings.ToLower(deployment.Status), query) ||
		strings.Contains(strings.ToLower(deployment.RequestedBy), query)
}
