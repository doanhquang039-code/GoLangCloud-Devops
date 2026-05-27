package repository

import (
	"context"
	"errors"
	"sync"

	"hr-cloud-service/internal/model"
)

var ErrEmployeeNotFound = errors.New("employee not found")

type EmployeeRepository interface {
	FindAll(ctx context.Context) ([]model.Employee, error)
	FindByID(ctx context.Context, id string) (model.Employee, error)
	Save(ctx context.Context, employee model.Employee) (model.Employee, error)
	DeleteByID(ctx context.Context, id string) error
}

type InMemoryEmployeeRepository struct {
	mu        sync.RWMutex
	employees map[string]model.Employee
}

func NewInMemoryEmployeeRepository() *InMemoryEmployeeRepository {
	return &InMemoryEmployeeRepository{
		employees: make(map[string]model.Employee),
	}
}

func (r *InMemoryEmployeeRepository) FindAll(ctx context.Context) ([]model.Employee, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	employees := make([]model.Employee, 0, len(r.employees))
	for _, employee := range r.employees {
		employees = append(employees, employee)
	}

	return employees, nil
}

func (r *InMemoryEmployeeRepository) FindByID(ctx context.Context, id string) (model.Employee, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	employee, ok := r.employees[id]
	if !ok {
		return model.Employee{}, ErrEmployeeNotFound
	}

	return employee, nil
}

func (r *InMemoryEmployeeRepository) Save(ctx context.Context, employee model.Employee) (model.Employee, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.employees[employee.ID] = employee
	return employee, nil
}

func (r *InMemoryEmployeeRepository) DeleteByID(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.employees[id]; !ok {
		return ErrEmployeeNotFound
	}

	delete(r.employees, id)
	return nil
}
