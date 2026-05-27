package repository

import (
	"context"
	"errors"
	"sync"

	"hr-cloud-service/internal/model"
)

var ErrPipelineRunNotFound = errors.New("pipeline run not found")

type PipelineRepository interface {
	FindAll(ctx context.Context) ([]model.PipelineRun, error)
	FindByID(ctx context.Context, id string) (model.PipelineRun, error)
	Save(ctx context.Context, pipelineRun model.PipelineRun) (model.PipelineRun, error)
	DeleteByID(ctx context.Context, id string) error
}

type InMemoryPipelineRepository struct {
	mu           sync.RWMutex
	pipelineRuns map[string]model.PipelineRun
}

func NewInMemoryPipelineRepository() *InMemoryPipelineRepository {
	return &InMemoryPipelineRepository{
		pipelineRuns: make(map[string]model.PipelineRun),
	}
}

func (r *InMemoryPipelineRepository) FindAll(ctx context.Context) ([]model.PipelineRun, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	pipelineRuns := make([]model.PipelineRun, 0, len(r.pipelineRuns))
	for _, pipelineRun := range r.pipelineRuns {
		pipelineRuns = append(pipelineRuns, pipelineRun)
	}

	return pipelineRuns, nil
}

func (r *InMemoryPipelineRepository) FindByID(ctx context.Context, id string) (model.PipelineRun, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	pipelineRun, ok := r.pipelineRuns[id]
	if !ok {
		return model.PipelineRun{}, ErrPipelineRunNotFound
	}

	return pipelineRun, nil
}

func (r *InMemoryPipelineRepository) Save(ctx context.Context, pipelineRun model.PipelineRun) (model.PipelineRun, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.pipelineRuns[pipelineRun.ID] = pipelineRun
	return pipelineRun, nil
}

func (r *InMemoryPipelineRepository) DeleteByID(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.pipelineRuns[id]; !ok {
		return ErrPipelineRunNotFound
	}

	delete(r.pipelineRuns, id)
	return nil
}
