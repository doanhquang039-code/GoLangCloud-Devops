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
		ApplicationID: "app-payroll-api",
		Name:          " payroll-api ",
		OwnerTeam:     " platform ",
		Protocol:      "HTTP",
		Endpoint:      " http://payroll-api:8080 ",
		Status:        "ACTIVE",
		Dependencies:  []string{" mongodb ", "", "mongodb", "audit-events"},
		Tags:          []string{" backend ", "backend", "payroll"},
	})
	if err != nil {
		t.Fatal(err)
	}

	if microservice.Protocol != "http" || microservice.Status != "active" {
		t.Fatalf("expected normalized protocol/status, got %q/%q", microservice.Protocol, microservice.Status)
	}
	if len(microservice.Dependencies) != 2 {
		t.Fatalf("expected normalized unique dependencies, got %#v", microservice.Dependencies)
	}
	if len(microservice.Tags) != 2 {
		t.Fatalf("expected normalized unique tags, got %#v", microservice.Tags)
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
		ApplicationID: "app-payroll-api",
		Name:          "payroll-api",
		OwnerTeam:     "Platform",
		Protocol:      "HTTP",
		Endpoint:      "http://payroll-api:8080",
		Status:        "Active",
		Dependencies:  []string{"mongodb"},
		Tags:          []string{"backend"},
	}); err != nil {
		t.Fatal(err)
	}

	filtered, err := microserviceService.GetMicroservices(ctx, model.MicroserviceFilter{
		Query:     "mongo",
		OwnerTeam: "platform",
		Protocol:  "http",
		Status:    "ACTIVE",
		Tag:       "BACKEND",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(filtered) != 1 || filtered[0].ID != "svc-payroll-api" {
		t.Fatalf("expected payroll microservice, got %#v", filtered)
	}
}
