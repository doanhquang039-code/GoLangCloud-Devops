package service

import (
	"context"
	"testing"

	"hr-cloud-service/internal/model"
	"hr-cloud-service/internal/repository"
)

func TestGetClustersFiltersByQueryProviderRegionAndStatus(t *testing.T) {
	ctx := context.Background()
	clusterRepository := repository.NewInMemoryClusterRepository()
	clusterService := NewClusterService(clusterRepository)

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

	filtered, err := clusterService.GetClusters(ctx, model.ClusterFilter{
		Query:    "staging",
		Provider: "aws",
		Region:   "AP-SOUTHEAST-1",
		Status:   "READY",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(filtered) != 1 || filtered[0].ID != "cls-staging-sg" {
		t.Fatalf("expected filtered cluster cls-staging-sg, got %#v", filtered)
	}
}
