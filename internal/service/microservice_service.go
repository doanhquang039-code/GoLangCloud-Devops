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

var ErrInvalidMicroservice = errors.New("invalid microservice input")

type MicroserviceService struct {
	applicationRepository  repository.ApplicationRepository
	microserviceRepository repository.MicroserviceRepository
}

func NewMicroserviceService(
	applicationRepository repository.ApplicationRepository,
	microserviceRepository repository.MicroserviceRepository,
) *MicroserviceService {
	return &MicroserviceService{
		applicationRepository:  applicationRepository,
		microserviceRepository: microserviceRepository,
	}
}

func (s *MicroserviceService) GetMicroservices(ctx context.Context, filter model.MicroserviceFilter) ([]model.Microservice, error) {
	filter.Query = strings.TrimSpace(filter.Query)
	filter.TenantID = strings.TrimSpace(filter.TenantID)
	filter.ApplicationID = strings.TrimSpace(filter.ApplicationID)
	filter.OwnerTeam = strings.TrimSpace(filter.OwnerTeam)
	filter.Protocol = strings.ToLower(strings.TrimSpace(filter.Protocol))
	filter.Status = strings.ToLower(strings.TrimSpace(filter.Status))
	filter.CloudProvider = strings.ToLower(strings.TrimSpace(filter.CloudProvider))
	filter.Region = strings.TrimSpace(filter.Region)
	filter.ClusterID = strings.TrimSpace(filter.ClusterID)
	filter.Namespace = strings.TrimSpace(filter.Namespace)
	filter.Environment = strings.ToLower(strings.TrimSpace(filter.Environment))
	filter.Runtime = strings.TrimSpace(filter.Runtime)
	filter.Tag = strings.TrimSpace(filter.Tag)
	filter.SortBy = strings.ToLower(strings.TrimSpace(filter.SortBy))
	filter.SortOrder = strings.ToLower(strings.TrimSpace(filter.SortOrder))
	filter.AfterID = strings.TrimSpace(filter.AfterID)
	if filter.Limit == 0 {
		filter.Limit = 100
	}

	if filter.Protocol != "" && !isValidMicroserviceProtocol(filter.Protocol) {
		return nil, ErrInvalidMicroservice
	}
	if filter.Status != "" && !isValidMicroserviceStatus(filter.Status) {
		return nil, ErrInvalidMicroservice
	}
	if filter.Environment != "" && !isValidMicroserviceEnvironment(filter.Environment) {
		return nil, ErrInvalidMicroservice
	}
	if filter.MinReplicas < 0 {
		return nil, ErrInvalidMicroservice
	}
	if filter.Offset < 0 || filter.Limit < 1 || filter.Limit > 1000 {
		return nil, ErrInvalidMicroservice
	}
	if filter.AfterID != "" && filter.Offset > 0 {
		return nil, ErrInvalidMicroservice
	}
	if filter.AfterID != "" && filter.SortBy != "" && filter.SortBy != "id" {
		return nil, ErrInvalidMicroservice
	}

	return s.microserviceRepository.FindByFilter(ctx, filter)
}

func (s *MicroserviceService) GetMicroserviceByID(ctx context.Context, id string) (model.Microservice, error) {
	if strings.TrimSpace(id) == "" {
		return model.Microservice{}, ErrInvalidMicroservice
	}

	return s.microserviceRepository.FindByID(ctx, id)
}

func (s *MicroserviceService) CreateMicroservice(ctx context.Context, request model.CreateMicroserviceRequest) (model.Microservice, error) {
	now := time.Now().UTC()
	microservice, err := s.buildMicroservice(ctx, model.Microservice{
		ID:        fmt.Sprintf("svc-%d", now.UnixNano()),
		CreatedAt: now,
	}, microserviceBuildInput{
		TenantID:             request.TenantID,
		ApplicationID:        request.ApplicationID,
		Name:                 request.Name,
		OwnerTeam:            request.OwnerTeam,
		Protocol:             request.Protocol,
		Endpoint:             request.Endpoint,
		Status:               request.Status,
		CloudProvider:        request.CloudProvider,
		Region:               request.Region,
		ClusterID:            request.ClusterID,
		Namespace:            request.Namespace,
		Environment:          request.Environment,
		Runtime:              request.Runtime,
		Image:                request.Image,
		Version:              request.Version,
		Replicas:             request.Replicas,
		CPURequest:           request.CPURequest,
		MemoryRequest:        request.MemoryRequest,
		HealthPath:           request.HealthPath,
		SLOTarget:            request.SLOTarget,
		ErrorBudgetRemaining: request.ErrorBudgetRemaining,
		Dependencies:         request.Dependencies,
		Config:               request.Config,
		Tags:                 request.Tags,
	})
	if err != nil {
		return model.Microservice{}, err
	}
	microservice.UpdatedAt = now

	return s.microserviceRepository.Save(ctx, microservice)
}

func (s *MicroserviceService) UpdateMicroservice(ctx context.Context, id string, request model.UpdateMicroserviceRequest) (model.Microservice, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return model.Microservice{}, ErrInvalidMicroservice
	}

	microservice, err := s.microserviceRepository.FindByID(ctx, id)
	if err != nil {
		return model.Microservice{}, err
	}

	microservice, err = s.buildMicroservice(ctx, microservice, microserviceBuildInput{
		TenantID:             request.TenantID,
		ApplicationID:        request.ApplicationID,
		Name:                 request.Name,
		OwnerTeam:            request.OwnerTeam,
		Protocol:             request.Protocol,
		Endpoint:             request.Endpoint,
		Status:               request.Status,
		CloudProvider:        request.CloudProvider,
		Region:               request.Region,
		ClusterID:            request.ClusterID,
		Namespace:            request.Namespace,
		Environment:          request.Environment,
		Runtime:              request.Runtime,
		Image:                request.Image,
		Version:              request.Version,
		Replicas:             request.Replicas,
		CPURequest:           request.CPURequest,
		MemoryRequest:        request.MemoryRequest,
		HealthPath:           request.HealthPath,
		SLOTarget:            request.SLOTarget,
		ErrorBudgetRemaining: request.ErrorBudgetRemaining,
		Dependencies:         request.Dependencies,
		Config:               request.Config,
		Tags:                 request.Tags,
	})
	if err != nil {
		return model.Microservice{}, err
	}
	microservice.UpdatedAt = time.Now().UTC()

	return s.microserviceRepository.Save(ctx, microservice)
}

func (s *MicroserviceService) UpdateMicroserviceStatus(ctx context.Context, id string, request model.UpdateMicroserviceStatusRequest) (model.Microservice, error) {
	id = strings.TrimSpace(id)
	status := strings.ToLower(strings.TrimSpace(request.Status))
	if id == "" || !isValidMicroserviceStatus(status) {
		return model.Microservice{}, ErrInvalidMicroservice
	}

	microservice, err := s.microserviceRepository.FindByID(ctx, id)
	if err != nil {
		return model.Microservice{}, err
	}

	microservice.Status = status
	microservice.UpdatedAt = time.Now().UTC()

	return s.microserviceRepository.Save(ctx, microservice)
}

func (s *MicroserviceService) DeleteMicroservice(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrInvalidMicroservice
	}

	return s.microserviceRepository.DeleteByID(ctx, id)
}

type microserviceBuildInput struct {
	TenantID             string
	ApplicationID        string
	Name                 string
	OwnerTeam            string
	Protocol             string
	Endpoint             string
	Status               string
	CloudProvider        string
	Region               string
	ClusterID            string
	Namespace            string
	Environment          string
	Runtime              string
	Image                string
	Version              string
	Replicas             int
	CPURequest           string
	MemoryRequest        string
	HealthPath           string
	SLOTarget            float64
	ErrorBudgetRemaining float64
	Dependencies         []string
	Config               map[string]string
	Tags                 []string
}

func (s *MicroserviceService) buildMicroservice(ctx context.Context, microservice model.Microservice, input microserviceBuildInput) (model.Microservice, error) {
	input.TenantID = strings.TrimSpace(input.TenantID)
	input.ApplicationID = strings.TrimSpace(input.ApplicationID)
	input.Name = strings.TrimSpace(input.Name)
	input.OwnerTeam = strings.TrimSpace(input.OwnerTeam)
	input.Protocol = strings.ToLower(strings.TrimSpace(input.Protocol))
	input.Endpoint = strings.TrimSpace(input.Endpoint)
	input.Status = strings.ToLower(strings.TrimSpace(input.Status))
	input.CloudProvider = strings.ToLower(strings.TrimSpace(input.CloudProvider))
	input.Region = strings.TrimSpace(input.Region)
	input.ClusterID = strings.TrimSpace(input.ClusterID)
	input.Namespace = strings.TrimSpace(input.Namespace)
	input.Environment = strings.ToLower(strings.TrimSpace(input.Environment))
	input.Runtime = strings.TrimSpace(input.Runtime)
	input.Image = strings.TrimSpace(input.Image)
	input.Version = strings.TrimSpace(input.Version)
	input.CPURequest = strings.TrimSpace(input.CPURequest)
	input.MemoryRequest = strings.TrimSpace(input.MemoryRequest)
	input.HealthPath = strings.TrimSpace(input.HealthPath)

	if input.ApplicationID == "" || input.Name == "" || input.OwnerTeam == "" || input.Protocol == "" || input.Endpoint == "" {
		return model.Microservice{}, ErrInvalidMicroservice
	}
	if microservice.TenantID != "" && input.TenantID != "" && input.TenantID != microservice.TenantID {
		return model.Microservice{}, ErrInvalidMicroservice
	}
	if input.TenantID == "" {
		input.TenantID = microservice.TenantID
	}
	if input.Status == "" {
		input.Status = "active"
	}
	if input.TenantID == "" {
		input.TenantID = "default"
	}
	if input.Namespace == "" {
		input.Namespace = "default"
	}
	if input.HealthPath == "" {
		input.HealthPath = "/healthz"
	}
	if input.Replicas == 0 {
		input.Replicas = 1
	}
	if input.SLOTarget == 0 {
		input.SLOTarget = 99.9
	}
	if !isValidMicroserviceProtocol(input.Protocol) || !isValidMicroserviceStatus(input.Status) {
		return model.Microservice{}, ErrInvalidMicroservice
	}
	if input.Environment != "" && !isValidMicroserviceEnvironment(input.Environment) {
		return model.Microservice{}, ErrInvalidMicroservice
	}
	if input.Replicas < 1 || input.SLOTarget < 0 || input.SLOTarget > 100 || input.ErrorBudgetRemaining < 0 || input.ErrorBudgetRemaining > 100 {
		return model.Microservice{}, ErrInvalidMicroservice
	}
	if _, err := s.applicationRepository.FindByID(ctx, input.ApplicationID); err != nil {
		return model.Microservice{}, err
	}

	microservice.ApplicationID = input.ApplicationID
	microservice.TenantID = input.TenantID
	microservice.Name = input.Name
	microservice.OwnerTeam = input.OwnerTeam
	microservice.Protocol = input.Protocol
	microservice.Endpoint = input.Endpoint
	microservice.Status = input.Status
	microservice.CloudProvider = input.CloudProvider
	microservice.Region = input.Region
	microservice.ClusterID = input.ClusterID
	microservice.Namespace = input.Namespace
	microservice.Environment = input.Environment
	microservice.Runtime = input.Runtime
	microservice.Image = input.Image
	microservice.Version = input.Version
	microservice.Replicas = input.Replicas
	microservice.CPURequest = input.CPURequest
	microservice.MemoryRequest = input.MemoryRequest
	microservice.HealthPath = input.HealthPath
	microservice.SLOTarget = input.SLOTarget
	microservice.ErrorBudgetRemaining = input.ErrorBudgetRemaining
	microservice.Dependencies = normalizeStringList(input.Dependencies)
	microservice.Config = input.Config
	microservice.Tags = normalizeTags(input.Tags)

	return microservice, nil
}

func isValidMicroserviceProtocol(protocol string) bool {
	return protocol == "http" || protocol == "grpc" || protocol == "event" || protocol == "worker"
}

func isValidMicroserviceStatus(status string) bool {
	return status == "active" || status == "degraded" || status == "offline" || status == "deprecated"
}

func isValidMicroserviceEnvironment(environment string) bool {
	return environment == "development" || environment == "staging" || environment == "production"
}

func normalizeStringList(values []string) []string {
	normalized := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		normalized = append(normalized, value)
	}
	return normalized
}

func microserviceMatchesQuery(microservice model.Microservice, query string) bool {
	query = strings.ToLower(query)
	if strings.Contains(strings.ToLower(microservice.ID), query) ||
		strings.Contains(strings.ToLower(microservice.ApplicationID), query) ||
		strings.Contains(strings.ToLower(microservice.Name), query) ||
		strings.Contains(strings.ToLower(microservice.OwnerTeam), query) ||
		strings.Contains(strings.ToLower(microservice.Protocol), query) ||
		strings.Contains(strings.ToLower(microservice.Endpoint), query) ||
		strings.Contains(strings.ToLower(microservice.Status), query) ||
		strings.Contains(strings.ToLower(microservice.CloudProvider), query) ||
		strings.Contains(strings.ToLower(microservice.Region), query) ||
		strings.Contains(strings.ToLower(microservice.ClusterID), query) ||
		strings.Contains(strings.ToLower(microservice.Namespace), query) ||
		strings.Contains(strings.ToLower(microservice.Environment), query) ||
		strings.Contains(strings.ToLower(microservice.Runtime), query) ||
		strings.Contains(strings.ToLower(microservice.Image), query) ||
		strings.Contains(strings.ToLower(microservice.Version), query) ||
		strings.Contains(strings.ToLower(microservice.CPURequest), query) ||
		strings.Contains(strings.ToLower(microservice.MemoryRequest), query) ||
		strings.Contains(strings.ToLower(microservice.HealthPath), query) {
		return true
	}
	for _, dependency := range microservice.Dependencies {
		if strings.Contains(strings.ToLower(dependency), query) {
			return true
		}
	}
	for _, tag := range microservice.Tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}
	for key, value := range microservice.Config {
		if strings.Contains(strings.ToLower(key), query) || strings.Contains(strings.ToLower(value), query) {
			return true
		}
	}

	return false
}
