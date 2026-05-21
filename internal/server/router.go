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
	deploymentController *controller.DeploymentController,
	pipelineController *controller.PipelineController,
	platformController *controller.PlatformController,
) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", healthController.Health)
	mux.HandleFunc("/readyz", healthController.Ready)

	mux.HandleFunc("/api/v1/employees", employeeController.Index)
	mux.HandleFunc("/api/v1/employees/", employeeController.Show)
	mux.HandleFunc("/api/v1/applications", applicationController.Index)
	mux.HandleFunc("/api/v1/applications/", applicationController.Show)
	mux.HandleFunc("/api/v1/clusters", clusterController.Index)
	mux.HandleFunc("/api/v1/clusters/", clusterController.ShowOrUpdateStatus)
	mux.HandleFunc("/api/v1/deployments", deploymentController.Index)
	mux.HandleFunc("/api/v1/deployments/", deploymentController.ShowOrUpdateStatus)
	mux.HandleFunc("/api/v1/pipelines", pipelineController.Index)
	mux.HandleFunc("/api/v1/pipelines/", pipelineController.ShowOrUpdateStatus)
	mux.HandleFunc("/api/v1/platform/summary", platformController.Summary)

	return WithRequestLogging(mux)
}
