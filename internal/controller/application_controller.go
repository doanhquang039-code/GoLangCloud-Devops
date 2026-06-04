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

type ApplicationController struct {
	applicationService *service.ApplicationService
}

func NewApplicationController(applicationService *service.ApplicationService) *ApplicationController {
	return &ApplicationController{applicationService: applicationService}
}

func (c *ApplicationController) Index(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		c.GetApplications(w, r)
	case http.MethodPost:
		c.CreateApplication(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (c *ApplicationController) GetApplications(w http.ResponseWriter, r *http.Request) {
	applications, err := c.applicationService.GetApplications(r.Context(), model.ApplicationFilter{
		Query:       r.URL.Query().Get("q"),
		OwnerTeam:   r.URL.Query().Get("owner_team"),
		Criticality: r.URL.Query().Get("criticality"),
		Runtime:     r.URL.Query().Get("runtime"),
		Tag:         r.URL.Query().Get("tag"),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list applications")
		return
	}

	writeJSON(w, http.StatusOK, applications)
}

func (c *ApplicationController) Show(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/applications/")

	switch r.Method {
	case http.MethodGet:
		c.GetApplication(w, r, id)
	case http.MethodPut:
		c.UpdateApplication(w, r, id)
	case http.MethodDelete:
		c.DeleteApplication(w, r, id)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (c *ApplicationController) GetApplication(w http.ResponseWriter, r *http.Request, id string) {
	application, err := c.applicationService.GetApplicationByID(r.Context(), id)
	if errors.Is(err, repository.ErrApplicationNotFound) {
		writeError(w, http.StatusNotFound, "application not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, application)
}

func (c *ApplicationController) UpdateApplication(w http.ResponseWriter, r *http.Request, id string) {
	defer r.Body.Close()

	var request model.UpdateApplicationRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	application, err := c.applicationService.UpdateApplication(r.Context(), id, request)
	if errors.Is(err, service.ErrInvalidApplication) {
		writeError(w, http.StatusBadRequest, "name, repository, runtime and owner_team are required; port must be 1-65535 and replicas must be positive")
		return
	}
	if errors.Is(err, repository.ErrApplicationNotFound) {
		writeError(w, http.StatusNotFound, "application not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not update application")
		return
	}

	writeJSON(w, http.StatusOK, application)
}

func (c *ApplicationController) DeleteApplication(w http.ResponseWriter, r *http.Request, id string) {
	err := c.applicationService.DeleteApplication(r.Context(), id)
	if errors.Is(err, service.ErrInvalidApplication) {
		writeError(w, http.StatusBadRequest, "application id is required")
		return
	}
	if errors.Is(err, repository.ErrApplicationNotFound) {
		writeError(w, http.StatusNotFound, "application not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not delete application")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *ApplicationController) CreateApplication(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var request model.CreateApplicationRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	application, err := c.applicationService.CreateApplication(r.Context(), request)
	if errors.Is(err, service.ErrInvalidApplication) {
		writeError(w, http.StatusBadRequest, "name, repository, runtime and owner_team are required; port must be 1-65535 and replicas must be positive")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create application")
		return
	}

	writeJSON(w, http.StatusCreated, application)
}
