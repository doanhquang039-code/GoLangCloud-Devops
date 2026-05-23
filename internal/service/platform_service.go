package service

import (
	"context"
	"math"
	"sort"

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
		Applications: len(applications),
		Clusters:     len(clusters),
		Environments: len(environments),
		Deployments:  len(deployments),
		PipelineRuns: len(pipelineRuns),
		Incidents:    len(incidents),
		OpenIncidents: openIncidents,
		ByStatus:     byStatus,
	}, nil
}

func (s *PlatformService) GetScorecards(ctx context.Context) ([]model.PlatformScorecard, error) {
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
			Criticality:    application.Criticality,
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

	sort.Slice(scorecards, func(i, j int) bool {
		if scorecards[i].OperationalReadinessScore == scorecards[j].OperationalReadinessScore {
			return scorecards[i].ApplicationName < scorecards[j].ApplicationName
		}
		return scorecards[i].OperationalReadinessScore < scorecards[j].OperationalReadinessScore
	})

	return scorecards, nil
}

func (s *PlatformService) GetEnvironmentDriftReports(ctx context.Context) ([]model.EnvironmentDriftReport, error) {
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

	sort.Slice(reports, func(i, j int) bool {
		if reports[i].DriftScore == reports[j].DriftScore {
			return reports[i].EnvironmentName < reports[j].EnvironmentName
		}
		return reports[i].DriftScore < reports[j].DriftScore
	})

	return reports, nil
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
