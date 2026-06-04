package controller

import (
	"net/http"
	"strconv"

	"hr-cloud-service/internal/model"
	"hr-cloud-service/internal/service"
)

type PlatformController struct {
	platformService *service.PlatformService
}

func NewPlatformController(platformService *service.PlatformService) *PlatformController {
	return &PlatformController{platformService: platformService}
}

func (c *PlatformController) Summary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	summary, err := c.platformService.GetSummary(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not build platform summary")
		return
	}

	writeJSON(w, http.StatusOK, summary)
}

func (c *PlatformController) Scorecards(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	filter, ok := parsePlatformScorecardFilter(w, r)
	if !ok {
		return
	}

	scorecards, err := c.platformService.GetScorecards(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not build platform scorecards")
		return
	}

	writeJSON(w, http.StatusOK, scorecards)
}

func (c *PlatformController) EnvironmentDrift(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	filter, ok := parseEnvironmentDriftReportFilter(w, r)
	if !ok {
		return
	}

	reports, err := c.platformService.GetEnvironmentDriftReports(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not build environment drift report")
		return
	}

	writeJSON(w, http.StatusOK, reports)
}

func parsePlatformScorecardFilter(w http.ResponseWriter, r *http.Request) (model.PlatformScorecardFilter, bool) {
	query := r.URL.Query()
	filter := model.PlatformScorecardFilter{
		Query:       query.Get("q"),
		OwnerTeam:   query.Get("owner_team"),
		Criticality: query.Get("criticality"),
		RiskLevel:   query.Get("risk_level"),
		SortBy:      query.Get("sort"),
		SortOrder:   query.Get("order"),
	}
	if raw := query.Get("min_score"); raw != "" {
		minScore, err := strconv.Atoi(raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "min_score must be a number")
			return model.PlatformScorecardFilter{}, false
		}
		filter.MinScore = minScore
	}
	return filter, true
}

func parseEnvironmentDriftReportFilter(w http.ResponseWriter, r *http.Request) (model.EnvironmentDriftReportFilter, bool) {
	query := r.URL.Query()
	filter := model.EnvironmentDriftReportFilter{
		Query:           query.Get("q"),
		ApplicationID:   query.Get("application_id"),
		EnvironmentType: query.Get("type"),
		Status:          query.Get("status"),
		DriftLevel:      query.Get("drift_level"),
		SortBy:          query.Get("sort"),
		SortOrder:       query.Get("order"),
	}
	if raw := query.Get("max_drift_score"); raw != "" {
		maxScore, err := strconv.Atoi(raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "max_drift_score must be a number")
			return model.EnvironmentDriftReportFilter{}, false
		}
		filter.MaxDriftScore = maxScore
	}
	return filter, true
}
