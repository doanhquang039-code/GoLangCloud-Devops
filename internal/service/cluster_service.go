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

var ErrInvalidCluster = errors.New("invalid cluster input")

type ClusterService struct {
	clusterRepository repository.ClusterRepository
}

func NewClusterService(clusterRepository repository.ClusterRepository) *ClusterService {
	return &ClusterService{clusterRepository: clusterRepository}
}

func (s *ClusterService) GetClusters(ctx context.Context, filter model.ClusterFilter) ([]model.Cluster, error) {
	filter.Provider = strings.TrimSpace(filter.Provider)
	filter.Region = strings.TrimSpace(filter.Region)
	filter.Status = strings.TrimSpace(filter.Status)

	if filter.Status != "" && !isValidClusterStatus(filter.Status) {
		return nil, ErrInvalidCluster
	}

	clusters, err := s.clusterRepository.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	if filter.Provider == "" && filter.Region == "" && filter.Status == "" {
		return clusters, nil
	}

	filtered := make([]model.Cluster, 0, len(clusters))
	for _, cluster := range clusters {
		if filter.Provider != "" && !strings.EqualFold(cluster.Provider, filter.Provider) {
			continue
		}
		if filter.Region != "" && !strings.EqualFold(cluster.Region, filter.Region) {
			continue
		}
		if filter.Status != "" && cluster.Status != filter.Status {
			continue
		}
		filtered = append(filtered, cluster)
	}

	return filtered, nil
}

func (s *ClusterService) GetClusterByID(ctx context.Context, id string) (model.Cluster, error) {
	if strings.TrimSpace(id) == "" {
		return model.Cluster{}, ErrInvalidCluster
	}

	return s.clusterRepository.FindByID(ctx, id)
}

func (s *ClusterService) CreateCluster(ctx context.Context, request model.CreateClusterRequest) (model.Cluster, error) {
	request.Name = strings.TrimSpace(request.Name)
	request.Provider = strings.TrimSpace(request.Provider)
	request.Region = strings.TrimSpace(request.Region)
	request.Endpoint = strings.TrimSpace(request.Endpoint)
	request.Version = strings.TrimSpace(request.Version)
	request.Status = strings.TrimSpace(request.Status)

	if request.Name == "" || request.Provider == "" || request.Region == "" || request.Endpoint == "" || request.Version == "" {
		return model.Cluster{}, ErrInvalidCluster
	}
	if request.Status == "" {
		request.Status = "ready"
	}
	if !isValidClusterStatus(request.Status) {
		return model.Cluster{}, ErrInvalidCluster
	}

	now := time.Now().UTC()
	cluster := model.Cluster{
		ID:        fmt.Sprintf("cls-%d", now.UnixNano()),
		Name:      request.Name,
		Provider:  request.Provider,
		Region:    request.Region,
		Endpoint:  request.Endpoint,
		Version:   request.Version,
		Status:    request.Status,
		CreatedAt: now,
		UpdatedAt: now,
	}

	return s.clusterRepository.Save(ctx, cluster)
}

func (s *ClusterService) UpdateCluster(ctx context.Context, id string, request model.UpdateClusterRequest) (model.Cluster, error) {
	id = strings.TrimSpace(id)
	request.Name = strings.TrimSpace(request.Name)
	request.Provider = strings.TrimSpace(request.Provider)
	request.Region = strings.TrimSpace(request.Region)
	request.Endpoint = strings.TrimSpace(request.Endpoint)
	request.Version = strings.TrimSpace(request.Version)
	request.Status = strings.TrimSpace(request.Status)

	if id == "" || request.Name == "" || request.Provider == "" || request.Region == "" || request.Endpoint == "" || request.Version == "" {
		return model.Cluster{}, ErrInvalidCluster
	}
	if request.Status == "" {
		request.Status = "ready"
	}
	if !isValidClusterStatus(request.Status) {
		return model.Cluster{}, ErrInvalidCluster
	}

	cluster, err := s.clusterRepository.FindByID(ctx, id)
	if err != nil {
		return model.Cluster{}, err
	}

	cluster.Name = request.Name
	cluster.Provider = request.Provider
	cluster.Region = request.Region
	cluster.Endpoint = request.Endpoint
	cluster.Version = request.Version
	cluster.Status = request.Status
	cluster.UpdatedAt = time.Now().UTC()

	return s.clusterRepository.Save(ctx, cluster)
}

func (s *ClusterService) UpdateClusterStatus(ctx context.Context, id string, request model.UpdateClusterStatusRequest) (model.Cluster, error) {
	status := strings.TrimSpace(request.Status)
	if strings.TrimSpace(id) == "" || !isValidClusterStatus(status) {
		return model.Cluster{}, ErrInvalidCluster
	}

	cluster, err := s.clusterRepository.FindByID(ctx, id)
	if err != nil {
		return model.Cluster{}, err
	}

	cluster.Status = status
	cluster.UpdatedAt = time.Now().UTC()

	return s.clusterRepository.Save(ctx, cluster)
}

func isValidClusterStatus(status string) bool {
	return status == "ready" || status == "degraded" || status == "maintenance" || status == "offline"
}
