#!/bin/bash
# Bootstrap verification script for Gitta
# Ensures the repository is properly initialized and ready for development

set -e

echo "🔍 Verifying Gitta bootstrap..."

# Check Go is installed and version
echo "Checking Go installation..."
if ! command -v go &> /dev/null; then
    echo "❌ Error: Go is not installed. Please install Go 1.21 or higher."
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
echo "✓ Go version: $GO_VERSION"

# Check go.mod exists
if [ ! -f "go.mod" ]; then
    echo "❌ Error: go.mod not found. Run 'go mod init github.com/gavin/gitta' first."
    exit 1
fi
echo "✓ go.mod found"

# Check module path
if ! grep -q "module github.com/gavin/gitta" go.mod; then
    echo "❌ Error: go.mod module path does not match expected 'github.com/gavin/gitta'"
    exit 1
fi
echo "✓ Module path correct"

# Check Go version in go.mod
if ! grep -q "go 1.21" go.mod && ! grep -q "go 1.22" go.mod; then
    echo "❌ Error: go.mod requires Go 1.21 or higher"
    exit 1
fi
echo "✓ Go version requirement met"

# Check GOEXPERIMENT is not set
if [ -n "$GOEXPERIMENT" ]; then
    echo "❌ Error: GOEXPERIMENT is set to '$GOEXPERIMENT' but must be disabled per constitution"
    exit 1
fi
echo "✓ GOEXPERIMENT not set"

# Run go mod tidy
echo "Running go mod tidy..."
go mod tidy
echo "✓ Dependencies resolved"

# Run make verify if Makefile exists
if [ -f "Makefile" ]; then
    echo "Running make verify..."
    make verify || {
        echo "⚠️  Warning: make verify failed. Some checks may need attention."
        exit 1
    }
    echo "✓ All verification checks passed"
else
    echo "⚠️  Warning: Makefile not found. Skipping verification."
fi

echo ""
echo "✅ Bootstrap verification complete! Repository is ready for development."

