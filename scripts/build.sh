#!/bin/bash

# Exit on any error
set -e

# Go to the project root directory
cd "$(dirname "$0")/.."

# Print the current directory
echo "Building from $(pwd)"

# Create the bin directory if it doesn't exist
mkdir -p bin

# Build with version information
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "unknown")
BUILD_TIME=$(date -u '+%Y-%m-%d %H:%M:%S')

echo "Building version: $VERSION"
echo "Build time: $BUILD_TIME"

# Build for multiple platforms
echo "Building Linux amd64..."
GOOS=linux GOARCH=amd64 go build -ldflags "-s -w -X main.Version=$VERSION -X 'main.BuildTime=$BUILD_TIME'" -o bin/wssht-linux-amd64 cmd/proxy/main.go

echo "Building Linux arm64..."
GOOS=linux GOARCH=arm64 go build -ldflags "-s -w -X main.Version=$VERSION -X 'main.BuildTime=$BUILD_TIME'" -o bin/wssht-linux-arm64 cmd/proxy/main.go

echo "Building Linux arm..."
GOOS=linux GOARCH=arm go build -ldflags "-s -w -X main.Version=$VERSION -X 'main.BuildTime=$BUILD_TIME'" -o bin/wssht-linux-arm cmd/proxy/main.go

# Make a symlink to the linux-amd64 version as the default
ln -sf wssht-linux-amd64 bin/wssht

echo "Build completed successfully!"
echo "Binary location: bin/wssht"