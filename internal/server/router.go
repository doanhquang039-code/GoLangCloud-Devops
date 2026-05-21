package server

import (
    "net/http"

    "hr-cloud-service/internal/controller"
)

func NewRouter(employeeController *controller.EmployeeController) http.Handler {
    mux := http.NewServeMux()

    mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodGet {
            w.WriteHeader(http.StatusMethodNotAllowed)
            return
        }
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte(`{"status":"ok"}`))
    })

    mux.HandleFunc("/api/v1/employees", employeeController.Index)
    mux.HandleFunc("/api/v1/employees/", employeeController.Show)

    return mux
}
