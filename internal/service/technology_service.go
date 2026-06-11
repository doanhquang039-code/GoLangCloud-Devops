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

var ErrInvalidTechnology = errors.New("invalid technology input")

type TechnologyService struct {
	technologyRepository repository.TechnologyRepository
}

func NewTechnologyService(technologyRepository repository.TechnologyRepository) *TechnologyService {
	return &TechnologyService{technologyRepository: technologyRepository}
}

func (s *TechnologyService) GetTechnologies(ctx context.Context, filter model.TechnologyFilter) ([]model.Technology, error) {
	filter = normalizeTechnologyFilter(filter)
	technologies, err := s.technologyRepository.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	if filter.Query == "" && filter.Category == "" && filter.OwnerTeam == "" && filter.Status == "" && filter.RiskLevel == "" && filter.AdoptionStage == "" && filter.Tag == "" {
		return technologies, nil
	}

	filtered := make([]model.Technology, 0, len(technologies))
	for _, technology := range technologies {
		if filter.Query != "" && !technologyMatchesQuery(technology, filter.Query) {
			continue
		}
		if filter.Category != "" && !strings.EqualFold(technology.Category, filter.Category) {
			continue
		}
		if filter.OwnerTeam != "" && !strings.EqualFold(technology.OwnerTeam, filter.OwnerTeam) {
			continue
		}
		if filter.Status != "" && !strings.EqualFold(technology.Status, filter.Status) {
			continue
		}
		if filter.RiskLevel != "" && !strings.EqualFold(technology.RiskLevel, filter.RiskLevel) {
			continue
		}
		if filter.AdoptionStage != "" && !strings.EqualFold(technology.AdoptionStage, filter.AdoptionStage) {
			continue
		}
		if filter.Tag != "" && !hasApplicationTag(technology.Tags, filter.Tag) {
			continue
		}
		filtered = append(filtered, technology)
	}

	return filtered, nil
}

func (s *TechnologyService) GetTechnologyByID(ctx context.Context, id string) (model.Technology, error) {
	if strings.TrimSpace(id) == "" {
		return model.Technology{}, ErrInvalidTechnology
	}
	return s.technologyRepository.FindByID(ctx, id)
}

func (s *TechnologyService) CreateTechnology(ctx context.Context, request model.CreateTechnologyRequest) (model.Technology, error) {
	technology, err := technologyFromRequest("", request)
	if err != nil {
		return model.Technology{}, err
	}
	now := time.Now().UTC()
	technology.ID = fmt.Sprintf("tech-%d", now.UnixNano())
	technology.CreatedAt = now
	technology.UpdatedAt = now
	return s.technologyRepository.Save(ctx, technology)
}

func (s *TechnologyService) UpdateTechnology(ctx context.Context, id string, request model.UpdateTechnologyRequest) (model.Technology, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return model.Technology{}, ErrInvalidTechnology
	}
	existing, err := s.technologyRepository.FindByID(ctx, id)
	if err != nil {
		return model.Technology{}, err
	}
	technology, err := technologyFromRequest(id, model.CreateTechnologyRequest(request))
	if err != nil {
		return model.Technology{}, err
	}
	technology.CreatedAt = existing.CreatedAt
	technology.UpdatedAt = time.Now().UTC()
	return s.technologyRepository.Save(ctx, technology)
}

func (s *TechnologyService) DeleteTechnology(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrInvalidTechnology
	}
	return s.technologyRepository.DeleteByID(ctx, id)
}

func technologyFromRequest(id string, request model.CreateTechnologyRequest) (model.Technology, error) {
	request.Name = strings.TrimSpace(request.Name)
	request.Category = strings.ToLower(strings.TrimSpace(request.Category))
	request.Version = strings.TrimSpace(request.Version)
	request.OwnerTeam = strings.TrimSpace(request.OwnerTeam)
	request.Status = strings.ToLower(strings.TrimSpace(request.Status))
	request.RiskLevel = strings.ToLower(strings.TrimSpace(request.RiskLevel))
	request.AdoptionStage = strings.ToLower(strings.TrimSpace(request.AdoptionStage))
	request.License = strings.TrimSpace(request.License)
	request.DocumentationURL = strings.TrimSpace(request.DocumentationURL)

	if request.Name == "" || request.Category == "" || request.Version == "" || request.OwnerTeam == "" {
		return model.Technology{}, ErrInvalidTechnology
	}
	if request.Status == "" {
		request.Status = "active"
	}
	if request.RiskLevel == "" {
		request.RiskLevel = "medium"
	}
	if request.AdoptionStage == "" {
		request.AdoptionStage = "adopt"
	}

	return model.Technology{
		ID:               id,
		Name:             request.Name,
		Category:         request.Category,
		Version:          request.Version,
		OwnerTeam:        request.OwnerTeam,
		Status:           request.Status,
		RiskLevel:        request.RiskLevel,
		AdoptionStage:    request.AdoptionStage,
		License:          request.License,
		DocumentationURL: request.DocumentationURL,
		Tags:             normalizeTags(request.Tags),
	}, nil
}

func normalizeTechnologyFilter(filter model.TechnologyFilter) model.TechnologyFilter {
	filter.Query = strings.TrimSpace(filter.Query)
	filter.Category = strings.TrimSpace(filter.Category)
	filter.OwnerTeam = strings.TrimSpace(filter.OwnerTeam)
	filter.Status = strings.TrimSpace(filter.Status)
	filter.RiskLevel = strings.TrimSpace(filter.RiskLevel)
	filter.AdoptionStage = strings.TrimSpace(filter.AdoptionStage)
	filter.Tag = strings.TrimSpace(filter.Tag)
	return filter
}

func technologyMatchesQuery(technology model.Technology, query string) bool {
	query = strings.ToLower(query)
	if strings.Contains(strings.ToLower(technology.ID), query) ||
		strings.Contains(strings.ToLower(technology.Name), query) ||
		strings.Contains(strings.ToLower(technology.Category), query) ||
		strings.Contains(strings.ToLower(technology.Version), query) ||
		strings.Contains(strings.ToLower(technology.OwnerTeam), query) ||
		strings.Contains(strings.ToLower(technology.Status), query) ||
		strings.Contains(strings.ToLower(technology.RiskLevel), query) ||
		strings.Contains(strings.ToLower(technology.AdoptionStage), query) ||
		strings.Contains(strings.ToLower(technology.License), query) ||
		strings.Contains(strings.ToLower(technology.DocumentationURL), query) {
		return true
	}
	for _, tag := range technology.Tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}
	return false
}
