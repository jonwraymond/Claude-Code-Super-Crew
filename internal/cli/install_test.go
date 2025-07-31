package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"github.com/jonwraymond/claude-code-super-crew/internal/core"
	"github.com/jonwraymond/claude-code-super-crew/internal/managers"
	"github.com/jonwraymond/claude-code-super-crew/pkg/logger"
)

func TestInstallCommand(t *testing.T) {
	// Save original flags and restore after test
	originalFlags := globalFlags
	defer func() { globalFlags = originalFlags }()
	
	// Set user home for tests
	userHome := os.Getenv("HOME")
	if userHome == "" {
		userHome = os.Getenv("USERPROFILE")
	}
	
	tests := []struct {
		name           string
		args           []string
		setupFunc      func(tempDir string) error
		validateFunc   func(tempDir string) error
		expectError    bool
		errorContains  string
	}{
		{
			name: "Quick Installation",
			args: []string{"--quick", "--yes"},
			validateFunc: func(tempDir string) error {
				// Check that core components are installed
				expectedDirs := []string{
					filepath.Join(tempDir, ".claude", "Core"),
					filepath.Join(tempDir, ".claude", "Commands"),
					filepath.Join(tempDir, ".claude", "hooks"),
				}
				for _, dir := range expectedDirs {
					if _, err := os.Stat(dir); os.IsNotExist(err) {
						return fmt.Errorf("expected directory not found: %s", dir)
					}
				}
				
				// Check VERSION file
				versionFile := filepath.Join(tempDir, ".claude", "VERSION")
				if _, err := os.Stat(versionFile); os.IsNotExist(err) {
					return fmt.Errorf("VERSION file not found")
				}
				
				return nil
			},
			expectError: false,
		},
		{
			name: "Minimal Installation",
			args: []string{"--minimal", "--yes"},
			validateFunc: func(tempDir string) error {
				// Check only core is installed
				coreDir := filepath.Join(tempDir, ".claude", "Core")
				if _, err := os.Stat(coreDir); os.IsNotExist(err) {
					return fmt.Errorf("Core directory not found")
				}
				
				// Hooks should not be installed in minimal mode
				hooksDir := filepath.Join(tempDir, ".claude", "hooks")
				if _, err := os.Stat(hooksDir); err == nil {
					return fmt.Errorf("hooks directory should not exist in minimal installation")
				}
				
				return nil
			},
			expectError: false,
		},
		{
			name: "Custom Components Installation",
			args: []string{"--components", "core,hooks", "--yes"},
			validateFunc: func(tempDir string) error {
				// Check specific components
				expectedDirs := []string{
					filepath.Join(tempDir, ".claude", "Core"),
					filepath.Join(tempDir, ".claude", "hooks"),
				}
				for _, dir := range expectedDirs {
					if _, err := os.Stat(dir); os.IsNotExist(err) {
						return fmt.Errorf("expected directory not found: %s", dir)
					}
				}
				
				// Commands should not be installed
				commandsDir := filepath.Join(tempDir, ".claude", "Commands")
				if _, err := os.Stat(commandsDir); err == nil {
					return fmt.Errorf("Commands directory should not exist")
				}
				
				return nil
			},
			expectError: false,
		},
		{
			name: "Installation with Existing Directory",
			setupFunc: func(tempDir string) error {
				// Create existing installation
				installDir := filepath.Join(tempDir, ".claude")
				if err := os.MkdirAll(installDir, 0755); err != nil {
					return err
				}
				// Create a test file to verify backup
				testFile := filepath.Join(installDir, "test.txt")
				return os.WriteFile(testFile, []byte("existing content"), 0644)
			},
			args: []string{"--quick", "--yes", "--force"},
			validateFunc: func(tempDir string) error {
				// Check that backup was created
				backupDir := filepath.Join(tempDir, ".claude", "backups")
				entries, err := os.ReadDir(backupDir)
				if err != nil {
					return fmt.Errorf("failed to read backup directory: %w", err)
				}
				if len(entries) == 0 {
					return fmt.Errorf("no backup created for existing installation")
				}
				return nil
			},
			expectError: false,
		},
		{
			name: "Dry Run Mode",
			args: []string{"--quick", "--yes", "--dry-run"},
			validateFunc: func(tempDir string) error {
				// Check that nothing was actually installed
				installDir := filepath.Join(tempDir, ".claude")
				entries, err := os.ReadDir(installDir)
				if err == nil && len(entries) > 0 {
					return fmt.Errorf("dry run should not create files")
				}
				return nil
			},
			expectError: false,
		},
		{
			name: "List Components",
			args: []string{"--list-components"},
			validateFunc: func(tempDir string) error {
				// Just check command runs without error
				return nil
			},
			expectError: false,
		},
		{
			name: "System Diagnostics",
			args: []string{"--diagnose"},
			validateFunc: func(tempDir string) error {
				// Just check command runs without error
				return nil
			},
			expectError: false,
		},
		{
			name: "Invalid Profile",
			args: []string{"--profile", "nonexistent", "--yes"},
			expectError:   true,
			errorContains: "unknown profile",
		},
		{
			name: "No Backup Flag",
			setupFunc: func(tempDir string) error {
				// Create existing installation
				installDir := filepath.Join(tempDir, ".claude")
				return os.MkdirAll(installDir, 0755)
			},
			args: []string{"--quick", "--yes", "--no-backup", "--force"},
			validateFunc: func(tempDir string) error {
				// Check that no backup was created
				backupDir := filepath.Join(tempDir, ".claude", "backups")
				if _, err := os.Stat(backupDir); err == nil {
					entries, _ := os.ReadDir(backupDir)
					if len(entries) > 0 {
						return fmt.Errorf("backup created despite --no-backup flag")
					}
				}
				return nil
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			tempDir, err := os.MkdirTemp("", "install-test-*")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tempDir)

			// Setup if needed
			if tt.setupFunc != nil {
				if err := tt.setupFunc(tempDir); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			// Set install directory within user home
			installDir := filepath.Join(userHome, ".test-claude", filepath.Base(tempDir))
			globalFlags.InstallDir = installDir
			globalFlags.Quiet = true // Reduce output noise
			
			// Parse global flags from args
			filteredArgs := []string{}
			for _, arg := range tt.args {
				switch arg {
				case "--yes":
					globalFlags.Yes = true
				case "--force":
					globalFlags.Force = true
				case "--dry-run":
					globalFlags.DryRun = true
				default:
					filteredArgs = append(filteredArgs, arg)
				}
			}
			
			// Clean up test directory after test
			defer os.RemoveAll(installDir)

			// Create command
			cmd := NewInstallCommand()
			cmd.SetArgs(filteredArgs)

			// Execute command
			err = cmd.Execute()

			// Check error expectation
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error containing '%s', got: %v", tt.errorContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			// Validate if no error expected and validation function provided
			if !tt.expectError && tt.validateFunc != nil {
				// Pass the install directory for validation
				if err := tt.validateFunc(filepath.Dir(installDir)); err != nil {
					t.Errorf("Validation failed: %v", err)
				}
			}
		})
	}
}

func TestGetComponentsToInstall(t *testing.T) {
	// Create a mock registry
	exe, _ := os.Executable()
	projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(exe)))
	registry := core.NewEnhancedComponentRegistry(filepath.Join(projectRoot, "setup", "components"))
	
	// Create a mock config manager
	configManager, _ := managers.NewConfigManager(filepath.Join(projectRoot, "config"), "")

	tests := []struct {
		name     string
		flags    InstallFlags
		expected []string
	}{
		{
			name: "Explicit Components",
			flags: InstallFlags{
				Components: []string{"core", "hooks"},
			},
			expected: []string{"core", "hooks"},
		},
		{
			name: "All Components",
			flags: InstallFlags{
				Components: []string{"all"},
			},
			expected: []string{"core", "commands", "hooks", "mcp"},
		},
		{
			name: "Quick Profile",
			flags: InstallFlags{
				Profile: "quick",
			},
			expected: []string{"core", "commands"},
		},
		{
			name: "Minimal Profile",
			flags: InstallFlags{
				Profile: "minimal",
			},
			expected: []string{"core"},
		},
		{
			name: "Developer Profile",
			flags: InstallFlags{
				Profile: "developer",
			},
			expected: []string{"core", "commands", "hooks", "mcp"},
		},
		{
			name: "Quick Flag",
			flags: InstallFlags{
				Quick: true,
			},
			expected: []string{"core", "commands", "hooks"},
		},
		{
			name: "Minimal Flag",
			flags: InstallFlags{
				Minimal: true,
			},
			expected: []string{"core"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			components, err := getComponentsToInstall(tt.flags, registry, configManager)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Check components match expected
			if len(components) != len(tt.expected) {
				t.Errorf("Expected %d components, got %d", len(tt.expected), len(components))
			}

			for i, comp := range components {
				if i < len(tt.expected) && comp != tt.expected[i] {
					t.Errorf("Expected component %s, got %s", tt.expected[i], comp)
				}
			}
		})
	}
}

func TestShouldInstallComponent(t *testing.T) {
	tests := []struct {
		name               string
		component          string
		selectedComponents []string
		expected           bool
	}{
		{
			name:               "Core Selected Directly",
			component:          "Core",
			selectedComponents: []string{"core"},
			expected:           true,
		},
		{
			name:               "Commands Selected",
			component:          "Commands",
			selectedComponents: []string{"commands"},
			expected:           true,
		},
		{
			name:               "All Selected",
			component:          "hooks",
			selectedComponents: []string{"all"},
			expected:           true,
		},
		{
			name:               "Not Selected",
			component:          "hooks",
			selectedComponents: []string{"core"},
			expected:           false,
		},
		{
			name:               "Empty Selection (Install All)",
			component:          "Core",
			selectedComponents: []string{},
			expected:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldInstallComponent(tt.component, tt.selectedComponents)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestCopyDirectoryRecursive(t *testing.T) {
	// Create source directory structure
	srcDir, err := os.MkdirTemp("", "copy-src-*")
	if err != nil {
		t.Fatalf("Failed to create source dir: %v", err)
	}
	defer os.RemoveAll(srcDir)

	// Create test files and directories
	testFiles := []string{
		"file1.txt",
		"subdir/file2.txt",
		"subdir/nested/file3.txt",
	}

	for _, file := range testFiles {
		fullPath := filepath.Join(srcDir, file)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
	}

	// Create destination directory
	dstDir, err := os.MkdirTemp("", "copy-dst-*")
	if err != nil {
		t.Fatalf("Failed to create destination dir: %v", err)
	}
	defer os.RemoveAll(dstDir)

	// Copy directory
	if err := copyDirectoryRecursive(srcDir, dstDir); err != nil {
		t.Fatalf("Copy failed: %v", err)
	}

	// Verify all files were copied
	for _, file := range testFiles {
		dstPath := filepath.Join(dstDir, file)
		if _, err := os.Stat(dstPath); os.IsNotExist(err) {
			t.Errorf("File not copied: %s", file)
		}
	}
}

func TestCreateSimpleBackup(t *testing.T) {
	// Create test installation directory
	installDir, err := os.MkdirTemp("", "backup-test-*")
	if err != nil {
		t.Fatalf("Failed to create install dir: %v", err)
	}
	defer os.RemoveAll(installDir)

	// Create some test files
	testFile := filepath.Join(installDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create backup
	if err := createSimpleBackup(installDir); err != nil {
		t.Fatalf("Backup creation failed: %v", err)
	}

	// Verify backup was created
	backupDir := filepath.Join(installDir, "backups")
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		t.Fatalf("Failed to read backup directory: %v", err)
	}

	if len(entries) == 0 {
		t.Error("No backup created")
	}

	// Verify backup contains the test file
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "crew-backup-") {
			backupTestFile := filepath.Join(backupDir, entry.Name(), "test.txt")
			if _, err := os.Stat(backupTestFile); os.IsNotExist(err) {
				t.Error("Test file not found in backup")
			}
		}
	}
}

func TestInstallCommandIntegration(t *testing.T) {
	// Skip if running in CI without proper setup
	if os.Getenv("CI") == "true" && os.Getenv("SKIP_INTEGRATION_TESTS") == "true" {
		t.Skip("Skipping integration test in CI")
	}

	// Create a test-specific logger
	testLogger := logger.NewLogger()
	testLogger.SetQuiet(true)

	// Test end-to-end installation workflow
	tempDir, err := os.MkdirTemp("", "integration-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test 1: Fresh installation
	t.Run("Fresh Installation Workflow", func(t *testing.T) {
		installDir := filepath.Join(tempDir, "fresh-install", ".claude")
		
		// Save and set global flags
		originalFlags := globalFlags
		globalFlags.InstallDir = installDir
		globalFlags.Quiet = true
		globalFlags.Yes = true
		defer func() { globalFlags = originalFlags }()

		// Run installation
		cmd := NewInstallCommand()
		cmd.SetArgs([]string{"--quick"})
		
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Installation failed: %v", err)
		}

		// Verify installation
		requiredFiles := []string{
			"VERSION",
			"Core/CLAUDE.md",
			"Core/FLAGS.md",
			"Commands/analyze.md",
		}

		for _, file := range requiredFiles {
			path := filepath.Join(installDir, file)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Errorf("Required file missing: %s", file)
			}
		}
	})

	// Test 2: Update existing installation
	t.Run("Update Existing Installation", func(t *testing.T) {
		installDir := filepath.Join(tempDir, "update-install", ".claude")
		
		// Create existing installation
		if err := os.MkdirAll(installDir, 0755); err != nil {
			t.Fatalf("Failed to create install dir: %v", err)
		}
		
		// Add custom file
		customFile := filepath.Join(installDir, "custom.txt")
		if err := os.WriteFile(customFile, []byte("custom content"), 0644); err != nil {
			t.Fatalf("Failed to create custom file: %v", err)
		}

		// Save and set global flags
		originalFlags := globalFlags
		globalFlags.InstallDir = installDir
		globalFlags.Quiet = true
		globalFlags.Yes = true
		globalFlags.Force = true
		defer func() { globalFlags = originalFlags }()

		// Run update
		cmd := NewInstallCommand()
		cmd.SetArgs([]string{"--quick"})
		
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Update failed: %v", err)
		}

		// Verify backup was created
		backupDir := filepath.Join(installDir, "backups")
		entries, err := os.ReadDir(backupDir)
		if err != nil || len(entries) == 0 {
			t.Error("No backup created during update")
		}
	})
}