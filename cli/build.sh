#!/bin/bash

# Build script for Shipyard CLI

set -e

echo "🏗️  Building Shipyard CLI..."

# Clean previous builds
rm -f shipyard

# Build for current platform
go build -ldflags="-s -w" -o shipyard main.go

echo "✅ Build complete: ./shipyard"

# Make executable
chmod +x shipyard

echo "🚀 Ready to use:"
echo "   ./shipyard --help"