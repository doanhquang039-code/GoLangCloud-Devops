# Cloud, Docker, Jenkins CI/CD

This project supports two deployment paths:

- Docker Compose on a cloud VM.
- Kubernetes on a managed cluster such as EKS, AKS, GKE, or a self-managed cluster.

## Docker Compose on a VM

Install Docker Engine or Docker Desktop on the VM, then create a `.env.prod` file:

```text
IMAGE_NAME=ghcr.io/owner/hr-cloud-service:main
API_PORT=8080
MONGO_URI=mongodb://mongo:27017
MONGO_DATABASE=hr_cloud
```

Start the service:

```sh
docker compose --env-file .env.prod -f docker-compose.prod.yml up -d
docker compose --env-file .env.prod -f docker-compose.prod.yml ps
```

Update to a new image:

```sh
docker compose --env-file .env.prod -f docker-compose.prod.yml pull
docker compose --env-file .env.prod -f docker-compose.prod.yml up -d
```

## Jenkins

Create these Jenkins credentials:

- `docker-registry-host`: Secret text, for example `ghcr.io` or `registry.example.com`.
- `docker-image-repo`: Secret text, for example `owner/hr-cloud-service`.
- `docker-registry-credentials`: Username/password for the registry.
- `kubeconfig`: Secret file containing kubeconfig for the target cluster.

The `Jenkinsfile` stages are:

- Checkout.
- Format, vet, and test Go code.
- Build the Go binary.
- Build the Docker image.
- Push image on `main` and release tags.
- Deploy to Kubernetes on `main` and release tags.

## Kubernetes

For production, create the MongoDB secret yourself before deploying:

```sh
kubectl create namespace hr-cloud --dry-run=client -o yaml | kubectl apply -f -
kubectl -n hr-cloud create secret generic hr-cloud-devops-service-secret \
  --from-literal=MONGO_URI='mongodb://user:password@mongo.example:27017/hr_cloud' \
  --dry-run=client -o yaml | kubectl apply -f -
```

Deploy a specific image:

```sh
./deploy/scripts/k8s-deploy.sh ghcr.io/owner/hr-cloud-service:main
```
