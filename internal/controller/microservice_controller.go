package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
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
	filter, ok := parseMicroserviceFilter(w, r)
	if !ok {
		return
	}

	microservices, err := c.microserviceService.GetMicroservices(r.Context(), filter)
	if errors.Is(err, service.ErrInvalidMicroservice) {
		writeError(w, http.StatusBadRequest, "protocol must be http, grpc, event or worker; status must be active, degraded, offline or deprecated; environment must be development, staging or production; limit must be 1 through 1000; offset must not be negative; after_id cannot be combined with offset or non-id sorting")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list microservices")
		return
	}

	writeJSON(w, http.StatusOK, microservices)
}

func parseMicroserviceFilter(w http.ResponseWriter, r *http.Request) (model.MicroserviceFilter, bool) {
	query := r.URL.Query()
	filter := model.MicroserviceFilter{
		Query:         r.URL.Query().Get("q"),
		TenantID:      r.URL.Query().Get("tenant_id"),
		ApplicationID: r.URL.Query().Get("application_id"),
		OwnerTeam:     r.URL.Query().Get("owner_team"),
		Protocol:      r.URL.Query().Get("protocol"),
		Status:        r.URL.Query().Get("status"),
		CloudProvider: r.URL.Query().Get("cloud_provider"),
		Region:        r.URL.Query().Get("region"),
		ClusterID:     r.URL.Query().Get("cluster_id"),
		Namespace:     r.URL.Query().Get("namespace"),
		Environment:   r.URL.Query().Get("environment"),
		Runtime:       r.URL.Query().Get("runtime"),
		Tag:           r.URL.Query().Get("tag"),
		SortBy:        r.URL.Query().Get("sort"),
		SortOrder:     r.URL.Query().Get("order"),
	}
	if raw := query.Get("min_replicas"); raw != "" {
		minReplicas, err := strconv.Atoi(raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "min_replicas must be a number")
			return model.MicroserviceFilter{}, false
		}
		filter.MinReplicas = minReplicas
	}
	if raw := query.Get("limit"); raw != "" {
		limit, err := strconv.Atoi(raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "limit must be a number")
			return model.MicroserviceFilter{}, false
		}
		filter.Limit = limit
	}
	if raw := query.Get("after_id"); raw != "" {
		filter.AfterID = raw
	}
	if raw := query.Get("offset"); raw != "" {
		offset, err := strconv.Atoi(raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "offset must be a number")
			return model.MicroserviceFilter{}, false
		}
		filter.Offset = offset
	}
	return filter, true
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
		writeError(w, http.StatusBadRequest, "application_id, name, owner_team, protocol and endpoint are required; protocol must be http, grpc, event or worker; status must be active, degraded, offline or deprecated; environment must be development, staging or production; replicas must be at least 1; slo_target and error_budget_remaining must be between 0 and 100")
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
