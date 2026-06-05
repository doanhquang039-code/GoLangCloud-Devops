package database

import (
	"context"
	"time"

	"hr-cloud-service/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func SeedMongoData(ctx context.Context, db *mongo.Database) error {
	now := time.Now().UTC()
	oneHourAgo := now.Add(-1 * time.Hour)
	twoHoursAgo := now.Add(-2 * time.Hour)
	threeHoursAgo := now.Add(-3 * time.Hour)
	finishedBuild := now.Add(-85 * time.Minute)
	finishedDeploy := now.Add(-25 * time.Minute)
	resolvedAt := now.Add(-30 * time.Minute)

	employees := []model.Employee{
		{ID: "emp-devops-001", Name: "Nguyen Van An", Email: "an.nguyen@company.com", Department: "Platform Engineering", Title: "DevOps Lead", CreatedAt: threeHoursAgo, UpdatedAt: now},
		{ID: "emp-backend-001", Name: "Tran Thi Bich", Email: "bich.tran@company.com", Department: "Engineering", Title: "Backend Engineer", CreatedAt: threeHoursAgo, UpdatedAt: now},
		{ID: "emp-sre-001", Name: "Le Minh Quan", Email: "quan.le@company.com", Department: "SRE", Title: "Site Reliability Engineer", CreatedAt: threeHoursAgo, UpdatedAt: now},
	}
	applications := []model.Application{
		{
			ID:             "app-payroll-api",
			Name:           "payroll-api",
			Repository:     "github.com/company/payroll-api",
			Runtime:        "go1.22",
			OwnerTeam:      "platform",
			Criticality:    "high",
			Port:           8080,
			Replicas:       3,
			HealthEndpoint: "/healthz",
			Environment:    map[string]string{"LOG_LEVEL": "info", "FEATURE_AUDIT": "enabled"},
			Tags:           []string{"go", "payroll", "backend"},
			CreatedAt:      threeHoursAgo,
			UpdatedAt:      now,
		},
		{
			ID:             "app-recruitment-web",
			Name:           "recruitment-web",
			Repository:     "github.com/company/recruitment-web",
			Runtime:        "node20",
			OwnerTeam:      "talent",
			Criticality:    "medium",
			Port:           3000,
			Replicas:       2,
			HealthEndpoint: "/ready",
			Environment:    map[string]string{"LOG_LEVEL": "info", "API_BASE_URL": "http://payroll-api:8080"},
			Tags:           []string{"frontend", "recruitment"},
			CreatedAt:      threeHoursAgo,
			UpdatedAt:      now,
		},
	}
	clusters := []model.Cluster{
		{ID: "cls-eks-staging", Name: "eks-staging-ap-southeast-1", Provider: "aws", Region: "ap-southeast-1", Endpoint: "https://staging.example.eks.amazonaws.com", Version: "1.30", Status: "ready", CreatedAt: threeHoursAgo, UpdatedAt: now},
		{ID: "cls-gke-prod", Name: "gke-prod-asia-southeast1", Provider: "gcp", Region: "asia-southeast1", Endpoint: "https://prod.example.gke.googleapis.com", Version: "1.30", Status: "ready", CreatedAt: threeHoursAgo, UpdatedAt: now},
	}
	cloudAccounts := []model.CloudAccount{
		{ID: "cloud-aws-hr-prod", Name: "hr-prod-aws", Provider: "aws", AccountID: "123456789012", Region: "ap-southeast-1", OwnerTeam: "platform", Environment: "production", Status: "active", MonthlyCostUSD: 18420.75, BudgetUSD: 25000, ComplianceScore: 91, BackupStatus: "protected", OpenSecurityFindings: 3, Tags: []string{"hr", "prod", "aws"}, CreatedAt: threeHoursAgo, UpdatedAt: now},
		{ID: "cloud-gcp-shared-prod", Name: "shared-prod-gcp", Provider: "gcp", AccountID: "company-prod-001", Region: "asia-southeast1", OwnerTeam: "sre", Environment: "production", Status: "active", MonthlyCostUSD: 22310.40, BudgetUSD: 30000, ComplianceScore: 88, BackupStatus: "protected", OpenSecurityFindings: 5, Tags: []string{"shared", "prod", "gcp"}, CreatedAt: threeHoursAgo, UpdatedAt: now},
		{ID: "cloud-azure-talent-staging", Name: "talent-staging-azure", Provider: "azure", AccountID: "az-talent-stg-01", Region: "southeastasia", OwnerTeam: "talent", Environment: "staging", Status: "restricted", MonthlyCostUSD: 3120.00, BudgetUSD: 5000, ComplianceScore: 76, BackupStatus: "partial", OpenSecurityFindings: 7, Tags: []string{"talent", "staging", "azure"}, CreatedAt: threeHoursAgo, UpdatedAt: now},
	}
	environments := []model.Environment{
		{ID: "env-payroll-staging", Name: "payroll-staging", Type: "staging", ApplicationID: "app-payroll-api", ClusterID: "cls-eks-staging", Namespace: "hr-staging", Status: "active", Variables: map[string]string{"LOG_LEVEL": "debug", "MONGO_DATABASE": "hr_cloud_staging"}, CreatedAt: twoHoursAgo, UpdatedAt: now},
		{ID: "env-payroll-prod", Name: "payroll-prod", Type: "production", ApplicationID: "app-payroll-api", ClusterID: "cls-gke-prod", Namespace: "hr-prod", Status: "active", Variables: map[string]string{"LOG_LEVEL": "info", "MONGO_DATABASE": "hr_cloud"}, CreatedAt: twoHoursAgo, UpdatedAt: now},
		{ID: "env-recruitment-staging", Name: "recruitment-staging", Type: "staging", ApplicationID: "app-recruitment-web", ClusterID: "cls-eks-staging", Namespace: "hr-staging", Status: "active", Variables: map[string]string{"NODE_ENV": "staging"}, CreatedAt: twoHoursAgo, UpdatedAt: now},
	}
	deployments := []model.Deployment{
		{ID: "dep-payroll-prod-141", ApplicationID: "app-payroll-api", ClusterID: "cls-gke-prod", Namespace: "hr-prod", Environment: "production", Version: "v1.4.1", Strategy: "rolling", Status: "succeeded", RequestedBy: "an.nguyen@company.com", StartedAt: twoHoursAgo, FinishedAt: &finishedBuild},
		{ID: "dep-payroll-staging-150", ApplicationID: "app-payroll-api", ClusterID: "cls-eks-staging", Namespace: "hr-staging", Environment: "staging", Version: "v1.5.0-rc1", Strategy: "canary", Status: "running", RequestedBy: "bich.tran@company.com", StartedAt: oneHourAgo},
		{ID: "dep-recruitment-staging-220", ApplicationID: "app-recruitment-web", ClusterID: "cls-eks-staging", Namespace: "hr-staging", Environment: "staging", Version: "v2.2.0", Strategy: "blue-green", Status: "succeeded", RequestedBy: "quan.le@company.com", StartedAt: now.Add(-45 * time.Minute), FinishedAt: &finishedDeploy},
	}
	pipelineRuns := []model.PipelineRun{
		{
			ID:            "pipe-payroll-main-141",
			ApplicationID: "app-payroll-api",
			Branch:        "main",
			CommitSHA:     "4f9a2c9",
			TriggeredBy:   "an.nguyen@company.com",
			Status:        "succeeded",
			Stages: []model.PipelineStage{
				{Name: "build", Status: "succeeded", StartedAt: twoHoursAgo, EndedAt: &finishedBuild},
				{Name: "unit-test", Status: "succeeded", StartedAt: twoHoursAgo.Add(10 * time.Minute), EndedAt: &finishedBuild},
				{Name: "security-scan", Status: "succeeded", StartedAt: twoHoursAgo.Add(20 * time.Minute), EndedAt: &finishedBuild},
			},
			StartedAt:  twoHoursAgo,
			FinishedAt: &finishedBuild,
		},
		{
			ID:            "pipe-recruitment-develop-220",
			ApplicationID: "app-recruitment-web",
			Branch:        "develop",
			CommitSHA:     "a18d7ef",
			TriggeredBy:   "quan.le@company.com",
			Status:        "running",
			Stages: []model.PipelineStage{
				{Name: "build", Status: "succeeded", StartedAt: oneHourAgo, EndedAt: &finishedDeploy},
				{Name: "containerize", Status: "running", StartedAt: now.Add(-20 * time.Minute)},
			},
			StartedAt: oneHourAgo,
		},
	}
	microservices := []model.Microservice{
		{
			ID:                   "svc-payroll-api",
			TenantID:             "tenant-hr",
			ApplicationID:        "app-payroll-api",
			Name:                 "payroll-api",
			OwnerTeam:            "platform",
			Protocol:             "http",
			Endpoint:             "http://payroll-api.hr-prod.svc.cluster.local:8080",
			Status:               "active",
			CloudProvider:        "gcp",
			Region:               "asia-southeast1",
			ClusterID:            "cls-gke-prod",
			Namespace:            "hr-prod",
			Environment:          "production",
			Runtime:              "go1.22",
			Image:                "ghcr.io/company/payroll-api:v1.4.1",
			Version:              "v1.4.1",
			Replicas:             4,
			CPURequest:           "500m",
			MemoryRequest:        "768Mi",
			HealthPath:           "/healthz",
			SLOTarget:            99.95,
			ErrorBudgetRemaining: 88.2,
			Dependencies:         []string{"mongodb", "audit-events"},
			Config:               map[string]string{"TIMEOUT_SECONDS": "5"},
			Tags:                 []string{"backend", "payroll"},
			CreatedAt:            threeHoursAgo,
			UpdatedAt:            now,
		},
		{
			ID:                   "svc-payroll-events",
			TenantID:             "tenant-hr",
			ApplicationID:        "app-payroll-api",
			Name:                 "payroll-events",
			OwnerTeam:            "platform",
			Protocol:             "event",
			Endpoint:             "topic://payroll-events",
			Status:               "active",
			CloudProvider:        "aws",
			Region:               "ap-southeast-1",
			ClusterID:            "cls-eks-staging",
			Namespace:            "hr-staging",
			Environment:          "staging",
			Runtime:              "go1.22",
			Image:                "ghcr.io/company/payroll-events:v1.5.0-rc1",
			Version:              "v1.5.0-rc1",
			Replicas:             2,
			CPURequest:           "250m",
			MemoryRequest:        "512Mi",
			HealthPath:           "/healthz",
			SLOTarget:            99.9,
			ErrorBudgetRemaining: 71.4,
			Dependencies:         []string{"payroll-api"},
			Tags:                 []string{"events", "payroll"},
			CreatedAt:            threeHoursAgo,
			UpdatedAt:            now,
		},
	}
	incidents := []model.Incident{
		{ID: "inc-payroll-error-rate", Title: "Payroll API error rate elevated", Summary: "5xx responses increased after the latest canary rollout.", Severity: "sev2", Status: "investigating", ApplicationID: "app-payroll-api", ClusterID: "cls-eks-staging", DeploymentID: "dep-payroll-staging-150", OwnerTeam: "platform", CreatedAt: oneHourAgo, UpdatedAt: now},
		{ID: "inc-recruitment-cache", Title: "Recruitment web cache warmed slowly", Summary: "Cache warmup exceeded the expected threshold during deployment.", Severity: "sev4", Status: "resolved", ApplicationID: "app-recruitment-web", ClusterID: "cls-eks-staging", DeploymentID: "dep-recruitment-staging-220", OwnerTeam: "talent", CreatedAt: twoHoursAgo, UpdatedAt: now, ResolvedAt: &resolvedAt},
	}

	seedSets := []struct {
		collection string
		documents  any
	}{
		{"employees", employees},
		{"applications", applications},
		{"clusters", clusters},
		{"cloud_accounts", cloudAccounts},
		{"environments", environments},
		{"deployments", deployments},
		{"pipeline_runs", pipelineRuns},
		{"microservices", microservices},
		{"incidents", incidents},
	}

	for _, seedSet := range seedSets {
		if err := upsertSeedDocuments(ctx, db.Collection(seedSet.collection), seedSet.documents); err != nil {
			return err
		}
	}

	return nil
}

func upsertSeedDocuments(ctx context.Context, collection *mongo.Collection, documents any) error {
	switch typedDocuments := documents.(type) {
	case []model.Employee:
		for _, document := range typedDocuments {
			if err := upsertSeedDocument(ctx, collection, document.ID, document); err != nil {
				return err
			}
		}
	case []model.Application:
		for _, document := range typedDocuments {
			if err := upsertSeedDocument(ctx, collection, document.ID, document); err != nil {
				return err
			}
		}
	case []model.Cluster:
		for _, document := range typedDocuments {
			if err := upsertSeedDocument(ctx, collection, document.ID, document); err != nil {
				return err
			}
		}
	case []model.CloudAccount:
		for _, document := range typedDocuments {
			if err := upsertSeedDocument(ctx, collection, document.ID, document); err != nil {
				return err
			}
		}
	case []model.Environment:
		for _, document := range typedDocuments {
			if err := upsertSeedDocument(ctx, collection, document.ID, document); err != nil {
				return err
			}
		}
	case []model.Deployment:
		for _, document := range typedDocuments {
			if err := upsertSeedDocument(ctx, collection, document.ID, document); err != nil {
				return err
			}
		}
	case []model.PipelineRun:
		for _, document := range typedDocuments {
			if err := upsertSeedDocument(ctx, collection, document.ID, document); err != nil {
				return err
			}
		}
	case []model.Microservice:
		for _, document := range typedDocuments {
			if err := upsertSeedDocument(ctx, collection, document.ID, document); err != nil {
				return err
			}
		}
	case []model.Incident:
		for _, document := range typedDocuments {
			if err := upsertSeedDocument(ctx, collection, document.ID, document); err != nil {
				return err
			}
		}
	}

	return nil
}

func upsertSeedDocument(ctx context.Context, collection *mongo.Collection, id string, document any) error {
	_, err := collection.ReplaceOne(ctx, bson.M{"id": id}, document, options.Replace().SetUpsert(true))
	return err
}
