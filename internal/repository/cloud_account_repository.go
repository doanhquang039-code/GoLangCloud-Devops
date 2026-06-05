package repository

import (
	"context"
	"errors"
	"sync"

	"hr-cloud-service/internal/model"
)

var ErrCloudAccountNotFound = errors.New("cloud account not found")

type CloudAccountRepository interface {
	FindAll(ctx context.Context) ([]model.CloudAccount, error)
	FindByID(ctx context.Context, id string) (model.CloudAccount, error)
	Save(ctx context.Context, account model.CloudAccount) (model.CloudAccount, error)
	DeleteByID(ctx context.Context, id string) error
}

type InMemoryCloudAccountRepository struct {
	mu       sync.RWMutex
	accounts map[string]model.CloudAccount
}

func NewInMemoryCloudAccountRepository() *InMemoryCloudAccountRepository {
	return &InMemoryCloudAccountRepository{
		accounts: make(map[string]model.CloudAccount),
	}
}

func (r *InMemoryCloudAccountRepository) FindAll(ctx context.Context) ([]model.CloudAccount, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	accounts := make([]model.CloudAccount, 0, len(r.accounts))
	for _, account := range r.accounts {
		accounts = append(accounts, account)
	}

	return accounts, nil
}

func (r *InMemoryCloudAccountRepository) FindByID(ctx context.Context, id string) (model.CloudAccount, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	account, ok := r.accounts[id]
	if !ok {
		return model.CloudAccount{}, ErrCloudAccountNotFound
	}

	return account, nil
}

func (r *InMemoryCloudAccountRepository) Save(ctx context.Context, account model.CloudAccount) (model.CloudAccount, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.accounts[account.ID] = account
	return account, nil
}

func (r *InMemoryCloudAccountRepository) DeleteByID(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.accounts[id]; !ok {
		return ErrCloudAccountNotFound
	}

	delete(r.accounts, id)
	return nil
}
