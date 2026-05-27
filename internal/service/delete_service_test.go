package service

import (
	"context"
	"errors"
	"testing"

	"hr-cloud-service/internal/model"
	"hr-cloud-service/internal/repository"
)

func TestEmployeeServiceDeletesEmployee(t *testing.T) {
	ctx := context.Background()
	employeeRepository := repository.NewInMemoryEmployeeRepository()
	employeeService := NewEmployeeService(employeeRepository)

	if _, err := employeeRepository.Save(ctx, model.Employee{ID: "emp-001", Name: "Nguyen Van A"}); err != nil {
		t.Fatal(err)
	}

	if err := employeeService.DeleteEmployee(ctx, "emp-001"); err != nil {
		t.Fatal(err)
	}

	if _, err := employeeRepository.FindByID(ctx, "emp-001"); !errors.Is(err, repository.ErrEmployeeNotFound) {
		t.Fatalf("expected deleted employee to be missing, got %v", err)
	}
}

func TestApplicationServiceDeleteRequiresID(t *testing.T) {
	applicationRepository := repository.NewInMemoryApplicationRepository()
	applicationService := NewApplicationService(applicationRepository)

	err := applicationService.DeleteApplication(context.Background(), " ")
	if !errors.Is(err, ErrInvalidApplication) {
		t.Fatalf("expected invalid application error, got %v", err)
	}
}
