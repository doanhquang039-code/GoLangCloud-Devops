package repository

import (
	"context"
	"errors"
	"sync"

	"hr-cloud-service/internal/model"
)

var ErrActivityNotFound = errors.New("activity not found")

type ActivityRepository interface {
	FindAll(ctx context.Context) ([]model.Activity, error)
	FindByID(ctx context.Context, id string) (model.Activity, error)
	Save(ctx context.Context, activity model.Activity) (model.Activity, error)
	DeleteByID(ctx context.Context, id string) error
}

type InMemoryActivityRepository struct {
	mu         sync.RWMutex
	activities map[string]model.Activity
}

func NewInMemoryActivityRepository() *InMemoryActivityRepository {
	return &InMemoryActivityRepository{
		activities: make(map[string]model.Activity),
	}
}

func (r *InMemoryActivityRepository) FindAll(ctx context.Context) ([]model.Activity, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	activities := make([]model.Activity, 0, len(r.activities))
	for _, activity := range r.activities {
		activities = append(activities, activity)
	}

	return activities, nil
}

func (r *InMemoryActivityRepository) FindByID(ctx context.Context, id string) (model.Activity, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	activity, ok := r.activities[id]
	if !ok {
		return model.Activity{}, ErrActivityNotFound
	}

	return activity, nil
}

func (r *InMemoryActivityRepository) Save(ctx context.Context, activity model.Activity) (model.Activity, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.activities[activity.ID] = activity
	return activity, nil
}

func (r *InMemoryActivityRepository) DeleteByID(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.activities[id]; !ok {
		return ErrActivityNotFound
	}

	delete(r.activities, id)
	return nil
}
