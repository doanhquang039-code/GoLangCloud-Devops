package service

import (
	"context"
	"errors"
	"testing"

	"hr-cloud-service/internal/model"
	"hr-cloud-service/internal/repository"
)

func TestCreateMicroserviceNormalizesAndValidatesApplication(t *testing.T) {
	ctx := context.Background()
	applicationRepository := repository.NewInMemoryApplicationRepository()
	microserviceRepository := repository.NewInMemoryMicroserviceRepository()
	microserviceService := NewMicroserviceService(applicationRepository, microserviceRepository)

	if _, err := applicationRepository.Save(ctx, model.Application{ID: "app-payroll-api"}); err != nil {
		t.Fatal(err)
	}

	microservice, err := microserviceService.CreateMicroservice(ctx, model.CreateMicroserviceRequest{
		TenantID:             " tenant-hr ",
		ApplicationID:        "app-payroll-api",
		Name:                 " payroll-api ",
		OwnerTeam:            " platform ",
		Protocol:             "HTTP",
		Endpoint:             " http://payroll-api:8080 ",
		Status:               "ACTIVE",
		CloudProvider:        "AWS",
		Region:               " ap-southeast-1 ",
		ClusterID:            "cls-eks-staging",
		Namespace:            " hr-staging ",
		Environment:          "STAGING",
		Runtime:              "go1.22",
		Image:                " ghcr.io/company/payroll-api:v1.2.3 ",
		Version:              " v1.2.3 ",
		Replicas:             3,
		CPURequest:           "250m",
		MemoryRequest:        "512Mi",
		HealthPath:           "/readyz",
		SLOTarget:            99.95,
		ErrorBudgetRemaining: 82.5,
		Dependencies:         []string{" mongodb ", "", "mongodb", "audit-events"},
		Tags:                 []string{" backend ", "backend", "payroll"},
	})
	if err != nil {
		t.Fatal(err)
	}

	if microservice.Protocol != "http" || microservice.Status != "active" {
		t.Fatalf("expected normalized protocol/status, got %q/%q", microservice.Protocol, microservice.Status)
	}
	if microservice.TenantID != "tenant-hr" {
		t.Fatalf("expected normalized tenant id, got %q", microservice.TenantID)
	}
	if len(microservice.Dependencies) != 2 {
		t.Fatalf("expected normalized unique dependencies, got %#v", microservice.Dependencies)
	}
	if len(microservice.Tags) != 2 {
		t.Fatalf("expected normalized unique tags, got %#v", microservice.Tags)
	}
	if microservice.CloudProvider != "aws" || microservice.Environment != "staging" || microservice.Replicas != 3 {
		t.Fatalf("expected normalized cloud metadata, got provider=%q environment=%q replicas=%d", microservice.CloudProvider, microservice.Environment, microservice.Replicas)
	}
	if microservice.SLOTarget != 99.95 || microservice.ErrorBudgetRemaining != 82.5 {
		t.Fatalf("expected SLO metadata, got target=%.2f budget=%.2f", microservice.SLOTarget, microservice.ErrorBudgetRemaining)
	}
}

func TestCreateMicroserviceRequiresExistingApplication(t *testing.T) {
	ctx := context.Background()
	microserviceService := NewMicroserviceService(
		repository.NewInMemoryApplicationRepository(),
		repository.NewInMemoryMicroserviceRepository(),
	)

	_, err := microserviceService.CreateMicroservice(ctx, model.CreateMicroserviceRequest{
		ApplicationID: "app-missing",
		Name:          "payroll-api",
		OwnerTeam:     "platform",
		Protocol:      "http",
		Endpoint:      "http://payroll-api:8080",
	})
	if !errors.Is(err, repository.ErrApplicationNotFound) {
		t.Fatalf("expected application not found, got %v", err)
	}
}

func TestGetMicroservicesFiltersCaseInsensitively(t *testing.T) {
	ctx := context.Background()
	microserviceRepository := repository.NewInMemoryMicroserviceRepository()
	microserviceService := NewMicroserviceService(repository.NewInMemoryApplicationRepository(), microserviceRepository)

	if _, err := microserviceRepository.Save(ctx, model.Microservice{
		ID:            "svc-payroll-api",
		TenantID:      "tenant-hr",
		ApplicationID: "app-payroll-api",
		Name:          "payroll-api",
		OwnerTeam:     "Platform",
		Protocol:      "HTTP",
		Endpoint:      "http://payroll-api:8080",
		Status:        "Active",
		CloudProvider: "AWS",
		Region:        "ap-southeast-1",
		ClusterID:     "cls-eks-staging",
		Namespace:     "HR-Staging",
		Environment:   "Staging",
		Runtime:       "go1.22",
		Image:         "ghcr.io/company/payroll-api:v1.2.3",
		Version:       "v1.2.3",
		Replicas:      3,
		Dependencies:  []string{"mongodb"},
		Tags:          []string{"backend"},
	}); err != nil {
		t.Fatal(err)
	}

	filtered, err := microserviceService.GetMicroservices(ctx, model.MicroserviceFilter{
		Query:         "mongo",
		TenantID:      "tenant-hr",
		OwnerTeam:     "platform",
		Protocol:      "http",
		Status:        "ACTIVE",
		CloudProvider: "aws",
		Region:        "AP-SOUTHEAST-1",
		ClusterID:     "cls-eks-staging",
		Namespace:     "hr-staging",
		Environment:   "STAGING",
		Runtime:       "GO1.22",
		Tag:           "BACKEND",
		MinReplicas:   2,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(filtered) != 1 || filtered[0].ID != "svc-payroll-api" {
		t.Fatalf("expected payroll microservice, got %#v", filtered)
	}
}

func TestGetMicroservicesPaginatesAndSortsAtRepositoryLevel(t *testing.T) {
	ctx := context.Background()
	microserviceRepository := repository.NewInMemoryMicroserviceRepository()
	microserviceService := NewMicroserviceService(repository.NewInMemoryApplicationRepository(), microserviceRepository)

	services := []model.Microservice{
		{ID: "svc-a", TenantID: "tenant-hr", Name: "alpha", Status: "active", Replicas: 1},
		{ID: "svc-b", TenantID: "tenant-hr", Name: "bravo", Status: "active", Replicas: 5},
		{ID: "svc-c", TenantID: "tenant-hr", Name: "charlie", Status: "active", Replicas: 3},
		{ID: "svc-z", TenantID: "tenant-finance", Name: "zulu", Status: "active", Replicas: 9},
	}
	for _, service := range services {
		if _, err := microserviceRepository.Save(ctx, service); err != nil {
			t.Fatal(err)
		}
	}

	filtered, err := microserviceService.GetMicroservices(ctx, model.MicroserviceFilter{
		TenantID:  "tenant-hr",
		Status:    "active",
		SortBy:    "replicas",
		SortOrder: "desc",
		Limit:     2,
		Offset:    1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(filtered) != 2 {
		t.Fatalf("expected 2 paginated services, got %d", len(filtered))
	}
	if filtered[0].ID != "svc-c" || filtered[1].ID != "svc-a" {
		t.Fatalf("expected second page sorted by replicas desc, got %#v", filtered)
	}
}

func TestGetMicroservicesUsesTenantAndCursorPagination(t *testing.T) {
	ctx := context.Background()
	microserviceRepository := repository.NewInMemoryMicroserviceRepository()
	microserviceService := NewMicroserviceService(repository.NewInMemoryApplicationRepository(), microserviceRepository)

	services := []model.Microservice{
		{ID: "svc-a", TenantID: "tenant-hr", Status: "active"},
		{ID: "svc-b", TenantID: "tenant-hr", Status: "active"},
		{ID: "svc-c", TenantID: "tenant-hr", Status: "active"},
		{ID: "svc-d", TenantID: "tenant-finance", Status: "active"},
	}
	for _, service := range services {
		if _, err := microserviceRepository.Save(ctx, service); err != nil {
			t.Fatal(err)
		}
	}

	filtered, err := microserviceService.GetMicroservices(ctx, model.MicroserviceFilter{
		TenantID:  "tenant-hr",
		Status:    "active",
		AfterID:   "svc-a",
		Limit:     2,
		SortBy:    "id",
		SortOrder: "asc",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(filtered) != 2 || filtered[0].ID != "svc-b" || filtered[1].ID != "svc-c" {
		t.Fatalf("expected tenant cursor page svc-b/svc-c, got %#v", filtered)
	}
}

func TestGetMicroservicesRejectsUnsafeCursorCombination(t *testing.T) {
	microserviceService := NewMicroserviceService(repository.NewInMemoryApplicationRepository(), repository.NewInMemoryMicroserviceRepository())

	_, err := microserviceService.GetMicroservices(context.Background(), model.MicroserviceFilter{
		AfterID: "svc-a",
		Offset:  10,
		Limit:   100,
	})
	if !errors.Is(err, ErrInvalidMicroservice) {
		t.Fatalf("expected invalid cursor and offset combination, got %v", err)
	}
}
