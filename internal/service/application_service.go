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

func (s *ApplicationService) GetApplications(ctx context.Context, filter model.ApplicationFilter) ([]model.Application, error) {
	filter.Query = strings.TrimSpace(filter.Query)
	filter.OwnerTeam = strings.TrimSpace(filter.OwnerTeam)
	filter.Criticality = strings.TrimSpace(filter.Criticality)
	filter.Runtime = strings.TrimSpace(filter.Runtime)
	filter.Tag = strings.TrimSpace(filter.Tag)

	applications, err := s.applicationRepository.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	if filter.Query == "" && filter.OwnerTeam == "" && filter.Criticality == "" && filter.Runtime == "" && filter.Tag == "" {
		return applications, nil
	}

	filtered := make([]model.Application, 0, len(applications))
	for _, application := range applications {
		if filter.Query != "" && !applicationMatchesQuery(application, filter.Query) {
			continue
		}
		if filter.OwnerTeam != "" && !strings.EqualFold(application.OwnerTeam, filter.OwnerTeam) {
			continue
		}
		if filter.Criticality != "" && !strings.EqualFold(application.Criticality, filter.Criticality) {
			continue
		}
		if filter.Runtime != "" && !strings.EqualFold(application.Runtime, filter.Runtime) {
			continue
		}
		if filter.Tag != "" && !hasApplicationTag(application.Tags, filter.Tag) {
			continue
		}
		filtered = append(filtered, application)
	}

	return filtered, nil
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
	request.Criticality = strings.ToLower(request.Criticality)
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
	request.Criticality = strings.ToLower(request.Criticality)
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

func (s *ApplicationService) DeleteApplication(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrInvalidApplication
	}

	return s.applicationRepository.DeleteByID(ctx, id)
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

func hasApplicationTag(tags []string, wanted string) bool {
	for _, tag := range tags {
		if strings.EqualFold(tag, wanted) {
			return true
		}
	}

	return false
}

func applicationMatchesQuery(application model.Application, query string) bool {
	query = strings.ToLower(query)
	if strings.Contains(strings.ToLower(application.ID), query) ||
		strings.Contains(strings.ToLower(application.Name), query) ||
		strings.Contains(strings.ToLower(application.Repository), query) ||
		strings.Contains(strings.ToLower(application.Runtime), query) ||
		strings.Contains(strings.ToLower(application.OwnerTeam), query) ||
		strings.Contains(strings.ToLower(application.Criticality), query) ||
		strings.Contains(strings.ToLower(application.HealthEndpoint), query) {
		return true
	}
	for _, tag := range application.Tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}
	for key, value := range application.Environment {
		if strings.Contains(strings.ToLower(key), query) || strings.Contains(strings.ToLower(value), query) {
			return true
		}
	}

	return false
}
