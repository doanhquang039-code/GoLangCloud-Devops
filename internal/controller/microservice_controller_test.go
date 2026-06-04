package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"hr-cloud-service/internal/model"
	"hr-cloud-service/internal/repository"
	"hr-cloud-service/internal/service"
)

func TestMicroserviceControllerCreateUpdateAndPatchStatus(t *testing.T) {
	ctx := t.Context()
	applicationRepository := repository.NewInMemoryApplicationRepository()
	microserviceRepository := repository.NewInMemoryMicroserviceRepository()
	microserviceController := NewMicroserviceController(service.NewMicroserviceService(applicationRepository, microserviceRepository))

	if _, err := applicationRepository.Save(ctx, model.Application{ID: "app-payroll-api"}); err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/v1/microservices", bytes.NewBufferString(`{
		"application_id": "app-payroll-api",
		"name": "payroll-api",
		"owner_team": "platform",
		"protocol": "HTTP",
		"endpoint": "http://payroll-api:8080"
	}`))
	response := httptest.NewRecorder()
	microserviceController.Index(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}

	var microservice model.Microservice
	if err := json.NewDecoder(response.Body).Decode(&microservice); err != nil {
		t.Fatal(err)
	}

	updateRequest := httptest.NewRequest(http.MethodPut, "/api/v1/microservices/"+microservice.ID, bytes.NewBufferString(`{
		"application_id": "app-payroll-api",
		"name": "payroll-worker",
		"owner_team": "platform",
		"protocol": "WORKER",
		"endpoint": "queue://payroll-jobs",
		"dependencies": ["mongodb", "payroll-events"],
		"config": {
			"CONCURRENCY": "4"
		},
		"tags": ["backend", "payroll"]
	}`))
	updateResponse := httptest.NewRecorder()
	microserviceController.ShowOrUpdate(updateResponse, updateRequest)
	if updateResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, updateResponse.Code, updateResponse.Body.String())
	}

	var updated model.Microservice
	if err := json.NewDecoder(updateResponse.Body).Decode(&updated); err != nil {
		t.Fatal(err)
	}
	if updated.Name != "payroll-worker" || updated.Protocol != "worker" || updated.Endpoint != "queue://payroll-jobs" {
		t.Fatalf("expected updated worker microservice, got %#v", updated)
	}

	patchRequest := httptest.NewRequest(http.MethodPatch, "/api/v1/microservices/"+microservice.ID, bytes.NewBufferString(`{"status":"DEGRADED"}`))
	patchResponse := httptest.NewRecorder()
	microserviceController.ShowOrUpdate(patchResponse, patchRequest)
	if patchResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, patchResponse.Code, patchResponse.Body.String())
	}

	var patched model.Microservice
	if err := json.NewDecoder(patchResponse.Body).Decode(&patched); err != nil {
		t.Fatal(err)
	}
	if patched.Status != "degraded" {
		t.Fatalf("expected degraded status, got %q", patched.Status)
	}

	listRequest := httptest.NewRequest(http.MethodGet, "/api/v1/microservices?q=payroll-events&status=degraded", nil)
	listResponse := httptest.NewRecorder()
	microserviceController.Index(listResponse, listRequest)
	if listResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, listResponse.Code, listResponse.Body.String())
	}

	var listed []model.Microservice
	if err := json.NewDecoder(listResponse.Body).Decode(&listed); err != nil {
		t.Fatal(err)
	}
	if len(listed) != 1 || listed[0].ID != microservice.ID {
		t.Fatalf("expected filtered microservice %q, got %#v", microservice.ID, listed)
	}
}
