package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jonwraymond/claude-code-super-crew/pkg/backup"
)

// createTestInstallation creates a minimal crew-metadata.json for testing
func createTestInstallation(installDir string) error {
	// Create .crew/config directory
	configDir := filepath.Join(installDir, ".crew", "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	metadataJSON := `{
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

	return os.WriteFile(filepath.Join(configDir, "crew-metadata.json"), []byte(metadataJSON), 0644)
}

func TestBackupCommand(t *testing.T) {
	// Save original flags and restore after test
	originalFlags := globalFlags
	defer func() { globalFlags = originalFlags }()

	tests := []struct {
		name           string
		args           []string
		setupFunc      func(tempDir string) error
		validateFunc   func(tempDir string) error
		expectError    bool
		errorContains  string
	}{
		{
			name: "Create Backup",
			setupFunc: func(tempDir string) error {
				// Create a mock installation
				installDir := filepath.Join(tempDir, ".claude")
				if err := os.MkdirAll(filepath.Join(installDir, "Core"), 0755); err != nil {
					return err
				}
				// Create installation.json
				if err := createTestInstallation(installDir); err != nil {
					return err
				}
				// Create test files
				files := []string{
					"Core/CLAUDE.md",
					"Core/FLAGS.md",
					"VERSION",
				}
				for _, file := range files {
					path := filepath.Join(installDir, file)
					if err := os.WriteFile(path, []byte("test content"), 0644); err != nil {
						return err
					}
				}
				return nil
			},
			args: []string{"--create"},
			validateFunc: func(tempDir string) error {
				// Check backup was created
				backupDir := filepath.Join(tempDir, ".claude", "backups")
				entries, err := os.ReadDir(backupDir)
				if err != nil {
					return fmt.Errorf("failed to read backup dir: %w", err)
				}
				if len(entries) == 0 {
					return fmt.Errorf("no backup created")
				}
				
				// Check backup file exists and has content
				for _, entry := range entries {
					if strings.HasPrefix(entry.Name(), "crew_backup_") {
						info, _ := entry.Info()
						if info.Size() == 0 {
							return fmt.Errorf("backup file is empty")
						}
						return nil
					}
				}
				return fmt.Errorf("no valid backup file found")
			},
			expectError: false,
		},
		{
			name: "Create Backup with Custom Name",
			setupFunc: func(tempDir string) error {
				// Create a mock installation
				installDir := filepath.Join(tempDir, ".claude")
				if err := os.MkdirAll(installDir, 0755); err != nil {
					return err
				}
				// Create installation.json
				if err := createTestInstallation(installDir); err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(installDir, "test.txt"), []byte("test"), 0644)
			},
			args: []string{"--create", "--name", "custom_backup"},
			validateFunc: func(tempDir string) error {
				backupDir := filepath.Join(tempDir, ".claude", "backups")
				entries, err := os.ReadDir(backupDir)
				if err != nil {
					return fmt.Errorf("failed to read backup dir: %w", err)
				}
				
				// Check for custom name
				found := false
				for _, entry := range entries {
					if strings.Contains(entry.Name(), "custom_backup") {
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("custom backup name not found")
				}
				return nil
			},
			expectError: false,
		},
		{
			name: "List Backups",
			setupFunc: func(tempDir string) error {
				// Create installation directory
				installDir := filepath.Join(tempDir, ".claude")
				if err := os.MkdirAll(installDir, 0755); err != nil {
					return err
				}
				// Create installation.json
				if err := createTestInstallation(installDir); err != nil {
					return err
				}
				// Create backup directory with test backups
				backupDir := filepath.Join(tempDir, ".claude", "backups")
				if err := os.MkdirAll(backupDir, 0755); err != nil {
					return err
				}
				
				// Create dummy backup files
				for i := 0; i < 3; i++ {
					filename := fmt.Sprintf("crew_backup_%d.tar.gz", i)
					path := filepath.Join(backupDir, filename)
					if err := os.WriteFile(path, []byte("backup data"), 0644); err != nil {
						return err
					}
				}
				return nil
			},
			args: []string{"--list"},
			validateFunc: func(tempDir string) error {
				// Just verify command runs without error
				return nil
			},
			expectError: false,
		},
		{
			name: "Restore Backup",
			setupFunc: func(tempDir string) error {
				// Create installation directory
				installDir := filepath.Join(tempDir, ".claude")
				if err := os.MkdirAll(installDir, 0755); err != nil {
					return err
				}
				// Create installation.json
				if err := createTestInstallation(installDir); err != nil {
					return err
				}
				
				// Create a simple backup using tar format
				backupDir := filepath.Join(installDir, "backups")
				if err := os.MkdirAll(backupDir, 0755); err != nil {
					return err
				}
				
				// For testing, create a mock backup file
				backupFile := filepath.Join(backupDir, "test_backup.tar.gz")
				// Note: In real test, this would be a proper tar.gz file
				return os.WriteFile(backupFile, []byte("mock backup data"), 0644)
			},
			args: []string{"--restore", "test_backup.tar.gz"},
			validateFunc: func(tempDir string) error {
				// Restore will fail with invalid archive
				return nil
			},
			expectError: true,
			errorContains: "invalid",
		},
		{
			name: "Show Backup Info",
			setupFunc: func(tempDir string) error {
				// Create backup directory with test backup
				backupDir := filepath.Join(tempDir, ".claude", "backups")
				if err := os.MkdirAll(backupDir, 0755); err != nil {
					return err
				}
				
				backupFile := filepath.Join(backupDir, "info_test.tar.gz")
				return os.WriteFile(backupFile, []byte("backup data"), 0644)
			},
			args: []string{"--info", "info_test.tar.gz"},
			validateFunc: func(tempDir string) error {
				// Just verify command runs
				return nil
			},
			expectError: false,
		},
		{
			name: "Cleanup Backups",
			setupFunc: func(tempDir string) error {
				// Create backup directory with multiple backups
				backupDir := filepath.Join(tempDir, ".claude", "backups")
				if err := os.MkdirAll(backupDir, 0755); err != nil {
					return err
				}
				
				// Create old backups
				for i := 0; i < 10; i++ {
					filename := fmt.Sprintf("crew_backup_%d.tar.gz", i)
					path := filepath.Join(backupDir, filename)
					if err := os.WriteFile(path, []byte("backup data"), 0644); err != nil {
						return err
					}
					// Set old modification time for some files
					if i < 5 {
						oldTime := time.Now().Add(-30 * 24 * time.Hour)
						os.Chtimes(path, oldTime, oldTime)
					}
				}
				return nil
			},
			args: []string{"--cleanup", "--keep", "3"},
			validateFunc: func(tempDir string) error {
				backupDir := filepath.Join(tempDir, ".claude", "backups")
				entries, err := os.ReadDir(backupDir)
				if err != nil {
					return fmt.Errorf("failed to read backup dir: %w", err)
				}
				
				// Should have at most 3 backups remaining
				if len(entries) > 3 {
					return fmt.Errorf("cleanup failed: %d backups remain, expected <= 3", len(entries))
				}
				return nil
			},
			expectError: false,
		},
		{
			name: "No Installation Error",
			args: []string{"--create"},
			expectError:   true,
			errorContains: "no installation found",
		},
		{
			name: "Create Backup Dry Run",
			setupFunc: func(tempDir string) error {
				// Create a mock installation
				installDir := filepath.Join(tempDir, ".claude")
				if err := os.MkdirAll(installDir, 0755); err != nil {
					return err
				}
				// Create installation.json
				if err := createTestInstallation(installDir); err != nil {
					return err
				}
				// Set dry-run mode
				globalFlags.DryRun = true
				return os.WriteFile(filepath.Join(installDir, "test.txt"), []byte("test"), 0644)
			},
			args: []string{"--create"},
			validateFunc: func(tempDir string) error {
				// Reset dry-run mode
				globalFlags.DryRun = false
				// Check no actual backup was created
				backupDir := filepath.Join(tempDir, ".claude", "backups")
				if _, err := os.Stat(backupDir); err == nil {
					entries, _ := os.ReadDir(backupDir)
					if len(entries) > 0 {
						return fmt.Errorf("backup created in dry-run mode")
					}
				}
				return nil
			},
			expectError: false,
		},
		{
			name: "Custom Backup Directory",
			setupFunc: func(tempDir string) error {
				// Create a mock installation
				installDir := filepath.Join(tempDir, ".claude")
				if err := os.MkdirAll(installDir, 0755); err != nil {
					return err
				}
				// Create installation.json
				if err := createTestInstallation(installDir); err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(installDir, "test.txt"), []byte("test"), 0644)
			},
			args: []string{"--create", "--backup-dir", filepath.Join("$TEMP_DIR", "custom-backups")},
			validateFunc: func(tempDir string) error {
				// Check backup in custom directory
				customBackupDir := filepath.Join(tempDir, "custom-backups")
				entries, err := os.ReadDir(customBackupDir)
				if err != nil {
					return fmt.Errorf("custom backup dir not found: %w", err)
				}
				if len(entries) == 0 {
					return fmt.Errorf("no backup in custom directory")
				}
				return nil
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			tempDir, err := os.MkdirTemp("", "backup-test-*")
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

			// Set global flags
			globalFlags.InstallDir = filepath.Join(tempDir, ".claude")
			globalFlags.Quiet = true

			// Replace $TEMP_DIR in args
			for i, arg := range tt.args {
				tt.args[i] = strings.ReplaceAll(arg, "$TEMP_DIR", tempDir)
			}

			// Create command
			cmd := NewBackupCommand()
			cmd.SetArgs(tt.args)

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
				if err := tt.validateFunc(tempDir); err != nil {
					t.Errorf("Validation failed: %v", err)
				}
			}
		})
	}
}

func TestGetBackupDirectory(t *testing.T) {
	// Save original flags
	originalFlags := backupFlags
	defer func() { backupFlags = originalFlags }()

	tests := []struct {
		name         string
		backupDir    string
		installDir   string
		expected     string
	}{
		{
			name:       "Default Backup Directory",
			backupDir:  "",
			installDir: "/home/user/.claude",
			expected:   "/home/user/.claude/backups",
		},
		{
			name:       "Custom Backup Directory",
			backupDir:  "/custom/backup/path",
			installDir: "/home/user/.claude",
			expected:   "/custom/backup/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backupFlags.BackupDir = tt.backupDir
			globalFlags.InstallDir = tt.installDir

			result := getBackupDirectory()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestCheckInstallationExists(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "install-check-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name     string
		setup    func() error
		expected bool
	}{
		{
			name: "Installation Exists",
			setup: func() error {
				// Create installation files
				dirs := []string{
					filepath.Join(tempDir, "Core"),
					filepath.Join(tempDir, "Commands"),
				}
				for _, dir := range dirs {
					if err := os.MkdirAll(dir, 0755); err != nil {
						return err
					}
				}
				// Create installation.json
				if err := createTestInstallation(tempDir); err != nil {
					return err
				}
				// Create VERSION file
				return os.WriteFile(filepath.Join(tempDir, "VERSION"), []byte("1.0.0"), 0644)
			},
			expected: true,
		},
		{
			name:     "No Installation",
			setup:    func() error { return nil },
			expected: false,
		},
		{
			name: "Partial Installation",
			setup: func() error {
				// Create only Core directory
				if err := os.MkdirAll(filepath.Join(tempDir, "Core"), 0755); err != nil {
					return err
				}
				// Create installation.json (required for detection)
				return createTestInstallation(tempDir)
			},
			expected: true, // Even partial installation returns true
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear directory
			os.RemoveAll(tempDir)
			os.MkdirAll(tempDir, 0755)

			// Setup
			if err := tt.setup(); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			// Save and restore global flags
			originalDir := globalFlags.InstallDir
			globalFlags.InstallDir = tempDir
			defer func() { globalFlags.InstallDir = originalDir }()

			result := checkInstallationExists()
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestBackupManagerIntegration(t *testing.T) {
	// Skip if running in CI without proper setup
	if os.Getenv("CI") == "true" && os.Getenv("SKIP_INTEGRATION_TESTS") == "true" {
		t.Skip("Skipping integration test in CI")
	}

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "backup-integration-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test complete backup workflow
	t.Run("Complete Backup Workflow", func(t *testing.T) {
		// Setup installation
		installDir := filepath.Join(tempDir, "test-install")
		if err := os.MkdirAll(filepath.Join(installDir, "Core"), 0755); err != nil {
			t.Fatalf("Failed to create install dir: %v", err)
		}

		// Create test files
		testFiles := map[string]string{
			"Core/CLAUDE.md": "# CLAUDE.md content",
			"Core/FLAGS.md":  "# FLAGS.md content",
			"VERSION":        "1.0.0",
			"custom.txt":     "Custom user content",
		}

		for file, content := range testFiles {
			path := filepath.Join(installDir, file)
			dir := filepath.Dir(path)
			if err := os.MkdirAll(dir, 0755); err != nil {
				t.Fatalf("Failed to create directory: %v", err)
			}
			if err := os.WriteFile(path, []byte(content), 0644); err != nil {
				t.Fatalf("Failed to write file %s: %v", file, err)
			}
		}

		// Create backup manager
		backupDir := filepath.Join(installDir, "backups")
		mgr := backup.NewManager(backup.Options{
			InstallDir: installDir,
			BackupDir:  backupDir,
			BackupName: "integration_test",
			Compress:   "gzip",
			Verbose:    false,
			DryRun:     false,
		})

		// Create backup
		backupFile, err := mgr.Create()
		if err != nil {
			t.Fatalf("Failed to create backup: %v", err)
		}

		// Verify backup file exists
		if _, err := os.Stat(backupFile); os.IsNotExist(err) {
			t.Error("Backup file does not exist")
		}

		// Get backup info
		info := mgr.GetBackupInfo(backupFile)
		if !info.Exists {
			t.Error("Backup info reports file doesn't exist")
		}
		if info.Size == 0 {
			t.Error("Backup file is empty")
		}

		// List backups
		backups, err := mgr.ListBackups()
		if err != nil {
			t.Fatalf("Failed to list backups: %v", err)
		}
		if len(backups) == 0 {
			t.Error("No backups found")
		}

		// Test restore (to different location)
		restoreDir := filepath.Join(tempDir, "restore-test")
		_ = backup.NewManager(backup.Options{
			InstallDir: restoreDir,
			BackupDir:  backupDir,
			Verbose:    false,
			DryRun:     false,
			Overwrite:  true,
		})

		// Note: Actual restore would require proper tar.gz implementation
		// For now, we're testing the workflow structure

		// Test cleanup
		// Create multiple backups
		for i := 0; i < 5; i++ {
			newMgr := backup.NewManager(backup.Options{
				InstallDir: installDir,
				BackupDir:  backupDir,
				BackupName: fmt.Sprintf("backup_%d", i),
				Compress:   "gzip",
				Verbose:    false,
				DryRun:     false,
			})
			_, err := newMgr.Create()
			if err != nil {
				t.Fatalf("Failed to create backup %d: %v", i, err)
			}
		}

		// Cleanup old backups
		removed, err := mgr.Cleanup(2, 0)
		if err != nil {
			t.Fatalf("Cleanup failed: %v", err)
		}
		if removed < 3 {
			t.Errorf("Expected at least 3 backups removed, got %d", removed)
		}

		// Verify only 2 backups remain
		remainingBackups, _ := mgr.ListBackups()
		if len(remainingBackups) > 2 {
			t.Errorf("Expected 2 backups remaining, got %d", len(remainingBackups))
		}
	})
}

func TestBackupCommandFlagValidation(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectError   bool
		errorContains string
	}{
		{
			name:        "No Operation Specified",
			args:        []string{},
			expectError: true,
			// The command requires one of the operation flags
		},
		{
			name: "Valid Create Operation",
			args: []string{"--create"},
			// Will fail due to no installation, but flag validation passes
			expectError: true,
		},
		{
			name: "Valid List Operation",
			args: []string{"--list"},
			// Should work even without installation
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			tempDir, err := os.MkdirTemp("", "flag-test-*")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tempDir)

			// Set global flags
			globalFlags.InstallDir = filepath.Join(tempDir, ".claude")
			globalFlags.Quiet = true

			// Create command
			cmd := NewBackupCommand()
			cmd.SetArgs(tt.args)

			// Execute command
			err = cmd.Execute()

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}