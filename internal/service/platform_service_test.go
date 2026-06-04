package service

import (
	"context"
	"testing"
	"time"

	"hr-cloud-service/internal/model"
	"hr-cloud-service/internal/repository"
)

func TestGetScorecardsBuildsOperationalReadinessMetrics(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 5, 23, 10, 0, 0, 0, time.UTC)
	finishedDeployment := now.Add(30 * time.Minute)
	finishedPipeline := now.Add(20 * time.Minute)
	resolvedAt := now.Add(45 * time.Minute)

	applicationRepository := repository.NewInMemoryApplicationRepository()
	clusterRepository := repository.NewInMemoryClusterRepository()
	environmentRepository := repository.NewInMemoryEnvironmentRepository()
	deploymentRepository := repository.NewInMemoryDeploymentRepository()
	pipelineRepository := repository.NewInMemoryPipelineRepository()
	incidentRepository := repository.NewInMemoryIncidentRepository()

	if _, err := applicationRepository.Save(ctx, model.Application{
		ID:          "app-payroll-api",
		Name:        "payroll-api",
		OwnerTeam:   "platform",
		Criticality: "high",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := clusterRepository.Save(ctx, model.Cluster{ID: "cls-prod", Status: "ready"}); err != nil {
		t.Fatal(err)
	}
	if _, err := environmentRepository.Save(ctx, model.Environment{
		ID:            "env-prod",
		ApplicationID: "app-payroll-api",
		ClusterID:     "cls-prod",
		Status:        "active",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := deploymentRepository.Save(ctx, model.Deployment{
		ID:            "dep-prod-100",
		ApplicationID: "app-payroll-api",
		Status:        "succeeded",
		StartedAt:     now,
		FinishedAt:    &finishedDeployment,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := pipelineRepository.Save(ctx, model.PipelineRun{
		ID:            "pipe-main-100",
		ApplicationID: "app-payroll-api",
		Status:        "succeeded",
		StartedAt:     now,
		FinishedAt:    &finishedPipeline,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := incidentRepository.Save(ctx, model.Incident{
		ID:            "inc-cache",
		ApplicationID: "app-payroll-api",
		Severity:      "sev2",
		Status:        "resolved",
		CreatedAt:     now,
		ResolvedAt:    &resolvedAt,
	}); err != nil {
		t.Fatal(err)
	}

	service := NewPlatformService(
		applicationRepository,
		clusterRepository,
		environmentRepository,
		deploymentRepository,
		pipelineRepository,
		incidentRepository,
	)

	scorecards, err := service.GetScorecards(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(scorecards) != 1 {
		t.Fatalf("expected 1 scorecard, got %d", len(scorecards))
	}

	scorecard := scorecards[0]
	if scorecard.ApplicationID != "app-payroll-api" {
		t.Fatalf("unexpected application id %q", scorecard.ApplicationID)
	}
	if scorecard.DeploymentSuccessRate != 100 {
		t.Fatalf("expected deployment success rate 100, got %.1f", scorecard.DeploymentSuccessRate)
	}
	if scorecard.PipelineSuccessRate != 100 {
		t.Fatalf("expected pipeline success rate 100, got %.1f", scorecard.PipelineSuccessRate)
	}
	if scorecard.MeanTimeToResolveMinutes != 45 {
		t.Fatalf("expected mttr 45 minutes, got %.1f", scorecard.MeanTimeToResolveMinutes)
	}
	if scorecard.OperationalReadinessScore != 85 {
		t.Fatalf("expected readiness score 85, got %d", scorecard.OperationalReadinessScore)
	}
	if scorecard.RiskLevel != "low" {
		t.Fatalf("expected low risk, got %q", scorecard.RiskLevel)
	}
}

func TestGetScorecardsFiltersAndSortsCloudReadiness(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 5, 23, 10, 0, 0, 0, time.UTC)
	finishedAt := now.Add(10 * time.Minute)

	applicationRepository := repository.NewInMemoryApplicationRepository()
	clusterRepository := repository.NewInMemoryClusterRepository()
	environmentRepository := repository.NewInMemoryEnvironmentRepository()
	deploymentRepository := repository.NewInMemoryDeploymentRepository()
	pipelineRepository := repository.NewInMemoryPipelineRepository()
	incidentRepository := repository.NewInMemoryIncidentRepository()

	if _, err := applicationRepository.Save(ctx, model.Application{ID: "app-payroll-api", Name: "payroll-api", OwnerTeam: "platform", Criticality: "high"}); err != nil {
		t.Fatal(err)
	}
	if _, err := applicationRepository.Save(ctx, model.Application{ID: "app-people-web", Name: "people-web", OwnerTeam: "people", Criticality: "medium"}); err != nil {
		t.Fatal(err)
	}
	if _, err := environmentRepository.Save(ctx, model.Environment{ID: "env-payroll-prod", ApplicationID: "app-payroll-api", ClusterID: "cls-prod", Status: "active"}); err != nil {
		t.Fatal(err)
	}
	if _, err := deploymentRepository.Save(ctx, model.Deployment{ID: "dep-payroll", ApplicationID: "app-payroll-api", Status: "succeeded", StartedAt: now, FinishedAt: &finishedAt}); err != nil {
		t.Fatal(err)
	}
	if _, err := pipelineRepository.Save(ctx, model.PipelineRun{ID: "pipe-payroll", ApplicationID: "app-payroll-api", Status: "succeeded", StartedAt: now, FinishedAt: &finishedAt}); err != nil {
		t.Fatal(err)
	}

	service := NewPlatformService(applicationRepository, clusterRepository, environmentRepository, deploymentRepository, pipelineRepository, incidentRepository)

	scorecards, err := service.GetScorecards(ctx, model.PlatformScorecardFilter{
		Query:     "payroll",
		OwnerTeam: "PLATFORM",
		RiskLevel: "LOW",
		MinScore:  80,
		SortBy:    "score",
		SortOrder: "desc",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(scorecards) != 1 || scorecards[0].ApplicationID != "app-payroll-api" {
		t.Fatalf("expected payroll scorecard, got %#v", scorecards)
	}
}

func TestGetEnvironmentDriftReportsFindsMissingChangedAndExtraVariables(t *testing.T) {
	ctx := context.Background()

	applicationRepository := repository.NewInMemoryApplicationRepository()
	clusterRepository := repository.NewInMemoryClusterRepository()
	environmentRepository := repository.NewInMemoryEnvironmentRepository()
	deploymentRepository := repository.NewInMemoryDeploymentRepository()
	pipelineRepository := repository.NewInMemoryPipelineRepository()
	incidentRepository := repository.NewInMemoryIncidentRepository()

	if _, err := applicationRepository.Save(ctx, model.Application{
		ID:   "app-payroll-api",
		Name: "payroll-api",
		Environment: map[string]string{
			"LOG_LEVEL":     "info",
			"FEATURE_AUDIT": "enabled",
			"MONGO_REGION":  "asia",
		},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := environmentRepository.Save(ctx, model.Environment{
		ID:            "env-payroll-prod",
		Name:          "payroll-prod",
		Type:          "production",
		ApplicationID: "app-payroll-api",
		ClusterID:     "cls-prod",
		Namespace:     "hr-prod",
		Status:        "active",
		Variables: map[string]string{
			"LOG_LEVEL":  "debug",
			"EXTRA_FLAG": "true",
		},
	}); err != nil {
		t.Fatal(err)
	}

	service := NewPlatformService(
		applicationRepository,
		clusterRepository,
		environmentRepository,
		deploymentRepository,
		pipelineRepository,
		incidentRepository,
	)

	reports, err := service.GetEnvironmentDriftReports(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(reports) != 1 {
		t.Fatalf("expected 1 drift report, got %d", len(reports))
	}

	report := reports[0]
	if len(report.MissingVariables) != 2 {
		t.Fatalf("expected 2 missing variables, got %d", len(report.MissingVariables))
	}
	if len(report.ChangedVariables) != 1 {
		t.Fatalf("expected 1 changed variable, got %d", len(report.ChangedVariables))
	}
	if report.ChangedVariables[0].Key != "LOG_LEVEL" {
		t.Fatalf("expected LOG_LEVEL changed variable, got %q", report.ChangedVariables[0].Key)
	}
	if len(report.ExtraVariables) != 1 {
		t.Fatalf("expected 1 extra variable, got %d", len(report.ExtraVariables))
	}
	if report.DriftLevel != "high" {
		t.Fatalf("expected high drift level, got %q", report.DriftLevel)
	}
}

func TestGetEnvironmentDriftReportsFiltersAndSortsCloudDrift(t *testing.T) {
	ctx := context.Background()

	applicationRepository := repository.NewInMemoryApplicationRepository()
	clusterRepository := repository.NewInMemoryClusterRepository()
	environmentRepository := repository.NewInMemoryEnvironmentRepository()
	deploymentRepository := repository.NewInMemoryDeploymentRepository()
	pipelineRepository := repository.NewInMemoryPipelineRepository()
	incidentRepository := repository.NewInMemoryIncidentRepository()

	if _, err := applicationRepository.Save(ctx, model.Application{
		ID:   "app-payroll-api",
		Name: "payroll-api",
		Environment: map[string]string{
			"LOG_LEVEL":    "info",
			"MONGO_REGION": "asia",
		},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := environmentRepository.Save(ctx, model.Environment{
		ID:            "env-payroll-prod",
		Name:          "payroll-prod",
		Type:          "Production",
		ApplicationID: "app-payroll-api",
		ClusterID:     "cls-prod",
		Namespace:     "hr-prod",
		Status:        "Active",
		Variables: map[string]string{
			"LOG_LEVEL":  "debug",
			"EXTRA_FLAG": "true",
		},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := environmentRepository.Save(ctx, model.Environment{
		ID:            "env-people-dev",
		Name:          "people-dev",
		Type:          "development",
		ApplicationID: "app-people-web",
		Status:        "inactive",
	}); err != nil {
		t.Fatal(err)
	}

	service := NewPlatformService(applicationRepository, clusterRepository, environmentRepository, deploymentRepository, pipelineRepository, incidentRepository)

	reports, err := service.GetEnvironmentDriftReports(ctx, model.EnvironmentDriftReportFilter{
		Query:           "LOG_LEVEL",
		EnvironmentType: "PRODUCTION",
		Status:          "ACTIVE",
		DriftLevel:      "HIGH",
		MaxDriftScore:   80,
		SortBy:          "score",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(reports) != 1 || reports[0].EnvironmentID != "env-payroll-prod" {
		t.Fatalf("expected payroll drift report, got %#v", reports)
	}
}
