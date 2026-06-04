package repository

import (
	"context"
	"errors"
	"sync"

	"hr-cloud-service/internal/model"
)

var ErrMicroserviceNotFound = errors.New("microservice not found")

type MicroserviceRepository interface {
	FindAll(ctx context.Context) ([]model.Microservice, error)
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
