package cli

import (
	"os"
	"path/filepath"
	"testing"
)

// createMockSuperCrewFiles creates a minimal SuperCrew structure for testing
func createMockSuperCrewFiles(t *testing.T, dir string) error {
	t.Helper()

	// Create SuperCrew directory structure
	superCrewDir := filepath.Join(dir, "SuperCrew")
	
	// Create directories
	dirs := []string{
		filepath.Join(superCrewDir, "Core"),
		filepath.Join(superCrewDir, "Commands"),
		filepath.Join(superCrewDir, "hooks"),
		filepath.Join(superCrewDir, "agents"),
	}
	
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return err
		}
	}
	
	// Create mock files
	files := map[string]string{
		filepath.Join(superCrewDir, "Core", "CLAUDE.md"): "# CLAUDE.md\nTest content",
		filepath.Join(superCrewDir, "Core", "FLAGS.md"): "# FLAGS.md\nTest flags",
		filepath.Join(superCrewDir, "Core", "PRINCIPLES.md"): "# PRINCIPLES.md\nTest principles",
		filepath.Join(superCrewDir, "Core", "RULES.md"): "# RULES.md\nTest rules",
		filepath.Join(superCrewDir, "Commands", "analyze.md"): "# analyze command",
		filepath.Join(superCrewDir, "Commands", "build.md"): "# build command",
		filepath.Join(superCrewDir, "hooks", "test-hook.sh"): "#!/bin/bash\necho test",
		filepath.Join(superCrewDir, "agents", "test.agent.md"): "# test agent",
	}
	
	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return err
		}
	}
	
	return nil
}

// setupTestEnvironment creates a complete test environment
func setupTestEnvironment(t *testing.T) (string, func()) {
	t.Helper()
	
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "crew-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	
	// Set environment variable to override project root detection
	oldPwd := os.Getenv("PWD")
	os.Setenv("PWD", tempDir)
	
	// Create mock SuperCrew files
	if err := createMockSuperCrewFiles(t, tempDir); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to create mock files: %v", err)
	}
	
	// Return cleanup function
	cleanup := func() {
		os.RemoveAll(tempDir)
		os.Setenv("PWD", oldPwd)
	}
	
	return tempDir, cleanup
}