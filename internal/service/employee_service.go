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

func (s *EmployeeService) UpdateEmployee(ctx context.Context, id string, request model.UpdateEmployeeRequest) (model.Employee, error) {
	id = strings.TrimSpace(id)
	request.Name = strings.TrimSpace(request.Name)
	request.Email = strings.TrimSpace(request.Email)
	request.Department = strings.TrimSpace(request.Department)
	request.Title = strings.TrimSpace(request.Title)

	if id == "" || request.Name == "" || request.Email == "" || request.Department == "" || request.Title == "" {
		return model.Employee{}, ErrInvalidEmployee
	}

	employee, err := s.employeeRepository.FindByID(ctx, id)
	if err != nil {
		return model.Employee{}, err
	}

	employee.Name = request.Name
	employee.Email = request.Email
	employee.Department = request.Department
	employee.Title = request.Title
	employee.UpdatedAt = time.Now().UTC()

	return s.employeeRepository.Save(ctx, employee)
}

func (s *EmployeeService) DeleteEmployee(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrInvalidEmployee
	}

	return s.employeeRepository.DeleteByID(ctx, id)
}
