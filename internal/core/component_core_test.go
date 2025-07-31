package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jonwraymond/claude-code-super-crew/pkg/logger"
)

func TestCoreComponent_InstallOrchestratorAgent(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "core-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create source directory structure that matches expected path
	// The core component expects the agent at sourceDir/../agents/orchestrator-agent.md
	coreSourceDir := filepath.Join(tempDir, "source", "core")
	agentsDir := filepath.Join(tempDir, "source", "agents")
	if err := os.MkdirAll(coreSourceDir, 0755); err != nil {
		t.Fatalf("Failed to create core source dir: %v", err)
	}
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("Failed to create agents dir: %v", err)
	}

	// Create a test orchestrator agent file
	orchestratorContent := `---
name: orchestrator-agent
description: Test orchestrator agent
version: "1.0.0"
---

Test orchestrator agent content`

	orchestratorPath := filepath.Join(agentsDir, "orchestrator-agent.md")
	if err := os.WriteFile(orchestratorPath, []byte(orchestratorContent), 0644); err != nil {
		t.Fatalf("Failed to create test orchestrator agent: %v", err)
	}

	// Create install directory
	installDir := filepath.Join(tempDir, ".claude")
	if err := os.MkdirAll(filepath.Join(installDir, "agents"), 0755); err != nil {
		t.Fatalf("Failed to create install agents dir: %v", err)
	}

	// Initialize logger
	log := logger.NewNamedLogger("test")
	log.SetLevel(logger.DebugLevel)

	// Create core component with coreSourceDir so it looks for agent at ../agents/
	c := NewCoreComponent(installDir, coreSourceDir)

	// Test installOrchestratorAgent
	if err := c.installOrchestratorAgent(installDir); err != nil {
		t.Errorf("installOrchestratorAgent failed: %v", err)
	}

	// Verify the file was copied
	targetPath := filepath.Join(installDir, "agents", "orchestrator-agent.md")
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		t.Error("Orchestrator agent was not copied to target location")
	}

	// Verify the content
	content, err := os.ReadFile(targetPath)
	if err != nil {
		t.Errorf("Failed to read installed orchestrator agent: %v", err)
	}

	if string(content) != orchestratorContent {
		t.Error("Installed orchestrator agent content does not match source")
	}

	// Verify permissions
	info, err := os.Stat(targetPath)
	if err != nil {
		t.Errorf("Failed to stat orchestrator agent: %v", err)
	}

	if info.Mode().Perm() != 0644 {
		t.Errorf("Orchestrator agent has incorrect permissions: %v", info.Mode().Perm())
	}
}

func TestCoreComponent_ValidateInstallation_WithOrchestratorAgent(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "validate-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize logger
	log := logger.NewNamedLogger("test")
	log.SetLevel(logger.DebugLevel)

	// Test validation without orchestrator agent
	installDir := filepath.Join(tempDir, ".claude")
	
	// Create core component with the correct install directory
	c := NewCoreComponent(installDir, "")
	if err := os.MkdirAll(filepath.Join(installDir, "agents"), 0755); err != nil {
		t.Fatalf("Failed to create agents dir: %v", err)
	}

	// Create crew-metadata.json for base validation to pass
	configDir := filepath.Join(installDir, ".crew", "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	metadataPath := filepath.Join(configDir, "crew-metadata.json")
	metadataContent := `{
		"framework": {
			"version": "1.0.0",
			"release_date": "2025-01-01",
			"updated_at": "2025-01-01T00:00:00Z"
		},
		"components": {
			"core": {
				"version": "1.0.0",
				"updated_at": "2025-01-01T00:00:00Z",
				"status": "installed"
			}
		},
		"documents": {},
		"features": null,
		"installation": {
			"install_dir": "` + installDir + `",
			"installed_at": "2025-01-01T00:00:00Z",
			"last_updated": "2025-01-01T00:00:00Z",
			"installer_version": "1.0.0"
		}
	}`
	if err := os.WriteFile(metadataPath, []byte(metadataContent), 0644); err != nil {
		t.Fatalf("Failed to create crew-metadata.json: %v", err)
	}

	// Debug: check what the component thinks its install dir is
	t.Logf("Component install dir: %s", c.InstallDir)
	t.Logf("Test install dir: %s", installDir)
	
	isValid, errors := c.ValidateInstallation(installDir)
	if isValid {
		t.Error("Validation should fail without orchestrator agent")
	}

	found := false
	for _, err := range errors {
		if err == "orchestrator-agent.md not found in agents directory" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Validation did not report missing orchestrator agent")
	}

	// Now add the orchestrator agent
	orchestratorPath := filepath.Join(installDir, "agents", "orchestrator-agent.md")
	if err := os.WriteFile(orchestratorPath, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create orchestrator agent: %v", err)
	}

	// Validation should now pass
	isValid, errors = c.ValidateInstallation(installDir)
	if !isValid {
		t.Errorf("Validation failed with orchestrator agent present: %v", errors)
	}
}

func TestCoreComponent_InstallOrchestratorAgent_MissingFile(t *testing.T) {
	// Test that the installer fails gracefully when orchestrator agent is missing
	tempDir, err := os.MkdirTemp("", "missing-file-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize logger
	log := logger.NewNamedLogger("test")
	log.SetLevel(logger.DebugLevel)

	// Create core component with source that doesn't have the agent file
	sourceDir := filepath.Join(tempDir, "source", "core")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("Failed to create source dir: %v", err)
	}

	c := NewCoreComponent(tempDir, sourceDir)

	// Create install directory
	installDir := filepath.Join(tempDir, ".claude")
	if err := os.MkdirAll(filepath.Join(installDir, "agents"), 0755); err != nil {
		t.Fatalf("Failed to create install agents dir: %v", err)
	}

	// Test should fail when orchestrator agent is not at expected path
	err = c.installOrchestratorAgent(installDir)
	if err == nil {
		t.Error("Expected error when orchestrator agent not found, got nil")
	}

	// Verify error message mentions the expected path
	expectedPath := filepath.Join(sourceDir, "..", "agents", "orchestrator-agent.md")
	if !strings.Contains(err.Error(), expectedPath) {
		t.Errorf("Error message should mention expected path %s, got: %v", expectedPath, err)
	}
}
