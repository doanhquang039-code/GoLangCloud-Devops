package service

import (
	"context"
	"math"
	"sort"
	"strings"

	"hr-cloud-service/internal/model"
	"hr-cloud-service/internal/repository"
)

type PlatformService struct {
	applicationRepository repository.ApplicationRepository
	clusterRepository     repository.ClusterRepository
	environmentRepository repository.EnvironmentRepository
	deploymentRepository  repository.DeploymentRepository
	pipelineRepository    repository.PipelineRepository
	incidentRepository    repository.IncidentRepository
}

func NewPlatformService(
	applicationRepository repository.ApplicationRepository,
	clusterRepository repository.ClusterRepository,
	environmentRepository repository.EnvironmentRepository,
	deploymentRepository repository.DeploymentRepository,
	pipelineRepository repository.PipelineRepository,
	incidentRepository repository.IncidentRepository,
) *PlatformService {
	return &PlatformService{
		applicationRepository: applicationRepository,
		clusterRepository:     clusterRepository,
		environmentRepository: environmentRepository,
		deploymentRepository:  deploymentRepository,
		pipelineRepository:    pipelineRepository,
		incidentRepository:    incidentRepository,
	}
}

func (s *PlatformService) GetSummary(ctx context.Context) (model.PlatformSummary, error) {
	applications, err := s.applicationRepository.FindAll(ctx)
	if err != nil {
		return model.PlatformSummary{}, err
	}

	clusters, err := s.clusterRepository.FindAll(ctx)
	if err != nil {
		return model.PlatformSummary{}, err
	}

	environments, err := s.environmentRepository.FindAll(ctx)
	if err != nil {
		return model.PlatformSummary{}, err
	}

	deployments, err := s.deploymentRepository.FindAll(ctx)
	if err != nil {
		return model.PlatformSummary{}, err
	}

	pipelineRuns, err := s.pipelineRepository.FindAll(ctx)
	if err != nil {
		return model.PlatformSummary{}, err
	}

	incidents, err := s.incidentRepository.FindAll(ctx)
	if err != nil {
		return model.PlatformSummary{}, err
	}

	byStatus := map[string]int{}
	for _, deployment := range deployments {
		byStatus[deployment.Status]++
	}

	openIncidents := 0
	for _, incident := range incidents {
		if incident.Status != "resolved" {
			openIncidents++
		}
	}

	return model.PlatformSummary{
		Applications:  len(applications),
		Clusters:      len(clusters),
		Environments:  len(environments),
		Deployments:   len(deployments),
		PipelineRuns:  len(pipelineRuns),
		Incidents:     len(incidents),
		OpenIncidents: openIncidents,
		ByStatus:      byStatus,
	}, nil
}

func (s *PlatformService) GetScorecards(ctx context.Context, filters ...model.PlatformScorecardFilter) ([]model.PlatformScorecard, error) {
	filter := model.PlatformScorecardFilter{}
	if len(filters) > 0 {
		filter = normalizePlatformScorecardFilter(filters[0])
	}

	applications, err := s.applicationRepository.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	environments, err := s.environmentRepository.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	deployments, err := s.deploymentRepository.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	pipelineRuns, err := s.pipelineRepository.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	incidents, err := s.incidentRepository.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	scorecards := make([]model.PlatformScorecard, 0, len(applications))
	for _, application := range applications {
		scorecard := model.PlatformScorecard{
			ApplicationID:   application.ID,
			ApplicationName: application.Name,
			OwnerTeam:       application.OwnerTeam,
			Criticality:     application.Criticality,
		}

		clusterIDs := map[string]struct{}{}
		for _, environment := range environments {
			if environment.ApplicationID != application.ID {
				continue
			}
			scorecard.EnvironmentCount++
			if environment.Status == "active" {
				scorecard.ActiveEnvironmentCount++
			}
			if environment.ClusterID != "" {
				clusterIDs[environment.ClusterID] = struct{}{}
			}
		}
		scorecard.ClusterCount = len(clusterIDs)

		var deploymentDurationMinutes float64
		var finishedDeployments int
		for _, deployment := range deployments {
			if deployment.ApplicationID != application.ID {
				continue
			}
			scorecard.DeploymentCount++
			switch deployment.Status {
			case "succeeded":
				scorecard.SuccessfulDeploymentCount++
			case "failed":
				scorecard.FailedDeploymentCount++
			case "running":
				scorecard.RunningDeploymentCount++
			}
			if deployment.FinishedAt != nil {
				finishedDeployments++
				deploymentDurationMinutes += deployment.FinishedAt.Sub(deployment.StartedAt).Minutes()
			}
		}
		scorecard.DeploymentSuccessRate = successRate(scorecard.SuccessfulDeploymentCount, scorecard.DeploymentCount)
		scorecard.AverageDeploymentDurationMinutes = averageMinutes(deploymentDurationMinutes, finishedDeployments)

		var pipelineDurationMinutes float64
		var finishedPipelineRuns int
		for _, pipelineRun := range pipelineRuns {
			if pipelineRun.ApplicationID != application.ID {
				continue
			}
			scorecard.PipelineRunCount++
			switch pipelineRun.Status {
			case "succeeded":
				scorecard.SuccessfulPipelineRunCount++
			case "failed":
				scorecard.FailedPipelineRunCount++
			case "running":
				scorecard.RunningPipelineRunCount++
			}
			if pipelineRun.FinishedAt != nil {
				finishedPipelineRuns++
				pipelineDurationMinutes += pipelineRun.FinishedAt.Sub(pipelineRun.StartedAt).Minutes()
			}
		}
		scorecard.PipelineSuccessRate = successRate(scorecard.SuccessfulPipelineRunCount, scorecard.PipelineRunCount)
		scorecard.AveragePipelineRunDurationMinutes = averageMinutes(pipelineDurationMinutes, finishedPipelineRuns)

		var resolveMinutes float64
		var resolvedIncidents int
		for _, incident := range incidents {
			if incident.ApplicationID != application.ID {
				continue
			}
			scorecard.IncidentCount++
			if incident.Status != "resolved" {
				scorecard.OpenIncidentCount++
			}
			switch incident.Severity {
			case "sev1":
				scorecard.Sev1IncidentCount++
			case "sev2":
				scorecard.Sev2IncidentCount++
			}
			if incident.ResolvedAt != nil {
				resolvedIncidents++
				resolveMinutes += incident.ResolvedAt.Sub(incident.CreatedAt).Minutes()
			}
		}
		scorecard.MeanTimeToResolveMinutes = averageMinutes(resolveMinutes, resolvedIncidents)
		scorecard.OperationalReadinessScore, scorecard.RiskLevel, scorecard.RiskReasons = scoreOperationalReadiness(scorecard)
		scorecards = append(scorecards, scorecard)
	}

	scorecards = filterPlatformScorecards(scorecards, filter)
	sortPlatformScorecards(scorecards, filter.SortBy, filter.SortOrder)

	return scorecards, nil
}

func (s *PlatformService) GetEnvironmentDriftReports(ctx context.Context, filters ...model.EnvironmentDriftReportFilter) ([]model.EnvironmentDriftReport, error) {
	filter := model.EnvironmentDriftReportFilter{}
	if len(filters) > 0 {
		filter = normalizeEnvironmentDriftReportFilter(filters[0])
	}

	applications, err := s.applicationRepository.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	environments, err := s.environmentRepository.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	applicationsByID := make(map[string]model.Application, len(applications))
	for _, application := range applications {
		applicationsByID[application.ID] = application
	}

	reports := make([]model.EnvironmentDriftReport, 0, len(environments))
	for _, environment := range environments {
		application, ok := applicationsByID[environment.ApplicationID]
		if !ok {
			reports = append(reports, model.EnvironmentDriftReport{
				EnvironmentID:   environment.ID,
				EnvironmentName: environment.Name,
				EnvironmentType: environment.Type,
				ApplicationID:   environment.ApplicationID,
				ClusterID:       environment.ClusterID,
				Namespace:       environment.Namespace,
				Status:          environment.Status,
				DriftScore:      0,
				DriftLevel:      "critical",
				DriftReasons:    []string{"application reference is missing"},
			})
			continue
		}

		report := model.EnvironmentDriftReport{
			EnvironmentID:   environment.ID,
			EnvironmentName: environment.Name,
			EnvironmentType: environment.Type,
			ApplicationID:   application.ID,
			ApplicationName: application.Name,
			ClusterID:       environment.ClusterID,
			Namespace:       environment.Namespace,
			Status:          environment.Status,
		}

		for key, expectedValue := range application.Environment {
			actualValue, exists := environment.Variables[key]
			if !exists {
				report.MissingVariables = append(report.MissingVariables, model.EnvironmentVariableDrift{
					Key:           key,
					ExpectedValue: expectedValue,
				})
				continue
			}
			if actualValue != expectedValue {
				report.ChangedVariables = append(report.ChangedVariables, model.EnvironmentVariableDrift{
					Key:           key,
					ExpectedValue: expectedValue,
					ActualValue:   actualValue,
				})
			}
		}

		for key, actualValue := range environment.Variables {
			if _, exists := application.Environment[key]; exists {
				continue
			}
			report.ExtraVariables = append(report.ExtraVariables, model.EnvironmentVariableDrift{
				Key:         key,
				ActualValue: actualValue,
			})
		}

		sortEnvironmentVariableDrift(report.MissingVariables)
		sortEnvironmentVariableDrift(report.ChangedVariables)
		sortEnvironmentVariableDrift(report.ExtraVariables)
		report.DriftScore, report.DriftLevel, report.DriftReasons = scoreEnvironmentDrift(report)
		reports = append(reports, report)
	}

	reports = filterEnvironmentDriftReports(reports, filter)
	sortEnvironmentDriftReports(reports, filter.SortBy, filter.SortOrder)

	return reports, nil
}

func normalizePlatformScorecardFilter(filter model.PlatformScorecardFilter) model.PlatformScorecardFilter {
	filter.Query = strings.TrimSpace(filter.Query)
	filter.OwnerTeam = strings.TrimSpace(filter.OwnerTeam)
	filter.Criticality = strings.ToLower(strings.TrimSpace(filter.Criticality))
	filter.RiskLevel = strings.ToLower(strings.TrimSpace(filter.RiskLevel))
	filter.SortBy = strings.ToLower(strings.TrimSpace(filter.SortBy))
	filter.SortOrder = strings.ToLower(strings.TrimSpace(filter.SortOrder))
	return filter
}

func filterPlatformScorecards(scorecards []model.PlatformScorecard, filter model.PlatformScorecardFilter) []model.PlatformScorecard {
	if filter.Query == "" && filter.OwnerTeam == "" && filter.Criticality == "" && filter.RiskLevel == "" && filter.MinScore == 0 {
		return scorecards
	}
	filtered := make([]model.PlatformScorecard, 0, len(scorecards))
	for _, scorecard := range scorecards {
		if filter.Query != "" && !platformScorecardMatchesQuery(scorecard, filter.Query) {
			continue
		}
		if filter.OwnerTeam != "" && !strings.EqualFold(scorecard.OwnerTeam, filter.OwnerTeam) {
			continue
		}
		if filter.Criticality != "" && !strings.EqualFold(scorecard.Criticality, filter.Criticality) {
			continue
		}
		if filter.RiskLevel != "" && !strings.EqualFold(scorecard.RiskLevel, filter.RiskLevel) {
			continue
		}
		if filter.MinScore > 0 && scorecard.OperationalReadinessScore < filter.MinScore {
			continue
		}
		filtered = append(filtered, scorecard)
	}
	return filtered
}

func platformScorecardMatchesQuery(scorecard model.PlatformScorecard, query string) bool {
	query = strings.ToLower(query)
	if strings.Contains(strings.ToLower(scorecard.ApplicationID), query) ||
		strings.Contains(strings.ToLower(scorecard.ApplicationName), query) ||
		strings.Contains(strings.ToLower(scorecard.OwnerTeam), query) ||
		strings.Contains(strings.ToLower(scorecard.Criticality), query) ||
		strings.Contains(strings.ToLower(scorecard.RiskLevel), query) {
		return true
	}
	for _, reason := range scorecard.RiskReasons {
		if strings.Contains(strings.ToLower(reason), query) {
			return true
		}
	}
	return false
}

func sortPlatformScorecards(scorecards []model.PlatformScorecard, sortBy string, sortOrder string) {
	desc := sortOrder == "desc"
	sort.Slice(scorecards, func(i, j int) bool {
		less, greater := false, false
		switch sortBy {
		case "name":
			less = scorecards[i].ApplicationName < scorecards[j].ApplicationName
			greater = scorecards[i].ApplicationName > scorecards[j].ApplicationName
		case "owner_team":
			less = scorecards[i].OwnerTeam < scorecards[j].OwnerTeam
			greater = scorecards[i].OwnerTeam > scorecards[j].OwnerTeam
		case "risk":
			less = scorecards[i].RiskLevel < scorecards[j].RiskLevel
			greater = scorecards[i].RiskLevel > scorecards[j].RiskLevel
		default:
			if scorecards[i].OperationalReadinessScore == scorecards[j].OperationalReadinessScore {
				less = scorecards[i].ApplicationName < scorecards[j].ApplicationName
				greater = scorecards[i].ApplicationName > scorecards[j].ApplicationName
			} else {
				less = scorecards[i].OperationalReadinessScore < scorecards[j].OperationalReadinessScore
				greater = scorecards[i].OperationalReadinessScore > scorecards[j].OperationalReadinessScore
			}
		}
		if desc {
			return greater
		}
		return less
	})
}

func normalizeEnvironmentDriftReportFilter(filter model.EnvironmentDriftReportFilter) model.EnvironmentDriftReportFilter {
	filter.Query = strings.TrimSpace(filter.Query)
	filter.ApplicationID = strings.TrimSpace(filter.ApplicationID)
	filter.EnvironmentType = strings.ToLower(strings.TrimSpace(filter.EnvironmentType))
	filter.Status = strings.ToLower(strings.TrimSpace(filter.Status))
	filter.DriftLevel = strings.ToLower(strings.TrimSpace(filter.DriftLevel))
	filter.SortBy = strings.ToLower(strings.TrimSpace(filter.SortBy))
	filter.SortOrder = strings.ToLower(strings.TrimSpace(filter.SortOrder))
	return filter
}

func filterEnvironmentDriftReports(reports []model.EnvironmentDriftReport, filter model.EnvironmentDriftReportFilter) []model.EnvironmentDriftReport {
	if filter.Query == "" && filter.ApplicationID == "" && filter.EnvironmentType == "" && filter.Status == "" && filter.DriftLevel == "" && filter.MaxDriftScore == 0 {
		return reports
	}
	filtered := make([]model.EnvironmentDriftReport, 0, len(reports))
	for _, report := range reports {
		if filter.Query != "" && !environmentDriftReportMatchesQuery(report, filter.Query) {
			continue
		}
		if filter.ApplicationID != "" && report.ApplicationID != filter.ApplicationID {
			continue
		}
		if filter.EnvironmentType != "" && !strings.EqualFold(report.EnvironmentType, filter.EnvironmentType) {
			continue
		}
		if filter.Status != "" && !strings.EqualFold(report.Status, filter.Status) {
			continue
		}
		if filter.DriftLevel != "" && !strings.EqualFold(report.DriftLevel, filter.DriftLevel) {
			continue
		}
		if filter.MaxDriftScore > 0 && report.DriftScore > filter.MaxDriftScore {
			continue
		}
		filtered = append(filtered, report)
	}
	return filtered
}

func environmentDriftReportMatchesQuery(report model.EnvironmentDriftReport, query string) bool {
	query = strings.ToLower(query)
	if strings.Contains(strings.ToLower(report.EnvironmentID), query) ||
		strings.Contains(strings.ToLower(report.EnvironmentName), query) ||
		strings.Contains(strings.ToLower(report.ApplicationID), query) ||
		strings.Contains(strings.ToLower(report.ApplicationName), query) ||
		strings.Contains(strings.ToLower(report.ClusterID), query) ||
		strings.Contains(strings.ToLower(report.Namespace), query) ||
		strings.Contains(strings.ToLower(report.DriftLevel), query) {
		return true
	}
	return driftListMatchesQuery(report.MissingVariables, query) ||
		driftListMatchesQuery(report.ChangedVariables, query) ||
		driftListMatchesQuery(report.ExtraVariables, query)
}

func driftListMatchesQuery(drifts []model.EnvironmentVariableDrift, query string) bool {
	for _, drift := range drifts {
		if strings.Contains(strings.ToLower(drift.Key), query) ||
			strings.Contains(strings.ToLower(drift.ExpectedValue), query) ||
			strings.Contains(strings.ToLower(drift.ActualValue), query) {
			return true
		}
	}
	return false
}

func sortEnvironmentDriftReports(reports []model.EnvironmentDriftReport, sortBy string, sortOrder string) {
	desc := sortOrder == "desc"
	sort.Slice(reports, func(i, j int) bool {
		less, greater := false, false
		switch sortBy {
		case "name":
			less = reports[i].EnvironmentName < reports[j].EnvironmentName
			greater = reports[i].EnvironmentName > reports[j].EnvironmentName
		case "application":
			less = reports[i].ApplicationName < reports[j].ApplicationName
			greater = reports[i].ApplicationName > reports[j].ApplicationName
		case "level":
			less = reports[i].DriftLevel < reports[j].DriftLevel
			greater = reports[i].DriftLevel > reports[j].DriftLevel
		default:
			if reports[i].DriftScore == reports[j].DriftScore {
				less = reports[i].EnvironmentName < reports[j].EnvironmentName
				greater = reports[i].EnvironmentName > reports[j].EnvironmentName
			} else {
				less = reports[i].DriftScore < reports[j].DriftScore
				greater = reports[i].DriftScore > reports[j].DriftScore
			}
		}
		if desc {
			return greater
		}
		return less
	})
}

func successRate(successful int, total int) float64 {
	if total == 0 {
		return 0
	}
	return roundOneDecimal(float64(successful) / float64(total) * 100)
}

func averageMinutes(totalMinutes float64, count int) float64 {
	if count == 0 {
		return 0
	}
	return roundOneDecimal(totalMinutes / float64(count))
}

func roundOneDecimal(value float64) float64 {
	return math.Round(value*10) / 10
}

func scoreOperationalReadiness(scorecard model.PlatformScorecard) (int, string, []string) {
	score := 100
	reasons := []string{}

	if scorecard.ActiveEnvironmentCount == 0 {
		score -= 20
		reasons = append(reasons, "no active environment")
	}
	if scorecard.ClusterCount == 0 {
		score -= 15
		reasons = append(reasons, "no target cluster coverage")
	}
	if scorecard.OpenIncidentCount > 0 {
		score -= min(scorecard.OpenIncidentCount*15, 30)
		reasons = append(reasons, "open incidents")
	}
	if scorecard.Sev1IncidentCount > 0 {
		score -= 25
		reasons = append(reasons, "sev1 incident history")
	}
	if scorecard.Sev2IncidentCount > 0 {
		score -= 15
		reasons = append(reasons, "sev2 incident history")
	}
	if scorecard.DeploymentCount == 0 {
		score -= 10
		reasons = append(reasons, "no deployment history")
	} else if scorecard.DeploymentSuccessRate < 80 {
		score -= 15
		reasons = append(reasons, "deployment success rate below 80 percent")
	}
	if scorecard.PipelineRunCount == 0 {
		score -= 10
		reasons = append(reasons, "no pipeline run history")
	} else if scorecard.PipelineSuccessRate < 80 {
		score -= 15
		reasons = append(reasons, "pipeline success rate below 80 percent")
	}

	if score < 0 {
		score = 0
	}

	switch {
	case score >= 85:
		return score, "low", reasons
	case score >= 65:
		return score, "medium", reasons
	default:
		return score, "high", reasons
	}
}

func sortEnvironmentVariableDrift(drifts []model.EnvironmentVariableDrift) {
	sort.Slice(drifts, func(i, j int) bool {
		return drifts[i].Key < drifts[j].Key
	})
}

func scoreEnvironmentDrift(report model.EnvironmentDriftReport) (int, string, []string) {
	score := 100
	reasons := []string{}

	if len(report.MissingVariables) > 0 {
		score -= min(len(report.MissingVariables)*25, 50)
		reasons = append(reasons, "missing baseline variables")
	}
	if len(report.ChangedVariables) > 0 {
		score -= min(len(report.ChangedVariables)*20, 40)
		reasons = append(reasons, "changed baseline variables")
	}
	if len(report.ExtraVariables) > 0 {
		score -= min(len(report.ExtraVariables)*5, 20)
		reasons = append(reasons, "extra runtime variables")
	}
	if report.Status != "active" {
		score -= 10
		reasons = append(reasons, "environment is not active")
	}
	if score < 0 {
		score = 0
	}

	switch {
	case score >= 95:
		return score, "none", reasons
	case score >= 80:
		return score, "low", reasons
	case score >= 60:
		return score, "medium", reasons
	default:
		return score, "high", reasons
	}
}
