#!/bin/bash
set -e

# Script to setup KinD clusters for infra validation testing with service accounts

echo "🚀 Setting up KinD test environment for infra validation..."

# Check if KinD is installed
if ! command -v kind &> /dev/null; then
    echo "❌ KinD is not installed. Please install it first:"
    echo "   https://kind.sigs.k8s.io/docs/user/quick-start/#installation"
    exit 1
fi

# Check if kubectl is installed
if ! command -v kubectl &> /dev/null; then
    echo "❌ kubectl is not installed. Please install it first."
    exit 1
fi

# Cleanup existing clusters if they exist
echo "🧹 Cleaning up existing test clusters..."
kind delete cluster --name test-baseline 2>/dev/null || true
kind delete cluster --name test-target1 2>/dev/null || true
kind delete cluster --name test-target2 2>/dev/null || true

# Create baseline cluster with custom API server port
echo "📦 Creating baseline cluster..."
cat << EOF | kind create cluster --name test-baseline --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
EOF

# Create target cluster 1
echo "📦 Creating target cluster 1..."
cat << EOF | kind create cluster --name test-target1 --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
EOF

# Create target cluster 2
echo "📦 Creating target cluster 2..."
cat << EOF | kind create cluster --name test-target2 --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
EOF

# Function to create service account and get token
create_service_account() {
    local CONTEXT=$1
    local NAMESPACE=$2
    local SA_NAME="mc-tool-validator"
    
    echo "  Creating service account in $CONTEXT/$NAMESPACE..."
    
    # Create namespace
    kubectl --context $CONTEXT create namespace $NAMESPACE 2>/dev/null || true
    
    # Create service account
    kubectl --context $CONTEXT create sa $SA_NAME -n $NAMESPACE 2>/dev/null || true
    
    # Create cluster role binding
    kubectl --context $CONTEXT create clusterrolebinding ${SA_NAME}-${NAMESPACE}-binding \
        --clusterrole=cluster-admin \
        --serviceaccount=${NAMESPACE}:${SA_NAME} 2>/dev/null || true
    
    # Create secret for service account (Kubernetes 1.24+)
    kubectl --context $CONTEXT apply -f - <<EOSECRET 2>/dev/null || true
apiVersion: v1
kind: Secret
metadata:
  name: ${SA_NAME}-token
  namespace: ${NAMESPACE}
  annotations:
    kubernetes.io/service-account.name: ${SA_NAME}
type: kubernetes.io/service-account-token
EOSECRET
    
    # Wait for token to be created
    echo "  Waiting for token..."
    sleep 3
    
    # Get token
    TOKEN=$(kubectl --context $CONTEXT get secret ${SA_NAME}-token -n $NAMESPACE -o jsonpath='{.data.token}' 2>/dev/null | base64 -d)
    echo "$TOKEN"
}

# Get cluster endpoints from kubeconfig
get_endpoint() {
    local CONTEXT=$1
    kubectl --context $CONTEXT config view --minify -o jsonpath='{.clusters[0].cluster.server}'
}

# Create namespaces and service accounts
echo "📝 Creating test namespaces and service accounts..."

# Baseline
BASELINE_ENDPOINT=$(get_endpoint kind-test-baseline)
BASELINE_TOKEN=$(create_service_account kind-test-baseline app-prod)

# Target 1
TARGET1_ENDPOINT=$(get_endpoint kind-test-target1)
TARGET1_TOKEN=$(create_service_account kind-test-target1 app-prod)

# Target 2
TARGET2_ENDPOINT=$(get_endpoint kind-test-target2)
TARGET2_TOKEN=$(create_service_account kind-test-target2 app-prod-replica)

echo "🎯 Deploying resources to clusters..."

# Deploy baseline resources
kubectl --context kind-test-baseline apply -n app-prod -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  database.host: "postgres.prod.svc.cluster.local"
  database.port: "5432"
  app.loglevel: "info"
---
apiVersion: v1
kind: Secret
metadata:
  name: app-secret
type: Opaque
stringData:
  db-password: "super-secret-password"
  api-key: "my-api-key-12345"
---
apiVersion: v1
kind: Service
metadata:
  name: app-service
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
spec:
  replicas: 3
  selector:
    matchLabels:
      app: myapp
  template:
    metadata:
      labels:
        app: myapp
    spec:
      containers:
      - name: app
        image: nginx:1.21
        ports:
        - containerPort: 8080
EOF

# Deploy matching resources to target1
kubectl --context kind-test-target1 apply -n app-prod -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  database.host: "postgres.prod.svc.cluster.local"
  database.port: "5432"
  app.loglevel: "info"
---
apiVersion: v1
kind: Secret
metadata:
  name: app-secret
type: Opaque
stringData:
  db-password: "super-secret-password"
  api-key: "my-api-key-12345"
---
apiVersion: v1
kind: Service
metadata:
  name: app-service
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
spec:
  replicas: 3
  selector:
    matchLabels:
      app: myapp
  template:
    metadata:
      labels:
        app: myapp
    spec:
      containers:
      - name: app
        image: nginx:1.21
        ports:
        - containerPort: 8080
EOF

# Deploy drifted resources to target2
kubectl --context kind-test-target2 apply -n app-prod-replica -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  database.host: "postgres.staging.svc.cluster.local"
  database.port: "5432"
  app.loglevel: "debug"
  extra.setting: "new-value"
---
apiVersion: v1
kind: Secret
metadata:
  name: app-secret
type: Opaque
stringData:
  db-password: "different-password"
  api-key: "my-api-key-12345"
---
apiVersion: v1
kind: Service
metadata:
  name: app-service
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
spec:
  replicas: 5
  selector:
    matchLabels:
      app: myapp
  template:
    metadata:
      labels:
        app: myapp
    spec:
      containers:
      - name: app
        image: nginx:1.22
        ports:
        - containerPort: 8080
EOF

# Generate infra config file
echo "📋 Generating infra config file..."
mkdir -p ~/.mc-tool

cat > ~/.mc-tool/infra-config.yaml <<EOCONFIG
sites:
  site1:
    name: site1
    endpoint: ${BASELINE_ENDPOINT}
    token: ${BASELINE_TOKEN}
    insecure: true
  site2:
    name: site2
    endpoint: ${TARGET1_ENDPOINT}
    token: ${TARGET1_TOKEN}
    insecure: true
  site3:
    name: site3
    endpoint: ${TARGET2_ENDPOINT}
    token: ${TARGET2_TOKEN}
    insecure: true
EOCONFIG

echo "✅ KinD test environment setup complete!"
echo ""
echo "Available sites:"
echo "  - site1 (kind-test-baseline, endpoint: $BASELINE_ENDPOINT, namespace: app-prod)"
echo "  - site2 (kind-test-target1, endpoint: $TARGET1_ENDPOINT, namespace: app-prod) - MATCHING"
echo "  - site3 (kind-test-target2, endpoint: $TARGET2_ENDPOINT, namespace: app-prod-replica) - DRIFTED"
echo ""
echo "Config file created at: ~/.mc-tool/infra-config.yaml"
echo ""
echo "Test with:"
echo "  ./mc-tool-test validate-infra site1/app-prod site2/app-prod site3/app-prod-replica"
