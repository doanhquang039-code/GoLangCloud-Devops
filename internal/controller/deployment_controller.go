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

type DeploymentController struct {
	deploymentService *service.DeploymentService
}

func NewDeploymentController(deploymentService *service.DeploymentService) *DeploymentController {
	return &DeploymentController{deploymentService: deploymentService}
}

func (c *DeploymentController) Index(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		c.GetDeployments(w, r)
	case http.MethodPost:
		c.CreateDeployment(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (c *DeploymentController) GetDeployments(w http.ResponseWriter, r *http.Request) {
	deployments, err := c.deploymentService.GetDeployments(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list deployments")
		return
	}

	writeJSON(w, http.StatusOK, deployments)
}

func (c *DeploymentController) ShowOrUpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/deployments/")

	switch r.Method {
	case http.MethodGet:
		c.Show(w, r, id)
	case http.MethodPatch:
		c.UpdateStatus(w, r, id)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (c *DeploymentController) Show(w http.ResponseWriter, r *http.Request, id string) {
	deployment, err := c.deploymentService.GetDeploymentByID(r.Context(), id)
	if errors.Is(err, repository.ErrDeploymentNotFound) {
		writeError(w, http.StatusNotFound, "deployment not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, deployment)
}

func (c *DeploymentController) CreateDeployment(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var request model.CreateDeploymentRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	deployment, err := c.deploymentService.CreateDeployment(r.Context(), request)
	if errors.Is(err, service.ErrInvalidDeployment) {
		writeError(w, http.StatusBadRequest, "application_id, cluster_id, environment, version and requested_by are required; strategy must be rolling, blue-green or canary")
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
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create deployment")
		return
	}

	writeJSON(w, http.StatusCreated, deployment)
}

func (c *DeploymentController) UpdateStatus(w http.ResponseWriter, r *http.Request, id string) {
	defer r.Body.Close()

	var request model.UpdateDeploymentStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	deployment, err := c.deploymentService.UpdateDeploymentStatus(r.Context(), id, request)
	if errors.Is(err, service.ErrInvalidDeployment) {
		writeError(w, http.StatusBadRequest, "status must be running, succeeded or failed")
		return
	}
	if errors.Is(err, repository.ErrDeploymentNotFound) {
		writeError(w, http.StatusNotFound, "deployment not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not update deployment")
		return
	}

	writeJSON(w, http.StatusOK, deployment)
}
