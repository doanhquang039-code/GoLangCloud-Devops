package main

import (
    "log"
    "net/http"
    "os"

    "hr-cloud-service/internal/controller"
    "hr-cloud-service/internal/repository"
    "hr-cloud-service/internal/server"
    "hr-cloud-service/internal/service"
)

func main() {
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    employeeRepository := repository.NewInMemoryEmployeeRepository()
    employeeService := service.NewEmployeeService(employeeRepository)
    employeeController := controller.NewEmployeeController(employeeService)

    router := server.NewRouter(employeeController)

    log.Printf("HR Cloud Service listening on :%s", port)
    if err := http.ListenAndServe(":"+port, router); err != nil {
        log.Fatal(err)
    }
}
