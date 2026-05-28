#!/usr/bin/env sh
set -eu

IMAGE="${1:?Usage: k8s-deploy.sh <image>}"
NAMESPACE="${NAMESPACE:-hr-cloud}"
DEPLOYMENT="${DEPLOYMENT:-hr-cloud-devops-service}"
CONTAINER="${CONTAINER:-api}"
MANIFEST_DIR="${MANIFEST_DIR:-deploy/k8s}"

kubectl apply -f "$MANIFEST_DIR/namespace.yaml"
kubectl apply -f "$MANIFEST_DIR/configmap.yaml"
kubectl apply -f "$MANIFEST_DIR/mongo.yaml"

if kubectl -n "$NAMESPACE" get secret hr-cloud-devops-service-secret >/dev/null 2>&1; then
  echo "Using existing secret hr-cloud-devops-service-secret"
else
  kubectl apply -f "$MANIFEST_DIR/secret.example.yaml"
fi

kubectl apply -f "$MANIFEST_DIR/deployment.yaml"
kubectl apply -f "$MANIFEST_DIR/service.yaml"
kubectl apply -f "$MANIFEST_DIR/hpa.yaml"
kubectl apply -f "$MANIFEST_DIR/pdb.yaml"
kubectl apply -f "$MANIFEST_DIR/networkpolicy.yaml"

kubectl -n "$NAMESPACE" set image "deployment/$DEPLOYMENT" "$CONTAINER=$IMAGE"
kubectl -n "$NAMESPACE" rollout status "deployment/$DEPLOYMENT" --timeout=180s
