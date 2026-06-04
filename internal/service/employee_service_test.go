package service

import (
	"context"
	"testing"

	"hr-cloud-service/internal/model"
	"hr-cloud-service/internal/repository"
)

func TestGetEmployeesFiltersByQueryDepartmentAndTitle(t *testing.T) {
	ctx := context.Background()
	employeeRepository := repository.NewInMemoryEmployeeRepository()
	employeeService := NewEmployeeService(employeeRepository)

	employees := []model.Employee{
		{ID: "emp-1", Name: "An Nguyen", Email: "an@example.com", Department: "Platform", Title: "SRE"},
		{ID: "emp-2", Name: "Binh Tran", Email: "binh@example.com", Department: "People", Title: "HRBP"},
		{ID: "emp-3", Name: "Chi Le", Email: "chi@example.com", Department: "Platform", Title: "Backend Engineer"},
	}
	for _, employee := range employees {
		if _, err := employeeRepository.Save(ctx, employee); err != nil {
			t.Fatal(err)
		}
	}

	filtered, err := employeeService.GetEmployees(ctx, model.EmployeeFilter{
		Query:      "an",
		Department: "platform",
		Title:      "sre",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(filtered) != 1 {
		t.Fatalf("expected 1 employee, got %d", len(filtered))
	}
	if filtered[0].ID != "emp-1" {
		t.Fatalf("expected emp-1, got %q", filtered[0].ID)
	}
}
