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

func TestClusterControllerFiltersClustersByQuery(t *testing.T) {
	ctx := t.Context()
	clusterRepository := repository.NewInMemoryClusterRepository()
	clusterController := NewClusterController(service.NewClusterService(clusterRepository))

	if _, err := clusterRepository.Save(ctx, model.Cluster{
		ID:       "cls-staging-sg",
		Name:     "eks-staging-ap-southeast-1",
		Provider: "AWS",
		Region:   "ap-southeast-1",
		Endpoint: "https://staging.example.eks.amazonaws.com",
		Version:  "1.31",
		Status:   "Ready",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := clusterRepository.Save(ctx, model.Cluster{
		ID:       "cls-prod-us",
		Name:     "gke-prod-us",
		Provider: "gcp",
		Region:   "us-central1",
		Endpoint: "https://prod.example.gke.googleapis.com",
		Version:  "1.30",
		Status:   "maintenance",
	}); err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/clusters?q=staging&provider=aws&status=ready", nil)
	response := httptest.NewRecorder()

	clusterController.Index(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var clusters []model.Cluster
	if err := json.NewDecoder(response.Body).Decode(&clusters); err != nil {
		t.Fatal(err)
	}
	if len(clusters) != 1 || clusters[0].ID != "cls-staging-sg" {
		t.Fatalf("expected filtered cluster cls-staging-sg, got %#v", clusters)
	}
}
