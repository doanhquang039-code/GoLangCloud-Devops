# HR Cloud DevOps Service

Go API service for HR/cloud platform operations. It tracks employees, applications, clusters, environments, deployments, pipeline runs, microservices, incidents, platform scorecards, and environment drift reports.

## Features

- REST API with CRUD for employees, applications, clusters, cloud accounts, environments, deployments, microservices, and incidents.
- Cloud account inventory with provider, region, owner, environment, monthly cost, budget, compliance, backup posture, and security finding metadata.
- Cloud microservice inventory with provider, region, cluster, namespace, environment, runtime, image, replicas, resource requests, health path, SLO, and error budget metadata.
- Pipeline run creation, status updates, stage updates, and deletion.
- Search and filters on list endpoints with `q`, status/type/team fields, and tags where applicable.
- Platform summary, operational readiness scorecards, and environment variable drift reports.
- MongoDB persistence with in-memory repositories for tests.
- Health, readiness, Prometheus-style metrics, Docker, Docker Compose, Jenkins, and Kubernetes manifests.

## Requirements

- Go 1.21+
- MongoDB 6+ for local or production runtime
- Docker optional for containerized runs

## Configuration

Environment variables:

| Name | Default | Description |
| --- | --- | --- |
| `PORT` | `8080` | HTTP port |
| `MONGO_URI` | `mongodb://localhost:27017` | MongoDB connection string |
| `MONGO_DATABASE` | `hr_cloud` | MongoDB database name |
| `SEED_DATA` | `true` | Seeds demo data on startup when enabled |

Copy `.env.example` or `.env.prod.example` as needed for local/prod deployment.

## Run Locally

```sh
go test ./...
go run ./cmd/api
```

Health checks:

```sh
curl http://localhost:8080/healthz
curl http://localhost:8080/readyz
curl http://localhost:8080/metrics
```

## Build

```sh
go build ./cmd/api
docker build -t hr-cloud-devops-service:local .
```

## Docker Compose

```sh
docker compose up -d
docker compose ps
```

Production compose:

```sh
docker compose --env-file .env.prod -f docker-compose.prod.yml up -d
```

## Kubernetes

Create the namespace and MongoDB secret, then deploy an image:

```sh
kubectl create namespace hr-cloud --dry-run=client -o yaml | kubectl apply -f -
kubectl -n hr-cloud create secret generic hr-cloud-devops-service-secret \
  --from-literal=MONGO_URI='mongodb://user:password@mongo.example:27017/hr_cloud' \
  --dry-run=client -o yaml | kubectl apply -f -
./deploy/scripts/k8s-deploy.sh ghcr.io/owner/hr-cloud-service:main
```

More CI/CD notes are in `docs/cloud-cicd.md`.

## Database Scale Notes

This service now pushes microservice inventory filters down to MongoDB and supports bounded list queries:

```http
GET /api/v1/microservices?tenant_id=tenant-hr&cloud_provider=aws&region=ap-southeast-1&environment=staging&limit=100&sort=id&after_id=svc-payroll-api
```

For very large traffic and data volumes, run MongoDB as a managed replica set or sharded cluster. Microservice inventory now stores `tenant_id`; a production shard key can start with tenant/account plus region or application identifiers, depending on access pattern.

Indexes included for microservices:

- Unique `tenant_id`, `id`.
- `tenant_id`, `application_id`, `protocol`, `status`.
- `tenant_id`, `cloud_provider`, `region`, `cluster_id`, `namespace`, `environment`, `runtime`.
- `tenant_id`, `updated_at`, `id` for stable latest-first paging.
- `tenant_id`, `replicas`, `id` for capacity-oriented sorting.

For “1 billion users” production readiness, plan these pieces outside this API:

- Global load balancing and multi-region active-active or active-passive topology.
- MongoDB sharding, backups, PITR, read replicas, connection pooling, and query dashboards.
- Cursor-based pagination for every high-cardinality endpoint. Microservices support `after_id`; offset is retained for compatibility but should not be used for deep pages.
- Dedicated search engine or MongoDB Atlas Search for free-text `q`; regex search is convenient but not the right path for billion-record scans.
- Redis or CDN caching for hot read paths.
- Event streaming with Kafka/Pub/Sub/SQS for writes, audits, metrics, and async workflows.
- Rate limiting, tenant isolation, authn/authz, secrets management, and SLO dashboards.

## API Overview

Base URL: `http://localhost:8080`

| Resource | Endpoints |
| --- | --- |
| Employees | `/api/v1/employees`, `/api/v1/employees/{id}` |
| Applications | `/api/v1/applications`, `/api/v1/applications/{id}` |
| Clusters | `/api/v1/clusters`, `/api/v1/clusters/{id}` |
| Cloud Accounts | `/api/v1/cloud-accounts`, `/api/v1/cloud-accounts/{id}`, `/api/v1/cloud/summary`, `/api/v1/cloud/policy-violations`, `/api/v1/cloud/remediation-plan` |
| Environments | `/api/v1/environments`, `/api/v1/environments/{id}` |
| Deployments | `/api/v1/deployments`, `/api/v1/deployments/{id}` |
| Pipelines | `/api/v1/pipelines`, `/api/v1/pipelines/{id}`, `/api/v1/pipelines/{id}/stages/{stage}` |
| Microservices | `/api/v1/microservices`, `/api/v1/microservices/{id}` |
| Incidents | `/api/v1/incidents`, `/api/v1/incidents/{id}` |
| Platform | `/api/v1/platform/summary`, `/api/v1/platform/scorecards`, `/api/v1/platform/environment-drift` |

Common list examples:

```http
GET /api/v1/applications?q=payroll&owner_team=platform
GET /api/v1/clusters?q=staging&provider=aws&status=ready
GET /api/v1/cloud-accounts?provider=aws&owner_team=platform&environment=production&backup_status=protected&tag=prod
GET /api/v1/cloud/summary
GET /api/v1/cloud/policy-violations?provider=aws&environment=production
GET /api/v1/cloud/remediation-plan?owner_team=platform
GET /api/v1/environments?q=payroll-v2&type=staging&status=active
GET /api/v1/deployments?q=canary&environment=staging&status=running
GET /api/v1/pipelines?q=security&status=running
GET /api/v1/microservices?tenant_id=tenant-hr&q=payroll&owner_team=platform&status=active&cloud_provider=aws&region=ap-southeast-1&namespace=hr-staging&environment=staging&min_replicas=2&limit=100&sort=id&after_id=svc-payroll-api
GET /api/v1/incidents?q=canary&severity=sev2&status=investigating
GET /api/v1/platform/scorecards?q=payroll&owner_team=platform&risk_level=low&min_score=80&sort=score&order=desc
GET /api/v1/platform/environment-drift?q=LOG_LEVEL&type=production&drift_level=high&max_drift_score=80
```

Full request samples are in `docs/api.http`.

## Validation

```sh
go test ./...
go build ./cmd/api
```
