package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"hr-cloud-service/internal/model"
	"hr-cloud-service/internal/repository"
	"hr-cloud-service/internal/service"
)

func TestEmployeeControllerFiltersEmployees(t *testing.T) {
	employeeRepository := repository.NewInMemoryEmployeeRepository()
	employeeService := service.NewEmployeeService(employeeRepository)
	employeeController := NewEmployeeController(employeeService)

	seed := []model.Employee{
		{ID: "emp-1", Name: "An Nguyen", Email: "an@example.com", Department: "Platform", Title: "SRE"},
		{ID: "emp-2", Name: "Binh Tran", Email: "binh@example.com", Department: "People", Title: "HRBP"},
	}
	for _, employee := range seed {
		if _, err := employeeRepository.Save(httptest.NewRequest(http.MethodGet, "/", nil).Context(), employee); err != nil {
			t.Fatal(err)
		}
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/employees?q=nguyen&department=platform&title=sre", nil)
	response := httptest.NewRecorder()

	employeeController.Index(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var employees []model.Employee
	if err := json.NewDecoder(response.Body).Decode(&employees); err != nil {
		t.Fatal(err)
	}
	if len(employees) != 1 {
		t.Fatalf("expected 1 employee, got %d", len(employees))
	}
	if employees[0].ID != "emp-1" {
		t.Fatalf("expected emp-1, got %q", employees[0].ID)
	}
}
