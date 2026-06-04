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

type MicroserviceController struct {
	microserviceService *service.MicroserviceService
}

func NewMicroserviceController(microserviceService *service.MicroserviceService) *MicroserviceController {
	return &MicroserviceController{microserviceService: microserviceService}
}

func (c *MicroserviceController) Index(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		c.GetMicroservices(w, r)
	case http.MethodPost:
		c.CreateMicroservice(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (c *MicroserviceController) GetMicroservices(w http.ResponseWriter, r *http.Request) {
	microservices, err := c.microserviceService.GetMicroservices(r.Context(), model.MicroserviceFilter{
		Query:         r.URL.Query().Get("q"),
		ApplicationID: r.URL.Query().Get("application_id"),
		OwnerTeam:     r.URL.Query().Get("owner_team"),
		Protocol:      r.URL.Query().Get("protocol"),
		Status:        r.URL.Query().Get("status"),
		Tag:           r.URL.Query().Get("tag"),
	})
	if errors.Is(err, service.ErrInvalidMicroservice) {
		writeError(w, http.StatusBadRequest, "protocol must be http, grpc, event or worker; status must be active, degraded, offline or deprecated")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list microservices")
		return
	}

	writeJSON(w, http.StatusOK, microservices)
}

func (c *MicroserviceController) ShowOrUpdate(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/microservices/")

	switch r.Method {
	case http.MethodGet:
		c.Show(w, r, id)
	case http.MethodPut:
		c.UpdateMicroservice(w, r, id)
	case http.MethodPatch:
		c.UpdateStatus(w, r, id)
	case http.MethodDelete:
		c.DeleteMicroservice(w, r, id)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (c *MicroserviceController) Show(w http.ResponseWriter, r *http.Request, id string) {
	microservice, err := c.microserviceService.GetMicroserviceByID(r.Context(), id)
	if errors.Is(err, repository.ErrMicroserviceNotFound) {
		writeError(w, http.StatusNotFound, "microservice not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, microservice)
}

func (c *MicroserviceController) CreateMicroservice(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var request model.CreateMicroserviceRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	microservice, err := c.microserviceService.CreateMicroservice(r.Context(), request)
	c.handleMutationResult(w, microservice, err, "could not create microservice", http.StatusCreated)
}

func (c *MicroserviceController) UpdateMicroservice(w http.ResponseWriter, r *http.Request, id string) {
	defer r.Body.Close()

	var request model.UpdateMicroserviceRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	microservice, err := c.microserviceService.UpdateMicroservice(r.Context(), id, request)
	c.handleMutationResult(w, microservice, err, "could not update microservice", http.StatusOK)
}

func (c *MicroserviceController) UpdateStatus(w http.ResponseWriter, r *http.Request, id string) {
	defer r.Body.Close()

	var request model.UpdateMicroserviceStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	microservice, err := c.microserviceService.UpdateMicroserviceStatus(r.Context(), id, request)
	c.handleMutationResult(w, microservice, err, "could not update microservice", http.StatusOK)
}

func (c *MicroserviceController) DeleteMicroservice(w http.ResponseWriter, r *http.Request, id string) {
	err := c.microserviceService.DeleteMicroservice(r.Context(), id)
	if errors.Is(err, service.ErrInvalidMicroservice) {
		writeError(w, http.StatusBadRequest, "microservice id is required")
		return
	}
	if errors.Is(err, repository.ErrMicroserviceNotFound) {
		writeError(w, http.StatusNotFound, "microservice not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not delete microservice")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *MicroserviceController) handleMutationResult(w http.ResponseWriter, microservice model.Microservice, err error, fallback string, status int) {
	if errors.Is(err, service.ErrInvalidMicroservice) {
		writeError(w, http.StatusBadRequest, "application_id, name, owner_team, protocol and endpoint are required; protocol must be http, grpc, event or worker; status must be active, degraded, offline or deprecated")
		return
	}
	if errors.Is(err, repository.ErrMicroserviceNotFound) {
		writeError(w, http.StatusNotFound, "microservice not found")
		return
	}
	if errors.Is(err, repository.ErrApplicationNotFound) {
		writeError(w, http.StatusNotFound, "application not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, fallback)
		return
	}

	writeJSON(w, status, microservice)
}
