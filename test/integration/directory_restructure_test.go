package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jonwraymond/claude-code-super-crew/internal/claude"
)

// TestData centralizes all test configuration and data
type TestData struct {
	// Directories that should be moved to .crew/
	UtilityDirs []string
	// Directories that should remain in .claude/
	CoreDirs []string
	// Claude Code directories that should remain untouched
	ClaudeCodeDirs []string
	// Framework files that should remain in .claude/
	FrameworkFiles []string
	// Test content templates
	TestContent    string
	CoreContent    string
	ClaudeContent  string
}

// GetTestData returns centralized test configuration
func GetTestData() *TestData {
	return &TestData{
		UtilityDirs: []string{
			"backups", "logs", "config", "completions", 
			"scripts", "workflows", "prompts",
		},
		CoreDirs: []string{
			"commands", "hooks", "agents",
		},
		ClaudeCodeDirs: []string{
			"statsig", "shell-snapshots",
		},
		FrameworkFiles: []string{
			"CLAUDE.md", "COMMANDS.md", "FLAGS.md", "PRINCIPLES.md", 
			"RULES.md", "MCP.md", "PERSONAS.md", "ORCHESTRATOR.md", "MODES.md",
		},
		TestContent:   "test content for validation",
		CoreContent:   "core supercrew content", 
		ClaudeContent: "claude code native content",
	}
}

// TestHelper provides utility functions for directory restructure testing
type TestHelper struct {
	t       *testing.T
	tmpDir  string
	claudeDir string
	data    *TestData
}

// NewTestHelper creates a new test helper with temporary directory setup
func NewTestHelper(t *testing.T) *TestHelper {
	tmpDir, err := os.MkdirTemp("", "supercrew_restructure_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	return &TestHelper{
		t:         t,
		tmpDir:    tmpDir,
		claudeDir: filepath.Join(tmpDir, ".claude"),
		data:      GetTestData(),
	}
}

// Cleanup removes the temporary test directory
func (th *TestHelper) Cleanup() {
	os.RemoveAll(th.tmpDir)
}

// SetupTestDirectories creates the complete test directory structure
func (th *TestHelper) SetupTestDirectories() {
	// Ensure base claude directory exists
	if err := os.MkdirAll(th.claudeDir, 0755); err != nil {
		th.t.Fatalf("Failed to create claude dir: %v", err)
	}

	// Create utility directories with test files
	th.createDirectoriesWithContent(th.data.UtilityDirs, "test_file.txt", th.data.TestContent)
	
	// Create core SuperCrew directories with test files
	th.createDirectoriesWithContent(th.data.CoreDirs, "core_file.txt", th.data.CoreContent)
	
	// Create Claude Code directories with test files
	th.createDirectoriesWithContent(th.data.ClaudeCodeDirs, "claude_code_file.txt", th.data.ClaudeContent)

	// Create framework files
	th.createFrameworkFiles()

	// Create crew-metadata.json (current framework version file)
	th.createMetadataFile()
}

// createDirectoriesWithContent creates directories and populates them with test files
func (th *TestHelper) createDirectoriesWithContent(dirs []string, fileName, content string) {
	for _, dir := range dirs {
		dirPath := filepath.Join(th.claudeDir, dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			th.t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		
		testFile := filepath.Join(dirPath, fileName)
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			th.t.Fatalf("Failed to create test file in %s: %v", dir, err)
		}
	}
}

// createFrameworkFiles creates the framework configuration files
func (th *TestHelper) createFrameworkFiles() {
	for _, fileName := range th.data.FrameworkFiles {
		filePath := filepath.Join(th.claudeDir, fileName)
		content := th.getFrameworkFileContent(fileName)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			th.t.Fatalf("Failed to create framework file %s: %v", fileName, err)
		}
	}
}

// createMetadataFile creates the crew-metadata.json file
func (th *TestHelper) createMetadataFile() {
	metadataPath := filepath.Join(th.claudeDir, ".crew", "config", "crew-metadata.json")
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(metadataPath), 0755); err != nil {
		th.t.Fatalf("Failed to create metadata directory: %v", err)
	}
	
	metadata := `{
		"framework": {
			"version": "1.0.0",
			"build": "test-build"
		},
		"components": {
			"core": "1.0.0",
			"commands": "1.0.0",
			"hooks": "1.0.0"
		}
	}`
	
	if err := os.WriteFile(metadataPath, []byte(metadata), 0644); err != nil {
		th.t.Fatalf("Failed to create metadata file: %v", err)
	}
}

// getFrameworkFileContent returns appropriate content for framework files
func (th *TestHelper) getFrameworkFileContent(fileName string) string {
	switch fileName {
	case "CLAUDE.md":
		return "# Claude Code Super Crew Framework\nTest framework configuration"
	case "COMMANDS.md":
		return "# Commands Reference\nTest commands documentation"
	default:
		return "# " + fileName + "\nTest configuration file"
	}
}

// ValidateDirectoryMove verifies that directories were moved correctly
func (th *TestHelper) ValidateDirectoryMove(dirs []string, fileName, expectedContent, testName string) {
	crewDir := filepath.Join(th.claudeDir, ".crew")
	
	for _, dir := range dirs {
		oldPath := filepath.Join(th.claudeDir, dir)
		newPath := filepath.Join(crewDir, dir)
		
		// Old path should not exist
		if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
			th.t.Errorf("%s: Old directory %s still exists after restructuring", testName, oldPath)
		}
		
		// New path should exist
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			th.t.Errorf("%s: New directory %s was not created", testName, newPath)
			continue
		}
		
		// Verify file content
		testFile := filepath.Join(newPath, fileName)
		if content, err := os.ReadFile(testFile); err != nil {
			th.t.Errorf("%s: Test file not found in moved directory %s: %v", testName, dir, err)
		} else if string(content) != expectedContent {
			th.t.Errorf("%s: File content corrupted in %s. Expected: %s, Got: %s", 
				testName, dir, expectedContent, string(content))
		}
	}
}

// ValidateDirectoryRemained verifies that directories remained in their original location
func (th *TestHelper) ValidateDirectoryRemained(dirs []string, fileName, expectedContent, testName string) {
	for _, dir := range dirs {
		dirPath := filepath.Join(th.claudeDir, dir)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			th.t.Errorf("%s: Directory %s was incorrectly moved", testName, dir)
			continue
		}
		
		// Verify file content
		testFile := filepath.Join(dirPath, fileName)
		if content, err := os.ReadFile(testFile); err != nil {
			th.t.Errorf("%s: File missing from %s: %v", testName, dir, err)
		} else if string(content) != expectedContent {
			th.t.Errorf("%s: File content corrupted in %s. Expected: %s, Got: %s", 
				testName, dir, expectedContent, string(content))
		}
	}
}

// TestDirectoryRestructuring tests the complete directory restructuring process
func TestDirectoryRestructuring(t *testing.T) {
	th := NewTestHelper(t)
	defer th.Cleanup()
	
	// Setup test environment
	th.SetupTestDirectories()

	t.Run("RestructureDirectories", func(t *testing.T) {
		// Perform restructuring
		restructurer := claude.NewDirectoryRestructurer(th.claudeDir)
		if err := restructurer.RestructureDirectories(); err != nil {
			t.Fatalf("Directory restructuring failed: %v", err)
		}

		// Verify .crew directory was created
		crewDir := filepath.Join(th.claudeDir, ".crew")
		if _, err := os.Stat(crewDir); os.IsNotExist(err) {
			t.Error(".crew directory was not created")
		}

		// Validate utility directories were moved to .crew/
		th.ValidateDirectoryMove(th.data.UtilityDirs, "test_file.txt", th.data.TestContent, "UtilityDirMove")

		// Validate core SuperCrew directories remained in place
		th.ValidateDirectoryRemained(th.data.CoreDirs, "core_file.txt", th.data.CoreContent, "CoreDirRemained")

		// Validate Claude Code directories remained untouched
		th.ValidateDirectoryRemained(th.data.ClaudeCodeDirs, "claude_code_file.txt", th.data.ClaudeContent, "ClaudeCodeDirRemained")

		// Verify framework files remained in place
		for _, fileName := range th.data.FrameworkFiles {
			filePath := filepath.Join(th.claudeDir, fileName)
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				t.Errorf("Framework file %s was incorrectly moved", fileName)
			}
		}
	})

	t.Run("ValidateRestructure", func(t *testing.T) {
		restructurer := claude.NewDirectoryRestructurer(th.claudeDir)
		issues := restructurer.ValidateRestructure()
		
		if len(issues) > 0 {
			t.Errorf("Validation found issues: %v", issues)
		}
	})

	t.Run("PathResolver", func(t *testing.T) {
		pathResolver := claude.NewPathResolver(th.claudeDir)
		
		// Test utility directory paths (should point to .crew/)
		expectedUtilityPaths := make(map[string]string)
		for _, dir := range th.data.UtilityDirs {
			expectedUtilityPaths[dir] = filepath.Join(th.claudeDir, ".crew", dir)
		}
		
		// Test specific path methods
		testCases := []struct {
			method   func() string
			expected string
			name     string
		}{
			{pathResolver.GetBackupsDir, expectedUtilityPaths["backups"], "BackupsDir"},
			{pathResolver.GetLogsDir, expectedUtilityPaths["logs"], "LogsDir"},
			{pathResolver.GetCommandsDir, filepath.Join(th.claudeDir, "commands"), "CommandsDir"},
			{pathResolver.GetAgentsDir, filepath.Join(th.claudeDir, "agents"), "AgentsDir"},
			{pathResolver.GetHooksDir, filepath.Join(th.claudeDir, "hooks"), "HooksDir"},
		}
		
		for _, tc := range testCases {
			actual := tc.method()
			if actual != tc.expected {
				t.Errorf("%s path incorrect: expected %s, got %s", tc.name, tc.expected, actual)
			}
		}
	})
}

// TestDirectoryRestructuringEdgeCases tests edge cases and error conditions
func TestDirectoryRestructuringEdgeCases(t *testing.T) {
	t.Run("EmptyDirectories", func(t *testing.T) {
		th := NewTestHelper(t)
		defer th.Cleanup()

		// Create only the base claude directory without subdirectories
		if err := os.MkdirAll(th.claudeDir, 0755); err != nil {
			t.Fatalf("Failed to create claude dir: %v", err)
		}

		restructurer := claude.NewDirectoryRestructurer(th.claudeDir)
		if err := restructurer.RestructureDirectories(); err != nil {
			t.Fatalf("Restructuring empty directories failed: %v", err)
		}

		// Should create .crew directory even if no directories to move
		crewDir := filepath.Join(th.claudeDir, ".crew")
		if _, err := os.Stat(crewDir); os.IsNotExist(err) {
			t.Error(".crew directory was not created for empty restructure")
		}
	})

	t.Run("NonExistentSourceDirectory", func(t *testing.T) {
		th := NewTestHelper(t)
		defer th.Cleanup()

		// Test with non-existent claude directory
		restructurer := claude.NewDirectoryRestructurer(th.claudeDir)
		if err := restructurer.RestructureDirectories(); err != nil {
			t.Fatalf("Restructuring should handle non-existent directories gracefully: %v", err)
		}
		
		// Directory should be created
		if _, err := os.Stat(th.claudeDir); os.IsNotExist(err) {
			t.Error("Claude directory should be created if it doesn't exist")
		}
	})

	t.Run("PermissionErrors", func(t *testing.T) {
		if os.Getuid() == 0 {
			t.Skip("Skipping permission test when running as root")
		}

		th := NewTestHelper(t)
		defer th.Cleanup()

		// Create claude directory and a subdirectory
		if err := os.MkdirAll(th.claudeDir, 0755); err != nil {
			t.Fatalf("Failed to create claude dir: %v", err)
		}

		testDir := filepath.Join(th.claudeDir, "logs")
		if err := os.MkdirAll(testDir, 0755); err != nil {
			t.Fatalf("Failed to create test dir: %v", err)
		}

		// Make parent directory read-only to prevent operations
		if err := os.Chmod(th.claudeDir, 0444); err != nil {
			t.Fatalf("Failed to change permissions: %v", err)
		}

		// Restore permissions for cleanup
		defer os.Chmod(th.claudeDir, 0755)

		restructurer := claude.NewDirectoryRestructurer(th.claudeDir)
		err := restructurer.RestructureDirectories()
		
		// Should handle permission errors gracefully
		if err != nil {
			if !strings.Contains(err.Error(), "permission") && 
			   !strings.Contains(err.Error(), "operation not permitted") {
				t.Errorf("Unexpected error type: %v", err)
			}
		}
	})

	t.Run("AlreadyRestructured", func(t *testing.T) {
		th := NewTestHelper(t)
		defer th.Cleanup()
		
		th.SetupTestDirectories()
		
		restructurer := claude.NewDirectoryRestructurer(th.claudeDir)
		
		// First restructuring
		if err := restructurer.RestructureDirectories(); err != nil {
			t.Fatalf("First restructuring failed: %v", err)
		}
		
		// Second restructuring should be idempotent
		if err := restructurer.RestructureDirectories(); err != nil {
			t.Fatalf("Second restructuring should be idempotent: %v", err)
		}
		
		// Validation should still pass
		issues := restructurer.ValidateRestructure()
		if len(issues) > 0 {
			t.Errorf("Validation after double restructuring found issues: %v", issues)
		}
	})
}

// TestPathResolverFunctionality tests the PathResolver utility comprehensively
func TestPathResolverFunctionality(t *testing.T) {
	th := NewTestHelper(t)
	defer th.Cleanup()

	pathResolver := claude.NewPathResolver(th.claudeDir)

	t.Run("EnsureCrewDirectories", func(t *testing.T) {
		if err := pathResolver.EnsureCrewDirectories(); err != nil {
			t.Fatalf("Failed to ensure crew directories: %v", err)
		}

		// Verify all crew directories were created
		for _, dir := range th.data.UtilityDirs {
			expectedPath := filepath.Join(th.claudeDir, ".crew", dir)
			if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
				t.Errorf("Utility directory %s was not created at %s", dir, expectedPath)
			}
		}
	})

	t.Run("EnsureCoreDirectories", func(t *testing.T) {
		if err := pathResolver.EnsureCoreDirectories(); err != nil {
			t.Fatalf("Failed to ensure core directories: %v", err)
		}

		// Verify all core directories were created
		for _, dir := range th.data.CoreDirs {
			expectedPath := filepath.Join(th.claudeDir, dir)
			if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
				t.Errorf("Core directory %s was not created at %s", dir, expectedPath)
			}
		}
	})

	t.Run("GetAllDirectoryPaths", func(t *testing.T) {
		allPaths := pathResolver.GetAllDirectoryPaths()

		expectedCount := len(th.data.UtilityDirs) + len(th.data.CoreDirs)
		if len(allPaths) != expectedCount {
			t.Errorf("Expected %d directory paths, got %d", expectedCount, len(allPaths))
		}

		// Verify utility directories point to .crew/
		for _, dir := range th.data.UtilityDirs {
			path, exists := allPaths[dir]
			if !exists {
				t.Errorf("Missing path for utility directory %s", dir)
				continue
			}
			
			expectedPath := filepath.Join(th.claudeDir, ".crew", dir)
			if path != expectedPath {
				t.Errorf("Wrong path for %s: expected %s, got %s", dir, expectedPath, path)
			}
		}

		// Verify core directories remain in main .claude/
		for _, dir := range th.data.CoreDirs {
			path, exists := allPaths[dir]
			if !exists {
				t.Errorf("Missing path for core directory %s", dir)
				continue
			}
			
			expectedPath := filepath.Join(th.claudeDir, dir)
			if path != expectedPath {
				t.Errorf("Wrong path for %s: expected %s, got %s", dir, expectedPath, path)
			}
		}
	})

	t.Run("FrameworkFiles", func(t *testing.T) {
		frameworkFiles := pathResolver.GetFrameworkFiles()
		
		if len(frameworkFiles) != len(th.data.FrameworkFiles) {
			t.Errorf("Expected %d framework files, got %d", len(th.data.FrameworkFiles), len(frameworkFiles))
		}
		
		for _, fileName := range th.data.FrameworkFiles {
			path, exists := frameworkFiles[fileName]
			if !exists {
				t.Errorf("Missing framework file %s", fileName)
				continue
			}
			
			expectedPath := filepath.Join(th.claudeDir, fileName)
			if path != expectedPath {
				t.Errorf("Wrong path for %s: expected %s, got %s", fileName, expectedPath, path)
			}
		}
	})

	t.Run("InstallationMetadata", func(t *testing.T) {
		metadataPath := pathResolver.GetInstallationMetadata()
		expectedPath := filepath.Join(th.claudeDir, ".crew", "config", "crew-metadata.json")
		
		if metadataPath != expectedPath {
			t.Errorf("Wrong metadata path: expected %s, got %s", expectedPath, metadataPath)
		}
	})
}

// BenchmarkDirectoryRestructuring provides performance benchmarks
func BenchmarkDirectoryRestructuring(b *testing.B) {
	for i := 0; i < b.N; i++ {
		th := &TestHelper{
			t:         &testing.T{}, // Minimal testing.T for benchmark
			tmpDir:    b.TempDir(),
		}
		th.claudeDir = filepath.Join(th.tmpDir, ".claude")
		th.data = GetTestData()
		
		th.SetupTestDirectories()
		
		restructurer := claude.NewDirectoryRestructurer(th.claudeDir)
		if err := restructurer.RestructureDirectories(); err != nil {
			b.Fatalf("Restructuring failed: %v", err)
		}
	}
}