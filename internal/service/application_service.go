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

var ErrInvalidApplication = errors.New("invalid application input")

type ApplicationService struct {
	applicationRepository repository.ApplicationRepository
}

func NewApplicationService(applicationRepository repository.ApplicationRepository) *ApplicationService {
	return &ApplicationService{applicationRepository: applicationRepository}
}

func (s *ApplicationService) GetApplications(ctx context.Context) ([]model.Application, error) {
	return s.applicationRepository.FindAll(ctx)
}

func (s *ApplicationService) GetApplicationByID(ctx context.Context, id string) (model.Application, error) {
	if strings.TrimSpace(id) == "" {
		return model.Application{}, ErrInvalidApplication
	}

	return s.applicationRepository.FindByID(ctx, id)
}

func (s *ApplicationService) CreateApplication(ctx context.Context, request model.CreateApplicationRequest) (model.Application, error) {
	request.Name = strings.TrimSpace(request.Name)
	request.Repository = strings.TrimSpace(request.Repository)
	request.Runtime = strings.TrimSpace(request.Runtime)
	request.OwnerTeam = strings.TrimSpace(request.OwnerTeam)
	request.Criticality = strings.TrimSpace(request.Criticality)

	if request.Name == "" || request.Repository == "" || request.Runtime == "" || request.OwnerTeam == "" {
		return model.Application{}, ErrInvalidApplication
	}
	if request.Criticality == "" {
		request.Criticality = "medium"
	}

	now := time.Now().UTC()
	application := model.Application{
		ID:          fmt.Sprintf("app-%d", now.UnixNano()),
		Name:        request.Name,
		Repository:  request.Repository,
		Runtime:     request.Runtime,
		OwnerTeam:   request.OwnerTeam,
		Criticality: request.Criticality,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	return s.applicationRepository.Save(ctx, application)
}
