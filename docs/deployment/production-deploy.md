# 🚀 Production Deployment

## 📋 Обзор production развертывания

Комплексное руководство по развертыванию микросервисной системы в production окружении с обеспечением высокой доступности, безопасности и производительности.

## 🏗️ Infrastructure Overview

### Cloud Infrastructure
- **Platform**: Kubernetes (GKE/AKS/EKS)
- **Load Balancer**: Nginx + CloudFlare
- **Database**: PostgreSQL 15+ (Cloud SQL/RDS)
- **Cache**: Redis (Cloud Memorystore/ElastiCache)
- **Message Queue**: Apache Kafka (Confluent Cloud/MSK)
- **Storage**: Google Cloud Storage/S3
- **Monitoring**: Prometheus + Grafana + Jaeger
- **CI/CD**: GitLab CI/CD

### Production Architecture
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   CloudFlare    │────│  Load Balancer  │────│  Kubernetes     │
│   (CDN/WAF)     │    │    (Nginx)      │    │   Cluster       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                                        │
        ┌───────────────────────────────────────────────┼───────────────┐
        │                                               │               │
┌───────▼───────┐  ┌──────────────┐  ┌─────────────┐  ┌▼─────────────┐ │
│  Auth Service │  │ User Service │  │Doc Service  │  │Other Services│ │
│   (3 replicas)│  │ (2 replicas) │  │(3 replicas) │  │              │ │
└───────────────┘  └──────────────┘  └─────────────┘  └──────────────┘ │
                                                                       │
┌─────────────────────────────────────────────────────────────────────┘
│  Data Layer
├── PostgreSQL (Primary + Replica)
├── Redis Cluster
├── Kafka Cluster
└── Object Storage
```

---

## 🐳 Kubernetes Configuration

### Namespace Setup
```yaml
# namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: inf-sys
  labels:
    name: inf-sys
    environment: production

---
apiVersion: v1
kind: ResourceQuota
metadata:
  name: inf-sys-quota
  namespace: inf-sys
spec:
  hard:
    requests.cpu: "10"
    requests.memory: 20Gi
    limits.cpu: "20"
    limits.memory: 40Gi
    pods: "50"
    services: "20"
    persistentvolumeclaims: "10"
```

### ConfigMaps and Secrets
```yaml
# configmap-production.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
  namespace: inf-sys
data:
  APP_ENV: "production"
  LOG_LEVEL: "info"
  DB_HOST: "postgres-service"
  REDIS_HOST: "redis-service"
  KAFKA_BROKERS: "kafka-service:9092"

---
apiVersion: v1
kind: Secret
metadata:
  name: app-secrets
  namespace: inf-sys
type: Opaque
data:
  db-password: <base64-encoded-password>
  jwt-secret: <base64-encoded-jwt-secret>
  google-client-secret: <base64-encoded-google-secret>
  azure-client-secret: <base64-encoded-azure-secret>
```

### Auth Service Deployment
```yaml
# auth-service-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: auth-service
  namespace: inf-sys
  labels:
    app: auth-service
    version: v1
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 1
  selector:
    matchLabels:
      app: auth-service
  template:
    metadata:
      labels:
        app: auth-service
        version: v1
    spec:
      serviceAccountName: inf-sys-service-account
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
        fsGroup: 2000
      containers:
      - name: auth-service
        image: inf-sys/auth-service:v1.2.3
        imagePullPolicy: Always
        ports:
        - containerPort: 8080
          name: http
        env:
        - name: APP_ENV
          value: "production"
        envFrom:
        - configMapRef:
            name: app-config
        - secretRef:
            name: app-secrets
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 500m
            memory: 512Mi
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 2
        securityContext:
          allowPrivilegeEscalation: false
          readOnlyRootFilesystem: true
          capabilities:
            drop:
            - ALL
        volumeMounts:
        - name: tmp-volume
          mountPath: /tmp
        - name: config-volume
          mountPath: /app/configs
          readOnly: true
      volumes:
      - name: tmp-volume
        emptyDir: {}
      - name: config-volume
        configMap:
          name: auth-service-config
      nodeSelector:
        kubernetes.io/os: linux
      tolerations:
      - key: "node.kubernetes.io/not-ready"
        operator: "Exists"
        effect: "NoExecute"
        tolerationSeconds: 300

---
apiVersion: v1
kind: Service
metadata:
  name: auth-service
  namespace: inf-sys
  labels:
    app: auth-service
spec:
  type: ClusterIP
  ports:
  - port: 80
    targetPort: 8080
    protocol: TCP
    name: http
  selector:
    app: auth-service

---
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: auth-service-pdb
  namespace: inf-sys
spec:
  minAvailable: 2
  selector:
    matchLabels:
      app: auth-service
```

---

## 🔄 Blue-Green Deployment

### Blue-Green Strategy Configuration
```yaml
# blue-green-deployment.yaml
apiVersion: argoproj.io/v1alpha1
kind: Rollout
metadata:
  name: auth-service-rollout
  namespace: inf-sys
spec:
  replicas: 3
  strategy:
    blueGreen:
      activeService: auth-service-active
      previewService: auth-service-preview
      autoPromotionEnabled: false
      scaleDownDelaySeconds: 30
      prePromotionAnalysis:
        templates:
        - templateName: success-rate
        args:
        - name: service-name
          value: auth-service-preview
      postPromotionAnalysis:
        templates:
        - templateName: success-rate
        args:
        - name: service-name
          value: auth-service-active
  selector:
    matchLabels:
      app: auth-service
  template:
    metadata:
      labels:
        app: auth-service
    spec:
      containers:
      - name: auth-service
        image: inf-sys/auth-service:v1.2.3
        ports:
        - containerPort: 8080

---
apiVersion: v1
kind: Service
metadata:
  name: auth-service-active
  namespace: inf-sys
spec:
  type: ClusterIP
  ports:
  - port: 80
    targetPort: 8080
  selector:
    app: auth-service

---
apiVersion: v1
kind: Service
metadata:
  name: auth-service-preview
  namespace: inf-sys
spec:
  type: ClusterIP
  ports:
  - port: 80
    targetPort: 8080
  selector:
    app: auth-service
```

### Analysis Template
```yaml
# analysis-template.yaml
apiVersion: argoproj.io/v1alpha1
kind: AnalysisTemplate
metadata:
  name: success-rate
  namespace: inf-sys
spec:
  args:
  - name: service-name
  metrics:
  - name: success-rate
    interval: 2m
    count: 5
    successCondition: result[0] >= 0.95
    failureLimit: 3
    provider:
      prometheus:
        address: http://prometheus:9090
        query: |
          sum(irate(
            http_requests_total{job="{{args.service-name}}",status!~"5.*"}[2m]
          )) /
          sum(irate(
            http_requests_total{job="{{args.service-name}}"}[2m]
          ))
```

---

## 🗄️ Database Configuration

### PostgreSQL High Availability
```yaml
# postgresql-ha.yaml
apiVersion: postgresql.cnpg.io/v1
kind: Cluster
metadata:
  name: postgres-cluster
  namespace: inf-sys
spec:
  instances: 3

  postgresql:
    parameters:
      max_connections: "200"
      shared_buffers: "256MB"
      effective_cache_size: "1GB"
      maintenance_work_mem: "64MB"
      checkpoint_completion_target: "0.9"
      wal_buffers: "16MB"
      default_statistics_target: "100"
      random_page_cost: "1.1"
      effective_io_concurrency: "200"
      work_mem: "4MB"
      min_wal_size: "1GB"
      max_wal_size: "4GB"

  bootstrap:
    initdb:
      database: inf_sys
      owner: inf_sys_user
      secret:
        name: postgres-credentials

  storage:
    size: 100Gi
    storageClass: fast-ssd

  monitoring:
    enabled: true

  backup:
    retentionPolicy: "30d"
    barmanObjectStore:
      destinationPath: "gs://inf-sys-backups/postgres"
      googleCredentials:
        applicationCredentials:
          name: backup-credentials
          key: service-account.json

---
apiVersion: v1
kind: Secret
metadata:
  name: postgres-credentials
  namespace: inf-sys
type: kubernetes.io/basic-auth
data:
  username: <base64-encoded-username>
  password: <base64-encoded-password>
```

### Redis Cluster
```yaml
# redis-cluster.yaml
apiVersion: redis.io/v1beta2
kind: RedisCluster
metadata:
  name: redis-cluster
  namespace: inf-sys
spec:
  numberOfMaster: 3
  replicationFactor: 1

  podTemplate:
    spec:
      containers:
      - name: redis
        image: redis:7-alpine
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 500m
            memory: 512Mi

  storage:
    volumeClaimTemplate:
      spec:
        accessModes: ["ReadWriteOnce"]
        resources:
          requests:
            storage: 10Gi
        storageClassName: fast-ssd

  securityContext:
    runAsUser: 999
    runAsGroup: 999
    fsGroup: 999
```

---

## 🌐 Ingress and Load Balancing

### Nginx Ingress Controller
```yaml
# ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: inf-sys-ingress
  namespace: inf-sys
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/force-ssl-redirect: "true"
    nginx.ingress.kubernetes.io/rate-limit: "100"
    nginx.ingress.kubernetes.io/rate-limit-window: "1m"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/cors-allow-origin: "https://app.inf-sys.example.com"
    nginx.ingress.kubernetes.io/enable-cors: "true"
spec:
  tls:
  - hosts:
    - api.inf-sys.example.com
    secretName: inf-sys-tls
  rules:
  - host: api.inf-sys.example.com
    http:
      paths:
      - path: /auth
        pathType: Prefix
        backend:
          service:
            name: auth-service-active
            port:
              number: 80
      - path: /users
        pathType: Prefix
        backend:
          service:
            name: user-service
            port:
              number: 80
      - path: /documents
        pathType: Prefix
        backend:
          service:
            name: document-service
            port:
              number: 80
      - path: /workflow
        pathType: Prefix
        backend:
          service:
            name: workflow-service
            port:
              number: 80

---
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@inf-sys.example.com
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: nginx
```

---

## 📊 Monitoring and Observability

### Prometheus Configuration
```yaml
# prometheus.yaml
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: prometheus
  namespace: monitoring
spec:
  serviceAccountName: prometheus
  serviceMonitorSelector:
    matchLabels:
      team: inf-sys
  ruleSelector:
    matchLabels:
      team: inf-sys
  resources:
    requests:
      memory: 400Mi
      cpu: 100m
    limits:
      memory: 2Gi
      cpu: 1
  retention: 30d
  storage:
    volumeClaimTemplate:
      spec:
        accessModes: ["ReadWriteOnce"]
        resources:
          requests:
            storage: 50Gi

---
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: inf-sys-services
  namespace: monitoring
  labels:
    team: inf-sys
spec:
  selector:
    matchLabels:
      monitoring: enabled
  endpoints:
  - port: http
    path: /metrics
    interval: 30s
```

### Grafana Dashboard
```yaml
# grafana.yaml
apiVersion: integreatly.org/v1alpha1
kind: Grafana
metadata:
  name: grafana
  namespace: monitoring
spec:
  config:
    auth:
      disable_login_form: false
    auth.anonymous:
      enabled: true
    security:
      admin_user: admin
      admin_password: admin
  dashboards:
    - name: inf-sys-dashboard
      datasources:
        - inputName: "DS_PROMETHEUS"
          datasourceName: "prometheus"
  datasources:
    - name: prometheus
      type: prometheus
      access: proxy
      url: http://prometheus:9090
      isDefault: true
```

### Jaeger Tracing
```yaml
# jaeger.yaml
apiVersion: jaegertracing.io/v1
kind: Jaeger
metadata:
  name: jaeger
  namespace: monitoring
spec:
  strategy: production
  storage:
    type: elasticsearch
    elasticsearch:
      nodeCount: 3
      redundancyPolicy: SingleRedundancy
      storage:
        size: 50Gi
  collector:
    maxReplicas: 5
    resources:
      limits:
        cpu: 500m
        memory: 512Mi
  query:
    replicas: 2
    resources:
      limits:
        cpu: 500m
        memory: 512Mi
```

---

## 🔄 CI/CD Pipeline

### GitLab CI Configuration
```yaml
# .gitlab-ci.yml
stages:
  - test
  - build
  - security
  - deploy-staging
  - deploy-production

variables:
  DOCKER_DRIVER: overlay2
  DOCKER_TLS_CERTDIR: "/certs"
  REGISTRY: ghcr.io/your-org/inf-sys

# Test Stage
test-backend:
  stage: test
  image: golang:1.25
  script:
    - make test-coverage
    - make lint
  artifacts:
    reports:
      coverage_report:
        coverage_format: cobertura
        path: coverage.xml
  coverage: '/coverage: \d+\.\d+% of statements/'

test-frontend:
  stage: test
  image: node:25
  script:
    - cd frontend
    - npm ci
    - npm run test:unit
    - npm run test:e2e
  artifacts:
    reports:
      junit: frontend/test-results.xml

# Build Stage
build-images:
  stage: build
  image: docker:latest
  services:
    - docker:dind
  before_script:
    - echo $CI_REGISTRY_PASSWORD | docker login -u $CI_REGISTRY_USER --password-stdin $REGISTRY
  script:
    - |
      for service in auth user document workflow; do
        docker build -t $REGISTRY/$service:$CI_COMMIT_SHA ./services/$service
        docker push $REGISTRY/$service:$CI_COMMIT_SHA
        docker tag $REGISTRY/$service:$CI_COMMIT_SHA $REGISTRY/$service:latest
        docker push $REGISTRY/$service:latest
      done
  only:
    - main
    - develop

# Security Stage
security-scan:
  stage: security
  image: aquasec/trivy:latest
  script:
    - trivy image --exit-code 1 --severity HIGH,CRITICAL $REGISTRY/auth:$CI_COMMIT_SHA
  only:
    - main

# Deploy Staging
deploy-staging:
  stage: deploy-staging
  image: bitnami/kubectl:latest
  before_script:
    - kubectl config use-context staging
  script:
    - |
      for service in auth user document workflow; do
        kubectl set image deployment/$service-service $service=$REGISTRY/$service:$CI_COMMIT_SHA -n inf-sys-staging
        kubectl rollout status deployment/$service-service -n inf-sys-staging --timeout=300s
      done
  environment:
    name: staging
    url: https://staging-api.inf-sys.example.com
  only:
    - develop

# Deploy Production
deploy-production:
  stage: deploy-production
  image: bitnami/kubectl:latest
  before_script:
    - kubectl config use-context production
  script:
    - |
      for service in auth user document workflow; do
        kubectl argo rollouts set image $service-rollout $service=$REGISTRY/$service:$CI_COMMIT_SHA -n inf-sys
        kubectl argo rollouts promote $service-rollout -n inf-sys
      done
  environment:
    name: production
    url: https://api.inf-sys.example.com
  when: manual
  only:
    - main
```

---

## 🔐 Security Configuration

### Network Policies
```yaml
# network-policies.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: inf-sys-network-policy
  namespace: inf-sys
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: nginx-ingress
  - from:
    - podSelector:
        matchLabels:
          app: auth-service
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - namespaceSelector:
        matchLabels:
          name: kube-system
  - to:
    - podSelector:
        matchLabels:
          app: postgres
    ports:
    - protocol: TCP
      port: 5432
  - to:
    - podSelector:
        matchLabels:
          app: redis
    ports:
    - protocol: TCP
      port: 6379
```

### Pod Security Policy
```yaml
# pod-security-policy.yaml
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: inf-sys-psp
spec:
  privileged: false
  allowPrivilegeEscalation: false
  requiredDropCapabilities:
    - ALL
  volumes:
    - 'configMap'
    - 'emptyDir'
    - 'projected'
    - 'secret'
    - 'downwardAPI'
    - 'persistentVolumeClaim'
  runAsUser:
    rule: 'MustRunAsNonRoot'
  seLinux:
    rule: 'RunAsAny'
  fsGroup:
    rule: 'RunAsAny'
  readOnlyRootFilesystem: true
```

---

## 📈 Scaling Configuration

### Horizontal Pod Autoscaler
```yaml
# hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: auth-service-hpa
  namespace: inf-sys
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: auth-service
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Percent
        value: 10
        periodSeconds: 60
    scaleUp:
      stabilizationWindowSeconds: 60
      policies:
      - type: Percent
        value: 50
        periodSeconds: 60
      - type: Pods
        value: 2
        periodSeconds: 60
      selectPolicy: Max
```

### Vertical Pod Autoscaler
```yaml
# vpa.yaml
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: auth-service-vpa
  namespace: inf-sys
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: auth-service
  updatePolicy:
    updateMode: "Auto"
  resourcePolicy:
    containerPolicies:
    - containerName: auth-service
      maxAllowed:
        cpu: 1
        memory: 2Gi
      minAllowed:
        cpu: 100m
        memory: 128Mi
```

---

## 🚀 Deployment Commands

### Initial Deployment
```bash
#!/bin/bash
# scripts/deploy-production.sh

set -e

echo "🚀 Starting production deployment..."

# Apply namespace and RBAC
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/rbac.yaml

# Apply secrets and configmaps
kubectl apply -f k8s/secrets.yaml
kubectl apply -f k8s/configmaps.yaml

# Deploy infrastructure
kubectl apply -f k8s/postgres/
kubectl apply -f k8s/redis/
kubectl apply -f k8s/kafka/

# Wait for infrastructure
kubectl wait --for=condition=ready pod -l app=postgres -n inf-sys --timeout=300s
kubectl wait --for=condition=ready pod -l app=redis -n inf-sys --timeout=300s

# Deploy services
for service in auth user document workflow schedule reporting task notification file integration; do
    echo "Deploying $service service..."
    kubectl apply -f k8s/services/$service/
    kubectl rollout status deployment/$service-service -n inf-sys --timeout=300s
done

# Deploy ingress
kubectl apply -f k8s/ingress.yaml

# Deploy monitoring
kubectl apply -f k8s/monitoring/

echo "✅ Production deployment completed successfully!"
```

### Rolling Update Script
```bash
#!/bin/bash
# scripts/rolling-update.sh

SERVICE=$1
IMAGE_TAG=$2

if [[ -z "$SERVICE" || -z "$IMAGE_TAG" ]]; then
    echo "Usage: $0 <service> <image_tag>"
    exit 1
fi

echo "🔄 Rolling update for $SERVICE to $IMAGE_TAG..."

# Update image
kubectl set image deployment/$SERVICE-service $SERVICE=inf-sys/$SERVICE:$IMAGE_TAG -n inf-sys

# Wait for rollout
kubectl rollout status deployment/$SERVICE-service -n inf-sys --timeout=300s

# Verify deployment
kubectl get pods -l app=$SERVICE-service -n inf-sys

echo "✅ Rolling update completed for $SERVICE"
```

---

## 🔧 Maintenance and Operations

### Backup Scripts
```bash
#!/bin/bash
# scripts/backup-database.sh

DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_NAME="inf-sys-backup-$DATE"

echo "📦 Creating database backup: $BACKUP_NAME"

# Create database backup
kubectl exec -n inf-sys postgres-cluster-1 -- pg_dump -U inf_sys_user inf_sys > /tmp/$BACKUP_NAME.sql

# Upload to cloud storage
gsutil cp /tmp/$BACKUP_NAME.sql gs://inf-sys-backups/database/

# Clean up local file
rm /tmp/$BACKUP_NAME.sql

echo "✅ Backup completed: $BACKUP_NAME"
```

### Health Check Script
```bash
#!/bin/bash
# scripts/health-check.sh

echo "🔍 Running production health checks..."

# Check all services
for service in auth user document workflow; do
    URL="https://api.inf-sys.example.com/$service/health"
    if curl -s -f "$URL" > /dev/null; then
        echo "✅ $service: healthy"
    else
        echo "❌ $service: unhealthy"
    fi
done

# Check database
kubectl exec -n inf-sys postgres-cluster-1 -- pg_isready -U inf_sys_user

# Check Redis
kubectl exec -n inf-sys redis-cluster-0 -- redis-cli ping

echo "✅ Health check completed"
```

Production deployment обеспечивает высокую доступность, безопасность и масштабируемость системы!
---

**📅 Актуальность документа**  
**Последнее обновление**: 2025-01-15  
**Версия проекта**: 0.1.0  
**Статус**: Актуальный

