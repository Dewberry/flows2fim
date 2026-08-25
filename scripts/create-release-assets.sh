#!/usr/bin/env bash

set -euo pipefail

# Script: create-release-assets.sh
# Description: Creates release archives and checksums for all platform binaries
# Usage: ./create-release-assets.sh

echo "Starting release asset creation process..."

# Create release assets directory
mkdir -p release-assets

echo "Creating release archives..."
zip -j release-assets/flows2fim-windows-amd64.zip builds/windows-amd64/flows2fim.exe
tar -czvf release-assets/flows2fim-darwin-arm64.tar.gz -C builds/darwin-arm64 flows2fim
tar -czvf release-assets/flows2fim-linux-amd64.tar.gz -C builds/linux-amd64 flows2fim
tar -czvf release-assets/flows2fim-linux-arm64.tar.gz -C builds/linux-arm64 flows2fim

echo "Release assets created:"
ls -lh release-assets/

echo "Generating checksums..."
cd release-assets
sha256sum *.tar.gz *.zip > checksums.txt
echo "Checksums generated:"
cat checksums.txt
cd ..

echo "Release assets creation completed successfully"
