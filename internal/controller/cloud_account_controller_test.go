package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"hr-cloud-service/internal/model"
	"hr-cloud-service/internal/repository"
	"hr-cloud-service/internal/service"
)

func TestCloudAccountControllerCRUDAndSummary(t *testing.T) {
	controller := NewCloudAccountController(service.NewCloudAccountService(repository.NewInMemoryCloudAccountRepository()))

	createBody := bytes.NewBufferString(`{
		"name": "hr-prod-aws",
		"provider": "aws",
		"account_id": "123456789012",
		"region": "ap-southeast-1",
		"owner_team": "platform",
		"environment": "production",
		"monthly_cost_usd": 1250.25,
		"budget_usd": 2000,
		"compliance_score": 92,
		"backup_status": "protected",
		"open_security_findings": 2,
		"tags": ["hr", "prod"]
	}`)
	createResponse := httptest.NewRecorder()
	controller.Index(createResponse, httptest.NewRequest(http.MethodPost, "/api/v1/cloud-accounts", createBody))
	if createResponse.Code != http.StatusCreated {
		t.Fatalf("expected create status %d, got %d: %s", http.StatusCreated, createResponse.Code, createResponse.Body.String())
	}

	var account model.CloudAccount
	if err := json.NewDecoder(createResponse.Body).Decode(&account); err != nil {
		t.Fatal(err)
	}
	if account.Provider != "aws" || account.Status != "active" {
		t.Fatalf("expected normalized account, got %#v", account)
	}

	listResponse := httptest.NewRecorder()
	controller.Index(listResponse, httptest.NewRequest(http.MethodGet, "/api/v1/cloud-accounts?provider=AWS&tag=prod", nil))
	if listResponse.Code != http.StatusOK {
		t.Fatalf("expected list status %d, got %d", http.StatusOK, listResponse.Code)
	}

	var accounts []model.CloudAccount
	if err := json.NewDecoder(listResponse.Body).Decode(&accounts); err != nil {
		t.Fatal(err)
	}
	if len(accounts) != 1 || accounts[0].ID != account.ID {
		t.Fatalf("expected filtered account %q, got %#v", account.ID, accounts)
	}

	patchResponse := httptest.NewRecorder()
	controller.ShowOrUpdate(patchResponse, httptest.NewRequest(http.MethodPatch, "/api/v1/cloud-accounts/"+account.ID, bytes.NewBufferString(`{"status":"restricted","backup_status":"partial"}`)))
	if patchResponse.Code != http.StatusOK {
		t.Fatalf("expected patch status %d, got %d: %s", http.StatusOK, patchResponse.Code, patchResponse.Body.String())
	}

	summaryResponse := httptest.NewRecorder()
	controller.Summary(summaryResponse, httptest.NewRequest(http.MethodGet, "/api/v1/cloud/summary", nil))
	if summaryResponse.Code != http.StatusOK {
		t.Fatalf("expected summary status %d, got %d", http.StatusOK, summaryResponse.Code)
	}

	var summary model.CloudAccountSummary
	if err := json.NewDecoder(summaryResponse.Body).Decode(&summary); err != nil {
		t.Fatal(err)
	}
	if summary.Accounts != 1 || summary.BackupStatus["partial"] != 1 {
		t.Fatalf("unexpected summary: %#v", summary)
	}

	violationsResponse := httptest.NewRecorder()
	controller.PolicyViolations(violationsResponse, httptest.NewRequest(http.MethodGet, "/api/v1/cloud/policy-violations", nil))
	if violationsResponse.Code != http.StatusOK {
		t.Fatalf("expected violations status %d, got %d", http.StatusOK, violationsResponse.Code)
	}

	var violations []model.CloudPolicyViolation
	if err := json.NewDecoder(violationsResponse.Body).Decode(&violations); err != nil {
		t.Fatal(err)
	}
	if len(violations) == 0 {
		t.Fatal("expected policy violations")
	}

	planResponse := httptest.NewRecorder()
	controller.RemediationPlan(planResponse, httptest.NewRequest(http.MethodGet, "/api/v1/cloud/remediation-plan", nil))
	if planResponse.Code != http.StatusOK {
		t.Fatalf("expected remediation plan status %d, got %d", http.StatusOK, planResponse.Code)
	}
}
