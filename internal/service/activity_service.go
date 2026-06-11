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

var ErrInvalidActivity = errors.New("invalid activity input")

type ActivityService struct {
	activityRepository repository.ActivityRepository
}

func NewActivityService(activityRepository repository.ActivityRepository) *ActivityService {
	return &ActivityService{activityRepository: activityRepository}
}

func (s *ActivityService) GetActivities(ctx context.Context, filter model.ActivityFilter) ([]model.Activity, error) {
	filter = normalizeActivityFilter(filter)
	activities, err := s.activityRepository.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	if filter.Query == "" && filter.Type == "" && filter.Action == "" && filter.Status == "" && filter.Actor == "" && filter.ResourceType == "" && filter.ResourceID == "" && filter.ApplicationID == "" && filter.OwnerTeam == "" && filter.Tag == "" {
		return activities, nil
	}

	filtered := make([]model.Activity, 0, len(activities))
	for _, activity := range activities {
		if filter.Query != "" && !activityMatchesQuery(activity, filter.Query) {
			continue
		}
		if filter.Type != "" && !strings.EqualFold(activity.Type, filter.Type) {
			continue
		}
		if filter.Action != "" && !strings.EqualFold(activity.Action, filter.Action) {
			continue
		}
		if filter.Status != "" && !strings.EqualFold(activity.Status, filter.Status) {
			continue
		}
		if filter.Actor != "" && !strings.EqualFold(activity.Actor, filter.Actor) {
			continue
		}
		if filter.ResourceType != "" && !strings.EqualFold(activity.ResourceType, filter.ResourceType) {
			continue
		}
		if filter.ResourceID != "" && activity.ResourceID != filter.ResourceID {
			continue
		}
		if filter.ApplicationID != "" && activity.ApplicationID != filter.ApplicationID {
			continue
		}
		if filter.OwnerTeam != "" && !strings.EqualFold(activity.OwnerTeam, filter.OwnerTeam) {
			continue
		}
		if filter.Tag != "" && !hasApplicationTag(activity.Tags, filter.Tag) {
			continue
		}
		filtered = append(filtered, activity)
	}

	return filtered, nil
}

func (s *ActivityService) GetActivityByID(ctx context.Context, id string) (model.Activity, error) {
	if strings.TrimSpace(id) == "" {
		return model.Activity{}, ErrInvalidActivity
	}
	return s.activityRepository.FindByID(ctx, id)
}

func (s *ActivityService) CreateActivity(ctx context.Context, request model.CreateActivityRequest) (model.Activity, error) {
	activity, err := activityFromRequest("", request)
	if err != nil {
		return model.Activity{}, err
	}
	now := time.Now().UTC()
	activity.ID = fmt.Sprintf("act-%d", now.UnixNano())
	activity.CreatedAt = now
	activity.UpdatedAt = now
	return s.activityRepository.Save(ctx, activity)
}

func (s *ActivityService) UpdateActivity(ctx context.Context, id string, request model.UpdateActivityRequest) (model.Activity, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return model.Activity{}, ErrInvalidActivity
	}
	existing, err := s.activityRepository.FindByID(ctx, id)
	if err != nil {
		return model.Activity{}, err
	}
	activity, err := activityFromRequest(id, model.CreateActivityRequest(request))
	if err != nil {
		return model.Activity{}, err
	}
	activity.CreatedAt = existing.CreatedAt
	activity.UpdatedAt = time.Now().UTC()
	return s.activityRepository.Save(ctx, activity)
}

func (s *ActivityService) DeleteActivity(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrInvalidActivity
	}
	return s.activityRepository.DeleteByID(ctx, id)
}

func activityFromRequest(id string, request model.CreateActivityRequest) (model.Activity, error) {
	request.Type = strings.ToLower(strings.TrimSpace(request.Type))
	request.Action = strings.ToLower(strings.TrimSpace(request.Action))
	request.Status = strings.ToLower(strings.TrimSpace(request.Status))
	request.Actor = strings.TrimSpace(request.Actor)
	request.ResourceType = strings.ToLower(strings.TrimSpace(request.ResourceType))
	request.ResourceID = strings.TrimSpace(request.ResourceID)
	request.ApplicationID = strings.TrimSpace(request.ApplicationID)
	request.OwnerTeam = strings.TrimSpace(request.OwnerTeam)
	request.Summary = strings.TrimSpace(request.Summary)

	if request.Type == "" || request.Action == "" || request.Actor == "" || request.ResourceType == "" || request.ResourceID == "" || request.Summary == "" {
		return model.Activity{}, ErrInvalidActivity
	}
	if request.Status == "" {
		request.Status = "succeeded"
	}

	return model.Activity{
		ID:            id,
		Type:          request.Type,
		Action:        request.Action,
		Status:        request.Status,
		Actor:         request.Actor,
		ResourceType:  request.ResourceType,
		ResourceID:    request.ResourceID,
		ApplicationID: request.ApplicationID,
		OwnerTeam:     request.OwnerTeam,
		Summary:       request.Summary,
		Metadata:      request.Metadata,
		Tags:          normalizeTags(request.Tags),
	}, nil
}

func normalizeActivityFilter(filter model.ActivityFilter) model.ActivityFilter {
	filter.Query = strings.TrimSpace(filter.Query)
	filter.Type = strings.TrimSpace(filter.Type)
	filter.Action = strings.TrimSpace(filter.Action)
	filter.Status = strings.TrimSpace(filter.Status)
	filter.Actor = strings.TrimSpace(filter.Actor)
	filter.ResourceType = strings.TrimSpace(filter.ResourceType)
	filter.ResourceID = strings.TrimSpace(filter.ResourceID)
	filter.ApplicationID = strings.TrimSpace(filter.ApplicationID)
	filter.OwnerTeam = strings.TrimSpace(filter.OwnerTeam)
	filter.Tag = strings.TrimSpace(filter.Tag)
	return filter
}

func activityMatchesQuery(activity model.Activity, query string) bool {
	query = strings.ToLower(query)
	if strings.Contains(strings.ToLower(activity.ID), query) ||
		strings.Contains(strings.ToLower(activity.Type), query) ||
		strings.Contains(strings.ToLower(activity.Action), query) ||
		strings.Contains(strings.ToLower(activity.Status), query) ||
		strings.Contains(strings.ToLower(activity.Actor), query) ||
		strings.Contains(strings.ToLower(activity.ResourceType), query) ||
		strings.Contains(strings.ToLower(activity.ResourceID), query) ||
		strings.Contains(strings.ToLower(activity.ApplicationID), query) ||
		strings.Contains(strings.ToLower(activity.OwnerTeam), query) ||
		strings.Contains(strings.ToLower(activity.Summary), query) {
		return true
	}
	for key, value := range activity.Metadata {
		if strings.Contains(strings.ToLower(key), query) || strings.Contains(strings.ToLower(value), query) {
			return true
		}
	}
	for _, tag := range activity.Tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}
	return false
}
