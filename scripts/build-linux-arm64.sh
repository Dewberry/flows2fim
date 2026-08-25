#!/bin/bash

# The script must be executed from the root of the repository & within the docker container
set -eo pipefail

echo "Building for Linux ARM64..."
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -buildvcs=false \
    -ldflags="-X main.GitTag=$(git describe --tags --always --dirty) -X main.GitCommit=$(git rev-parse --short HEAD) -X main.BuildDate=$(date +%Y-%m-%d)" \
    -o builds/linux-arm64/flows2fim main.go
echo "Linux ARM64 build completed."

chmod +x builds/linux-arm64/flows2fim
