package repository

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"

	"hr-cloud-service/internal/model"
)

var ErrMicroserviceNotFound = errors.New("microservice not found")

type MicroserviceRepository interface {
	FindAll(ctx context.Context) ([]model.Microservice, error)
	FindByFilter(ctx context.Context, filter model.MicroserviceFilter) ([]model.Microservice, error)
	FindByID(ctx context.Context, id string) (model.Microservice, error)
	Save(ctx context.Context, microservice model.Microservice) (model.Microservice, error)
	DeleteByID(ctx context.Context, id string) error
}

type InMemoryMicroserviceRepository struct {
	mu            sync.RWMutex
	microservices map[string]model.Microservice
}

func NewInMemoryMicroserviceRepository() *InMemoryMicroserviceRepository {
	return &InMemoryMicroserviceRepository{
		microservices: make(map[string]model.Microservice),
	}
}

func (r *InMemoryMicroserviceRepository) FindAll(ctx context.Context) ([]model.Microservice, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	microservices := make([]model.Microservice, 0, len(r.microservices))
	for _, microservice := range r.microservices {
		microservices = append(microservices, microservice)
	}

	return microservices, nil
}

func (r *InMemoryMicroserviceRepository) FindByFilter(ctx context.Context, filter model.MicroserviceFilter) ([]model.Microservice, error) {
	microservices, err := r.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	filtered := make([]model.Microservice, 0, len(microservices))
	for _, microservice := range microservices {
		if !MicroserviceMatchesFilter(microservice, filter) {
			continue
		}
		filtered = append(filtered, microservice)
	}

	SortMicroservices(filtered, filter.SortBy, filter.SortOrder)
	filtered = CursorMicroservices(filtered, filter.AfterID, filter.SortOrder)
	return PaginateMicroservices(filtered, filter.Offset, filter.Limit), nil
}

func MicroserviceMatchesFilter(microservice model.Microservice, filter model.MicroserviceFilter) bool {
	if filter.Query != "" && !microserviceMatchesRepositoryQuery(microservice, filter.Query) {
		return false
	}
	if filter.TenantID != "" && microservice.TenantID != filter.TenantID {
		return false
	}
	if filter.ApplicationID != "" && microservice.ApplicationID != filter.ApplicationID {
		return false
	}
	if filter.OwnerTeam != "" && !strings.EqualFold(microservice.OwnerTeam, filter.OwnerTeam) {
		return false
	}
	if filter.Protocol != "" && !strings.EqualFold(microservice.Protocol, filter.Protocol) {
		return false
	}
	if filter.Status != "" && !strings.EqualFold(microservice.Status, filter.Status) {
		return false
	}
	if filter.CloudProvider != "" && !strings.EqualFold(microservice.CloudProvider, filter.CloudProvider) {
		return false
	}
	if filter.Region != "" && !strings.EqualFold(microservice.Region, filter.Region) {
		return false
	}
	if filter.ClusterID != "" && microservice.ClusterID != filter.ClusterID {
		return false
	}
	if filter.Namespace != "" && !strings.EqualFold(microservice.Namespace, filter.Namespace) {
		return false
	}
	if filter.Environment != "" && !strings.EqualFold(microservice.Environment, filter.Environment) {
		return false
	}
	if filter.Runtime != "" && !strings.EqualFold(microservice.Runtime, filter.Runtime) {
		return false
	}
	if filter.Tag != "" && !stringSliceContainsFold(microservice.Tags, filter.Tag) {
		return false
	}
	return filter.MinReplicas == 0 || microservice.Replicas >= filter.MinReplicas
}

func SortMicroservices(microservices []model.Microservice, sortBy string, sortOrder string) {
	desc := strings.EqualFold(sortOrder, "desc")
	sort.Slice(microservices, func(i, j int) bool {
		less, greater := false, false
		switch sortBy {
		case "id":
			less = microservices[i].ID < microservices[j].ID
			greater = microservices[i].ID > microservices[j].ID
		case "name":
			less = microservices[i].Name < microservices[j].Name
			greater = microservices[i].Name > microservices[j].Name
		case "updated_at":
			less = microservices[i].UpdatedAt.Before(microservices[j].UpdatedAt)
			greater = microservices[i].UpdatedAt.After(microservices[j].UpdatedAt)
		case "replicas":
			less = microservices[i].Replicas < microservices[j].Replicas
			greater = microservices[i].Replicas > microservices[j].Replicas
		default:
			less = microservices[i].ID < microservices[j].ID
			greater = microservices[i].ID > microservices[j].ID
		}
		if desc {
			return greater
		}
		return less
	})
}

func CursorMicroservices(microservices []model.Microservice, afterID string, sortOrder string) []model.Microservice {
	if afterID == "" {
		return microservices
	}
	cursor := 0
	desc := strings.EqualFold(sortOrder, "desc")
	for cursor < len(microservices) {
		if desc {
			if microservices[cursor].ID < afterID {
				break
			}
		} else if microservices[cursor].ID > afterID {
			break
		}
		cursor++
	}
	return microservices[cursor:]
}

func PaginateMicroservices(microservices []model.Microservice, offset int, limit int) []model.Microservice {
	if offset >= len(microservices) {
		return []model.Microservice{}
	}
	end := offset + limit
	if end > len(microservices) {
		end = len(microservices)
	}
	return microservices[offset:end]
}

func stringSliceContainsFold(values []string, expected string) bool {
	for _, value := range values {
		if strings.EqualFold(value, expected) {
			return true
		}
	}
	return false
}

func microserviceMatchesRepositoryQuery(microservice model.Microservice, query string) bool {
	query = strings.ToLower(query)
	if strings.Contains(strings.ToLower(microservice.ID), query) ||
		strings.Contains(strings.ToLower(microservice.TenantID), query) ||
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

func (r *InMemoryMicroserviceRepository) FindByID(ctx context.Context, id string) (model.Microservice, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	microservice, ok := r.microservices[id]
	if !ok {
		return model.Microservice{}, ErrMicroserviceNotFound
	}

	return microservice, nil
}

func (r *InMemoryMicroserviceRepository) Save(ctx context.Context, microservice model.Microservice) (model.Microservice, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.microservices[microservice.ID] = microservice
	return microservice, nil
}

func (r *InMemoryMicroserviceRepository) DeleteByID(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.microservices[id]; !ok {
		return ErrMicroserviceNotFound
	}

	delete(r.microservices, id)
	return nil
}
