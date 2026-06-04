package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"hr-cloud-service/internal/controller"
	"hr-cloud-service/internal/model"
	"hr-cloud-service/internal/repository"
	"hr-cloud-service/internal/service"
)

func TestRouterApplicationPipelineFlow(t *testing.T) {
	router := newTestRouter()

	appBody := bytes.NewBufferString(`{
		"name": "payroll-api",
		"repository": "git@example.com/payroll-api.git",
		"runtime": "go1.22",
		"owner_team": "platform",
		"tags": ["backend"]
	}`)
	appResponse := httptest.NewRecorder()
	router.ServeHTTP(appResponse, httptest.NewRequest(http.MethodPost, "/api/v1/applications", appBody))
	if appResponse.Code != http.StatusCreated {
		t.Fatalf("expected application create status %d, got %d: %s", http.StatusCreated, appResponse.Code, appResponse.Body.String())
	}

	var application model.Application
	if err := json.NewDecoder(appResponse.Body).Decode(&application); err != nil {
		t.Fatal(err)
	}

	pipelineBody := bytes.NewBufferString(`{
		"application_id": "` + application.ID + `",
		"branch": "main",
		"commit_sha": "abc123",
		"triggered_by": "devops@example.com",
		"stages": ["build", "unit-test"]
	}`)
	pipelineResponse := httptest.NewRecorder()
	router.ServeHTTP(pipelineResponse, httptest.NewRequest(http.MethodPost, "/api/v1/pipelines", pipelineBody))
	if pipelineResponse.Code != http.StatusCreated {
		t.Fatalf("expected pipeline create status %d, got %d: %s", http.StatusCreated, pipelineResponse.Code, pipelineResponse.Body.String())
	}

	var pipelineRun model.PipelineRun
	if err := json.NewDecoder(pipelineResponse.Body).Decode(&pipelineRun); err != nil {
		t.Fatal(err)
	}

	stageResponse := httptest.NewRecorder()
	router.ServeHTTP(stageResponse, httptest.NewRequest(http.MethodPatch, "/api/v1/pipelines/"+pipelineRun.ID+"/stages/build", bytes.NewBufferString(`{"status":"SUCCEEDED"}`)))
	if stageResponse.Code != http.StatusOK {
		t.Fatalf("expected stage update status %d, got %d: %s", http.StatusOK, stageResponse.Code, stageResponse.Body.String())
	}

	var updatedRun model.PipelineRun
	if err := json.NewDecoder(stageResponse.Body).Decode(&updatedRun); err != nil {
		t.Fatal(err)
	}
	if updatedRun.Stages[0].Status != "succeeded" {
		t.Fatalf("expected normalized stage status, got %q", updatedRun.Stages[0].Status)
	}

	appListResponse := httptest.NewRecorder()
	router.ServeHTTP(appListResponse, httptest.NewRequest(http.MethodGet, "/api/v1/applications?owner_team=PLATFORM&tag=BACKEND", nil))
	if appListResponse.Code != http.StatusOK {
		t.Fatalf("expected application list status %d, got %d: %s", http.StatusOK, appListResponse.Code, appListResponse.Body.String())
	}

	var applications []model.Application
	if err := json.NewDecoder(appListResponse.Body).Decode(&applications); err != nil {
		t.Fatal(err)
	}
	if len(applications) != 1 || applications[0].ID != application.ID {
		t.Fatalf("expected filtered application %q, got %#v", application.ID, applications)
	}

	pipelineListResponse := httptest.NewRecorder()
	router.ServeHTTP(pipelineListResponse, httptest.NewRequest(http.MethodGet, "/api/v1/pipelines?branch=MAIN&status=RUNNING&triggered_by=DEVOPS@example.com", nil))
	if pipelineListResponse.Code != http.StatusOK {
		t.Fatalf("expected pipeline list status %d, got %d: %s", http.StatusOK, pipelineListResponse.Code, pipelineListResponse.Body.String())
	}

	var pipelineRuns []model.PipelineRun
	if err := json.NewDecoder(pipelineListResponse.Body).Decode(&pipelineRuns); err != nil {
		t.Fatal(err)
	}
	if len(pipelineRuns) != 1 || pipelineRuns[0].ID != pipelineRun.ID {
		t.Fatalf("expected filtered pipeline %q, got %#v", pipelineRun.ID, pipelineRuns)
	}

	serviceBody := bytes.NewBufferString(`{
		"application_id": "` + application.ID + `",
		"name": "payroll-api",
		"owner_team": "platform",
		"protocol": "HTTP",
		"endpoint": "http://payroll-api:8080",
		"tags": ["backend"]
	}`)
	serviceResponse := httptest.NewRecorder()
	router.ServeHTTP(serviceResponse, httptest.NewRequest(http.MethodPost, "/api/v1/microservices", serviceBody))
	if serviceResponse.Code != http.StatusCreated {
		t.Fatalf("expected microservice create status %d, got %d: %s", http.StatusCreated, serviceResponse.Code, serviceResponse.Body.String())
	}

	var microservice model.Microservice
	if err := json.NewDecoder(serviceResponse.Body).Decode(&microservice); err != nil {
		t.Fatal(err)
	}
	if microservice.Protocol != "http" || microservice.Status != "active" {
		t.Fatalf("expected normalized microservice protocol/status, got %q/%q", microservice.Protocol, microservice.Status)
	}

	serviceListResponse := httptest.NewRecorder()
	router.ServeHTTP(serviceListResponse, httptest.NewRequest(http.MethodGet, "/api/v1/microservices?owner_team=PLATFORM&protocol=HTTP&status=ACTIVE&tag=BACKEND", nil))
	if serviceListResponse.Code != http.StatusOK {
		t.Fatalf("expected microservice list status %d, got %d: %s", http.StatusOK, serviceListResponse.Code, serviceListResponse.Body.String())
	}

	var microservices []model.Microservice
	if err := json.NewDecoder(serviceListResponse.Body).Decode(&microservices); err != nil {
		t.Fatal(err)
	}
	if len(microservices) != 1 || microservices[0].ID != microservice.ID {
		t.Fatalf("expected filtered microservice %q, got %#v", microservice.ID, microservices)
	}
}

func newTestRouter() http.Handler {
	employeeRepository := repository.NewInMemoryEmployeeRepository()
	applicationRepository := repository.NewInMemoryApplicationRepository()
	clusterRepository := repository.NewInMemoryClusterRepository()
	environmentRepository := repository.NewInMemoryEnvironmentRepository()
	deploymentRepository := repository.NewInMemoryDeploymentRepository()
	pipelineRepository := repository.NewInMemoryPipelineRepository()
	microserviceRepository := repository.NewInMemoryMicroserviceRepository()
	incidentRepository := repository.NewInMemoryIncidentRepository()

	employeeService := service.NewEmployeeService(employeeRepository)
	applicationService := service.NewApplicationService(applicationRepository)
	clusterService := service.NewClusterService(clusterRepository)
	environmentService := service.NewEnvironmentService(applicationRepository, clusterRepository, environmentRepository)
	deploymentService := service.NewDeploymentService(applicationRepository, clusterRepository, deploymentRepository)
	pipelineService := service.NewPipelineService(applicationRepository, pipelineRepository)
	microserviceService := service.NewMicroserviceService(applicationRepository, microserviceRepository)
	incidentService := service.NewIncidentService(applicationRepository, clusterRepository, deploymentRepository, incidentRepository)
	platformService := service.NewPlatformService(applicationRepository, clusterRepository, environmentRepository, deploymentRepository, pipelineRepository, incidentRepository)

	return NewRouter(
		controller.NewHealthController(nil),
		controller.NewEmployeeController(employeeService),
		controller.NewApplicationController(applicationService),
		controller.NewClusterController(clusterService),
		controller.NewEnvironmentController(environmentService),
		controller.NewDeploymentController(deploymentService),
		controller.NewPipelineController(pipelineService),
		controller.NewMicroserviceController(microserviceService),
		controller.NewIncidentController(incidentService),
		controller.NewPlatformController(platformService),
	)
}
