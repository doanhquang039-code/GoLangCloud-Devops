package repository

import (
	"context"
	"errors"
	"sync"

	"hr-cloud-service/internal/model"
)

var ErrClusterNotFound = errors.New("cluster not found")

type ClusterRepository interface {
	FindAll(ctx context.Context) ([]model.Cluster, error)
	FindByID(ctx context.Context, id string) (model.Cluster, error)
	Save(ctx context.Context, cluster model.Cluster) (model.Cluster, error)
}

type InMemoryClusterRepository struct {
	mu       sync.RWMutex
	clusters map[string]model.Cluster
}

func NewInMemoryClusterRepository() *InMemoryClusterRepository {
	return &InMemoryClusterRepository{
		clusters: make(map[string]model.Cluster),
	}
}

func (r *InMemoryClusterRepository) FindAll(ctx context.Context) ([]model.Cluster, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	clusters := make([]model.Cluster, 0, len(r.clusters))
	for _, cluster := range r.clusters {
		clusters = append(clusters, cluster)
	}

	return clusters, nil
}

func (r *InMemoryClusterRepository) FindByID(ctx context.Context, id string) (model.Cluster, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cluster, ok := r.clusters[id]
	if !ok {
		return model.Cluster{}, ErrClusterNotFound
	}

	return cluster, nil
}

func (r *InMemoryClusterRepository) Save(ctx context.Context, cluster model.Cluster) (model.Cluster, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.clusters[cluster.ID] = cluster
	return cluster, nil
}
