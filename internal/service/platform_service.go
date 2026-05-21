package service

import (
	"context"

	"hr-cloud-service/internal/model"
	"hr-cloud-service/internal/repository"
)

type PlatformService struct {
	applicationRepository repository.ApplicationRepository
	clusterRepository     repository.ClusterRepository
	deploymentRepository  repository.DeploymentRepository
	pipelineRepository    repository.PipelineRepository
}

func NewPlatformService(
	applicationRepository repository.ApplicationRepository,
	clusterRepository repository.ClusterRepository,
	deploymentRepository repository.DeploymentRepository,
	pipelineRepository repository.PipelineRepository,
) *PlatformService {
	return &PlatformService{
		applicationRepository: applicationRepository,
		clusterRepository:     clusterRepository,
		deploymentRepository:  deploymentRepository,
		pipelineRepository:    pipelineRepository,
	}
}

func (s *PlatformService) GetSummary(ctx context.Context) (model.PlatformSummary, error) {
	applications, err := s.applicationRepository.FindAll(ctx)
	if err != nil {
		return model.PlatformSummary{}, err
	}

	clusters, err := s.clusterRepository.FindAll(ctx)
	if err != nil {
		return model.PlatformSummary{}, err
	}

	deployments, err := s.deploymentRepository.FindAll(ctx)
	if err != nil {
		return model.PlatformSummary{}, err
	}

	pipelineRuns, err := s.pipelineRepository.FindAll(ctx)
	if err != nil {
		return model.PlatformSummary{}, err
	}

	byStatus := map[string]int{}
	for _, deployment := range deployments {
		byStatus[deployment.Status]++
	}

	return model.PlatformSummary{
		Applications: len(applications),
		Clusters:     len(clusters),
		Deployments:  len(deployments),
		PipelineRuns: len(pipelineRuns),
		ByStatus:     byStatus,
	}, nil
}
