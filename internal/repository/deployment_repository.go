package repository

import (
	"context"
	"errors"
	"sync"

	"hr-cloud-service/internal/model"
)

var ErrDeploymentNotFound = errors.New("deployment not found")

type DeploymentRepository interface {
	FindAll(ctx context.Context) ([]model.Deployment, error)
	FindByID(ctx context.Context, id string) (model.Deployment, error)
	Save(ctx context.Context, deployment model.Deployment) (model.Deployment, error)
}

type InMemoryDeploymentRepository struct {
	mu          sync.RWMutex
	deployments map[string]model.Deployment
}

func NewInMemoryDeploymentRepository() *InMemoryDeploymentRepository {
	return &InMemoryDeploymentRepository{
		deployments: make(map[string]model.Deployment),
	}
}

func (r *InMemoryDeploymentRepository) FindAll(ctx context.Context) ([]model.Deployment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	deployments := make([]model.Deployment, 0, len(r.deployments))
	for _, deployment := range r.deployments {
		deployments = append(deployments, deployment)
	}

	return deployments, nil
}

func (r *InMemoryDeploymentRepository) FindByID(ctx context.Context, id string) (model.Deployment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	deployment, ok := r.deployments[id]
	if !ok {
		return model.Deployment{}, ErrDeploymentNotFound
	}

	return deployment, nil
}

func (r *InMemoryDeploymentRepository) Save(ctx context.Context, deployment model.Deployment) (model.Deployment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.deployments[deployment.ID] = deployment
	return deployment, nil
}
