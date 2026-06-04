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

type IncidentController struct {
	incidentService *service.IncidentService
}

func NewIncidentController(incidentService *service.IncidentService) *IncidentController {
	return &IncidentController{incidentService: incidentService}
}

func (c *IncidentController) Index(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		c.GetIncidents(w, r)
	case http.MethodPost:
		c.CreateIncident(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (c *IncidentController) GetIncidents(w http.ResponseWriter, r *http.Request) {
	incidents, err := c.incidentService.GetIncidents(r.Context(), model.IncidentFilter{
		Query:         r.URL.Query().Get("q"),
		ApplicationID: r.URL.Query().Get("application_id"),
		ClusterID:     r.URL.Query().Get("cluster_id"),
		DeploymentID:  r.URL.Query().Get("deployment_id"),
		Severity:      r.URL.Query().Get("severity"),
		Status:        r.URL.Query().Get("status"),
		OwnerTeam:     r.URL.Query().Get("owner_team"),
	})
	if errors.Is(err, service.ErrInvalidIncident) {
		writeError(w, http.StatusBadRequest, "severity must be sev1, sev2, sev3 or sev4; status must be open, investigating, mitigated or resolved")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list incidents")
		return
	}

	writeJSON(w, http.StatusOK, incidents)
}

func (c *IncidentController) ShowOrUpdate(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/incidents/")

	switch r.Method {
	case http.MethodGet:
		c.Show(w, r, id)
	case http.MethodPut:
		c.UpdateIncident(w, r, id)
	case http.MethodPatch:
		c.UpdateStatus(w, r, id)
	case http.MethodDelete:
		c.DeleteIncident(w, r, id)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (c *IncidentController) Show(w http.ResponseWriter, r *http.Request, id string) {
	incident, err := c.incidentService.GetIncidentByID(r.Context(), id)
	if errors.Is(err, repository.ErrIncidentNotFound) {
		writeError(w, http.StatusNotFound, "incident not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, incident)
}

func (c *IncidentController) DeleteIncident(w http.ResponseWriter, r *http.Request, id string) {
	err := c.incidentService.DeleteIncident(r.Context(), id)
	c.writeIncidentError(w, err, "could not delete incident")
	if err != nil {
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *IncidentController) CreateIncident(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var request model.CreateIncidentRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	incident, err := c.incidentService.CreateIncident(r.Context(), request)
	c.writeIncidentError(w, err, "could not create incident")
	if err != nil {
		return
	}

	writeJSON(w, http.StatusCreated, incident)
}

func (c *IncidentController) UpdateIncident(w http.ResponseWriter, r *http.Request, id string) {
	defer r.Body.Close()

	var request model.UpdateIncidentRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	incident, err := c.incidentService.UpdateIncident(r.Context(), id, request)
	c.writeIncidentError(w, err, "could not update incident")
	if err != nil {
		return
	}

	writeJSON(w, http.StatusOK, incident)
}

func (c *IncidentController) UpdateStatus(w http.ResponseWriter, r *http.Request, id string) {
	defer r.Body.Close()

	var request model.UpdateIncidentStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	incident, err := c.incidentService.UpdateIncidentStatus(r.Context(), id, request)
	c.writeIncidentError(w, err, "could not update incident")
	if err != nil {
		return
	}

	writeJSON(w, http.StatusOK, incident)
}

func (c *IncidentController) writeIncidentError(w http.ResponseWriter, err error, fallback string) {
	if err == nil {
		return
	}
	if errors.Is(err, service.ErrInvalidIncident) {
		writeError(w, http.StatusBadRequest, "title, summary, severity and owner_team are required; severity must be sev1, sev2, sev3 or sev4; status must be open, investigating, mitigated or resolved")
		return
	}
	if errors.Is(err, repository.ErrIncidentNotFound) {
		writeError(w, http.StatusNotFound, "incident not found")
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
	if errors.Is(err, repository.ErrDeploymentNotFound) {
		writeError(w, http.StatusNotFound, "deployment not found")
		return
	}
	writeError(w, http.StatusInternalServerError, fallback)
}
