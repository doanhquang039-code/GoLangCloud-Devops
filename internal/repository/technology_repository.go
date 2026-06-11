package repository

import (
	"context"
	"errors"
	"sync"

	"hr-cloud-service/internal/model"
)

var ErrTechnologyNotFound = errors.New("technology not found")

type TechnologyRepository interface {
	FindAll(ctx context.Context) ([]model.Technology, error)
	FindByID(ctx context.Context, id string) (model.Technology, error)
	Save(ctx context.Context, technology model.Technology) (model.Technology, error)
	DeleteByID(ctx context.Context, id string) error
}

type InMemoryTechnologyRepository struct {
	mu           sync.RWMutex
	technologies map[string]model.Technology
}

func NewInMemoryTechnologyRepository() *InMemoryTechnologyRepository {
	return &InMemoryTechnologyRepository{
		technologies: make(map[string]model.Technology),
	}
}

func (r *InMemoryTechnologyRepository) FindAll(ctx context.Context) ([]model.Technology, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	technologies := make([]model.Technology, 0, len(r.technologies))
	for _, technology := range r.technologies {
		technologies = append(technologies, technology)
	}

	return technologies, nil
}

func (r *InMemoryTechnologyRepository) FindByID(ctx context.Context, id string) (model.Technology, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	technology, ok := r.technologies[id]
	if !ok {
		return model.Technology{}, ErrTechnologyNotFound
	}

	return technology, nil
}

func (r *InMemoryTechnologyRepository) Save(ctx context.Context, technology model.Technology) (model.Technology, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.technologies[technology.ID] = technology
	return technology, nil
}

func (r *InMemoryTechnologyRepository) DeleteByID(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.technologies[id]; !ok {
		return ErrTechnologyNotFound
	}

	delete(r.technologies, id)
	return nil
}
