//go:build tools
// +build tools

// Package tools manages tool dependencies via go.mod.
// To install all tools, run: go install ./tools
package tools

import (
	_ "golang.org/x/tools/cmd/goimports"
	_ "golang.org/x/vuln/cmd/govulncheck"
	_ "honnef.co/go/tools/cmd/staticcheck"
)
