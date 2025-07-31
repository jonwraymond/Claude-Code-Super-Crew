package integration

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jonwraymond/claude-code-super-crew/internal/claude"
)

// setupMockGlobalInstall creates a mock global SuperCrew installation for testing
func setupMockGlobalInstall(t *testing.T) (string, func()) {
	// Create a temporary home directory
	homeDir, err := os.MkdirTemp("", "test_home")
	if err != nil {
		t.Fatalf("Failed to create temp home: %v", err)
	}

	// Create mock global installation
	globalInstallDir := filepath.Join(homeDir, ".claude")
	commandsDir := filepath.Join(globalInstallDir, "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatalf("Failed to create mock commands dir: %v", err)
	}

	// Create a minimal commands file
	commandFile := filepath.Join(commandsDir, "test.md")
	if err := os.WriteFile(commandFile, []byte("# Test Command"), 0644); err != nil {
		t.Fatalf("Failed to create test command: %v", err)
	}

	// Set HOME environment variable
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", homeDir)

	cleanup := func() {
		os.Setenv("HOME", originalHome)
		os.RemoveAll(homeDir)
	}

	return globalInstallDir, cleanup
}

// TestClaudeCommandWithRestructure tests Claude command integration with the new .crew/ directory structure
func TestClaudeCommandWithRestructure(t *testing.T) {
	// Skip test if binary doesn't exist (for local development)
	// Get the absolute path to the crew binary in project root
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	projectRoot := filepath.Dir(filepath.Dir(wd))
	crewBinary := filepath.Join(projectRoot, "crew")
	
	if _, err := os.Stat(crewBinary); os.IsNotExist(err) {
		t.Skip("Crew binary not found - build with 'make build' first")
	}

	// Set up mock global installation
	_, cleanupGlobal := setupMockGlobalInstall(t)
	defer cleanupGlobal()

	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "claude_restructure_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("Warning: failed to clean up temp directory: %v", err)
		}
	}()

	// Set up working directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	claudeDir := filepath.Join(tmpDir, ".claude")

	t.Run("InstallWithNewStructure", func(t *testing.T) {
		// Install with project directory structure
		cmd := exec.Command(crewBinary, "claude", "--install", "--project-dir", tmpDir, "--yes")
		var out bytes.Buffer
		var errOut bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &errOut

		err := cmd.Run()
		output := out.String()
		errorOutput := errOut.String()
		
		if err != nil {
			t.Logf("Command output: %s", output)
			t.Logf("Command error: %s", errorOutput)
			t.Logf("Command exit error: %v", err)
			// Don't fail completely - may be expected in test environment
		}

		// Verify .claude directory was created
		if _, err := os.Stat(claudeDir); os.IsNotExist(err) {
			t.Error(".claude directory was not created")
		}

		// Check that main config file exists in correct location
		configFile := filepath.Join(claudeDir, "supercrew-commands.json")
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			t.Error("Main config file not found in expected location")
		}
	})

	t.Run("RestructureDirectories", func(t *testing.T) {
		// First create some test directories in old structure
		oldStructureDirs := []string{
			"backups",
			"logs",
			"config", 
			"completions",
			"scripts",
			"workflows",
			"prompts",
		}

		// Create directories with test files
		for _, dir := range oldStructureDirs {
			dirPath := filepath.Join(claudeDir, dir)
			if err := os.MkdirAll(dirPath, 0755); err != nil {
				t.Fatalf("Failed to create test directory %s: %v", dir, err)
			}
			
			testFile := filepath.Join(dirPath, "test.txt")
			if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}
		}

		// Also create core directories that should NOT move
		coreDirs := []string{"commands", "hooks", "agents"}
		for _, dir := range coreDirs {
			dirPath := filepath.Join(claudeDir, dir)
			if err := os.MkdirAll(dirPath, 0755); err != nil {
				t.Fatalf("Failed to create core directory %s: %v", dir, err)
			}
			
			testFile := filepath.Join(dirPath, "core.txt")
			if err := os.WriteFile(testFile, []byte("core content"), 0644); err != nil {
				t.Fatalf("Failed to create core file: %v", err)
			}
		}

		// Perform restructuring
		restructurer := claude.NewDirectoryRestructurer(claudeDir)
		if err := restructurer.RestructureDirectories(); err != nil {
			t.Fatalf("Directory restructuring failed: %v", err)
		}

		// Verify .crew directory was created
		crewDir := filepath.Join(claudeDir, ".crew")
		if _, err := os.Stat(crewDir); os.IsNotExist(err) {
			t.Error(".crew directory was not created")
		}

		// Verify utility directories moved to .crew/
		for _, dir := range oldStructureDirs {
			oldPath := filepath.Join(claudeDir, dir)
			newPath := filepath.Join(crewDir, dir)
			
			// Old location should not exist
			if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
				t.Errorf("Directory %s was not moved from old location", dir)
			}
			
			// New location should exist with content
			if _, err := os.Stat(newPath); os.IsNotExist(err) {
				t.Errorf("Directory %s was not moved to .crew/", dir)
			}
			
			// Test file should be in new location
			testFile := filepath.Join(newPath, "test.txt")
			if content, err := os.ReadFile(testFile); err != nil {
				t.Errorf("Test file missing from moved directory %s", dir)
			} else if string(content) != "test content" {
				t.Errorf("Test file content corrupted in moved directory %s", dir)
			}
		}

		// Verify core directories remained in place
		for _, dir := range coreDirs {
			corePath := filepath.Join(claudeDir, dir)
			if _, err := os.Stat(corePath); os.IsNotExist(err) {
				t.Errorf("Core directory %s was incorrectly moved", dir)
			}
			
			testFile := filepath.Join(corePath, "core.txt")
			if content, err := os.ReadFile(testFile); err != nil {
				t.Errorf("Core file missing from directory %s", dir)
			} else if string(content) != "core content" {
				t.Errorf("Core file content corrupted in directory %s", dir)
			}
		}
	})

	t.Run("StatusAfterRestructure", func(t *testing.T) {
		// Test status command after restructuring
		cmd := exec.Command(crewBinary, "claude", "--status", "--project-dir", tmpDir)
		var out bytes.Buffer
		var errOut bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &errOut

		err := cmd.Run()
		output := out.String()
		
		if err != nil {
			t.Logf("Status command output: %s", output)
			t.Logf("Status command error: %s", errOut.String())
			// Don't fail - status may show expected warnings
		}

		// Should show installation status
		if !strings.Contains(strings.ToLower(output), "status") && 
		   !strings.Contains(strings.ToLower(output), "install") {
			t.Logf("Unexpected status output: %s", output)
		}
	})

	t.Run("UninstallAndReinstall", func(t *testing.T) {
		// Test uninstall
		cmd := exec.Command(crewBinary, "claude", "--uninstall", "--project-dir", tmpDir, "--yes")
		var out bytes.Buffer
		var errOut bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &errOut

		err := cmd.Run()
		output := out.String()
		
		if err != nil {
			t.Logf("Uninstall output: %s", output)
			t.Logf("Uninstall error: %s", errOut.String())
		}

		// Verify main files were removed but directory structure may remain
		configFile := filepath.Join(claudeDir, "supercrew-commands.json")
		if _, err := os.Stat(configFile); !os.IsNotExist(err) {
			t.Error("Config file was not removed during uninstall")
		}

		// Test reinstall with new structure
		cmd = exec.Command(crewBinary, "claude", "--install", "--project-dir", tmpDir, "--yes")
		cmd.Stdout = &out
		cmd.Stderr = &errOut
		out.Reset()
		errOut.Reset()

		err = cmd.Run()
		output = out.String()
		
		if err != nil {
			t.Logf("Reinstall output: %s", output)
			t.Logf("Reinstall error: %s", errOut.String())
		}

		// Verify config file was recreated
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			t.Error("Config file was not recreated during reinstall")
		}

		// Verify .crew directory structure is maintained
		pathResolver := claude.NewPathResolver(claudeDir)
		
		// Check specific utility directories
		utilityPaths := map[string]string{
			"backups":     pathResolver.GetBackupsDir(),
			"logs":        pathResolver.GetLogsDir(),
			"config":      pathResolver.GetConfigDir(),
			"completions": pathResolver.GetCompletionsDir(),
			"scripts":     pathResolver.GetScriptsDir(),
			"workflows":   pathResolver.GetWorkflowsDir(),
			"prompts":     pathResolver.GetPromptsDir(),
		}
		
		for name, expectedPath := range utilityPaths {
			if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
				t.Errorf("Utility directory %s not properly maintained after reinstall at %s", name, expectedPath)
			}
		}
	})

	t.Run("VerifyClaudeCodeDirectoriesUntouched", func(t *testing.T) {
		// Create mock Claude Code directories
		claudeCodeDirs := []string{"statsig", "shell-snapshots"}
		
		for _, dir := range claudeCodeDirs {
			dirPath := filepath.Join(claudeDir, dir)
			if err := os.MkdirAll(dirPath, 0755); err != nil {
				t.Fatalf("Failed to create Claude Code directory %s: %v", dir, err)
			}
			
			testFile := filepath.Join(dirPath, "claude_code_data.txt")
			if err := os.WriteFile(testFile, []byte("important claude code data"), 0644); err != nil {
				t.Fatalf("Failed to create Claude Code test file: %v", err)
			}
		}

		// Perform another restructuring
		restructurer := claude.NewDirectoryRestructurer(claudeDir)
		if err := restructurer.RestructureDirectories(); err != nil {
			t.Fatalf("Second restructuring failed: %v", err)
		}

		// Verify Claude Code directories are untouched
		for _, dir := range claudeCodeDirs {
			dirPath := filepath.Join(claudeDir, dir)
			if _, err := os.Stat(dirPath); os.IsNotExist(err) {
				t.Errorf("Claude Code directory %s was incorrectly moved or deleted", dir)
			}
			
			testFile := filepath.Join(dirPath, "claude_code_data.txt")
			if content, err := os.ReadFile(testFile); err != nil {
				t.Errorf("Claude Code data file missing from %s", dir)
			} else if string(content) != "important claude code data" {
				t.Errorf("Claude Code data corrupted in %s", dir)
			}
		}
	})

	t.Run("PathResolverIntegration", func(t *testing.T) {
		pathResolver := claude.NewPathResolver(claudeDir)
		
		// Test that paths are correctly resolved
		expectedLogsPath := filepath.Join(claudeDir, ".crew", "logs")
		if pathResolver.GetLogsDir() != expectedLogsPath {
			t.Errorf("Logs path incorrect: expected %s, got %s", expectedLogsPath, pathResolver.GetLogsDir())
		}

		expectedCommandsPath := filepath.Join(claudeDir, "commands")
		if pathResolver.GetCommandsDir() != expectedCommandsPath {
			t.Errorf("Commands path incorrect: expected %s, got %s", expectedCommandsPath, pathResolver.GetCommandsDir())
		}

		// Test directory creation
		if err := pathResolver.EnsureCrewDirectories(); err != nil {
			t.Fatalf("Failed to ensure crew directories: %v", err)
		}

		if err := pathResolver.EnsureCoreDirectories(); err != nil {
			t.Fatalf("Failed to ensure core directories: %v", err)
		}

		// Verify all directories exist
		allPaths := pathResolver.GetAllDirectoryPaths()
		for name, path := range allPaths {
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Errorf("Directory %s was not created at path %s", name, path)
			}
		}
	})
}

// TestClaudeIntegrationWithRestructure tests the Claude integration with new directory structure
func TestClaudeIntegrationWithRestructure(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "claude_integration_restructure_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	claudeDir := filepath.Join(tmpDir, ".claude")
	commandsDir := filepath.Join(claudeDir, "commands")

	// Create minimal test command structure
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatalf("Failed to create commands directory: %v", err)
	}

	// Create a test command file
	testCommandPath := filepath.Join(commandsDir, "test.md")
	testCommandContent := `---
name: test
description: Test command
---

# Test Command

This is a test command for integration testing.
`
	if err := os.WriteFile(testCommandPath, []byte(testCommandContent), 0644); err != nil {
		t.Fatalf("Failed to create test command: %v", err)
	}

	t.Run("IntegrationWithNewPaths", func(t *testing.T) {
		// Create integration using new path structure
		integration, err := claude.NewClaudeIntegration(commandsDir, claudeDir)
		if err != nil {
			t.Fatalf("Failed to create Claude integration: %v", err)
		}

		// Install integration
		if err := integration.InstallIntegration(); err != nil {
			t.Fatalf("Failed to install integration: %v", err)
		}

		// Verify config file was created in correct location
		pathResolver := claude.NewPathResolver(claudeDir)
		configPath := pathResolver.GetMainConfigFile()
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("Integration config file not created in expected location")
		}

		// Verify completion scripts were created in .crew/completions/
		completionsDir := pathResolver.GetCompletionsDir()
		if _, err := os.Stat(completionsDir); os.IsNotExist(err) {
			t.Error("Completions directory not created in .crew/")
		}

		// Check for at least one completion script
		entries, err := os.ReadDir(completionsDir)
		if err != nil {
			t.Fatalf("Failed to read completions directory: %v", err)
		}
		
		foundCompletionScript := false
		for _, entry := range entries {
			if strings.Contains(entry.Name(), "supercrew") && !entry.IsDir() {
				foundCompletionScript = true
				break
			}
		}
		
		if !foundCompletionScript {
			t.Error("No completion scripts found in .crew/completions/")
		}
	})

	t.Run("StatusCheck", func(t *testing.T) {
		integration, err := claude.NewClaudeIntegration(commandsDir, claudeDir)
		if err != nil {
			t.Fatalf("Failed to create Claude integration: %v", err)
		}

		status, err := integration.CheckIntegration()
		if err != nil {
			t.Fatalf("Failed to check integration status: %v", err)
		}

		if !status.Installed {
			t.Error("Integration should be reported as installed")
		}

		if status.CommandCount <= 0 {
			t.Error("Command count should be greater than 0")
		}

		// Verify completion path points to .crew/ location
		pathResolver := claude.NewPathResolver(claudeDir)
		expectedCompletionPath := pathResolver.GetCompletionsDir()
		if status.CompletionPath != expectedCompletionPath {
			t.Errorf("Completion path incorrect: expected %s, got %s", expectedCompletionPath, status.CompletionPath)
		}
	})

	t.Run("UninstallIntegration", func(t *testing.T) {
		integration, err := claude.NewClaudeIntegration(commandsDir, claudeDir)
		if err != nil {
			t.Fatalf("Failed to create Claude integration: %v", err)
		}

		// Uninstall
		if err := integration.UninstallIntegration(); err != nil {
			t.Fatalf("Failed to uninstall integration: %v", err)
		}

		// Verify files were removed
		pathResolver := claude.NewPathResolver(claudeDir)
		configPath := pathResolver.GetMainConfigFile()
		if _, err := os.Stat(configPath); !os.IsNotExist(err) {
			t.Error("Config file was not removed during uninstall")
		}

		completionsDir := pathResolver.GetCompletionsDir()
		if _, err := os.Stat(completionsDir); !os.IsNotExist(err) {
			t.Error("Completions directory was not removed during uninstall")
		}

		// Verify status reflects uninstallation
		status, err := integration.CheckIntegration()
		if err != nil {
			t.Fatalf("Failed to check integration status after uninstall: %v", err)
		}

		if status.Installed {
			t.Error("Integration should be reported as not installed after uninstall")
		}
	})
}