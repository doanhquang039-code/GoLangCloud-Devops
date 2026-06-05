package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"hr-cloud-service/internal/model"
	"hr-cloud-service/internal/repository"
	"hr-cloud-service/internal/service"
)

type CloudAccountController struct {
	cloudAccountService *service.CloudAccountService
}

func NewCloudAccountController(cloudAccountService *service.CloudAccountService) *CloudAccountController {
	return &CloudAccountController{cloudAccountService: cloudAccountService}
}

func (c *CloudAccountController) Index(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		c.GetCloudAccounts(w, r)
	case http.MethodPost:
		c.CreateCloudAccount(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (c *CloudAccountController) Summary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	summary, err := c.cloudAccountService.GetCloudAccountSummary(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not summarize cloud accounts")
		return
	}

	writeJSON(w, http.StatusOK, summary)
}

func (c *CloudAccountController) PolicyViolations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	violations, err := c.cloudAccountService.GetPolicyViolations(r.Context(), cloudAccountFilterFromRequest(r))
	if errors.Is(err, service.ErrInvalidCloudAccount) {
		writeError(w, http.StatusBadRequest, "invalid cloud account filter")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not evaluate cloud policies")
		return
	}

	writeJSON(w, http.StatusOK, violations)
}

func (c *CloudAccountController) RemediationPlan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	plan, err := c.cloudAccountService.GetRemediationPlan(r.Context(), cloudAccountFilterFromRequest(r))
	if errors.Is(err, service.ErrInvalidCloudAccount) {
		writeError(w, http.StatusBadRequest, "invalid cloud account filter")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not build cloud remediation plan")
		return
	}

	writeJSON(w, http.StatusOK, plan)
}

func (c *CloudAccountController) GetCloudAccounts(w http.ResponseWriter, r *http.Request) {
	accounts, err := c.cloudAccountService.GetCloudAccounts(r.Context(), cloudAccountFilterFromRequest(r))
	if errors.Is(err, service.ErrInvalidCloudAccount) {
		writeError(w, http.StatusBadRequest, "invalid cloud account filter")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list cloud accounts")
		return
	}

	writeJSON(w, http.StatusOK, accounts)
}

func (c *CloudAccountController) ShowOrUpdate(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/cloud-accounts/")

	switch r.Method {
	case http.MethodGet:
		c.Show(w, r, id)
	case http.MethodPut:
		c.UpdateCloudAccount(w, r, id)
	case http.MethodPatch:
		c.UpdateStatus(w, r, id)
	case http.MethodDelete:
		c.DeleteCloudAccount(w, r, id)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (c *CloudAccountController) Show(w http.ResponseWriter, r *http.Request, id string) {
	account, err := c.cloudAccountService.GetCloudAccountByID(r.Context(), id)
	if errors.Is(err, repository.ErrCloudAccountNotFound) {
		writeError(w, http.StatusNotFound, "cloud account not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, account)
}

func (c *CloudAccountController) CreateCloudAccount(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var request model.CreateCloudAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	account, err := c.cloudAccountService.CreateCloudAccount(r.Context(), request)
	if errors.Is(err, service.ErrInvalidCloudAccount) {
		writeError(w, http.StatusBadRequest, "name, provider, account_id, region, owner_team and environment are required")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create cloud account")
		return
	}

	writeJSON(w, http.StatusCreated, account)
}

func (c *CloudAccountController) UpdateCloudAccount(w http.ResponseWriter, r *http.Request, id string) {
	defer r.Body.Close()

	var request model.UpdateCloudAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	account, err := c.cloudAccountService.UpdateCloudAccount(r.Context(), id, request)
	if errors.Is(err, service.ErrInvalidCloudAccount) {
		writeError(w, http.StatusBadRequest, "invalid cloud account input")
		return
	}
	if errors.Is(err, repository.ErrCloudAccountNotFound) {
		writeError(w, http.StatusNotFound, "cloud account not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not update cloud account")
		return
	}

	writeJSON(w, http.StatusOK, account)
}

func (c *CloudAccountController) UpdateStatus(w http.ResponseWriter, r *http.Request, id string) {
	defer r.Body.Close()

	var request model.UpdateCloudAccountStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	account, err := c.cloudAccountService.UpdateCloudAccountStatus(r.Context(), id, request)
	if errors.Is(err, service.ErrInvalidCloudAccount) {
		writeError(w, http.StatusBadRequest, "status or backup_status is invalid")
		return
	}
	if errors.Is(err, repository.ErrCloudAccountNotFound) {
		writeError(w, http.StatusNotFound, "cloud account not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not update cloud account")
		return
	}

	writeJSON(w, http.StatusOK, account)
}

func (c *CloudAccountController) DeleteCloudAccount(w http.ResponseWriter, r *http.Request, id string) {
	err := c.cloudAccountService.DeleteCloudAccount(r.Context(), id)
	if errors.Is(err, service.ErrInvalidCloudAccount) {
		writeError(w, http.StatusBadRequest, "cloud account id is required")
		return
	}
	if errors.Is(err, repository.ErrCloudAccountNotFound) {
		writeError(w, http.StatusNotFound, "cloud account not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not delete cloud account")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func cloudAccountFilterFromRequest(r *http.Request) model.CloudAccountFilter {
	return model.CloudAccountFilter{
		Query:        r.URL.Query().Get("q"),
		Provider:     r.URL.Query().Get("provider"),
		Region:       r.URL.Query().Get("region"),
		OwnerTeam:    r.URL.Query().Get("owner_team"),
		Environment:  r.URL.Query().Get("environment"),
		Status:       r.URL.Query().Get("status"),
		BackupStatus: r.URL.Query().Get("backup_status"),
		Tag:          r.URL.Query().Get("tag"),
	}
}
