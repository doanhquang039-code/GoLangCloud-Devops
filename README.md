# HR Cloud DevOps Service

A small Go backend example for HR and DevOps-style service management. The project uses a traditional MVC-style structure, close to what you may know from Java/Spring:

- `cmd/api`: application entrypoint
- `internal/model`: data models and request DTOs
- `internal/controller`: HTTP controllers
- `internal/service`: business logic
- `internal/repository`: data access layer
- `internal/server`: route registration and server wiring
- `deploy/k8s`: Kubernetes manifests

Request flow:

```text
HTTP request -> Controller -> Service -> Repository -> Model
```

## Run

The service uses MongoDB. Your local MongoDB Compass screenshot shows MongoDB is already available at:

```text
mongodb://localhost:27017
```

Install Go, then run:

```powershell
cd D:\GoLang\hr-cloud-service
$env:MONGO_URI="mongodb://localhost:27017"
$env:MONGO_DATABASE="hr_cloud"
go run .\cmd\api
```

The API listens on `http://localhost:8080`.

## Domain

This project now has cloud and DevOps-oriented modules:

- Employees: HR data owned by a team.
- Applications: services/microservices managed by DevOps, including runtime, port, replicas, environment variables, health endpoint, and tags.
- Clusters: Kubernetes-style target clusters with provider, region, API endpoint, version, and operational status.
- Environments: runtime environments that map applications to clusters and namespaces, with environment variables and lifecycle status.
- Deployments: deployment records for each application, cluster, namespace, environment, version, and rollout strategy.
- Pipeline runs: CI/CD executions for each application, including branch, commit, stages, trigger owner, and final status.
- Incidents: operational incidents linked to applications, clusters, or deployments, with severity, status, owner team, and resolution time.
- Platform summary: lightweight operational counts grouped by deployment status.
- Platform scorecards: per-application operational readiness, release health, incident posture, environment coverage, and duration metrics.
- Environment drift detection: compares desired application variables with actual environment variables.
- Observability: Prometheus-compatible HTTP metrics for request counts and latency.

## Endpoints

```http
GET  /healthz
GET  /readyz
GET  /metrics

GET  /api/v1/employees
POST /api/v1/employees
GET  /api/v1/employees/{id}
PUT  /api/v1/employees/{id}
DELETE /api/v1/employees/{id}

GET  /api/v1/applications
POST /api/v1/applications
GET  /api/v1/applications/{id}
PUT  /api/v1/applications/{id}
DELETE /api/v1/applications/{id}

GET   /api/v1/clusters
POST  /api/v1/clusters
GET   /api/v1/clusters/{id}
PUT   /api/v1/clusters/{id}
PATCH /api/v1/clusters/{id}
DELETE /api/v1/clusters/{id}

GET   /api/v1/environments
POST  /api/v1/environments
GET   /api/v1/environments/{id}
PUT   /api/v1/environments/{id}
DELETE /api/v1/environments/{id}

GET   /api/v1/deployments
POST  /api/v1/deployments
GET   /api/v1/deployments/{id}
PUT   /api/v1/deployments/{id}
PATCH /api/v1/deployments/{id}
DELETE /api/v1/deployments/{id}

GET   /api/v1/pipelines
POST  /api/v1/pipelines
GET   /api/v1/pipelines/{id}
PATCH /api/v1/pipelines/{id}
DELETE /api/v1/pipelines/{id}

GET   /api/v1/incidents
POST  /api/v1/incidents
GET   /api/v1/incidents/{id}
PUT   /api/v1/incidents/{id}
PATCH /api/v1/incidents/{id}
DELETE /api/v1/incidents/{id}

GET  /api/v1/platform/summary
GET  /api/v1/platform/scorecards
GET  /api/v1/platform/environment-drift
```

`GET /metrics` exposes Prometheus text-format metrics:

```text
hr_cloud_service_up 1
http_requests_total{method="GET",route="/api/v1/platform/scorecards",status="200"} 1
http_request_duration_seconds_sum{method="GET",route="/api/v1/platform/scorecards",status="200"} 0.004215
http_request_duration_seconds_count{method="GET",route="/api/v1/platform/scorecards",status="200"} 1
```

`GET /api/v1/clusters` supports optional filters:

```http
GET /api/v1/clusters?provider=aws&region=ap-southeast-1&status=ready
```

`GET /api/v1/environments` supports optional filters:

```http
GET /api/v1/environments?application_id=app-123&cluster_id=cls-123&type=staging&status=active
```

`GET /api/v1/deployments` supports optional filters:

```http
GET /api/v1/deployments?application_id=app-123&cluster_id=cls-123&environment=staging&status=running
```

`GET /api/v1/pipelines` supports optional filters:

```http
GET /api/v1/pipelines?application_id=app-123&branch=main&status=running&triggered_by=devops@example.com
```

`GET /api/v1/incidents` supports optional filters:

```http
GET /api/v1/incidents?severity=sev2&status=investigating&owner_team=platform
```

`GET /api/v1/platform/scorecards` returns application-level DevOps readiness metrics:

```json
[
  {
    "application_id": "app-payroll-api",
    "application_name": "payroll-api",
    "owner_team": "platform",
    "criticality": "high",
    "environment_count": 2,
    "active_environment_count": 2,
    "cluster_count": 2,
    "deployment_success_rate": 50,
    "pipeline_success_rate": 100,
    "open_incident_count": 1,
    "operational_readiness_score": 40,
    "risk_level": "high",
    "risk_reasons": ["open incidents", "sev2 incident history", "deployment success rate below 80 percent"]
  }
]
```

`GET /api/v1/platform/environment-drift` returns runtime configuration drift reports:

```json
[
  {
    "environment_id": "env-payroll-prod",
    "environment_name": "payroll-prod",
    "application_id": "app-payroll-api",
    "application_name": "payroll-api",
    "missing_variables": [
      {
        "key": "FEATURE_AUDIT",
        "expected_value": "enabled"
      }
    ],
    "changed_variables": [
      {
        "key": "LOG_LEVEL",
        "expected_value": "info",
        "actual_value": "debug"
      }
    ],
    "drift_score": 55,
    "drift_level": "high",
    "drift_reasons": ["missing baseline variables", "changed baseline variables"]
  }
]
```

Example create employee request:

```json
{
  "name": "Nguyen Van A",
  "email": "a@example.com",
  "department": "Engineering",
  "title": "Backend Engineer"
}
```

Example create application request:

```json
{
  "name": "payroll-api",
  "repository": "github.com/company/payroll-api",
  "runtime": "go1.22",
  "owner_team": "platform",
  "criticality": "high",
  "port": 8080,
  "replicas": 3,
  "health_endpoint": "/healthz",
  "environment": {
    "LOG_LEVEL": "info"
  },
  "tags": ["go", "payroll", "backend"]
}
```

Example create cluster request:

```json
{
  "name": "eks-staging-ap-southeast-1",
  "provider": "aws",
  "region": "ap-southeast-1",
  "endpoint": "https://staging.example.eks.amazonaws.com",
  "version": "1.30",
  "status": "ready"
}
```

Example create deployment request:

```json
{
  "application_id": "app-123",
  "cluster_id": "cls-123",
  "namespace": "hr-staging",
  "environment": "staging",
  "version": "v1.3.0",
  "strategy": "canary",
  "requested_by": "devops@company.com"
}
```

Example create environment request:

```json
{
  "name": "payroll-staging",
  "type": "staging",
  "application_id": "app-123",
  "cluster_id": "cls-123",
  "namespace": "hr-staging",
  "status": "active",
  "variables": {
    "LOG_LEVEL": "debug"
  }
}
```

Example update deployment status request:

```json
{
  "status": "succeeded"
}
```

Example create incident request:

```json
{
  "title": "Payroll API error rate elevated",
  "summary": "5xx responses are above the platform threshold after the latest rollout.",
  "severity": "sev2",
  "status": "investigating",
  "application_id": "app-123",
  "cluster_id": "cls-123",
  "deployment_id": "dep-123",
  "owner_team": "platform"
}
```

Example create pipeline run request:

```json
{
  "application_id": "app-123",
  "branch": "main",
  "commit_sha": "4f9a2c9",
  "triggered_by": "devops@company.com",
  "stages": ["build", "unit-test", "security-scan", "containerize"]
}
```

## Docker

Build and run:

```powershell
docker build -t hr-cloud-devops-service .
docker run --rm -p 8080:8080 -e MONGO_URI="mongodb://host.docker.internal:27017" -e MONGO_DATABASE="hr_cloud" hr-cloud-devops-service
```

Or with Docker Compose:

```powershell
docker compose up --build
```

If port `8080` is already used, run Compose with another host port:

```powershell
$env:API_PORT="8081"
docker compose up --build
```

`/healthz` checks whether the process is alive. `/readyz` checks whether the API can ping MongoDB.

## API Examples

Open `docs/api.http` in VS Code or IntelliJ HTTP Client to call the sample APIs.

## CI/CD Docker

The GitHub Actions workflow in `.github/workflows/ci.yml` runs on pull requests and pushes to `main`:

- `gofmt` check
- `go vet ./...`
- `go test ./...`
- `govulncheck` dependency and code vulnerability scan
- Docker Buildx image build
- Docker image publish to GitHub Container Registry on `main` and `v*` tags

Published images use this registry path:

```text
ghcr.io/<owner>/<repo>
```

Common tags:

- `main` for the latest `main` branch build
- `sha-<commit>` for immutable commit builds
- `<version>` and `<major>.<minor>` for Git tags like `v1.2.3`

Run a published image with Compose:

```powershell
$env:IMAGE_NAME="ghcr.io/<owner>/<repo>:main"
docker compose -f docker-compose.prod.yml up -d
```

Create a release image:

```powershell
git tag v1.0.0
git push origin v1.0.0
```

## Kubernetes

After building and pushing the image, update the image name in `deploy/k8s/deployment.yaml`, then run:

```powershell
kubectl apply -f .\deploy\k8s
```

For clusters without Prometheus Operator, apply the core manifests without `servicemonitor.yaml`:

```powershell
kubectl apply -f .\deploy\k8s\namespace.yaml
kubectl apply -f .\deploy\k8s\configmap.yaml
kubectl apply -f .\deploy\k8s\secret.example.yaml
kubectl apply -f .\deploy\k8s\deployment.yaml
kubectl apply -f .\deploy\k8s\service.yaml
kubectl apply -f .\deploy\k8s\hpa.yaml
kubectl apply -f .\deploy\k8s\pdb.yaml
kubectl apply -f .\deploy\k8s\networkpolicy.yaml
```

The Kubernetes manifests include:

- ConfigMap-based runtime settings and a Secret-based MongoDB URI.
- Restricted pod/container security contexts.
- Rolling updates with zero unavailable pods.
- Prometheus scrape annotations on the API pods.
- A `ServiceMonitor` for clusters using Prometheus Operator.
- A CPU-based `HorizontalPodAutoscaler` with 2 to 6 replicas.
- A `PodDisruptionBudget` to keep at least one API pod available during voluntary disruptions.
- A `NetworkPolicy` that limits API pod traffic to HTTP ingress, DNS egress, and MongoDB egress.

If your cluster does not have Prometheus Operator installed, skip `deploy/k8s/servicemonitor.yaml` or install the CRD first.
For production, replace `deploy/k8s/secret.example.yaml` with a real Secret from your secret manager or CI/CD pipeline before applying manifests.

## Jenkins and Cloud Deployment

This repo includes a `Jenkinsfile` for Go quality gates, Docker image build/push, and Kubernetes rollout. See `docs/cloud-cicd.md` for the required Jenkins credentials and the Docker Compose production path for a cloud VM.

## Note

This is MVC-style, but adapted to Go. The service layer keeps business rules out of controllers, and the repository layer keeps data access replaceable when you move from in-memory storage to PostgreSQL, MySQL, MongoDB, or Redis.
