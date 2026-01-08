#!/bin/bash
set -e

# Script to cleanup KinD test environment

echo "🧹 Cleaning up KinD test environment..."

kind delete cluster --name infra-test 2>/dev/null || true

echo "✅ Cleanup complete!"
