package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"hr-cloud-service/internal/model"
	"hr-cloud-service/internal/repository"
)

var ErrInvalidCloudAccount = errors.New("invalid cloud account input")

type CloudAccountService struct {
	cloudAccountRepository repository.CloudAccountRepository
}

func NewCloudAccountService(cloudAccountRepository repository.CloudAccountRepository) *CloudAccountService {
	return &CloudAccountService{cloudAccountRepository: cloudAccountRepository}
}

func (s *CloudAccountService) GetCloudAccounts(ctx context.Context, filter model.CloudAccountFilter) ([]model.CloudAccount, error) {
	filter.Query = strings.TrimSpace(filter.Query)
	filter.Provider = strings.ToLower(strings.TrimSpace(filter.Provider))
	filter.Region = strings.TrimSpace(filter.Region)
	filter.OwnerTeam = strings.ToLower(strings.TrimSpace(filter.OwnerTeam))
	filter.Environment = strings.ToLower(strings.TrimSpace(filter.Environment))
	filter.Status = strings.ToLower(strings.TrimSpace(filter.Status))
	filter.BackupStatus = strings.ToLower(strings.TrimSpace(filter.BackupStatus))
	filter.Tag = strings.ToLower(strings.TrimSpace(filter.Tag))

	if filter.Status != "" && !isValidCloudAccountStatus(filter.Status) {
		return nil, ErrInvalidCloudAccount
	}
	if filter.BackupStatus != "" && !isValidCloudBackupStatus(filter.BackupStatus) {
		return nil, ErrInvalidCloudAccount
	}

	accounts, err := s.cloudAccountRepository.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	filtered := make([]model.CloudAccount, 0, len(accounts))
	for _, account := range accounts {
		if filter.Query != "" && !cloudAccountMatchesQuery(account, filter.Query) {
			continue
		}
		if filter.Provider != "" && !strings.EqualFold(account.Provider, filter.Provider) {
			continue
		}
		if filter.Region != "" && !strings.EqualFold(account.Region, filter.Region) {
			continue
		}
		if filter.OwnerTeam != "" && !strings.EqualFold(account.OwnerTeam, filter.OwnerTeam) {
			continue
		}
		if filter.Environment != "" && !strings.EqualFold(account.Environment, filter.Environment) {
			continue
		}
		if filter.Status != "" && !strings.EqualFold(account.Status, filter.Status) {
			continue
		}
		if filter.BackupStatus != "" && !strings.EqualFold(account.BackupStatus, filter.BackupStatus) {
			continue
		}
		if filter.Tag != "" && !containsCloudAccountTag(account.Tags, filter.Tag) {
			continue
		}
		filtered = append(filtered, account)
	}

	return filtered, nil
}

func (s *CloudAccountService) GetCloudAccountByID(ctx context.Context, id string) (model.CloudAccount, error) {
	if strings.TrimSpace(id) == "" {
		return model.CloudAccount{}, ErrInvalidCloudAccount
	}

	return s.cloudAccountRepository.FindByID(ctx, id)
}

func (s *CloudAccountService) CreateCloudAccount(ctx context.Context, request model.CreateCloudAccountRequest) (model.CloudAccount, error) {
	account, err := cloudAccountFromRequest("", request)
	if err != nil {
		return model.CloudAccount{}, err
	}

	now := time.Now().UTC()
	account.ID = cloudAccountID(account.Provider, account.AccountID)
	account.CreatedAt = now
	account.UpdatedAt = now

	return s.cloudAccountRepository.Save(ctx, account)
}

func (s *CloudAccountService) UpdateCloudAccount(ctx context.Context, id string, request model.UpdateCloudAccountRequest) (model.CloudAccount, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return model.CloudAccount{}, ErrInvalidCloudAccount
	}

	existing, err := s.cloudAccountRepository.FindByID(ctx, id)
	if err != nil {
		return model.CloudAccount{}, err
	}

	account, err := cloudAccountFromRequest(id, model.CreateCloudAccountRequest(request))
	if err != nil {
		return model.CloudAccount{}, err
	}
	account.CreatedAt = existing.CreatedAt
	account.UpdatedAt = time.Now().UTC()

	return s.cloudAccountRepository.Save(ctx, account)
}

func (s *CloudAccountService) UpdateCloudAccountStatus(ctx context.Context, id string, request model.UpdateCloudAccountStatusRequest) (model.CloudAccount, error) {
	status := strings.ToLower(strings.TrimSpace(request.Status))
	backupStatus := strings.ToLower(strings.TrimSpace(request.BackupStatus))
	if strings.TrimSpace(id) == "" {
		return model.CloudAccount{}, ErrInvalidCloudAccount
	}
	if status != "" && !isValidCloudAccountStatus(status) {
		return model.CloudAccount{}, ErrInvalidCloudAccount
	}
	if backupStatus != "" && !isValidCloudBackupStatus(backupStatus) {
		return model.CloudAccount{}, ErrInvalidCloudAccount
	}

	account, err := s.cloudAccountRepository.FindByID(ctx, id)
	if err != nil {
		return model.CloudAccount{}, err
	}
	if status != "" {
		account.Status = status
	}
	if backupStatus != "" {
		account.BackupStatus = backupStatus
	}
	account.UpdatedAt = time.Now().UTC()

	return s.cloudAccountRepository.Save(ctx, account)
}

func (s *CloudAccountService) DeleteCloudAccount(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrInvalidCloudAccount
	}

	return s.cloudAccountRepository.DeleteByID(ctx, id)
}

func (s *CloudAccountService) GetCloudAccountSummary(ctx context.Context) (model.CloudAccountSummary, error) {
	accounts, err := s.cloudAccountRepository.FindAll(ctx)
	if err != nil {
		return model.CloudAccountSummary{}, err
	}

	summary := model.CloudAccountSummary{
		ByProvider:   make(map[string]int),
		ByStatus:     make(map[string]int),
		BackupStatus: make(map[string]int),
	}
	for _, account := range accounts {
		summary.Accounts++
		summary.TotalMonthlyCostUSD += account.MonthlyCostUSD
		summary.TotalBudgetUSD += account.BudgetUSD
		summary.AverageComplianceScore += float64(account.ComplianceScore)
		summary.OpenSecurityFindings += account.OpenSecurityFindings
		summary.ByProvider[strings.ToLower(account.Provider)]++
		summary.ByStatus[strings.ToLower(account.Status)]++
		summary.BackupStatus[strings.ToLower(account.BackupStatus)]++
	}

	if summary.Accounts > 0 {
		summary.AverageComplianceScore = roundFloat(summary.AverageComplianceScore / float64(summary.Accounts))
	}
	if summary.TotalBudgetUSD > 0 {
		summary.BudgetUtilization = roundFloat(summary.TotalMonthlyCostUSD / summary.TotalBudgetUSD * 100)
	}
	summary.TotalMonthlyCostUSD = roundFloat(summary.TotalMonthlyCostUSD)
	summary.TotalBudgetUSD = roundFloat(summary.TotalBudgetUSD)

	return summary, nil
}

func (s *CloudAccountService) GetPolicyViolations(ctx context.Context, filter model.CloudAccountFilter) ([]model.CloudPolicyViolation, error) {
	accounts, err := s.GetCloudAccounts(ctx, filter)
	if err != nil {
		return nil, err
	}

	violations := make([]model.CloudPolicyViolation, 0)
	for _, account := range accounts {
		violations = append(violations, cloudPolicyViolationsForAccount(account)...)
	}

	return violations, nil
}

func (s *CloudAccountService) GetRemediationPlan(ctx context.Context, filter model.CloudAccountFilter) (model.CloudRemediationPlan, error) {
	violations, err := s.GetPolicyViolations(ctx, filter)
	if err != nil {
		return model.CloudRemediationPlan{}, err
	}

	plan := model.CloudRemediationPlan{
		Actions:     make([]model.CloudRemediationAction, 0, len(violations)),
		GeneratedAt: time.Now().UTC(),
	}
	seen := make(map[string]struct{})
	for _, violation := range violations {
		key := violation.AccountID + ":" + violation.Policy
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}

		action := model.CloudRemediationAction{
			AccountID:       violation.AccountID,
			Action:          remediationActionForPolicy(violation.Policy),
			Priority:        violation.Severity,
			OwnerTeam:       violation.OwnerTeam,
			EstimatedImpact: violation.Remediation,
			SavingsUSD:      violation.EstimatedWaste,
		}
		if violation.Policy == "security-findings" {
			plan.SecurityFindingDelta++
		}
		plan.EstimatedSavingsUSD += violation.EstimatedWaste
		plan.Actions = append(plan.Actions, action)
	}
	plan.EstimatedSavingsUSD = roundFloat(plan.EstimatedSavingsUSD)

	return plan, nil
}

func cloudAccountFromRequest(id string, request model.CreateCloudAccountRequest) (model.CloudAccount, error) {
	request.Name = strings.TrimSpace(request.Name)
	request.Provider = strings.ToLower(strings.TrimSpace(request.Provider))
	request.AccountID = strings.TrimSpace(request.AccountID)
	request.Region = strings.TrimSpace(request.Region)
	request.OwnerTeam = strings.ToLower(strings.TrimSpace(request.OwnerTeam))
	request.Environment = strings.ToLower(strings.TrimSpace(request.Environment))
	request.Status = strings.ToLower(strings.TrimSpace(request.Status))
	request.BackupStatus = strings.ToLower(strings.TrimSpace(request.BackupStatus))

	if request.Name == "" || request.Provider == "" || request.AccountID == "" || request.Region == "" || request.OwnerTeam == "" || request.Environment == "" {
		return model.CloudAccount{}, ErrInvalidCloudAccount
	}
	if request.Status == "" {
		request.Status = "active"
	}
	if request.BackupStatus == "" {
		request.BackupStatus = "protected"
	}
	if !isValidCloudProvider(request.Provider) || !isValidCloudAccountStatus(request.Status) || !isValidCloudBackupStatus(request.BackupStatus) {
		return model.CloudAccount{}, ErrInvalidCloudAccount
	}
	if request.MonthlyCostUSD < 0 || request.BudgetUSD < 0 || request.ComplianceScore < 0 || request.ComplianceScore > 100 || request.OpenSecurityFindings < 0 {
		return model.CloudAccount{}, ErrInvalidCloudAccount
	}

	return model.CloudAccount{
		ID:                   id,
		Name:                 request.Name,
		Provider:             request.Provider,
		AccountID:            request.AccountID,
		Region:               request.Region,
		OwnerTeam:            request.OwnerTeam,
		Environment:          request.Environment,
		Status:               request.Status,
		MonthlyCostUSD:       roundFloat(request.MonthlyCostUSD),
		BudgetUSD:            roundFloat(request.BudgetUSD),
		ComplianceScore:      request.ComplianceScore,
		BackupStatus:         request.BackupStatus,
		OpenSecurityFindings: request.OpenSecurityFindings,
		Tags:                 normalizeTags(request.Tags),
	}, nil
}

func isValidCloudProvider(provider string) bool {
	return provider == "aws" || provider == "gcp" || provider == "azure" || provider == "onprem"
}

func isValidCloudAccountStatus(status string) bool {
	return status == "active" || status == "restricted" || status == "suspended" || status == "decommissioning"
}

func isValidCloudBackupStatus(status string) bool {
	return status == "protected" || status == "partial" || status == "missing"
}

func cloudAccountMatchesQuery(account model.CloudAccount, query string) bool {
	query = strings.ToLower(query)
	return strings.Contains(strings.ToLower(account.ID), query) ||
		strings.Contains(strings.ToLower(account.Name), query) ||
		strings.Contains(strings.ToLower(account.Provider), query) ||
		strings.Contains(strings.ToLower(account.AccountID), query) ||
		strings.Contains(strings.ToLower(account.Region), query) ||
		strings.Contains(strings.ToLower(account.OwnerTeam), query) ||
		strings.Contains(strings.ToLower(account.Environment), query) ||
		strings.Contains(strings.ToLower(account.Status), query)
}

func containsCloudAccountTag(tags []string, tag string) bool {
	for _, candidate := range tags {
		if strings.EqualFold(candidate, tag) {
			return true
		}
	}
	return false
}

func roundFloat(value float64) float64 {
	return math.Round(value*100) / 100
}

func cloudAccountID(provider string, accountID string) string {
	replacer := strings.NewReplacer(" ", "-", "_", "-", ".", "-", "/", "-", ":", "-")
	normalizedAccountID := strings.ToLower(replacer.Replace(strings.TrimSpace(accountID)))
	normalizedProvider := strings.ToLower(replacer.Replace(strings.TrimSpace(provider)))
	return fmt.Sprintf("cloud-%s-%s", normalizedProvider, normalizedAccountID)
}

func cloudPolicyViolationsForAccount(account model.CloudAccount) []model.CloudPolicyViolation {
	violations := []model.CloudPolicyViolation{}
	budgetUtilization := 0.0
	if account.BudgetUSD > 0 {
		budgetUtilization = account.MonthlyCostUSD / account.BudgetUSD * 100
	}

	if budgetUtilization >= 90 {
		violations = append(violations, cloudPolicyViolation(account, "budget-utilization", severityForThreshold(budgetUtilization, 110, 95), fmt.Sprintf("Budget utilization is %.1f%%", budgetUtilization), "Right-size workloads and review reserved capacity commitments.", roundFloat(math.Max(0, account.MonthlyCostUSD-account.BudgetUSD*0.85))))
	}
	if account.ComplianceScore < 80 {
		violations = append(violations, cloudPolicyViolation(account, "compliance-score", severityForThreshold(float64(80-account.ComplianceScore), 20, 10), fmt.Sprintf("Compliance score is %d", account.ComplianceScore), "Run CIS control remediation and re-run cloud posture checks.", 0))
	}
	if account.BackupStatus != "protected" {
		violations = append(violations, cloudPolicyViolation(account, "backup-posture", severityForBackup(account.BackupStatus), "Backup posture is "+account.BackupStatus, "Enable protected backup policy for critical resources.", 0))
	}
	if account.OpenSecurityFindings > 0 {
		violations = append(violations, cloudPolicyViolation(account, "security-findings", severityForFindings(account.OpenSecurityFindings), fmt.Sprintf("%d open security findings", account.OpenSecurityFindings), "Triage critical findings and close public exposure risks first.", 0))
	}
	if account.Status == "restricted" || account.Status == "suspended" {
		violations = append(violations, cloudPolicyViolation(account, "account-status", "high", "Account status is "+account.Status, "Review restrictions and restore approved operating state.", 0))
	}

	return violations
}

func cloudPolicyViolation(account model.CloudAccount, policy string, severity string, message string, remediation string, estimatedWaste float64) model.CloudPolicyViolation {
	return model.CloudPolicyViolation{
		ID:             fmt.Sprintf("%s-%s", account.ID, policy),
		AccountID:      account.ID,
		AccountName:    account.Name,
		Provider:       account.Provider,
		Region:         account.Region,
		OwnerTeam:      account.OwnerTeam,
		Policy:         policy,
		Severity:       severity,
		Message:        message,
		Remediation:    remediation,
		EstimatedWaste: estimatedWaste,
	}
}

func severityForThreshold(value float64, highThreshold float64, mediumThreshold float64) string {
	if value >= highThreshold {
		return "high"
	}
	if value >= mediumThreshold {
		return "medium"
	}
	return "low"
}

func severityForBackup(status string) string {
	if status == "missing" {
		return "high"
	}
	return "medium"
}

func severityForFindings(findings int) string {
	if findings >= 5 {
		return "high"
	}
	if findings >= 2 {
		return "medium"
	}
	return "low"
}

func remediationActionForPolicy(policy string) string {
	switch policy {
	case "budget-utilization":
		return "optimize-cost"
	case "compliance-score":
		return "run-compliance-remediation"
	case "backup-posture":
		return "enforce-backup-policy"
	case "security-findings":
		return "triage-security-findings"
	case "account-status":
		return "review-account-restrictions"
	default:
		return "review-policy"
	}
}
