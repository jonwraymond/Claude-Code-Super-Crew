package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/jonwraymond/claude-code-super-crew/internal/cli"
)

func getVersion() string {
	// Use embedded version for security and reliability
	const defaultVersion = "1.0.0"
	
	// Only read VERSION file if it exists in a safe location
	versionFile := "VERSION"
	
	// Validate the path to prevent directory traversal
	if strings.Contains(versionFile, "..") || strings.ContainsAny(versionFile, "/\\") {
		return defaultVersion
	}
	
	content, err := os.ReadFile(versionFile)
	if err != nil {
		// Log error but don't expose file system details
		return defaultVersion
	}
	
	version := strings.TrimSpace(string(content))
	// Validate version format (basic semver pattern)
	if len(version) == 0 || len(version) > 20 {
		return defaultVersion
	}
	
	// Basic validation for version format (x.y.z)
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return defaultVersion
	}
	
	return version
}

func main() {
	rootCmd := cli.NewRootCommand(getVersion())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
