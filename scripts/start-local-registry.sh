#!/bin/bash
# Start a local Docker registry for Tilt/KIND development

set -e

REGISTRY_NAME="registry"
REGISTRY_PORT="5000"

# Check if registry is already running
if docker ps | grep -q "$REGISTRY_NAME"; then
    echo "✅ Registry '$REGISTRY_NAME' is already running"
else
    echo "🚀 Starting local registry on port $REGISTRY_PORT..."
    docker run -d \
        --name "$REGISTRY_NAME" \
        --restart=unless-stopped \
        -p "127.0.0.1:${REGISTRY_PORT}:5000" \
        registry:2

    echo "✅ Registry started at localhost:$REGISTRY_PORT"
fi

# Optionally start port-forwarding for the cluster
# Uncomment if you need the registry accessible from within Kubernetes
# kubectl port-forward -n kube-system svc/docker-registry 5000:5000 &

echo ""
echo "📋 To use with Tilt/KIND, ensure your cluster can reach:"
echo "   - localhost:$REGISTRY_PORT (from Docker host)"
echo "   - or registry.docker-registry:5000 (from within Kubernetes)"
