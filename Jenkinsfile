pipeline {
  agent any

  options {
    timestamps()
    disableConcurrentBuilds()
    buildDiscarder(logRotator(numToKeepStr: '20'))
  }

  environment {
    REGISTRY = credentials('docker-registry-host')
    IMAGE_REPO = credentials('docker-image-repo')
    DOCKER_CREDS = 'docker-registry-credentials'
    KUBECONFIG_CREDS = 'kubeconfig'
    GOPATH = "${WORKSPACE}/.tools/gopath"
    GOCACHE = "${WORKSPACE}/.tools/gocache"
  }

  stages {
    stage('Checkout') {
      steps {
        checkout scm
      }
    }

    stage('Go Quality Gates') {
      steps {
        sh '''
          set -eu
          test -z "$(git ls-files '*.go' | xargs gofmt -l)"
          go vet ./...
          go test ./...
        '''
      }
    }

    stage('Build Binary') {
      steps {
        sh '''
          set -eu
          mkdir -p bin
          CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/hr-cloud-service ./cmd/api
        '''
      }
    }

    stage('Build Docker Image') {
      steps {
        script {
          env.IMAGE_TAG = env.TAG_NAME ?: "sha-${env.GIT_COMMIT.take(12)}"
          env.IMAGE = "${env.REGISTRY}/${env.IMAGE_REPO}:${env.IMAGE_TAG}"
        }
        sh 'docker build -t "$IMAGE" .'
      }
    }

    stage('Push Docker Image') {
      when {
        anyOf {
          branch 'main'
          buildingTag()
        }
      }
      steps {
        withCredentials([usernamePassword(credentialsId: "${DOCKER_CREDS}", usernameVariable: 'DOCKER_USER', passwordVariable: 'DOCKER_PASSWORD')]) {
          sh '''
            set -eu
            echo "$DOCKER_PASSWORD" | docker login "$REGISTRY" -u "$DOCKER_USER" --password-stdin
            docker push "$IMAGE"
          '''
        }
      }
    }

    stage('Deploy Kubernetes') {
      when {
        anyOf {
          branch 'main'
          buildingTag()
        }
      }
      steps {
        withCredentials([file(credentialsId: "${KUBECONFIG_CREDS}", variable: 'KUBECONFIG_FILE')]) {
          sh '''
            set -eu
            export KUBECONFIG="$KUBECONFIG_FILE"
            sh deploy/scripts/k8s-deploy.sh "$IMAGE"
          '''
        }
      }
    }
  }

  post {
    always {
      sh 'docker logout "$REGISTRY" || true'
    }
  }
}
