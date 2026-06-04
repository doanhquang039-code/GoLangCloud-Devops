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

type ClusterController struct {
	clusterService *service.ClusterService
}

func NewClusterController(clusterService *service.ClusterService) *ClusterController {
	return &ClusterController{clusterService: clusterService}
}

func (c *ClusterController) Index(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		c.GetClusters(w, r)
	case http.MethodPost:
		c.CreateCluster(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (c *ClusterController) GetClusters(w http.ResponseWriter, r *http.Request) {
	clusters, err := c.clusterService.GetClusters(r.Context(), model.ClusterFilter{
		Query:    r.URL.Query().Get("q"),
		Provider: r.URL.Query().Get("provider"),
		Region:   r.URL.Query().Get("region"),
		Status:   r.URL.Query().Get("status"),
	})
	if errors.Is(err, service.ErrInvalidCluster) {
		writeError(w, http.StatusBadRequest, "status must be ready, degraded, maintenance or offline")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list clusters")
		return
	}

	writeJSON(w, http.StatusOK, clusters)
}

func (c *ClusterController) ShowOrUpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/clusters/")

	switch r.Method {
	case http.MethodGet:
		c.Show(w, r, id)
	case http.MethodPut:
		c.UpdateCluster(w, r, id)
	case http.MethodPatch:
		c.UpdateStatus(w, r, id)
	case http.MethodDelete:
		c.DeleteCluster(w, r, id)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (c *ClusterController) Show(w http.ResponseWriter, r *http.Request, id string) {
	cluster, err := c.clusterService.GetClusterByID(r.Context(), id)
	if errors.Is(err, repository.ErrClusterNotFound) {
		writeError(w, http.StatusNotFound, "cluster not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, cluster)
}

func (c *ClusterController) DeleteCluster(w http.ResponseWriter, r *http.Request, id string) {
	err := c.clusterService.DeleteCluster(r.Context(), id)
	if errors.Is(err, service.ErrInvalidCluster) {
		writeError(w, http.StatusBadRequest, "cluster id is required")
		return
	}
	if errors.Is(err, repository.ErrClusterNotFound) {
		writeError(w, http.StatusNotFound, "cluster not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not delete cluster")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *ClusterController) CreateCluster(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var request model.CreateClusterRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	cluster, err := c.clusterService.CreateCluster(r.Context(), request)
	if errors.Is(err, service.ErrInvalidCluster) {
		writeError(w, http.StatusBadRequest, "name, provider, region, endpoint and version are required")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create cluster")
		return
	}

	writeJSON(w, http.StatusCreated, cluster)
}

func (c *ClusterController) UpdateCluster(w http.ResponseWriter, r *http.Request, id string) {
	defer r.Body.Close()

	var request model.UpdateClusterRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	cluster, err := c.clusterService.UpdateCluster(r.Context(), id, request)
	if errors.Is(err, service.ErrInvalidCluster) {
		writeError(w, http.StatusBadRequest, "name, provider, region, endpoint and version are required; status must be ready, degraded, maintenance or offline")
		return
	}
	if errors.Is(err, repository.ErrClusterNotFound) {
		writeError(w, http.StatusNotFound, "cluster not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not update cluster")
		return
	}

	writeJSON(w, http.StatusOK, cluster)
}

func (c *ClusterController) UpdateStatus(w http.ResponseWriter, r *http.Request, id string) {
	defer r.Body.Close()

	var request model.UpdateClusterStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	cluster, err := c.clusterService.UpdateClusterStatus(r.Context(), id, request)
	if errors.Is(err, service.ErrInvalidCluster) {
		writeError(w, http.StatusBadRequest, "status must be ready, degraded, maintenance or offline")
		return
	}
	if errors.Is(err, repository.ErrClusterNotFound) {
		writeError(w, http.StatusNotFound, "cluster not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not update cluster")
		return
	}

	writeJSON(w, http.StatusOK, cluster)
}
