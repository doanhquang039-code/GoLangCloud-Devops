package repository

import (
	"context"
	"errors"
	"sync"

	"hr-cloud-service/internal/model"
)

var ErrEnvironmentNotFound = errors.New("environment not found")

type EnvironmentRepository interface {
	FindAll(ctx context.Context) ([]model.Environment, error)
	FindByID(ctx context.Context, id string) (model.Environment, error)
	Save(ctx context.Context, environment model.Environment) (model.Environment, error)
}

type InMemoryEnvironmentRepository struct {
	mu           sync.RWMutex
	environments map[string]model.Environment
}

func NewInMemoryEnvironmentRepository() *InMemoryEnvironmentRepository {
	return &InMemoryEnvironmentRepository{
		environments: make(map[string]model.Environment),
	}
}

func (r *InMemoryEnvironmentRepository) FindAll(ctx context.Context) ([]model.Environment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	environments := make([]model.Environment, 0, len(r.environments))
	for _, environment := range r.environments {
		environments = append(environments, environment)
	}

	return environments, nil
}

func (r *InMemoryEnvironmentRepository) FindByID(ctx context.Context, id string) (model.Environment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	environment, ok := r.environments[id]
	if !ok {
		return model.Environment{}, ErrEnvironmentNotFound
	}

	return environment, nil
}

func (r *InMemoryEnvironmentRepository) Save(ctx context.Context, environment model.Environment) (model.Environment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.environments[environment.ID] = environment
	return environment, nil
}
