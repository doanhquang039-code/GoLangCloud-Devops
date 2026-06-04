package server

import (
	"net/http"

	"hr-cloud-service/internal/controller"
)

func NewRouter(
	healthController *controller.HealthController,
	employeeController *controller.EmployeeController,
	applicationController *controller.ApplicationController,
	clusterController *controller.ClusterController,
	environmentController *controller.EnvironmentController,
	deploymentController *controller.DeploymentController,
	pipelineController *controller.PipelineController,
	microserviceController *controller.MicroserviceController,
	incidentController *controller.IncidentController,
	platformController *controller.PlatformController,
) http.Handler {
	mux := http.NewServeMux()
	metrics := NewMetrics()

	mux.HandleFunc("/healthz", healthController.Health)
	mux.HandleFunc("/readyz", healthController.Ready)
	mux.HandleFunc("/metrics", metrics.Handler)

	mux.HandleFunc("/api/v1/employees", employeeController.Index)
	mux.HandleFunc("/api/v1/employees/", employeeController.Show)
	mux.HandleFunc("/api/v1/applications", applicationController.Index)
	mux.HandleFunc("/api/v1/applications/", applicationController.Show)
	mux.HandleFunc("/api/v1/clusters", clusterController.Index)
	mux.HandleFunc("/api/v1/clusters/", clusterController.ShowOrUpdateStatus)
	mux.HandleFunc("/api/v1/environments", environmentController.Index)
	mux.HandleFunc("/api/v1/environments/", environmentController.ShowOrUpdate)
	mux.HandleFunc("/api/v1/deployments", deploymentController.Index)
	mux.HandleFunc("/api/v1/deployments/", deploymentController.ShowOrUpdateStatus)
	mux.HandleFunc("/api/v1/pipelines", pipelineController.Index)
	mux.HandleFunc("/api/v1/pipelines/", pipelineController.ShowOrUpdateStatus)
	mux.HandleFunc("/api/v1/microservices", microserviceController.Index)
	mux.HandleFunc("/api/v1/microservices/", microserviceController.ShowOrUpdate)
	mux.HandleFunc("/api/v1/incidents", incidentController.Index)
	mux.HandleFunc("/api/v1/incidents/", incidentController.ShowOrUpdate)
	mux.HandleFunc("/api/v1/platform/summary", platformController.Summary)
	mux.HandleFunc("/api/v1/platform/scorecards", platformController.Scorecards)
	mux.HandleFunc("/api/v1/platform/environment-drift", platformController.EnvironmentDrift)

	return WithRequestLogging(metrics.Middleware(mux))
}
