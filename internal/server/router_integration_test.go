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

func TestRouterCloudAccountFlow(t *testing.T) {
	router := newTestRouter()

	createBody := bytes.NewBufferString(`{
		"name": "hr-prod-aws",
		"provider": "aws",
		"account_id": "123456789012",
		"region": "ap-southeast-1",
		"owner_team": "platform",
		"environment": "production",
		"monthly_cost_usd": 1250,
		"budget_usd": 2000,
		"compliance_score": 90,
		"backup_status": "protected",
		"tags": ["hr", "prod"]
	}`)
	createResponse := httptest.NewRecorder()
	router.ServeHTTP(createResponse, httptest.NewRequest(http.MethodPost, "/api/v1/cloud-accounts", createBody))
	if createResponse.Code != http.StatusCreated {
		t.Fatalf("expected cloud account create status %d, got %d: %s", http.StatusCreated, createResponse.Code, createResponse.Body.String())
	}

	var account model.CloudAccount
	if err := json.NewDecoder(createResponse.Body).Decode(&account); err != nil {
		t.Fatal(err)
	}

	patchResponse := httptest.NewRecorder()
	router.ServeHTTP(patchResponse, httptest.NewRequest(http.MethodPatch, "/api/v1/cloud-accounts/"+account.ID, bytes.NewBufferString(`{"status":"restricted","backup_status":"partial"}`)))
	if patchResponse.Code != http.StatusOK {
		t.Fatalf("expected cloud account patch status %d, got %d: %s", http.StatusOK, patchResponse.Code, patchResponse.Body.String())
	}

	listResponse := httptest.NewRecorder()
	router.ServeHTTP(listResponse, httptest.NewRequest(http.MethodGet, "/api/v1/cloud-accounts?provider=aws&backup_status=partial&tag=prod", nil))
	if listResponse.Code != http.StatusOK {
		t.Fatalf("expected cloud account list status %d, got %d: %s", http.StatusOK, listResponse.Code, listResponse.Body.String())
	}

	var accounts []model.CloudAccount
	if err := json.NewDecoder(listResponse.Body).Decode(&accounts); err != nil {
		t.Fatal(err)
	}
	if len(accounts) != 1 || accounts[0].ID != account.ID {
		t.Fatalf("expected filtered cloud account %q, got %#v", account.ID, accounts)
	}

	summaryResponse := httptest.NewRecorder()
	router.ServeHTTP(summaryResponse, httptest.NewRequest(http.MethodGet, "/api/v1/cloud/summary", nil))
	if summaryResponse.Code != http.StatusOK {
		t.Fatalf("expected cloud summary status %d, got %d: %s", http.StatusOK, summaryResponse.Code, summaryResponse.Body.String())
	}

	violationsResponse := httptest.NewRecorder()
	router.ServeHTTP(violationsResponse, httptest.NewRequest(http.MethodGet, "/api/v1/cloud/policy-violations?provider=aws", nil))
	if violationsResponse.Code != http.StatusOK {
		t.Fatalf("expected cloud policy violations status %d, got %d: %s", http.StatusOK, violationsResponse.Code, violationsResponse.Body.String())
	}

	planResponse := httptest.NewRecorder()
	router.ServeHTTP(planResponse, httptest.NewRequest(http.MethodGet, "/api/v1/cloud/remediation-plan", nil))
	if planResponse.Code != http.StatusOK {
		t.Fatalf("expected cloud remediation plan status %d, got %d: %s", http.StatusOK, planResponse.Code, planResponse.Body.String())
	}
}

func TestRouterTechnologyActivityFlow(t *testing.T) {
	router := newTestRouter()

	techBody := bytes.NewBufferString(`{
		"name": "Go",
		"category": "Language",
		"version": "1.22",
		"owner_team": "platform",
		"risk_level": "LOW",
		"tags": ["backend", "cloud"]
	}`)
	techResponse := httptest.NewRecorder()
	router.ServeHTTP(techResponse, httptest.NewRequest(http.MethodPost, "/api/v1/technologies", techBody))
	if techResponse.Code != http.StatusCreated {
		t.Fatalf("expected technology create status %d, got %d: %s", http.StatusCreated, techResponse.Code, techResponse.Body.String())
	}

	var technology model.Technology
	if err := json.NewDecoder(techResponse.Body).Decode(&technology); err != nil {
		t.Fatal(err)
	}
	if technology.Category != "language" || technology.Status != "active" || technology.RiskLevel != "low" {
		t.Fatalf("expected normalized technology fields, got %#v", technology)
	}

	activityBody := bytes.NewBufferString(`{
		"type": "Deployment",
		"action": "Rollout",
		"actor": "devops@example.com",
		"resource_type": "technology",
		"resource_id": "` + technology.ID + `",
		"owner_team": "platform",
		"summary": "Technology baseline approved.",
		"tags": ["audit"]
	}`)
	activityResponse := httptest.NewRecorder()
	router.ServeHTTP(activityResponse, httptest.NewRequest(http.MethodPost, "/api/v1/activities", activityBody))
	if activityResponse.Code != http.StatusCreated {
		t.Fatalf("expected activity create status %d, got %d: %s", http.StatusCreated, activityResponse.Code, activityResponse.Body.String())
	}

	var activity model.Activity
	if err := json.NewDecoder(activityResponse.Body).Decode(&activity); err != nil {
		t.Fatal(err)
	}
	if activity.Type != "deployment" || activity.Action != "rollout" || activity.Status != "succeeded" {
		t.Fatalf("expected normalized activity fields, got %#v", activity)
	}

	techListResponse := httptest.NewRecorder()
	router.ServeHTTP(techListResponse, httptest.NewRequest(http.MethodGet, "/api/v1/technologies?category=LANGUAGE&owner_team=platform&tag=backend", nil))
	if techListResponse.Code != http.StatusOK {
		t.Fatalf("expected technology list status %d, got %d: %s", http.StatusOK, techListResponse.Code, techListResponse.Body.String())
	}

	var technologies []model.Technology
	if err := json.NewDecoder(techListResponse.Body).Decode(&technologies); err != nil {
		t.Fatal(err)
	}
	if len(technologies) != 1 || technologies[0].ID != technology.ID {
		t.Fatalf("expected filtered technology %q, got %#v", technology.ID, technologies)
	}

	activityListResponse := httptest.NewRecorder()
	router.ServeHTTP(activityListResponse, httptest.NewRequest(http.MethodGet, "/api/v1/activities?type=DEPLOYMENT&resource_type=technology&owner_team=PLATFORM&tag=audit", nil))
	if activityListResponse.Code != http.StatusOK {
		t.Fatalf("expected activity list status %d, got %d: %s", http.StatusOK, activityListResponse.Code, activityListResponse.Body.String())
	}

	var activities []model.Activity
	if err := json.NewDecoder(activityListResponse.Body).Decode(&activities); err != nil {
		t.Fatal(err)
	}
	if len(activities) != 1 || activities[0].ID != activity.ID {
		t.Fatalf("expected filtered activity %q, got %#v", activity.ID, activities)
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
	cloudAccountRepository := repository.NewInMemoryCloudAccountRepository()
	technologyRepository := repository.NewInMemoryTechnologyRepository()
	activityRepository := repository.NewInMemoryActivityRepository()

	employeeService := service.NewEmployeeService(employeeRepository)
	applicationService := service.NewApplicationService(applicationRepository)
	clusterService := service.NewClusterService(clusterRepository)
	environmentService := service.NewEnvironmentService(applicationRepository, clusterRepository, environmentRepository)
	deploymentService := service.NewDeploymentService(applicationRepository, clusterRepository, deploymentRepository)
	pipelineService := service.NewPipelineService(applicationRepository, pipelineRepository)
	microserviceService := service.NewMicroserviceService(applicationRepository, microserviceRepository)
	incidentService := service.NewIncidentService(applicationRepository, clusterRepository, deploymentRepository, incidentRepository)
	cloudAccountService := service.NewCloudAccountService(cloudAccountRepository)
	technologyService := service.NewTechnologyService(technologyRepository)
	activityService := service.NewActivityService(activityRepository)
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
		controller.NewCloudAccountController(cloudAccountService),
		controller.NewTechnologyController(technologyService),
		controller.NewActivityController(activityService),
		controller.NewPlatformController(platformService),
	)
}
