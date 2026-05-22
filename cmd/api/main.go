package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"hr-cloud-service/internal/controller"
	"hr-cloud-service/internal/database"
	"hr-cloud-service/internal/repository"
	"hr-cloud-service/internal/server"
	"hr-cloud-service/internal/service"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	mongoDatabase := os.Getenv("MONGO_DATABASE")
	if mongoDatabase == "" {
		mongoDatabase = "hr_cloud"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db, disconnect, err := database.ConnectMongo(ctx, database.MongoConfig{
		URI:      mongoURI,
		Database: mongoDatabase,
	})
	if err != nil {
		log.Fatalf("could not connect to MongoDB: %v", err)
	}
	defer func() {
		if err := disconnect(context.Background()); err != nil {
			log.Printf("could not disconnect MongoDB: %v", err)
		}
	}()

	indexCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := database.EnsureMongoIndexes(indexCtx, db); err != nil {
		log.Fatalf("could not ensure MongoDB indexes: %v", err)
	}

	employeeRepository := repository.NewMongoEmployeeRepository(db)
	applicationRepository := repository.NewMongoApplicationRepository(db)
	clusterRepository := repository.NewMongoClusterRepository(db)
	environmentRepository := repository.NewMongoEnvironmentRepository(db)
	deploymentRepository := repository.NewMongoDeploymentRepository(db)
	pipelineRepository := repository.NewMongoPipelineRepository(db)
	incidentRepository := repository.NewMongoIncidentRepository(db)

	employeeService := service.NewEmployeeService(employeeRepository)
	applicationService := service.NewApplicationService(applicationRepository)
	clusterService := service.NewClusterService(clusterRepository)
	environmentService := service.NewEnvironmentService(applicationRepository, clusterRepository, environmentRepository)
	deploymentService := service.NewDeploymentService(applicationRepository, clusterRepository, deploymentRepository)
	pipelineService := service.NewPipelineService(applicationRepository, pipelineRepository)
	incidentService := service.NewIncidentService(applicationRepository, clusterRepository, deploymentRepository, incidentRepository)
	platformService := service.NewPlatformService(applicationRepository, clusterRepository, environmentRepository, deploymentRepository, pipelineRepository, incidentRepository)

	healthController := controller.NewHealthController(db)
	employeeController := controller.NewEmployeeController(employeeService)
	applicationController := controller.NewApplicationController(applicationService)
	clusterController := controller.NewClusterController(clusterService)
	environmentController := controller.NewEnvironmentController(environmentService)
	deploymentController := controller.NewDeploymentController(deploymentService)
	pipelineController := controller.NewPipelineController(pipelineService)
	incidentController := controller.NewIncidentController(incidentService)
	platformController := controller.NewPlatformController(platformService)

	router := server.NewRouter(healthController, employeeController, applicationController, clusterController, environmentController, deploymentController, pipelineController, incidentController, platformController)

	httpServer := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("connected to MongoDB database %q", mongoDatabase)
	log.Printf("HR Cloud DevOps Service listening on :%s", port)

	errCh := make(chan error, 1)
	go func() {
		errCh <- httpServer.ListenAndServe()
	}()

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	case <-stopCh:
		log.Println("shutdown signal received")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("could not shutdown server cleanly: %v", err)
	}

	log.Println("server stopped")
}
