#!/usr/bin/env bash

set -euo pipefail

# Script: create-release-binaries.sh
# Description: Builds all platform binaries for release using Docker Compose
# Usage: ./create-release-binaries.sh

echo "Starting release binary build process..."

# Get container ID using compose project label for reliability
CONTAINER_ID=$(docker compose ps -q | head -n 1)
if [ -z "$CONTAINER_ID" ]; then
  echo "Error: No running containers found from docker compose"
  exit 1
fi

echo "Using container: $CONTAINER_ID"

# Track build failures
BUILD_FAILED=0

echo "Building Darwin ARM64..."
if ! docker exec $CONTAINER_ID /bin/bash -c "./scripts/build-darwin-arm64.sh"; then
  echo "Error: Darwin ARM64 build failed"
  BUILD_FAILED=1
fi

echo "Building Linux AMD64..."
if ! docker exec $CONTAINER_ID /bin/bash -c "./scripts/build-linux-amd64.sh"; then
  echo "Error: Linux AMD64 build failed"
  BUILD_FAILED=1
fi

echo "Building Linux ARM64..."
if ! docker exec $CONTAINER_ID /bin/bash -c "./scripts/build-linux-arm64.sh"; then
  echo "Error: Linux ARM64 build failed"
  BUILD_FAILED=1
fi

echo "Building Windows AMD64..."
if ! docker exec $CONTAINER_ID /bin/bash -c "./scripts/build-windows-amd64.sh"; then
  echo "Error: Windows AMD64 build failed"
  BUILD_FAILED=1
fi

docker compose down

if [ $BUILD_FAILED -eq 1 ]; then
  echo "One or more builds failed"
  exit 1
fi

echo "All builds completed successfully"
