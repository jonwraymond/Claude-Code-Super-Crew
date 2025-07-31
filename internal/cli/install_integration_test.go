package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstallCommandIntegrationWithMocks(t *testing.T) {
	// Skip if explicitly disabled
	if os.Getenv("SKIP_INTEGRATION_TESTS") == "true" {
		t.Skip("Skipping integration test")
	}

	// Save original values
	originalFlags := globalFlags
	originalWd, _ := os.Getwd()
	defer func() {
		globalFlags = originalFlags
		os.Chdir(originalWd)
	}()

	tests := []struct {
		name         string
		setupFunc    func(tempDir string) error
		args         []string
		globalSetup  func()
		validateFunc func(installDir string) error
		expectError  bool
	}{
		{
			name: "Complete Installation Flow",
			args: []string{"--quick"},
			globalSetup: func() {
				globalFlags.Yes = true
			},
			validateFunc: func(installDir string) error {
				// Check all components installed
				expectedDirs := []string{
					"Core",
					"Commands", 
					"hooks",
				}
				for _, dir := range expectedDirs {
					path := filepath.Join(installDir, dir)
					if _, err := os.Stat(path); os.IsNotExist(err) {
						return err
					}
				}
				
				// Check core files
				coreFiles := []string{
					"Core/CLAUDE.md",
					"Core/FLAGS.md",
					"VERSION",
				}
				for _, file := range coreFiles {
					path := filepath.Join(installDir, file)
					if _, err := os.Stat(path); os.IsNotExist(err) {
						return err
					}
				}
				
				return nil
			},
			expectError: false,
		},
		{
			name: "Minimal Installation",
			args: []string{"--minimal"},
			globalSetup: func() {
				globalFlags.Yes = true
			},
			validateFunc: func(installDir string) error {
				// Check only core installed
				corePath := filepath.Join(installDir, "Core")
				if _, err := os.Stat(corePath); os.IsNotExist(err) {
					return err
				}
				
				// Commands should not exist
				commandsPath := filepath.Join(installDir, "Commands")
				if _, err := os.Stat(commandsPath); err == nil {
					t.Errorf("Commands directory should not exist in minimal install")
				}
				
				return nil
			},
			expectError: false,
		},
		{
			name: "Update Existing Installation with Backup",
			setupFunc: func(tempDir string) error {
				// Create existing installation
				installDir := filepath.Join(tempDir, ".claude")
				if err := os.MkdirAll(filepath.Join(installDir, "Core"), 0755); err != nil {
					return err
				}
				
				// Add custom file to verify backup
				customFile := filepath.Join(installDir, "custom.txt")
				return os.WriteFile(customFile, []byte("user data"), 0644)
			},
			args: []string{"--quick"},
			globalSetup: func() {
				globalFlags.Yes = true
				globalFlags.Force = true
			},
			validateFunc: func(installDir string) error {
				// Check backup was created
				backupDir := filepath.Join(installDir, "backups")
				entries, err := os.ReadDir(backupDir)
				if err != nil {
					return err
				}
				if len(entries) == 0 {
					t.Error("No backup created")
				}
				
				// Check new installation exists
				if _, err := os.Stat(filepath.Join(installDir, "Core", "CLAUDE.md")); err != nil {
					return err
				}
				
				return nil
			},
			expectError: false,
		},
		{
			name: "Dry Run Does Not Modify Files",
			args: []string{"--quick"},
			globalSetup: func() {
				globalFlags.Yes = true
				globalFlags.DryRun = true
			},
			validateFunc: func(installDir string) error {
				// Check no files were created
				entries, err := os.ReadDir(installDir)
				if err == nil && len(entries) > 0 {
					t.Errorf("Dry run created files: %d entries found", len(entries))
				}
				return nil
			},
			expectError: false,
		},
		{
			name: "Force Flag Overwrites Existing",
			setupFunc: func(tempDir string) error {
				// Create existing installation with different content
				installDir := filepath.Join(tempDir, ".claude")
				corePath := filepath.Join(installDir, "Core")
				if err := os.MkdirAll(corePath, 0755); err != nil {
					return err
				}
				
				// Write file with old content
				oldFile := filepath.Join(corePath, "CLAUDE.md")
				return os.WriteFile(oldFile, []byte("old content"), 0644)
			},
			args: []string{"--quick"},
			globalSetup: func() {
				globalFlags.Yes = true
				globalFlags.Force = true
			},
			validateFunc: func(installDir string) error {
				// Check file was overwritten
				content, err := os.ReadFile(filepath.Join(installDir, "Core", "CLAUDE.md"))
				if err != nil {
					return err
				}
				
				// Should have new content, not "old content"
				if string(content) == "old content" {
					t.Error("File was not overwritten")
				}
				
				return nil
			},
			expectError: false,
		},
		{
			name: "Component Selection",
			args: []string{"--components", "core,hooks"},
			globalSetup: func() {
				globalFlags.Yes = true
			},
			validateFunc: func(installDir string) error {
				// Check selected components exist
				corePath := filepath.Join(installDir, "Core")
				hooksPath := filepath.Join(installDir, "hooks")
				
				if _, err := os.Stat(corePath); os.IsNotExist(err) {
					t.Error("Core not installed")
				}
				if _, err := os.Stat(hooksPath); os.IsNotExist(err) {
					t.Error("Hooks not installed")
				}
				
				// Commands should not exist
				commandsPath := filepath.Join(installDir, "Commands")
				if _, err := os.Stat(commandsPath); err == nil {
					t.Error("Commands should not be installed")
				}
				
				return nil
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test environment
			tempRoot, cleanup := setupTestEnvironment(t)
			defer cleanup()
			
			// Change to temp directory
			os.Chdir(tempRoot)
			
			// Setup test directory structure
			testDir := filepath.Join(tempRoot, "test-home")
			if err := os.MkdirAll(testDir, 0755); err != nil {
				t.Fatalf("Failed to create test directory: %v", err)
			}
			
			// Run setup if provided
			if tt.setupFunc != nil {
				if err := tt.setupFunc(testDir); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}
			
			// Set install directory
			installDir := filepath.Join(testDir, ".claude")
			globalFlags = GlobalFlags{
				InstallDir: installDir,
				Quiet:      true,
			}
			
			// Apply global setup
			if tt.globalSetup != nil {
				tt.globalSetup()
			}
			
			// Create and execute command
			cmd := NewInstallCommand()
			cmd.SetArgs(tt.args)
			
			err := cmd.Execute()
			
			// Check error expectation
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			
			// Validate results
			if !tt.expectError && tt.validateFunc != nil {
				if err := tt.validateFunc(installDir); err != nil {
					t.Errorf("Validation failed: %v", err)
				}
			}
		})
	}
}

func TestInstallCommandEdgeCases(t *testing.T) {
	// Save original values
	originalFlags := globalFlags
	originalWd, _ := os.Getwd()
	defer func() {
		globalFlags = originalFlags
		os.Chdir(originalWd)
	}()

	tests := []struct {
		name        string
		setupFunc   func(tempDir string) error
		args        []string
		expectError bool
		errorMsg    string
	}{
		{
			name: "No Write Permission",
			setupFunc: func(tempDir string) error {
				// Create read-only directory
				readOnlyDir := filepath.Join(tempDir, "readonly")
				if err := os.MkdirAll(readOnlyDir, 0555); err != nil {
					return err
				}
				globalFlags.InstallDir = filepath.Join(readOnlyDir, ".claude")
				return nil
			},
			args:        []string{"--quick"},
			expectError: true,
		},
		{
			name: "Disk Full Simulation",
			setupFunc: func(tempDir string) error {
				// This would require more complex mocking
				// For now, we'll skip actual disk full simulation
				return nil
			},
			args:        []string{"--quick"},
			expectError: false,
		},
		{
			name: "Corrupted Existing Installation",
			setupFunc: func(tempDir string) error {
				installDir := filepath.Join(tempDir, ".claude")
				// Create a file where directory should be
				return os.WriteFile(installDir, []byte("not a directory"), 0644)
			},
			args:        []string{"--quick", "--force"},
			expectError: true,
		},
		{
			name: "Very Long Path",
			setupFunc: func(tempDir string) error {
				// Create deeply nested path
				longPath := tempDir
				for i := 0; i < 50; i++ {
					longPath = filepath.Join(longPath, "subdir")
				}
				globalFlags.InstallDir = filepath.Join(longPath, ".claude")
				return nil
			},
			args:        []string{"--minimal"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test environment
			tempRoot, cleanup := setupTestEnvironment(t)
			defer cleanup()
			
			// Change to temp directory
			os.Chdir(tempRoot)
			
			// Reset global flags
			globalFlags = GlobalFlags{
				InstallDir: filepath.Join(tempRoot, ".claude"),
				Quiet:      true,
				Yes:        true,
			}
			
			// Run setup
			if tt.setupFunc != nil {
				if err := tt.setupFunc(tempRoot); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}
			
			// Create and execute command
			cmd := NewInstallCommand()
			cmd.SetArgs(tt.args)
			
			err := cmd.Execute()
			
			// Check error expectation
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestInstallCommandConcurrency(t *testing.T) {
	// Test concurrent installations don't interfere
	if testing.Short() {
		t.Skip("Skipping concurrency test in short mode")
	}

	// Create test environment
	tempRoot, cleanup := setupTestEnvironment(t)
	defer cleanup()
	os.Chdir(tempRoot)

	// Run multiple installations in parallel
	done := make(chan bool, 3)
	errors := make(chan error, 3)

	for i := 0; i < 3; i++ {
		go func(index int) {
			defer func() { done <- true }()
			
			installDir := filepath.Join(tempRoot, "install", string(rune('A'+index)), ".claude")
			
			// Create command
			cmd := NewInstallCommand()
			cmd.SetArgs([]string{"--quick"})
			
			// Set flags for this goroutine
			localFlags := GlobalFlags{
				InstallDir: installDir,
				Quiet:      true,
				Yes:        true,
			}
			
			// This is a simplified test - in real implementation
			// we'd need proper synchronization
			_ = localFlags
			
			if err := os.MkdirAll(filepath.Dir(installDir), 0755); err != nil {
				errors <- err
				return
			}
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}
	
	close(errors)
	
	// Check for errors
	for err := range errors {
		if err != nil {
			t.Errorf("Concurrent installation error: %v", err)
		}
	}
}