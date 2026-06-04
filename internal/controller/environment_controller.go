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

type EnvironmentController struct {
	environmentService *service.EnvironmentService
}

func NewEnvironmentController(environmentService *service.EnvironmentService) *EnvironmentController {
	return &EnvironmentController{environmentService: environmentService}
}

func (c *EnvironmentController) Index(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		c.GetEnvironments(w, r)
	case http.MethodPost:
		c.CreateEnvironment(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (c *EnvironmentController) GetEnvironments(w http.ResponseWriter, r *http.Request) {
	environments, err := c.environmentService.GetEnvironments(r.Context(), model.EnvironmentFilter{
		Query:         r.URL.Query().Get("q"),
		ApplicationID: r.URL.Query().Get("application_id"),
		ClusterID:     r.URL.Query().Get("cluster_id"),
		Type:          r.URL.Query().Get("type"),
		Status:        r.URL.Query().Get("status"),
	})
	if errors.Is(err, service.ErrInvalidEnvironment) {
		writeError(w, http.StatusBadRequest, "type must be development, staging or production; status must be active, inactive or deprecated")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list environments")
		return
	}

	writeJSON(w, http.StatusOK, environments)
}

func (c *EnvironmentController) ShowOrUpdate(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/environments/")

	switch r.Method {
	case http.MethodGet:
		c.Show(w, r, id)
	case http.MethodPut:
		c.UpdateEnvironment(w, r, id)
	case http.MethodDelete:
		c.DeleteEnvironment(w, r, id)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (c *EnvironmentController) Show(w http.ResponseWriter, r *http.Request, id string) {
	environment, err := c.environmentService.GetEnvironmentByID(r.Context(), id)
	if errors.Is(err, repository.ErrEnvironmentNotFound) {
		writeError(w, http.StatusNotFound, "environment not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, environment)
}

func (c *EnvironmentController) DeleteEnvironment(w http.ResponseWriter, r *http.Request, id string) {
	err := c.environmentService.DeleteEnvironment(r.Context(), id)
	c.writeEnvironmentError(w, err, "could not delete environment")
	if err != nil {
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *EnvironmentController) CreateEnvironment(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var request model.CreateEnvironmentRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	environment, err := c.environmentService.CreateEnvironment(r.Context(), request)
	c.writeEnvironmentError(w, err, "could not create environment")
	if err != nil {
		return
	}

	writeJSON(w, http.StatusCreated, environment)
}

func (c *EnvironmentController) UpdateEnvironment(w http.ResponseWriter, r *http.Request, id string) {
	defer r.Body.Close()

	var request model.UpdateEnvironmentRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	environment, err := c.environmentService.UpdateEnvironment(r.Context(), id, request)
	c.writeEnvironmentError(w, err, "could not update environment")
	if err != nil {
		return
	}

	writeJSON(w, http.StatusOK, environment)
}

func (c *EnvironmentController) writeEnvironmentError(w http.ResponseWriter, err error, fallback string) {
	if err == nil {
		return
	}
	if errors.Is(err, service.ErrInvalidEnvironment) {
		writeError(w, http.StatusBadRequest, "name, type, application_id and cluster_id are required; type must be development, staging or production; status must be active, inactive or deprecated")
		return
	}
	if errors.Is(err, repository.ErrEnvironmentNotFound) {
		writeError(w, http.StatusNotFound, "environment not found")
		return
	}
	if errors.Is(err, repository.ErrApplicationNotFound) {
		writeError(w, http.StatusNotFound, "application not found")
		return
	}
	if errors.Is(err, repository.ErrClusterNotFound) {
		writeError(w, http.StatusNotFound, "cluster not found")
		return
	}
	writeError(w, http.StatusInternalServerError, fallback)
}
