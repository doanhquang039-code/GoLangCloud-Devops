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
	filter.ApplicationID = strings.TrimSpace(filter.ApplicationID)
	filter.OwnerTeam = strings.TrimSpace(filter.OwnerTeam)
	filter.Protocol = strings.ToLower(strings.TrimSpace(filter.Protocol))
	filter.Status = strings.ToLower(strings.TrimSpace(filter.Status))
	filter.Tag = strings.TrimSpace(filter.Tag)

	if filter.Protocol != "" && !isValidMicroserviceProtocol(filter.Protocol) {
		return nil, ErrInvalidMicroservice
	}
	if filter.Status != "" && !isValidMicroserviceStatus(filter.Status) {
		return nil, ErrInvalidMicroservice
	}

	microservices, err := s.microserviceRepository.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	if filter.Query == "" && filter.ApplicationID == "" && filter.OwnerTeam == "" && filter.Protocol == "" && filter.Status == "" && filter.Tag == "" {
		return microservices, nil
	}

	filtered := make([]model.Microservice, 0, len(microservices))
	for _, microservice := range microservices {
		if filter.Query != "" && !microserviceMatchesQuery(microservice, filter.Query) {
			continue
		}
		if filter.ApplicationID != "" && microservice.ApplicationID != filter.ApplicationID {
			continue
		}
		if filter.OwnerTeam != "" && !strings.EqualFold(microservice.OwnerTeam, filter.OwnerTeam) {
			continue
		}
		if filter.Protocol != "" && !strings.EqualFold(microservice.Protocol, filter.Protocol) {
			continue
		}
		if filter.Status != "" && !strings.EqualFold(microservice.Status, filter.Status) {
			continue
		}
		if filter.Tag != "" && !hasApplicationTag(microservice.Tags, filter.Tag) {
			continue
		}
		filtered = append(filtered, microservice)
	}

	return filtered, nil
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
	}, request.ApplicationID, request.Name, request.OwnerTeam, request.Protocol, request.Endpoint, request.Status, request.Dependencies, request.Config, request.Tags)
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

	microservice, err = s.buildMicroservice(ctx, microservice, request.ApplicationID, request.Name, request.OwnerTeam, request.Protocol, request.Endpoint, request.Status, request.Dependencies, request.Config, request.Tags)
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

func (s *MicroserviceService) buildMicroservice(ctx context.Context, microservice model.Microservice, applicationID, name, ownerTeam, protocol, endpoint, status string, dependencies []string, config map[string]string, tags []string) (model.Microservice, error) {
	applicationID = strings.TrimSpace(applicationID)
	name = strings.TrimSpace(name)
	ownerTeam = strings.TrimSpace(ownerTeam)
	protocol = strings.ToLower(strings.TrimSpace(protocol))
	endpoint = strings.TrimSpace(endpoint)
	status = strings.ToLower(strings.TrimSpace(status))

	if applicationID == "" || name == "" || ownerTeam == "" || protocol == "" || endpoint == "" {
		return model.Microservice{}, ErrInvalidMicroservice
	}
	if status == "" {
		status = "active"
	}
	if !isValidMicroserviceProtocol(protocol) || !isValidMicroserviceStatus(status) {
		return model.Microservice{}, ErrInvalidMicroservice
	}
	if _, err := s.applicationRepository.FindByID(ctx, applicationID); err != nil {
		return model.Microservice{}, err
	}

	microservice.ApplicationID = applicationID
	microservice.Name = name
	microservice.OwnerTeam = ownerTeam
	microservice.Protocol = protocol
	microservice.Endpoint = endpoint
	microservice.Status = status
	microservice.Dependencies = normalizeStringList(dependencies)
	microservice.Config = config
	microservice.Tags = normalizeTags(tags)

	return microservice, nil
}

func isValidMicroserviceProtocol(protocol string) bool {
	return protocol == "http" || protocol == "grpc" || protocol == "event" || protocol == "worker"
}

func isValidMicroserviceStatus(status string) bool {
	return status == "active" || status == "degraded" || status == "offline" || status == "deprecated"
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
		strings.Contains(strings.ToLower(microservice.Status), query) {
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
