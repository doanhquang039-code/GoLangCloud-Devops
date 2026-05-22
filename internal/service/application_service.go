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

var ErrInvalidApplication = errors.New("invalid application input")

type ApplicationService struct {
	applicationRepository repository.ApplicationRepository
}

func NewApplicationService(applicationRepository repository.ApplicationRepository) *ApplicationService {
	return &ApplicationService{applicationRepository: applicationRepository}
}

func (s *ApplicationService) GetApplications(ctx context.Context) ([]model.Application, error) {
	return s.applicationRepository.FindAll(ctx)
}

func (s *ApplicationService) GetApplicationByID(ctx context.Context, id string) (model.Application, error) {
	if strings.TrimSpace(id) == "" {
		return model.Application{}, ErrInvalidApplication
	}

	return s.applicationRepository.FindByID(ctx, id)
}

func (s *ApplicationService) CreateApplication(ctx context.Context, request model.CreateApplicationRequest) (model.Application, error) {
	request.Name = strings.TrimSpace(request.Name)
	request.Repository = strings.TrimSpace(request.Repository)
	request.Runtime = strings.TrimSpace(request.Runtime)
	request.OwnerTeam = strings.TrimSpace(request.OwnerTeam)
	request.Criticality = strings.TrimSpace(request.Criticality)
	request.HealthEndpoint = strings.TrimSpace(request.HealthEndpoint)

	if request.Name == "" || request.Repository == "" || request.Runtime == "" || request.OwnerTeam == "" {
		return model.Application{}, ErrInvalidApplication
	}
	if request.Criticality == "" {
		request.Criticality = "medium"
	}
	if request.Port == 0 {
		request.Port = 8080
	}
	if request.Replicas == 0 {
		request.Replicas = 1
	}
	if request.HealthEndpoint == "" {
		request.HealthEndpoint = "/healthz"
	}
	if request.Port < 1 || request.Port > 65535 || request.Replicas < 1 {
		return model.Application{}, ErrInvalidApplication
	}

	now := time.Now().UTC()
	application := model.Application{
		ID:             fmt.Sprintf("app-%d", now.UnixNano()),
		Name:           request.Name,
		Repository:     request.Repository,
		Runtime:        request.Runtime,
		OwnerTeam:      request.OwnerTeam,
		Criticality:    request.Criticality,
		Port:           request.Port,
		Replicas:       request.Replicas,
		HealthEndpoint: request.HealthEndpoint,
		Environment:    request.Environment,
		Tags:           normalizeTags(request.Tags),
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	return s.applicationRepository.Save(ctx, application)
}

func (s *ApplicationService) UpdateApplication(ctx context.Context, id string, request model.UpdateApplicationRequest) (model.Application, error) {
	id = strings.TrimSpace(id)
	request.Name = strings.TrimSpace(request.Name)
	request.Repository = strings.TrimSpace(request.Repository)
	request.Runtime = strings.TrimSpace(request.Runtime)
	request.OwnerTeam = strings.TrimSpace(request.OwnerTeam)
	request.Criticality = strings.TrimSpace(request.Criticality)
	request.HealthEndpoint = strings.TrimSpace(request.HealthEndpoint)

	if id == "" || request.Name == "" || request.Repository == "" || request.Runtime == "" || request.OwnerTeam == "" {
		return model.Application{}, ErrInvalidApplication
	}
	if request.Criticality == "" {
		request.Criticality = "medium"
	}
	if request.Port == 0 {
		request.Port = 8080
	}
	if request.Replicas == 0 {
		request.Replicas = 1
	}
	if request.HealthEndpoint == "" {
		request.HealthEndpoint = "/healthz"
	}
	if request.Port < 1 || request.Port > 65535 || request.Replicas < 1 {
		return model.Application{}, ErrInvalidApplication
	}

	application, err := s.applicationRepository.FindByID(ctx, id)
	if err != nil {
		return model.Application{}, err
	}

	application.Name = request.Name
	application.Repository = request.Repository
	application.Runtime = request.Runtime
	application.OwnerTeam = request.OwnerTeam
	application.Criticality = request.Criticality
	application.Port = request.Port
	application.Replicas = request.Replicas
	application.HealthEndpoint = request.HealthEndpoint
	application.Environment = request.Environment
	application.Tags = normalizeTags(request.Tags)
	application.UpdatedAt = time.Now().UTC()

	return s.applicationRepository.Save(ctx, application)
}

func normalizeTags(tags []string) []string {
	normalized := make([]string, 0, len(tags))
	seen := map[string]struct{}{}
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		normalized = append(normalized, tag)
	}

	return normalized
}
