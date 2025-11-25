package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// Architecture rules: adapters cannot import domain, domain cannot import adapters
var (
	adapterPaths = []string{
		"cmd/",
		"infra/",
		"ui/",
	}
	domainPaths = []string{
		"internal/core",
		"internal/services",
	}
)

func main() {
	errors := []string{}

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories and vendor
		if strings.HasPrefix(path, ".") || strings.Contains(path, "vendor") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Only process Go files
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Skip test files for now (can be added later)
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}

		// Parse the file
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return nil // Skip files that can't be parsed
		}

		// Determine if this file is in an adapter or domain path
		isAdapter := false
		isDomain := false
		for _, adapterPath := range adapterPaths {
			if strings.HasPrefix(path, adapterPath) {
				isAdapter = true
				break
			}
		}
		if !isAdapter {
			for _, domainPath := range domainPaths {
				if strings.HasPrefix(path, domainPath) {
					isDomain = true
					break
				}
			}
		}

		// Check imports
		for _, imp := range node.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)

			// Skip standard library and external packages
			if !strings.Contains(importPath, "github.com/gavin/gitta") {
				continue
			}

			// Extract the package path relative to module root
			modulePrefix := "github.com/gavin/gitta/"
			if !strings.HasPrefix(importPath, modulePrefix) {
				continue
			}
			relativePath := strings.TrimPrefix(importPath, modulePrefix)

			// Check violations
			if isAdapter {
				// Adapters cannot import domain
				for _, domainPath := range domainPaths {
					if strings.HasPrefix(relativePath, domainPath) {
						errors = append(errors, fmt.Sprintf(
							"%s: adapter cannot import domain package %s",
							path, importPath,
						))
					}
				}
			} else if isDomain {
				// Domain cannot import adapters
				for _, adapterPath := range adapterPaths {
					if strings.HasPrefix(relativePath, adapterPath) {
						errors = append(errors, fmt.Sprintf(
							"%s: domain cannot import adapter package %s",
							path, importPath,
						))
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking directory: %v\n", err)
		os.Exit(1)
	}

	if len(errors) > 0 {
		fmt.Fprintf(os.Stderr, "Architecture violations detected:\n\n")
		for _, err := range errors {
			fmt.Fprintf(os.Stderr, "  %s\n", err)
		}
		os.Exit(1)
	}

	fmt.Println("Architecture check passed!")
}
