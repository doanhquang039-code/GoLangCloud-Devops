package repository

import (
	"context"
	"errors"
	"sync"

	"hr-cloud-service/internal/model"
)

var ErrIncidentNotFound = errors.New("incident not found")

type IncidentRepository interface {
	FindAll(ctx context.Context) ([]model.Incident, error)
	FindByID(ctx context.Context, id string) (model.Incident, error)
	Save(ctx context.Context, incident model.Incident) (model.Incident, error)
}

type InMemoryIncidentRepository struct {
	mu        sync.RWMutex
	incidents map[string]model.Incident
}

func NewInMemoryIncidentRepository() *InMemoryIncidentRepository {
	return &InMemoryIncidentRepository{
		incidents: make(map[string]model.Incident),
	}
}

func (r *InMemoryIncidentRepository) FindAll(ctx context.Context) ([]model.Incident, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	incidents := make([]model.Incident, 0, len(r.incidents))
	for _, incident := range r.incidents {
		incidents = append(incidents, incident)
	}

	return incidents, nil
}

func (r *InMemoryIncidentRepository) FindByID(ctx context.Context, id string) (model.Incident, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	incident, ok := r.incidents[id]
	if !ok {
		return model.Incident{}, ErrIncidentNotFound
	}

	return incident, nil
}

func (r *InMemoryIncidentRepository) Save(ctx context.Context, incident model.Incident) (model.Incident, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.incidents[incident.ID] = incident
	return incident, nil
}
