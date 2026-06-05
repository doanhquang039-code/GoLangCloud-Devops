package service

import (
	"context"
	"testing"

	"hr-cloud-service/internal/model"
	"hr-cloud-service/internal/repository"
)

func TestCloudAccountServiceFiltersAndSummary(t *testing.T) {
	repo := repository.NewInMemoryCloudAccountRepository()
	service := NewCloudAccountService(repo)
	ctx := context.Background()

	aws, err := service.CreateCloudAccount(ctx, model.CreateCloudAccountRequest{
		Name:                 "hr-prod-aws",
		Provider:             "AWS",
		AccountID:            "123456789012",
		Region:               "ap-southeast-1",
		OwnerTeam:            "platform",
		Environment:          "production",
		MonthlyCostUSD:       1200.50,
		BudgetUSD:            2000,
		ComplianceScore:      91,
		BackupStatus:         "protected",
		OpenSecurityFindings: 2,
		Tags:                 []string{"hr", "prod"},
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := service.CreateCloudAccount(ctx, model.CreateCloudAccountRequest{
		Name:                 "talent-gcp-staging",
		Provider:             "gcp",
		AccountID:            "talent-staging",
		Region:               "asia-southeast1",
		OwnerTeam:            "talent",
		Environment:          "staging",
		Status:               "restricted",
		MonthlyCostUSD:       300,
		BudgetUSD:            500,
		ComplianceScore:      74,
		BackupStatus:         "partial",
		OpenSecurityFindings: 5,
		Tags:                 []string{"talent"},
	}); err != nil {
		t.Fatal(err)
	}

	filtered, err := service.GetCloudAccounts(ctx, model.CloudAccountFilter{Provider: "aws", Tag: "HR"})
	if err != nil {
		t.Fatal(err)
	}
	if len(filtered) != 1 || filtered[0].ID != aws.ID {
		t.Fatalf("expected aws account %q, got %#v", aws.ID, filtered)
	}

	summary, err := service.GetCloudAccountSummary(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if summary.Accounts != 2 || summary.ByProvider["aws"] != 1 || summary.BackupStatus["partial"] != 1 {
		t.Fatalf("unexpected summary: %#v", summary)
	}
	if summary.BudgetUtilization <= 0 || summary.AverageComplianceScore <= 0 {
		t.Fatalf("expected computed summary values, got %#v", summary)
	}

	violations, err := service.GetPolicyViolations(ctx, model.CloudAccountFilter{Provider: "gcp"})
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) == 0 {
		t.Fatal("expected policy violations for restricted partial-backup account")
	}

	plan, err := service.GetRemediationPlan(ctx, model.CloudAccountFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Actions) == 0 {
		t.Fatal("expected remediation actions")
	}
}

func TestCloudAccountServiceRejectsInvalidInput(t *testing.T) {
	service := NewCloudAccountService(repository.NewInMemoryCloudAccountRepository())

	_, err := service.CreateCloudAccount(context.Background(), model.CreateCloudAccountRequest{
		Name:            "bad",
		Provider:        "invalid",
		AccountID:       "123",
		Region:          "ap-southeast-1",
		OwnerTeam:       "platform",
		Environment:     "production",
		ComplianceScore: 101,
	})
	if err != ErrInvalidCloudAccount {
		t.Fatalf("expected ErrInvalidCloudAccount, got %v", err)
	}
}
