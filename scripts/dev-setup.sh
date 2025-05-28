#!/bin/bash

# Development setup script
set -e

echo "🚀 Setting up development environment for System Monitor with Templ..."

# Check if Go is installed
if ! command -v go &>/dev/null; then
	echo "❌ Go is not installed. Please install Go first."
	exit 1
fi

echo "✅ Go is installed: $(go version)"

# Install templ
echo "📦 Installing templ..."
go install github.com/a-h/templ/cmd/templ@latest

# Install Air for live reloading
echo "📦 Installing Air for live reloading..."
go install github.com/air-verse/air@latest

# Install golangci-lint for linting
echo "📦 Installing golangci-lint..."
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Download dependencies
echo "📦 Downloading Go dependencies..."
go mod tidy
go mod download

# Generate templ files
echo "🔧 Generating templ files..."
templ generate

# Create necessary directories
echo "📁 Creating necessary directories..."
mkdir -p tmp
mkdir -p bin
mkdir -p static

# Copy static files
echo "📄 Setting up static files..."
cp styles.css static/ 2>/dev/null || true
cp script.js static/ 2>/dev/null || true

# Set executable permissions
chmod +x scripts/dev-setup.sh

echo "✅ Development environment setup complete!"
echo ""
echo "🎯 Quick start commands:"
echo "  make dev          - Start development server with live reload"
echo "  make dev-verbose  - Start with verbose output"
echo "  make templ-generate - Generate templ files"
echo "  make test         - Run tests"
echo "  make fmt          - Format code (Go + Templ)"
echo "  make lint         - Lint code"
echo ""
echo "🔥 Run 'make dev' to start developing!"
