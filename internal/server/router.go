package server

import (
	"net/http"

	"hr-cloud-service/internal/controller"
)

func NewRouter(
	healthController *controller.HealthController,
	employeeController *controller.EmployeeController,
	applicationController *controller.ApplicationController,
	deploymentController *controller.DeploymentController,
) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", healthController.Health)
	mux.HandleFunc("/readyz", healthController.Ready)

	mux.HandleFunc("/api/v1/employees", employeeController.Index)
	mux.HandleFunc("/api/v1/employees/", employeeController.Show)
	mux.HandleFunc("/api/v1/applications", applicationController.Index)
	mux.HandleFunc("/api/v1/applications/", applicationController.Show)
	mux.HandleFunc("/api/v1/deployments", deploymentController.Index)
	mux.HandleFunc("/api/v1/deployments/", deploymentController.ShowOrUpdateStatus)

	return WithRequestLogging(mux)
}
