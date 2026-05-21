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

var ErrInvalidEmployee = errors.New("invalid employee input")

type EmployeeService struct {
	employeeRepository repository.EmployeeRepository
}

func NewEmployeeService(employeeRepository repository.EmployeeRepository) *EmployeeService {
	return &EmployeeService{employeeRepository: employeeRepository}
}

func (s *EmployeeService) GetEmployees(ctx context.Context) ([]model.Employee, error) {
	return s.employeeRepository.FindAll(ctx)
}

func (s *EmployeeService) GetEmployeeByID(ctx context.Context, id string) (model.Employee, error) {
	if strings.TrimSpace(id) == "" {
		return model.Employee{}, ErrInvalidEmployee
	}

	return s.employeeRepository.FindByID(ctx, id)
}

func (s *EmployeeService) CreateEmployee(ctx context.Context, request model.CreateEmployeeRequest) (model.Employee, error) {
	request.Name = strings.TrimSpace(request.Name)
	request.Email = strings.TrimSpace(request.Email)
	request.Department = strings.TrimSpace(request.Department)
	request.Title = strings.TrimSpace(request.Title)

	if request.Name == "" || request.Email == "" || request.Department == "" || request.Title == "" {
		return model.Employee{}, ErrInvalidEmployee
	}

	now := time.Now().UTC()
	employee := model.Employee{
		ID:         fmt.Sprintf("emp-%d", now.UnixNano()),
		Name:       request.Name,
		Email:      request.Email,
		Department: request.Department,
		Title:      request.Title,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	return s.employeeRepository.Save(ctx, employee)
}
