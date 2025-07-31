package cli

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestBackupCommandIntegrationWithMocks(t *testing.T) {
	// Skip if explicitly disabled
	if os.Getenv("SKIP_INTEGRATION_TESTS") == "true" {
		t.Skip("Skipping integration test")
	}

	// Save original values
	originalFlags := globalFlags
	originalBackupFlags := backupFlags
	defer func() {
		globalFlags = originalFlags
		backupFlags = originalBackupFlags
	}()

	tests := []struct {
		name         string
		setupFunc    func(tempDir string) error
		args         []string
		validateFunc func(tempDir string) error
		expectError  bool
	}{
		{
			name: "Create Basic Backup",
			setupFunc: func(tempDir string) error {
				// Create installation to backup
				installDir := filepath.Join(tempDir, ".claude")
				dirs := []string{
					filepath.Join(installDir, "Core"),
					filepath.Join(installDir, "Commands"),
				}
				for _, dir := range dirs {
					if err := os.MkdirAll(dir, 0755); err != nil {
						return err
					}
				}
				
				// Create test files
				files := map[string]string{
					filepath.Join(installDir, "VERSION"):         "1.0.0",
					filepath.Join(installDir, "Core/CLAUDE.md"):  "# Test CLAUDE",
					filepath.Join(installDir, "Core/FLAGS.md"):   "# Test FLAGS",
					filepath.Join(installDir, "custom.txt"):      "User custom file",
				}
				
				for path, content := range files {
					if err := os.WriteFile(path, []byte(content), 0644); err != nil {
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
					return err
				}
				
				if len(entries) == 0 {
					t.Error("No backup created")
					return nil
				}
				
				// Verify backup file has content
				for _, entry := range entries {
					if strings.HasPrefix(entry.Name(), "crew_backup_") {
						info, _ := entry.Info()
						if info.Size() == 0 {
							t.Error("Backup file is empty")
						}
					}
				}
				
				return nil
			},
			expectError: false,
		},
		{
			name: "Create Backup with Custom Name",
			setupFunc: func(tempDir string) error {
				// Create minimal installation
				installDir := filepath.Join(tempDir, ".claude")
				if err := os.MkdirAll(installDir, 0755); err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(installDir, "test.txt"), []byte("test"), 0644)
			},
			args: []string{"--create", "--name", "my_custom_backup"},
			validateFunc: func(tempDir string) error {
				backupDir := filepath.Join(tempDir, ".claude", "backups")
				entries, err := os.ReadDir(backupDir)
				if err != nil {
					return err
				}
				
				found := false
				for _, entry := range entries {
					if strings.Contains(entry.Name(), "my_custom_backup") {
						found = true
						break
					}
				}
				
				if !found {
					t.Error("Custom backup name not found")
				}
				
				return nil
			},
			expectError: false,
		},
		{
			name: "List Multiple Backups",
			setupFunc: func(tempDir string) error {
				// Create backup directory with multiple backups
				backupDir := filepath.Join(tempDir, ".claude", "backups")
				if err := os.MkdirAll(backupDir, 0755); err != nil {
					return err
				}
				
				// Create mock backup files
				for i := 0; i < 5; i++ {
					filename := fmt.Sprintf("crew_backup_%d_%s.tar.gz", i, time.Now().Format("20060102_150405"))
					path := filepath.Join(backupDir, filename)
					
					// Create a simple tar.gz file
					if err := createMockTarGz(path); err != nil {
						return err
					}
					
					// Set different modification times
					modTime := time.Now().Add(-time.Duration(i) * 24 * time.Hour)
					if err := os.Chtimes(path, modTime, modTime); err != nil {
						return fmt.Errorf("failed to set modification time: %w", err)
					}
				}
				
				return nil
			},
			args: []string{"--list"},
			validateFunc: func(tempDir string) error {
				// Just verify command runs without error
				// Output validation would be done in unit tests
				return nil
			},
			expectError: false,
		},
		{
			name: "Restore from Backup",
			setupFunc: func(tempDir string) error {
				// Create a backup to restore from
				backupDir := filepath.Join(tempDir, ".claude", "backups")
				if err := os.MkdirAll(backupDir, 0755); err != nil {
					return err
				}
				
				// Create a proper tar.gz backup
				backupFile := filepath.Join(backupDir, "test_restore.tar.gz")
				if err := createBackupArchive(backupFile, map[string]string{
					"VERSION":        "1.0.0",
					"Core/CLAUDE.md": "Restored content",
				}); err != nil {
					return err
				}
				
				return nil
			},
			args: []string{"--restore", "test_restore.tar.gz", "--overwrite"},
			validateFunc: func(tempDir string) error {
				// Check restored files
				restoredFile := filepath.Join(tempDir, ".claude", "Core", "CLAUDE.md")
				if content, err := os.ReadFile(restoredFile); err == nil {
					if string(content) != "Restored content" {
						t.Error("Restored content doesn't match")
					}
				}
				return nil
			},
			expectError: false,
		},
		{
			name: "Cleanup Old Backups",
			setupFunc: func(tempDir string) error {
				backupDir := filepath.Join(tempDir, ".claude", "backups")
				if err := os.MkdirAll(backupDir, 0755); err != nil {
					return err
				}
				
				// Create 10 backups with different ages
				for i := 0; i < 10; i++ {
					filename := fmt.Sprintf("crew_backup_%02d.tar.gz", i)
					path := filepath.Join(backupDir, filename)
					
					if err := createMockTarGz(path); err != nil {
						return err
					}
					
					// Set older modification times for first 5
					if i < 5 {
						oldTime := time.Now().Add(-time.Duration(30+i) * 24 * time.Hour)
						if err := os.Chtimes(path, oldTime, oldTime); err != nil {
							return fmt.Errorf("failed to set old modification time: %w", err)
						}
					}
				}
				
				return nil
			},
			args: []string{"--cleanup", "--keep", "3"},
			validateFunc: func(tempDir string) error {
				backupDir := filepath.Join(tempDir, ".claude", "backups")
				entries, err := os.ReadDir(backupDir)
				if err != nil {
					return err
				}
				
				// Should keep only 3 most recent backups
				if len(entries) > 3 {
					t.Errorf("Expected 3 backups, found %d", len(entries))
				}
				
				return nil
			},
			expectError: false,
		},
		{
			name: "Backup with No Compression",
			setupFunc: func(tempDir string) error {
				installDir := filepath.Join(tempDir, ".claude")
				if err := os.MkdirAll(installDir, 0755); err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(installDir, "test.txt"), []byte("test"), 0644)
			},
			args: []string{"--create", "--compress", "none"},
			validateFunc: func(tempDir string) error {
				backupDir := filepath.Join(tempDir, ".claude", "backups")
				entries, err := os.ReadDir(backupDir)
				if err != nil {
					return err
				}
				
				// Check that backup is not compressed
				for _, entry := range entries {
					if !strings.HasSuffix(entry.Name(), ".gz") {
						return nil // Found uncompressed backup
					}
				}
				
				t.Error("No uncompressed backup found")
				return nil
			},
			expectError: false,
		},
		{
			name: "Backup Info Display",
			setupFunc: func(tempDir string) error {
				backupDir := filepath.Join(tempDir, ".claude", "backups")
				if err := os.MkdirAll(backupDir, 0755); err != nil {
					return err
				}
				
				// Create a backup with metadata
				backupFile := filepath.Join(backupDir, "info_test.tar.gz")
				return createBackupArchive(backupFile, map[string]string{
					"VERSION":        "1.0.0",
					"Core/CLAUDE.md": "Test content",
				})
			},
			args: []string{"--info", "info_test.tar.gz"},
			validateFunc: func(tempDir string) error {
				// Command should run without error
				// Output validation would be in unit tests
				return nil
			},
			expectError: false,
		},
		{
			name: "Error on Missing Installation",
			args: []string{"--create"},
			validateFunc: func(tempDir string) error {
				// Should fail because no installation exists
				return nil
			},
			expectError: true,
		},
		{
			name: "Cleanup with Age Filter",
			setupFunc: func(tempDir string) error {
				backupDir := filepath.Join(tempDir, ".claude", "backups")
				if err := os.MkdirAll(backupDir, 0755); err != nil {
					return err
				}
				
				// Create backups of different ages
				for i := 0; i < 5; i++ {
					filename := fmt.Sprintf("crew_backup_age_%d.tar.gz", i)
					path := filepath.Join(backupDir, filename)
					
					if err := createMockTarGz(path); err != nil {
						return err
					}
					
					// Make some backups old
					if i < 3 {
						oldTime := time.Now().Add(-time.Duration(40) * 24 * time.Hour)
						if err := os.Chtimes(path, oldTime, oldTime); err != nil {
							return fmt.Errorf("failed to set old modification time: %w", err)
						}
					}
				}
				
				return nil
			},
			args: []string{"--cleanup", "--older-than", "30"},
			validateFunc: func(tempDir string) error {
				backupDir := filepath.Join(tempDir, ".claude", "backups")
				entries, err := os.ReadDir(backupDir)
				if err != nil {
					return err
				}
				
				// Should have removed backups older than 30 days
				oldCount := 0
				for _, entry := range entries {
					info, _ := entry.Info()
					if time.Since(info.ModTime()) > 30*24*time.Hour {
						oldCount++
					}
				}
				
				if oldCount > 0 {
					t.Errorf("Found %d old backups that should have been removed", oldCount)
				}
				
				return nil
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			tempDir, err := os.MkdirTemp("", "backup-integration-*")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tempDir)
			
			// Reset flags
			globalFlags = GlobalFlags{
				InstallDir: filepath.Join(tempDir, ".claude"),
				Quiet:      true,
			}
			backupFlags = BackupFlags{}
			
			// Run setup
			if tt.setupFunc != nil {
				if err := tt.setupFunc(tempDir); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}
			
			// Create and execute command
			cmd := NewBackupCommand()
			cmd.SetArgs(tt.args)
			
			err = cmd.Execute()
			
			// Check error expectation
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			
			// Validate results
			if !tt.expectError && tt.validateFunc != nil {
				if err := tt.validateFunc(tempDir); err != nil {
					t.Errorf("Validation failed: %v", err)
				}
			}
		})
	}
}

// Helper function to create a mock tar.gz file
func createMockTarGz(filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()
	
	gw := gzip.NewWriter(file)
	defer gw.Close()
	
	tw := tar.NewWriter(gw)
	defer tw.Close()
	
	// Add a simple file to the archive
	header := &tar.Header{
		Name: "test.txt",
		Mode: 0644,
		Size: 4,
	}
	
	if err := tw.WriteHeader(header); err != nil {
		return err
	}
	
	if _, err := tw.Write([]byte("test")); err != nil {
		return err
	}
	
	return nil
}

// Helper function to create a proper backup archive
func createBackupArchive(filepath string, files map[string]string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()
	
	gw := gzip.NewWriter(file)
	defer gw.Close()
	
	tw := tar.NewWriter(gw)
	defer tw.Close()
	
	for name, content := range files {
		header := &tar.Header{
			Name:    name,
			Mode:    0644,
			Size:    int64(len(content)),
			ModTime: time.Now(),
		}
		
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		
		if _, err := io.WriteString(tw, content); err != nil {
			return err
		}
	}
	
	return nil
}

func TestBackupCommandErrorScenarios(t *testing.T) {
	originalFlags := globalFlags
	defer func() { globalFlags = originalFlags }()

	tests := []struct {
		name      string
		setupFunc func(tempDir string) error
		args      []string
		errorMsg  string
	}{
		{
			name: "Restore Non-existent Backup",
			args: []string{"--restore", "nonexistent.tar.gz"},
			errorMsg: "backup file not found",
		},
		{
			name: "Create Backup with Read-only Directory",
			setupFunc: func(tempDir string) error {
				installDir := filepath.Join(tempDir, ".claude")
				if err := os.MkdirAll(installDir, 0755); err != nil {
					return err
				}
				
				// Create backup dir as read-only
				backupDir := filepath.Join(installDir, "backups")
				if err := os.MkdirAll(backupDir, 0555); err != nil {
					return err
				}
				
				return os.WriteFile(filepath.Join(installDir, "test.txt"), []byte("test"), 0644)
			},
			args: []string{"--create"},
			errorMsg: "permission denied",
		},
		{
			name: "Invalid Compression Method",
			setupFunc: func(tempDir string) error {
				installDir := filepath.Join(tempDir, ".claude")
				if err := os.MkdirAll(installDir, 0755); err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(installDir, "test.txt"), []byte("test"), 0644)
			},
			args: []string{"--create", "--compress", "invalid"},
			errorMsg: "invalid compression",
		},
		{
			name: "Cleanup with Invalid Keep Value",
			setupFunc: func(tempDir string) error {
				backupDir := filepath.Join(tempDir, ".claude", "backups") 
				return os.MkdirAll(backupDir, 0755)
			},
			args: []string{"--cleanup", "--keep", "-1"},
			errorMsg: "invalid keep value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			tempDir, err := os.MkdirTemp("", "backup-error-*")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tempDir)
			
			// Reset flags
			globalFlags = GlobalFlags{
				InstallDir: filepath.Join(tempDir, ".claude"),
				Quiet:      true,
			}
			
			// Run setup
			if tt.setupFunc != nil {
				if err := tt.setupFunc(tempDir); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}
			
			// Create and execute command
			cmd := NewBackupCommand()
			cmd.SetArgs(tt.args)
			
			err = cmd.Execute()
			
			// Should always error
			if err == nil {
				t.Error("Expected error but got none")
			}
		})
	}
}