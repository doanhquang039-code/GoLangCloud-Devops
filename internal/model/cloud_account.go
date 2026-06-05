package model

import "time"

type CloudAccount struct {
	ID                   string    `json:"id" bson:"id"`
	Name                 string    `json:"name" bson:"name"`
	Provider             string    `json:"provider" bson:"provider"`
	AccountID            string    `json:"account_id" bson:"account_id"`
	Region               string    `json:"region" bson:"region"`
	OwnerTeam            string    `json:"owner_team" bson:"owner_team"`
	Environment          string    `json:"environment" bson:"environment"`
	Status               string    `json:"status" bson:"status"`
	MonthlyCostUSD       float64   `json:"monthly_cost_usd" bson:"monthly_cost_usd"`
	BudgetUSD            float64   `json:"budget_usd" bson:"budget_usd"`
	ComplianceScore      int       `json:"compliance_score" bson:"compliance_score"`
	BackupStatus         string    `json:"backup_status" bson:"backup_status"`
	OpenSecurityFindings int       `json:"open_security_findings" bson:"open_security_findings"`
	Tags                 []string  `json:"tags,omitempty" bson:"tags,omitempty"`
	CreatedAt            time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt            time.Time `json:"updated_at" bson:"updated_at"`
}

type CloudAccountFilter struct {
	Query        string
	Provider     string
	Region       string
	OwnerTeam    string
	Environment  string
	Status       string
	BackupStatus string
	Tag          string
}

type CreateCloudAccountRequest struct {
	Name                 string   `json:"name"`
	Provider             string   `json:"provider"`
	AccountID            string   `json:"account_id"`
	Region               string   `json:"region"`
	OwnerTeam            string   `json:"owner_team"`
	Environment          string   `json:"environment"`
	Status               string   `json:"status"`
	MonthlyCostUSD       float64  `json:"monthly_cost_usd"`
	BudgetUSD            float64  `json:"budget_usd"`
	ComplianceScore      int      `json:"compliance_score"`
	BackupStatus         string   `json:"backup_status"`
	OpenSecurityFindings int      `json:"open_security_findings"`
	Tags                 []string `json:"tags"`
}

type UpdateCloudAccountRequest CreateCloudAccountRequest

type UpdateCloudAccountStatusRequest struct {
	Status       string `json:"status"`
	BackupStatus string `json:"backup_status"`
}

type CloudAccountSummary struct {
	Accounts               int            `json:"accounts"`
	TotalMonthlyCostUSD    float64        `json:"total_monthly_cost_usd"`
	TotalBudgetUSD         float64        `json:"total_budget_usd"`
	BudgetUtilization      float64        `json:"budget_utilization"`
	AverageComplianceScore float64        `json:"average_compliance_score"`
	OpenSecurityFindings   int            `json:"open_security_findings"`
	ByProvider             map[string]int `json:"by_provider"`
	ByStatus               map[string]int `json:"by_status"`
	BackupStatus           map[string]int `json:"backup_status"`
}

type CloudPolicyViolation struct {
	ID             string  `json:"id"`
	AccountID      string  `json:"account_id"`
	AccountName    string  `json:"account_name"`
	Provider       string  `json:"provider"`
	Region         string  `json:"region"`
	OwnerTeam      string  `json:"owner_team"`
	Policy         string  `json:"policy"`
	Severity       string  `json:"severity"`
	Message        string  `json:"message"`
	Remediation    string  `json:"remediation"`
	EstimatedWaste float64 `json:"estimated_waste_usd,omitempty"`
}

type CloudRemediationAction struct {
	AccountID       string  `json:"account_id"`
	Action          string  `json:"action"`
	Priority        string  `json:"priority"`
	OwnerTeam       string  `json:"owner_team"`
	EstimatedImpact string  `json:"estimated_impact"`
	SavingsUSD      float64 `json:"savings_usd,omitempty"`
}

type CloudRemediationPlan struct {
	Actions              []CloudRemediationAction `json:"actions"`
	EstimatedSavingsUSD  float64                  `json:"estimated_savings_usd"`
	SecurityFindingDelta int                      `json:"security_finding_delta"`
	GeneratedAt          time.Time                `json:"generated_at"`
}
