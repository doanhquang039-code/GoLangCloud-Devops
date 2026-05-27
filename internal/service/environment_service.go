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

var ErrInvalidEnvironment = errors.New("invalid environment input")

type EnvironmentService struct {
	applicationRepository repository.ApplicationRepository
	clusterRepository     repository.ClusterRepository
	environmentRepository repository.EnvironmentRepository
}

func NewEnvironmentService(
	applicationRepository repository.ApplicationRepository,
	clusterRepository repository.ClusterRepository,
	environmentRepository repository.EnvironmentRepository,
) *EnvironmentService {
	return &EnvironmentService{
		applicationRepository: applicationRepository,
		clusterRepository:     clusterRepository,
		environmentRepository: environmentRepository,
	}
}

func (s *EnvironmentService) GetEnvironments(ctx context.Context, filter model.EnvironmentFilter) ([]model.Environment, error) {
	filter.ApplicationID = strings.TrimSpace(filter.ApplicationID)
	filter.ClusterID = strings.TrimSpace(filter.ClusterID)
	filter.Type = strings.TrimSpace(filter.Type)
	filter.Status = strings.TrimSpace(filter.Status)

	if filter.Type != "" && !isValidEnvironmentType(filter.Type) {
		return nil, ErrInvalidEnvironment
	}
	if filter.Status != "" && !isValidEnvironmentStatus(filter.Status) {
		return nil, ErrInvalidEnvironment
	}

	environments, err := s.environmentRepository.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	if filter.ApplicationID == "" && filter.ClusterID == "" && filter.Type == "" && filter.Status == "" {
		return environments, nil
	}

	filtered := make([]model.Environment, 0, len(environments))
	for _, environment := range environments {
		if filter.ApplicationID != "" && environment.ApplicationID != filter.ApplicationID {
			continue
		}
		if filter.ClusterID != "" && environment.ClusterID != filter.ClusterID {
			continue
		}
		if filter.Type != "" && environment.Type != filter.Type {
			continue
		}
		if filter.Status != "" && environment.Status != filter.Status {
			continue
		}
		filtered = append(filtered, environment)
	}

	return filtered, nil
}

func (s *EnvironmentService) GetEnvironmentByID(ctx context.Context, id string) (model.Environment, error) {
	if strings.TrimSpace(id) == "" {
		return model.Environment{}, ErrInvalidEnvironment
	}

	return s.environmentRepository.FindByID(ctx, id)
}

func (s *EnvironmentService) CreateEnvironment(ctx context.Context, request model.CreateEnvironmentRequest) (model.Environment, error) {
	environment, err := s.buildEnvironment(ctx, model.Environment{}, request.Name, request.Type, request.ApplicationID, request.ClusterID, request.Namespace, request.Status, request.Variables)
	if err != nil {
		return model.Environment{}, err
	}

	now := time.Now().UTC()
	environment.ID = fmt.Sprintf("env-%d", now.UnixNano())
	environment.CreatedAt = now
	environment.UpdatedAt = now

	return s.environmentRepository.Save(ctx, environment)
}

func (s *EnvironmentService) UpdateEnvironment(ctx context.Context, id string, request model.UpdateEnvironmentRequest) (model.Environment, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return model.Environment{}, ErrInvalidEnvironment
	}

	existing, err := s.environmentRepository.FindByID(ctx, id)
	if err != nil {
		return model.Environment{}, err
	}

	environment, err := s.buildEnvironment(ctx, existing, request.Name, request.Type, request.ApplicationID, request.ClusterID, request.Namespace, request.Status, request.Variables)
	if err != nil {
		return model.Environment{}, err
	}
	environment.UpdatedAt = time.Now().UTC()

	return s.environmentRepository.Save(ctx, environment)
}

func (s *EnvironmentService) DeleteEnvironment(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrInvalidEnvironment
	}

	return s.environmentRepository.DeleteByID(ctx, id)
}

func (s *EnvironmentService) buildEnvironment(ctx context.Context, environment model.Environment, name, environmentType, applicationID, clusterID, namespace, status string, variables map[string]string) (model.Environment, error) {
	name = strings.TrimSpace(name)
	environmentType = strings.TrimSpace(environmentType)
	applicationID = strings.TrimSpace(applicationID)
	clusterID = strings.TrimSpace(clusterID)
	namespace = strings.TrimSpace(namespace)
	status = strings.TrimSpace(status)

	if name == "" || environmentType == "" || applicationID == "" || clusterID == "" {
		return model.Environment{}, ErrInvalidEnvironment
	}
	if namespace == "" {
		namespace = "default"
	}
	if status == "" {
		status = "active"
	}
	if !isValidEnvironmentType(environmentType) || !isValidEnvironmentStatus(status) {
		return model.Environment{}, ErrInvalidEnvironment
	}

	if _, err := s.applicationRepository.FindByID(ctx, applicationID); err != nil {
		return model.Environment{}, err
	}
	if _, err := s.clusterRepository.FindByID(ctx, clusterID); err != nil {
		return model.Environment{}, err
	}

	environment.Name = name
	environment.Type = environmentType
	environment.ApplicationID = applicationID
	environment.ClusterID = clusterID
	environment.Namespace = namespace
	environment.Status = status
	environment.Variables = variables

	return environment, nil
}

func isValidEnvironmentType(environmentType string) bool {
	return environmentType == "development" || environmentType == "staging" || environmentType == "production"
}

func isValidEnvironmentStatus(status string) bool {
	return status == "active" || status == "inactive" || status == "deprecated"
}
