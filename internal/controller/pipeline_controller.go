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

type PipelineController struct {
	pipelineService *service.PipelineService
}

func NewPipelineController(pipelineService *service.PipelineService) *PipelineController {
	return &PipelineController{pipelineService: pipelineService}
}

func (c *PipelineController) Index(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		c.GetPipelineRuns(w, r)
	case http.MethodPost:
		c.CreatePipelineRun(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (c *PipelineController) GetPipelineRuns(w http.ResponseWriter, r *http.Request) {
	pipelineRuns, err := c.pipelineService.GetPipelineRuns(r.Context(), model.PipelineRunFilter{
		Query:         r.URL.Query().Get("q"),
		ApplicationID: r.URL.Query().Get("application_id"),
		Branch:        r.URL.Query().Get("branch"),
		Status:        r.URL.Query().Get("status"),
		TriggeredBy:   r.URL.Query().Get("triggered_by"),
	})
	if errors.Is(err, service.ErrInvalidPipelineRun) {
		writeError(w, http.StatusBadRequest, "status must be running, succeeded, failed or cancelled")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list pipeline runs")
		return
	}

	writeJSON(w, http.StatusOK, pipelineRuns)
}

func (c *PipelineController) ShowOrUpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/pipelines/")
	if pipelineID, stageName, ok := strings.Cut(id, "/stages/"); ok {
		if r.Method == http.MethodPatch {
			c.UpdateStageStatus(w, r, pipelineID, stageName)
			return
		}
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	switch r.Method {
	case http.MethodGet:
		c.Show(w, r, id)
	case http.MethodPatch:
		c.UpdateStatus(w, r, id)
	case http.MethodDelete:
		c.DeletePipelineRun(w, r, id)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (c *PipelineController) Show(w http.ResponseWriter, r *http.Request, id string) {
	pipelineRun, err := c.pipelineService.GetPipelineRunByID(r.Context(), id)
	if errors.Is(err, repository.ErrPipelineRunNotFound) {
		writeError(w, http.StatusNotFound, "pipeline run not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, pipelineRun)
}

func (c *PipelineController) UpdateStageStatus(w http.ResponseWriter, r *http.Request, id string, stageName string) {
	defer r.Body.Close()

	var request model.UpdatePipelineStageStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	pipelineRun, err := c.pipelineService.UpdatePipelineStageStatus(r.Context(), id, stageName, request)
	if errors.Is(err, service.ErrInvalidPipelineRun) {
		writeError(w, http.StatusBadRequest, "stage status must be pending, running, succeeded, failed or skipped")
		return
	}
	if errors.Is(err, repository.ErrPipelineRunNotFound) {
		writeError(w, http.StatusNotFound, "pipeline run not found")
		return
	}
	if errors.Is(err, service.ErrPipelineStageNotFound) {
		writeError(w, http.StatusNotFound, "pipeline stage not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not update pipeline stage")
		return
	}

	writeJSON(w, http.StatusOK, pipelineRun)
}

func (c *PipelineController) DeletePipelineRun(w http.ResponseWriter, r *http.Request, id string) {
	err := c.pipelineService.DeletePipelineRun(r.Context(), id)
	if errors.Is(err, service.ErrInvalidPipelineRun) {
		writeError(w, http.StatusBadRequest, "pipeline run id is required")
		return
	}
	if errors.Is(err, repository.ErrPipelineRunNotFound) {
		writeError(w, http.StatusNotFound, "pipeline run not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not delete pipeline run")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *PipelineController) CreatePipelineRun(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var request model.CreatePipelineRunRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	pipelineRun, err := c.pipelineService.CreatePipelineRun(r.Context(), request)
	if errors.Is(err, service.ErrInvalidPipelineRun) {
		writeError(w, http.StatusBadRequest, "application_id, branch, commit_sha and triggered_by are required")
		return
	}
	if errors.Is(err, repository.ErrApplicationNotFound) {
		writeError(w, http.StatusNotFound, "application not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create pipeline run")
		return
	}

	writeJSON(w, http.StatusCreated, pipelineRun)
}

func (c *PipelineController) UpdateStatus(w http.ResponseWriter, r *http.Request, id string) {
	defer r.Body.Close()

	var request model.UpdatePipelineRunStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	pipelineRun, err := c.pipelineService.UpdatePipelineRunStatus(r.Context(), id, request)
	if errors.Is(err, service.ErrInvalidPipelineRun) {
		writeError(w, http.StatusBadRequest, "status must be running, succeeded, failed or cancelled")
		return
	}
	if errors.Is(err, repository.ErrPipelineRunNotFound) {
		writeError(w, http.StatusNotFound, "pipeline run not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not update pipeline run")
		return
	}

	writeJSON(w, http.StatusOK, pipelineRun)
}
