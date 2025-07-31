package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jonwraymond/claude-code-super-crew/internal/metadata"
	"github.com/jonwraymond/claude-code-super-crew/pkg/logger"
)

// CoreComponent implements the core framework component.
// It manages the essential Claude Code Super Crew framework files including
// CLAUDE.md, COMMANDS.md, FLAGS.md, and other core documentation.
type CoreComponent struct {
	BaseComponent
	sourceDir string
	log       logger.Logger
}

// NewCoreComponent creates a new core component instance
func NewCoreComponent(installDir, sourceDir string) *CoreComponent {
	c := &CoreComponent{
		BaseComponent: BaseComponent{
			InstallDir: installDir,
			Metadata: ComponentMetadata{
				Name:        "core",
				Version:     CoreComponentVersion,
				Description: "Core Claude Code Super Crew framework files",
				Category:    "core",
				Author:      "Claude Code Super Crew Team",
				Tags:        []string{"essential", "framework"},
			},
		},
		sourceDir: sourceDir,
		log:       logger.GetLogger(),
	}

	// Initialize managers
	c.InitManagers(installDir)

	// Discover component files
	if sourceDir != "" {
		excludePatterns := []string{"README.md", "CHANGELOG.md", "LICENSE.md"}
		if files, err := c.DiscoverFiles(sourceDir, ".md", excludePatterns); err == nil {
			c.ComponentFiles = files
			c.log.Debug(fmt.Sprintf("Discovered %d core files: %v", len(files), files))
		} else {
			c.log.Error(fmt.Sprintf("Failed to discover core files: %v", err))
		}
	}

	return c
}

// GetFilesToInstall returns list of files to install
func (c *CoreComponent) GetFilesToInstall() []FilePair {
	pairs := make([]FilePair, 0, len(c.ComponentFiles))

	for _, file := range c.ComponentFiles {
		source := filepath.Join(c.sourceDir, file)
		target := filepath.Join(c.InstallDir, file)
		pairs = append(pairs, FilePair{
			Source: source,
			Target: target,
		})
	}

	return pairs
}

// ValidatePrerequisites checks if the core component can be installed
func (c *CoreComponent) ValidatePrerequisites(installDir string) (bool, []string) {
	var errors []string

	// Call base validation
	isValid, baseErrors := c.BaseComponent.ValidatePrerequisites(installDir)
	if !isValid {
		errors = append(errors, baseErrors...)
	}

	// Check source directory exists
	if c.sourceDir == "" {
		errors = append(errors, "Source directory not specified")
	} else if !c.FileManager.IsDirectory(c.sourceDir) {
		errors = append(errors, fmt.Sprintf("Source directory not found: %s", c.sourceDir))
	}

	// Check all source files exist
	for _, file := range c.ComponentFiles {
		sourcePath := filepath.Join(c.sourceDir, file)
		if !c.FileManager.FileExists(sourcePath) {
			errors = append(errors, fmt.Sprintf("Missing source file: %s", file))
		}
	}

	// Validate all files for security
	filesToInstall := c.GetFilesToInstall()
	fileNames := make([]string, len(filesToInstall))
	for i, pair := range filesToInstall {
		fileNames[i] = filepath.Base(pair.Source)
	}

	targetDir := installDir
	isSecure, securityErrors := c.SecurityValidator.ValidateComponentFiles(fileNames, c.sourceDir, targetDir)
	if !isSecure {
		errors = append(errors, securityErrors...)
	}

	return len(errors) == 0, errors
}

// Install creates the core directory structure and installs framework files
func (c *CoreComponent) Install(installDir string, config map[string]interface{}) error {
	c.log.Info("=== CORE COMPONENT INSTALL METHOD CALLED ===")
	c.log.Info(fmt.Sprintf("Installing core component version %s", c.Metadata.Version))

	// Check for dry-run mode
	dryRun := false
	if dryRunVal, exists := config["dry_run"]; exists {
		if dryRunBool, ok := dryRunVal.(bool); ok {
			dryRun = dryRunBool
		}
	}

	if dryRun {
		c.log.Info("[DRY RUN] Would install core component files")
		return nil
	}

	// FileManager is already initialized with metadata tracking via InitManagers

	// Validate prerequisites
	isValid, errors := c.ValidatePrerequisites(installDir)
	if !isValid {
		for _, err := range errors {
			c.log.Error(fmt.Sprintf("Validation error: %s", err))
		}
		return fmt.Errorf("prerequisites validation failed")
	}

	// Create main directory and .crew subdirectory for utilities
	dirs := []string{
		installDir,
		// Core SuperCrew directories (remain in main .claude/)
		filepath.Join(installDir, "commands"),
		filepath.Join(installDir, "hooks"),
		filepath.Join(installDir, "agents"),
		// Utility directories (moved to .crew/)
		filepath.Join(installDir, ".crew"),
		filepath.Join(installDir, ".crew", "backups"),
		filepath.Join(installDir, ".crew", "logs"),
		filepath.Join(installDir, ".crew", "workflows"),
		filepath.Join(installDir, ".crew", "scripts"),
		filepath.Join(installDir, ".crew", "config"),
		filepath.Join(installDir, ".crew", "prompts"),
		filepath.Join(installDir, ".crew", "completions"),
	}

	for _, dir := range dirs {
		c.log.Infof("Creating directory: %s", dir)
		if err := c.FileManager.EnsureDirectoryWithInventory(dir); err != nil {
			c.log.Errorf("Failed to create directory %s: %v", dir, err)
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
		c.log.Infof("Successfully created directory: %s", dir)
	}

	// Copy all framework files
	filesToInstall := c.GetFilesToInstall()
	successCount := 0

	// Check CLAUDE.md handling flags
	claudeOverwrite := false
	claudeSkip := false
	if overwriteVal, exists := config["claude_overwrite"]; exists {
		if overwriteBool, ok := overwriteVal.(bool); ok {
			claudeOverwrite = overwriteBool
		}
	}
	if skipVal, exists := config["claude_skip"]; exists {
		if skipBool, ok := skipVal.(bool); ok {
			claudeSkip = skipBool
		}
	}

	for _, pair := range filesToInstall {
		c.log.Debug(fmt.Sprintf("Copying file from %s to %s", pair.Source, pair.Target))

		// Handle CLAUDE.md files based on flags
		if filepath.Base(pair.Target) == "CLAUDE.md" {
			if claudeSkip {
				// Skip CLAUDE.md installation
				c.log.Info("Skipping CLAUDE.md installation as requested")
				successCount++
				continue
			} else if claudeOverwrite {
				// Overwrite existing CLAUDE.md
				if err := c.FileManager.CopyFileWithInventory(pair.Source, pair.Target); err != nil {
					c.log.Error(fmt.Sprintf("Failed to copy file %s: %v", filepath.Base(pair.Source), err))
					continue
				}
				c.log.Debug(fmt.Sprintf("Successfully overwrote file %s", filepath.Base(pair.Source)))
			} else {
				// Default: merge functionality for CLAUDE.md files to preserve user content
				if err := c.FileManager.CopyFileWithMerge(pair.Source, pair.Target); err != nil {
					c.log.Error(fmt.Sprintf("Failed to copy/merge file %s: %v", filepath.Base(pair.Source), err))
					continue
				}
				c.log.Debug(fmt.Sprintf("Successfully copied/merged file %s", filepath.Base(pair.Source)))
			}
		} else {
			if err := c.FileManager.CopyFileWithInventory(pair.Source, pair.Target); err != nil {
				c.log.Error(fmt.Sprintf("Failed to copy file %s: %v", filepath.Base(pair.Source), err))
				continue
			}
			c.log.Debug(fmt.Sprintf("Successfully copied file %s", filepath.Base(pair.Source)))
		}

		successCount++
	}

	if successCount != len(filesToInstall) {
		return fmt.Errorf("only %d/%d files copied successfully", successCount, len(filesToInstall))
	}

	// Install orchestrator agent
	if err := c.installOrchestratorAgent(installDir); err != nil {
		c.log.Error(fmt.Sprintf("Failed to install orchestrator agent: %v", err))
		// Don't fail the entire installation, just log the error
	}

	// Create default config.json if it doesn't exist
	configFile := filepath.Join(installDir, ".crew", "config", "config.json")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		defaultConfig := map[string]interface{}{
			"version":     c.Metadata.Version,
			"install_dir": installDir,
			"components":  []string{"core"},
			"settings": map[string]interface{}{
				"auto_update":      false,
				"telemetry":        false,
				"log_level":        "info",
				"backup_on_update": true,
				"theme":            "dark",
			},
			"last_updated": time.Now().Format(time.RFC3339),
		}

		configData, err := json.MarshalIndent(defaultConfig, "", "  ")
		if err == nil {
			if err := os.WriteFile(configFile, configData, 0644); err == nil {
				// Track config file in inventory
				if c.FileManager.HasMetadataManager() {
					if metadataManager := metadata.NewMetadataManager(installDir); metadataManager != nil {
						metadataManager.AddToInventory(configFile, false)
					}
				}
			}
		}
	}

	// Update settings.json
	if err := c.SettingsManager.UpdateComponentVersion(c.Metadata.Name, c.Metadata.Version); err != nil {
		c.log.Error(fmt.Sprintf("Failed to update settings.json: %v", err))
		// Don't fail installation for settings update failure
	}

	c.log.Info(fmt.Sprintf("Core component installed successfully with %d files", successCount))
	return nil
}

// Update backs up existing files and installs the new version
func (c *CoreComponent) Update(installDir string, config map[string]interface{}) error {
	// For now, update is the same as install
	// In production, this would backup existing customizations
	return c.Install(installDir, config)
}

// Uninstall removes core framework files while preserving user data
func (c *CoreComponent) Uninstall(installDir string, config map[string]interface{}) error {
	// List of core files to remove (preserve user-created content)
	coreFiles := []string{
		"CLAUDE.md",
		"COMMANDS.md",
		"FLAGS.md",
		"PRINCIPLES.md",
		"RULES.md",
		"MCP.md",
		"PERSONAS.md",
		"ORCHESTRATOR.md",
		"MODES.md",
		"AGENTS_INDEX.md",
	}

	claudeDir := filepath.Join(installDir, ".claude")
	for _, file := range coreFiles {
		filePath := filepath.Join(claudeDir, file)
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			// Log error but continue with other files
			fmt.Printf("Warning: failed to remove %s: %v\n", file, err)
		}
	}

	return nil
}

// Validate ensures the installation directory is writable
func (c *CoreComponent) Validate(installDir string) error {
	isValid, errors := c.ValidatePrerequisites(installDir)
	if !isValid {
		return fmt.Errorf("validation failed: %v", errors)
	}
	return nil
}

// IsInstalled checks if the core component is installed by looking for marker files
func (c *CoreComponent) IsInstalled(installDir string) bool {
	// Check settings.json first
	if !c.BaseComponent.IsInstalled(installDir) {
		return false
	}

	// Also check if CLAUDE.md exists as a marker file (now in root)
	markerFile := filepath.Join(installDir, "CLAUDE.md")
	return c.FileManager.FileExists(markerFile)
}

// GetInstalledVersion reads the version from the installed files
func (c *CoreComponent) GetInstalledVersion(installDir string) string {
	return c.BaseComponent.GetInstalledVersion(installDir)
}

// GetSizeEstimate returns the estimated installation size
func (c *CoreComponent) GetSizeEstimate() int64 {
	var totalSize int64

	for _, pair := range c.GetFilesToInstall() {
		if size, err := c.FileManager.GetFileSize(pair.Source); err == nil {
			totalSize += size
		}
	}

	// Return calculated size or default if no files found
	if totalSize > 0 {
		return totalSize
	}
	return 2 * 1024 * 1024 // 2MB default
}

// installOrchestratorAgent copies the orchestrator agent file to the agents directory
func (c *CoreComponent) installOrchestratorAgent(installDir string) error {
	// Use the single canonical source path for the orchestrator agent
	sourceFile := filepath.Join(c.sourceDir, "..", "agents", "orchestrator-agent.md")

	// Check if the file exists at the expected location
	if !c.FileManager.FileExists(sourceFile) {
		c.log.Warn(fmt.Sprintf("Orchestrator agent file not found at expected path: %s", sourceFile))
		return fmt.Errorf("orchestrator-agent.md not found at %s", sourceFile)
	}

	// Target location
	targetFile := filepath.Join(installDir, "agents", "orchestrator-agent.md")

	// Copy the file with inventory tracking
	if err := c.FileManager.CopyFileWithInventory(sourceFile, targetFile); err != nil {
		return fmt.Errorf("failed to copy orchestrator agent: %w", err)
	}

	// Set permissions to 0644
	if err := os.Chmod(targetFile, 0644); err != nil {
		c.log.Warn(fmt.Sprintf("Failed to set permissions on orchestrator agent: %v", err))
		// Don't fail installation for permission issues
	}

	c.log.Info(fmt.Sprintf("Successfully installed orchestrator agent from %s", sourceFile))
	return nil
}

// ValidateInstallation checks if component is correctly installed
func (c *CoreComponent) ValidateInstallation(installDir string) (bool, []string) {
	isValid, errors := c.BaseComponent.ValidateInstallation(installDir)

	// Also check for orchestrator agent
	orchestratorPath := filepath.Join(installDir, "agents", "orchestrator-agent.md")
	if !c.FileManager.FileExists(orchestratorPath) {
		errors = append(errors, "orchestrator-agent.md not found in agents directory")
		isValid = false
	}

	return isValid, errors
}
