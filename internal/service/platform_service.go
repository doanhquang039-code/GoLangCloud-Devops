package service

import (
	"context"

	"hr-cloud-service/internal/model"
	"hr-cloud-service/internal/repository"
)

type PlatformService struct {
	applicationRepository repository.ApplicationRepository
	clusterRepository     repository.ClusterRepository
	environmentRepository repository.EnvironmentRepository
	deploymentRepository  repository.DeploymentRepository
	pipelineRepository    repository.PipelineRepository
	incidentRepository    repository.IncidentRepository
}

func NewPlatformService(
	applicationRepository repository.ApplicationRepository,
	clusterRepository repository.ClusterRepository,
	environmentRepository repository.EnvironmentRepository,
	deploymentRepository repository.DeploymentRepository,
	pipelineRepository repository.PipelineRepository,
	incidentRepository repository.IncidentRepository,
) *PlatformService {
	return &PlatformService{
		applicationRepository: applicationRepository,
		clusterRepository:     clusterRepository,
		environmentRepository: environmentRepository,
		deploymentRepository:  deploymentRepository,
		pipelineRepository:    pipelineRepository,
		incidentRepository:    incidentRepository,
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

	environments, err := s.environmentRepository.FindAll(ctx)
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

	incidents, err := s.incidentRepository.FindAll(ctx)
	if err != nil {
		return model.PlatformSummary{}, err
	}

	byStatus := map[string]int{}
	for _, deployment := range deployments {
		byStatus[deployment.Status]++
	}

	openIncidents := 0
	for _, incident := range incidents {
		if incident.Status != "resolved" {
			openIncidents++
		}
	}

	return model.PlatformSummary{
		Applications: len(applications),
		Clusters:     len(clusters),
		Environments: len(environments),
		Deployments:  len(deployments),
		PipelineRuns: len(pipelineRuns),
		Incidents:    len(incidents),
		OpenIncidents: openIncidents,
		ByStatus:     byStatus,
	}, nil
}
