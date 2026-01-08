#!/bin/bash
set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}🚀 Setting up KinD cluster for infravalidation testing${NC}"

# Check if kind is installed
if ! command -v kind &> /dev/null; then
    echo -e "${RED}❌ kind is not installed. Please install it first:${NC}"
    echo "   https://kind.sigs.k8s.io/docs/user/quick-start/#installation"
    exit 1
fi

# Check if kubectl is installed
if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}❌ kubectl is not installed. Please install it first.${NC}"
    exit 1
fi

CLUSTER_NAME="kind-infra-test"

# Check if cluster already exists
if kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
    echo -e "${YELLOW}⚠️  Cluster ${CLUSTER_NAME} already exists. Deleting...${NC}"
    kind delete cluster --name "${CLUSTER_NAME}"
fi

# Create KinD cluster
echo -e "${GREEN}📦 Creating KinD cluster: ${CLUSTER_NAME}${NC}"
cat <<EOF | kind create cluster --name "${CLUSTER_NAME}" --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
EOF

# Wait for cluster to be ready
echo "⏳ Waiting for cluster to be ready..."
kubectl wait --for=condition=Ready nodes --all --timeout=60s

# Get API server endpoint
API_SERVER=$(kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}')
echo "✅ Cluster ready. API Server: $API_SERVER"

# Create namespaces
echo -e "\n${GREEN}📁 Creating namespaces...${NC}"
kubectl create namespace app-prod
kubectl create namespace app-staging
kubectl create namespace app-dev

# Create service accounts and get tokens
echo -e "\n${GREEN}🔑 Creating service accounts...${NC}"

# Function to create service account and get token
create_sa_and_get_token() {
    local namespace=$1
    local sa_name="infravalidation-sa"
    
    # Create service account
    kubectl create serviceaccount "${sa_name}" -n "${namespace}" >/dev/null 2>&1
    
    # Create ClusterRoleBinding for admin access
    kubectl create clusterrolebinding "${sa_name}-${namespace}-admin" \
        --clusterrole=cluster-admin \
        --serviceaccount="${namespace}:${sa_name}" >/dev/null 2>&1
    
    # Create token secret
    kubectl apply -f - >/dev/null 2>&1 <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: ${sa_name}-token
  namespace: ${namespace}
  annotations:
    kubernetes.io/service-account.name: ${sa_name}
type: kubernetes.io/service-account-token
EOF
    
    # Wait for token to be populated
    for i in {1..30}; do
        TOKEN=$(kubectl get secret "${sa_name}-token" -n "${namespace}" -o jsonpath='{.data.token}' 2>/dev/null | base64 -d)
        if [ -n "$TOKEN" ]; then
            echo "$TOKEN"
            return 0
        fi
        sleep 1
    done
    
    echo "" >&2
    return 1
}

PROD_TOKEN=$(create_sa_and_get_token "app-prod")
STAGING_TOKEN=$(create_sa_and_get_token "app-staging")
DEV_TOKEN=$(create_sa_and_get_token "app-dev")

echo "✅ Service accounts created and tokens extracted"

# Deploy baseline resources to app-prod
echo -e "\n${GREEN}📦 Deploying baseline resources to app-prod...${NC}"
kubectl apply -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
  namespace: app-prod
data:
  database.host: "postgres.prod.svc.cluster.local"
  database.port: "5432"
  app.loglevel: "info"
---
apiVersion: v1
kind: Secret
metadata:
  name: app-secret
  namespace: app-prod
type: Opaque
stringData:
  db-password: "secret-password-123"
  api-key: "my-api-key-12345"
---
apiVersion: v1
kind: Service
metadata:
  name: app-service
  namespace: app-prod
spec:
  selector:
    app: myapp
  ports:
  - name: http
    port: 80
    targetPort: 8080
  - name: metrics
    port: 9090
    targetPort: 9090
  type: ClusterIP
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app-deployment
  namespace: app-prod
  labels:
    app: myapp
    version: v1.0.0
spec:
  replicas: 3
  selector:
    matchLabels:
      app: myapp
  template:
    metadata:
      labels:
        app: myapp
        version: v1.0.0
    spec:
      containers:
      - name: app
        image: nginx:1.21
        ports:
        - containerPort: 8080
        env:
        - name: DB_HOST
          valueFrom:
            configMapKeyRef:
              name: app-config
              key: database.host
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: app-secret
              key: db-password
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: app-stateful
  namespace: app-prod
spec:
  serviceName: app-stateful
  replicas: 2
  selector:
    matchLabels:
      app: stateful-app
  template:
    metadata:
      labels:
        app: stateful-app
    spec:
      containers:
      - name: app
        image: redis:6.2
        ports:
        - containerPort: 6379
EOF

# Deploy matching resources to app-staging (identical to prod)
echo -e "\n${GREEN}📦 Deploying matching resources to app-staging...${NC}"
kubectl apply -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
  namespace: app-staging
data:
  database.host: "postgres.prod.svc.cluster.local"
  database.port: "5432"
  app.loglevel: "info"
---
apiVersion: v1
kind: Secret
metadata:
  name: app-secret
  namespace: app-staging
type: Opaque
stringData:
  db-password: "secret-password-123"
  api-key: "my-api-key-12345"
---
apiVersion: v1
kind: Service
metadata:
  name: app-service
  namespace: app-staging
spec:
  selector:
    app: myapp
  ports:
  - name: http
    port: 80
    targetPort: 8080
  - name: metrics
    port: 9090
    targetPort: 9090
  type: ClusterIP
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app-deployment
  namespace: app-staging
  labels:
    app: myapp
    version: v1.0.0
spec:
  replicas: 3
  selector:
    matchLabels:
      app: myapp
  template:
    metadata:
      labels:
        app: myapp
        version: v1.0.0
    spec:
      containers:
      - name: app
        image: nginx:1.21
        ports:
        - containerPort: 8080
        env:
        - name: DB_HOST
          valueFrom:
            configMapKeyRef:
              name: app-config
              key: database.host
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: app-secret
              key: db-password
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: app-stateful
  namespace: app-staging
spec:
  serviceName: app-stateful
  replicas: 2
  selector:
    matchLabels:
      app: stateful-app
  template:
    metadata:
      labels:
        app: stateful-app
    spec:
      containers:
      - name: app
        image: redis:6.2
        ports:
        - containerPort: 6379
EOF

# Deploy drifted resources to app-dev (with differences)
echo -e "\n${GREEN}📦 Deploying drifted resources to app-dev...${NC}"
kubectl apply -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
  namespace: app-dev
data:
  database.host: "postgres.dev.svc.cluster.local"
  database.port: "5432"
  app.loglevel: "debug"
  extra.setting: "new-value"
---
apiVersion: v1
kind: Secret
metadata:
  name: app-secret
  namespace: app-dev
type: Opaque
stringData:
  db-password: "different-password"
  api-key: "my-api-key-12345"
---
apiVersion: v1
kind: Service
metadata:
  name: app-service
  namespace: app-dev
spec:
  selector:
    app: myapp
  ports:
  - name: http
    port: 80
    targetPort: 8080
  type: ClusterIP
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app-deployment
  namespace: app-dev
  labels:
    app: myapp
    version: v1.1.0
spec:
  replicas: 5
  selector:
    matchLabels:
      app: myapp
  template:
    metadata:
      labels:
        app: myapp
        version: v1.1.0
    spec:
      containers:
      - name: app
        image: nginx:1.22
        ports:
        - containerPort: 8080
        env:
        - name: DB_HOST
          valueFrom:
            configMapKeyRef:
              name: app-config
              key: database.host
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: app-secret
              key: db-password
EOF

# Wait for deployments to be ready
echo -e "\n⏳ Waiting for deployments to be ready..."
kubectl wait --for=condition=available --timeout=60s deployment/app-deployment -n app-prod
kubectl wait --for=condition=available --timeout=60s deployment/app-deployment -n app-staging
kubectl wait --for=condition=available --timeout=60s deployment/app-deployment -n app-dev

# Generate infra config file
echo -e "\n${GREEN}📋 Generating infra config file...${NC}"
mkdir -p ~/.mc-tool

cat > ~/.mc-tool/infra-config.yaml <<EOCONFIG
sites:
  site1:
    name: site1
    endpoint: ${API_SERVER}
    token: ${PROD_TOKEN}
    insecure: true
  site2:
    name: site2
    endpoint: ${API_SERVER}
    token: ${STAGING_TOKEN}
    insecure: true
  site3:
    name: site3
    endpoint: ${API_SERVER}
    token: ${DEV_TOKEN}
    insecure: true
EOCONFIG

echo -e "${GREEN}✅ KinD test environment setup complete!${NC}"
echo ""
echo "Cluster: ${CLUSTER_NAME}"
echo "API Server: $API_SERVER"
echo ""
echo "Available sites (namespaces):"
echo "  - site1 → namespace: app-prod (BASELINE)"
echo "  - site2 → namespace: app-staging (MATCHING - identical to prod)"
echo "  - site3 → namespace: app-dev (DRIFTED - has differences)"
echo ""
echo "Config file created at: ~/.mc-tool/infra-config.yaml"
echo ""
echo -e "${YELLOW}Test with:${NC}"
echo "  ./mc-tool validate-infra site1/app-prod site2/app-staging"
echo "  ./mc-tool validate-infra site1/app-prod site3/app-dev"
echo "  ./mc-tool validate-infra site1/app-prod site2/app-staging site3/app-dev"
echo ""
echo -e "${YELLOW}Verify resources:${NC}"
echo "  kubectl get all -n app-prod"
echo "  kubectl get all -n app-staging"
echo "  kubectl get all -n app-dev"
echo ""
echo -e "${YELLOW}View differences:${NC}"
echo "  kubectl get configmap app-config -n app-prod -o yaml"
echo "  kubectl get configmap app-config -n app-dev -o yaml"
