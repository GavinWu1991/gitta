.PHONY: verify fmt vet test staticcheck govulncheck check-architecture install-tools release-snapshot

# Default target
.DEFAULT_GOAL := verify

# Install required tools
install-tools:
	@echo "Installing required tools..."
	@go install ./tools
	@echo "Tools installed successfully"

# Format code
fmt:
	@echo "Running go fmt..."
	@go fmt ./...

# Run goimports (requires installed tool)
imports:
	@echo "Running goimports..."
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	else \
		echo "Warning: goimports not found. Install with: go install golang.org/x/tools/cmd/goimports@latest"; \
	fi

# Run go vet
vet:
	@echo "Running go vet..."
	@go vet ./...

# Run staticcheck (requires installed tool)
staticcheck:
	@echo "Running staticcheck..."
	@if command -v staticcheck >/dev/null 2>&1; then \
		staticcheck ./...; \
	else \
		echo "Warning: staticcheck not found. Install with: go install honnef.co/go/tools/cmd/staticcheck@latest"; \
	fi

# Run govulncheck (requires installed tool)
govulncheck:
	@echo "Running govulncheck..."
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./...; \
	else \
		echo "Warning: govulncheck not found. Install with: go install golang.org/x/vuln/cmd/govulncheck@latest"; \
	fi

# Run tests
test:
	@echo "Running tests..."
	@go test ./...

# Check architecture boundaries
check-architecture:
	@echo "Checking architecture boundaries..."
	@go run ./tools/check-architecture.go

# GoReleaser snapshot (dry-run, no publish)
release-snapshot:
	@echo "Running GoReleaser snapshot (no publish)..."
	@if command -v goreleaser >/dev/null 2>&1; then \
		goreleaser release --snapshot --skip-publish --clean; \
	else \
		echo "Warning: goreleaser not found. Install with: go install github.com/goreleaser/goreleaser@latest"; \
	fi

# Run all verification steps
verify: fmt imports vet staticcheck govulncheck check-architecture test
	@echo "All verification checks passed!"

