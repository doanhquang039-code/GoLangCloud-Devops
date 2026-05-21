package controller

import (
    "encoding/json"
    "errors"
    "net/http"
    "strings"

    "hr-cloud-service/internal/model"
    "hr-cloud-service/internal/repository"
    "hr-cloud-service/internal/service"
)

type EmployeeController struct {
    employeeService *service.EmployeeService
}

func NewEmployeeController(employeeService *service.EmployeeService) *EmployeeController {
    return &EmployeeController{employeeService: employeeService}
}

func (c *EmployeeController) Index(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        c.GetEmployees(w, r)
    case http.MethodPost:
        c.CreateEmployee(w, r)
    default:
        writeError(w, http.StatusMethodNotAllowed, "method not allowed")
    }
}

func (c *EmployeeController) GetEmployees(w http.ResponseWriter, r *http.Request) {
    employees, err := c.employeeService.GetEmployees(r.Context())
    if err != nil {
        writeError(w, http.StatusInternalServerError, "could not list employees")
        return
    }

    writeJSON(w, http.StatusOK, employees)
}

func (c *EmployeeController) Show(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        writeError(w, http.StatusMethodNotAllowed, "method not allowed")
        return
    }

    id := strings.TrimPrefix(r.URL.Path, "/api/v1/employees/")
    employee, err := c.employeeService.GetEmployeeByID(r.Context(), id)
    if errors.Is(err, repository.ErrEmployeeNotFound) {
        writeError(w, http.StatusNotFound, "employee not found")
        return
    }
    if err != nil {
        writeError(w, http.StatusBadRequest, err.Error())
        return
    }

    writeJSON(w, http.StatusOK, employee)
}

func (c *EmployeeController) CreateEmployee(w http.ResponseWriter, r *http.Request) {
    defer r.Body.Close()

    var request model.CreateEmployeeRequest
    if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
        writeError(w, http.StatusBadRequest, "invalid json body")
        return
    }

    employee, err := c.employeeService.CreateEmployee(r.Context(), request)
    if errors.Is(err, service.ErrInvalidEmployee) {
        writeError(w, http.StatusBadRequest, "name, email, department and title are required")
        return
    }
    if err != nil {
        writeError(w, http.StatusInternalServerError, "could not create employee")
        return
    }

    writeJSON(w, http.StatusCreated, employee)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
    writeJSON(w, status, map[string]string{"error": message})
}
