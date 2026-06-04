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

var ErrInvalidIncident = errors.New("invalid incident input")

type IncidentService struct {
	applicationRepository repository.ApplicationRepository
	clusterRepository     repository.ClusterRepository
	deploymentRepository  repository.DeploymentRepository
	incidentRepository    repository.IncidentRepository
}

func NewIncidentService(
	applicationRepository repository.ApplicationRepository,
	clusterRepository repository.ClusterRepository,
	deploymentRepository repository.DeploymentRepository,
	incidentRepository repository.IncidentRepository,
) *IncidentService {
	return &IncidentService{
		applicationRepository: applicationRepository,
		clusterRepository:     clusterRepository,
		deploymentRepository:  deploymentRepository,
		incidentRepository:    incidentRepository,
	}
}

func (s *IncidentService) GetIncidents(ctx context.Context, filter model.IncidentFilter) ([]model.Incident, error) {
	filter.Query = strings.TrimSpace(filter.Query)
	filter.ApplicationID = strings.TrimSpace(filter.ApplicationID)
	filter.ClusterID = strings.TrimSpace(filter.ClusterID)
	filter.DeploymentID = strings.TrimSpace(filter.DeploymentID)
	filter.Severity = strings.ToLower(strings.TrimSpace(filter.Severity))
	filter.Status = strings.ToLower(strings.TrimSpace(filter.Status))
	filter.OwnerTeam = strings.TrimSpace(filter.OwnerTeam)

	if filter.Severity != "" && !isValidIncidentSeverity(filter.Severity) {
		return nil, ErrInvalidIncident
	}
	if filter.Status != "" && !isValidIncidentStatus(filter.Status) {
		return nil, ErrInvalidIncident
	}

	incidents, err := s.incidentRepository.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	if filter.Query == "" && filter.ApplicationID == "" && filter.ClusterID == "" && filter.DeploymentID == "" && filter.Severity == "" && filter.Status == "" && filter.OwnerTeam == "" {
		return incidents, nil
	}

	filtered := make([]model.Incident, 0, len(incidents))
	for _, incident := range incidents {
		if filter.Query != "" && !incidentMatchesQuery(incident, filter.Query) {
			continue
		}
		if filter.ApplicationID != "" && incident.ApplicationID != filter.ApplicationID {
			continue
		}
		if filter.ClusterID != "" && incident.ClusterID != filter.ClusterID {
			continue
		}
		if filter.DeploymentID != "" && incident.DeploymentID != filter.DeploymentID {
			continue
		}
		if filter.Severity != "" && !strings.EqualFold(incident.Severity, filter.Severity) {
			continue
		}
		if filter.Status != "" && !strings.EqualFold(incident.Status, filter.Status) {
			continue
		}
		if filter.OwnerTeam != "" && !strings.EqualFold(incident.OwnerTeam, filter.OwnerTeam) {
			continue
		}
		filtered = append(filtered, incident)
	}

	return filtered, nil
}

func (s *IncidentService) GetIncidentByID(ctx context.Context, id string) (model.Incident, error) {
	if strings.TrimSpace(id) == "" {
		return model.Incident{}, ErrInvalidIncident
	}

	return s.incidentRepository.FindByID(ctx, id)
}

func (s *IncidentService) CreateIncident(ctx context.Context, request model.CreateIncidentRequest) (model.Incident, error) {
	incident, err := s.buildIncident(ctx, model.Incident{}, request.Title, request.Summary, request.Severity, request.Status, request.ApplicationID, request.ClusterID, request.DeploymentID, request.OwnerTeam)
	if err != nil {
		return model.Incident{}, err
	}

	now := time.Now().UTC()
	incident.ID = fmt.Sprintf("inc-%d", now.UnixNano())
	incident.CreatedAt = now
	incident.UpdatedAt = now
	if incident.Status == "resolved" {
		incident.ResolvedAt = &now
	}

	return s.incidentRepository.Save(ctx, incident)
}

func (s *IncidentService) UpdateIncident(ctx context.Context, id string, request model.UpdateIncidentRequest) (model.Incident, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return model.Incident{}, ErrInvalidIncident
	}

	existing, err := s.incidentRepository.FindByID(ctx, id)
	if err != nil {
		return model.Incident{}, err
	}

	incident, err := s.buildIncident(ctx, existing, request.Title, request.Summary, request.Severity, request.Status, request.ApplicationID, request.ClusterID, request.DeploymentID, request.OwnerTeam)
	if err != nil {
		return model.Incident{}, err
	}
	now := time.Now().UTC()
	incident.UpdatedAt = now
	if incident.Status == "resolved" && incident.ResolvedAt == nil {
		incident.ResolvedAt = &now
	}
	if incident.Status != "resolved" {
		incident.ResolvedAt = nil
	}

	return s.incidentRepository.Save(ctx, incident)
}

func (s *IncidentService) UpdateIncidentStatus(ctx context.Context, id string, request model.UpdateIncidentStatusRequest) (model.Incident, error) {
	id = strings.TrimSpace(id)
	status := strings.TrimSpace(request.Status)
	if id == "" || !isValidIncidentStatus(status) {
		return model.Incident{}, ErrInvalidIncident
	}

	incident, err := s.incidentRepository.FindByID(ctx, id)
	if err != nil {
		return model.Incident{}, err
	}

	now := time.Now().UTC()
	incident.Status = status
	incident.UpdatedAt = now
	if status == "resolved" {
		incident.ResolvedAt = &now
	} else {
		incident.ResolvedAt = nil
	}

	return s.incidentRepository.Save(ctx, incident)
}

func (s *IncidentService) DeleteIncident(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrInvalidIncident
	}

	return s.incidentRepository.DeleteByID(ctx, id)
}

func (s *IncidentService) buildIncident(ctx context.Context, incident model.Incident, title, summary, severity, status, applicationID, clusterID, deploymentID, ownerTeam string) (model.Incident, error) {
	title = strings.TrimSpace(title)
	summary = strings.TrimSpace(summary)
	severity = strings.TrimSpace(severity)
	status = strings.TrimSpace(status)
	applicationID = strings.TrimSpace(applicationID)
	clusterID = strings.TrimSpace(clusterID)
	deploymentID = strings.TrimSpace(deploymentID)
	ownerTeam = strings.TrimSpace(ownerTeam)

	if title == "" || summary == "" || severity == "" || ownerTeam == "" {
		return model.Incident{}, ErrInvalidIncident
	}
	if status == "" {
		status = "open"
	}
	if !isValidIncidentSeverity(severity) || !isValidIncidentStatus(status) {
		return model.Incident{}, ErrInvalidIncident
	}

	if applicationID != "" {
		if _, err := s.applicationRepository.FindByID(ctx, applicationID); err != nil {
			return model.Incident{}, err
		}
	}
	if clusterID != "" {
		if _, err := s.clusterRepository.FindByID(ctx, clusterID); err != nil {
			return model.Incident{}, err
		}
	}
	if deploymentID != "" {
		if _, err := s.deploymentRepository.FindByID(ctx, deploymentID); err != nil {
			return model.Incident{}, err
		}
	}

	incident.Title = title
	incident.Summary = summary
	incident.Severity = severity
	incident.Status = status
	incident.ApplicationID = applicationID
	incident.ClusterID = clusterID
	incident.DeploymentID = deploymentID
	incident.OwnerTeam = ownerTeam

	return incident, nil
}

func isValidIncidentSeverity(severity string) bool {
	return severity == "sev1" || severity == "sev2" || severity == "sev3" || severity == "sev4"
}

func isValidIncidentStatus(status string) bool {
	return status == "open" || status == "investigating" || status == "mitigated" || status == "resolved"
}

func incidentMatchesQuery(incident model.Incident, query string) bool {
	query = strings.ToLower(query)
	return strings.Contains(strings.ToLower(incident.ID), query) ||
		strings.Contains(strings.ToLower(incident.Title), query) ||
		strings.Contains(strings.ToLower(incident.Summary), query) ||
		strings.Contains(strings.ToLower(incident.Severity), query) ||
		strings.Contains(strings.ToLower(incident.Status), query) ||
		strings.Contains(strings.ToLower(incident.ApplicationID), query) ||
		strings.Contains(strings.ToLower(incident.ClusterID), query) ||
		strings.Contains(strings.ToLower(incident.DeploymentID), query) ||
		strings.Contains(strings.ToLower(incident.OwnerTeam), query)
}
