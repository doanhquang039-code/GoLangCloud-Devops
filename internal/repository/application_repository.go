package repository

import (
	"context"
	"errors"
	"sync"

	"hr-cloud-service/internal/model"
)

var ErrApplicationNotFound = errors.New("application not found")

type ApplicationRepository interface {
	FindAll(ctx context.Context) ([]model.Application, error)
	FindByID(ctx context.Context, id string) (model.Application, error)
	Save(ctx context.Context, application model.Application) (model.Application, error)
}

type InMemoryApplicationRepository struct {
	mu           sync.RWMutex
	applications map[string]model.Application
}

func NewInMemoryApplicationRepository() *InMemoryApplicationRepository {
	return &InMemoryApplicationRepository{
		applications: make(map[string]model.Application),
	}
}

func (r *InMemoryApplicationRepository) FindAll(ctx context.Context) ([]model.Application, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	applications := make([]model.Application, 0, len(r.applications))
	for _, application := range r.applications {
		applications = append(applications, application)
	}

	return applications, nil
}

func (r *InMemoryApplicationRepository) FindByID(ctx context.Context, id string) (model.Application, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	application, ok := r.applications[id]
	if !ok {
		return model.Application{}, ErrApplicationNotFound
	}

	return application, nil
}

func (r *InMemoryApplicationRepository) Save(ctx context.Context, application model.Application) (model.Application, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.applications[application.ID] = application
	return application, nil
}
