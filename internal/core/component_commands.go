package core

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jonwraymond/claude-code-super-crew/pkg/logger"
)

// CommandsComponent implements the commands library component.
// It provides the slash command definitions and implementations
// for Claude Code Super Crew operations.
type CommandsComponent struct {
	BaseComponent
	sourceDir string
	log       logger.Logger
}

// NewCommandsComponent creates a new commands component instance
func NewCommandsComponent(installDir, sourceDir string) *CommandsComponent {
	c := &CommandsComponent{
		BaseComponent: BaseComponent{
			InstallDir: installDir,
			Metadata: ComponentMetadata{
				Name:         "commands",
				Version:      CommandsComponentVersion,
				Description:  "Claude Code Super Crew command library",
				Category:     "extension",
				Author:       "Claude Code Super Crew Team",
				Tags:         []string{"commands", "slash-commands"},
				Dependencies: []string{"core"},
			},
		},
		sourceDir: sourceDir,
		log:       logger.GetLogger(),
	}

	// Initialize managers
	c.InitManagers(installDir)

	// Discover command files if source directory provided
	if sourceDir != "" {
		// Commands are typically .md files and .sh scripts
		if files, err := c.DiscoverFiles(sourceDir, ".md", []string{"README.md"}); err == nil {
			c.ComponentFiles = files
		}
		// Also discover shell scripts
		if scripts, err := c.DiscoverFiles(sourceDir, ".sh", nil); err == nil {
			c.ComponentFiles = append(c.ComponentFiles, scripts...)
		}
	}

	return c
}

// GetFilesToInstall returns list of files to install
func (c *CommandsComponent) GetFilesToInstall() []FilePair {
	pairs := make([]FilePair, 0, len(c.ComponentFiles))

	for _, file := range c.ComponentFiles {
		source := filepath.Join(c.sourceDir, file)
		// Install to ~/.claude/commands/crew/ for proper namespacing
		target := filepath.Join(c.InstallDir, "commands", "crew", file)
		pairs = append(pairs, FilePair{
			Source: source,
			Target: target,
		})
	}

	return pairs
}

// ValidatePrerequisites checks if the commands component can be installed
func (c *CommandsComponent) ValidatePrerequisites(installDir string) (bool, []string) {
	var errors []string

	// Call base validation
	isValid, baseErrors := c.BaseComponent.ValidatePrerequisites(installDir)
	if !isValid {
		errors = append(errors, baseErrors...)
	}

	// Check core component is installed (dependency)
	coreMarker := filepath.Join(installDir, "CLAUDE.md")
	if !c.FileManager.FileExists(coreMarker) {
		errors = append(errors, "Core component must be installed first")
	}

	// Check source directory
	if c.sourceDir == "" {
		errors = append(errors, "Source directory not specified")
	} else if !c.FileManager.IsDirectory(c.sourceDir) {
		errors = append(errors, fmt.Sprintf("Source directory not found: %s", c.sourceDir))
	}

	return len(errors) == 0, errors
}

// Install creates the commands directory and installs command files
func (c *CommandsComponent) Install(installDir string, config map[string]interface{}) error {
	// Check for dry-run mode
	dryRun := false
	if dryRunVal, exists := config["dry_run"]; exists {
		if dryRunBool, ok := dryRunVal.(bool); ok {
			dryRun = dryRunBool
		}
	}

	if dryRun {
		c.log.Info("[DRY RUN] Would install commands component files")
		return nil
	}

	c.log.Info(fmt.Sprintf("Installing commands component version %s", c.Metadata.Version))

	// Validate prerequisites
	isValid, errors := c.ValidatePrerequisites(installDir)
	if !isValid {
		for _, err := range errors {
			c.log.Error(fmt.Sprintf("Validation error: %s", err))
		}
		return fmt.Errorf("prerequisites validation failed")
	}

	// Create commands directory with crew subdirectory for proper namespacing
	cmdDir := filepath.Join(installDir, "commands", "crew")
	if err := c.FileManager.EnsureDirectoryWithInventory(cmdDir); err != nil {
		return fmt.Errorf("failed to create commands directory: %w", err)
	}

	// Copy all command files
	filesToInstall := c.GetFilesToInstall()
	successCount := 0

	for _, pair := range filesToInstall {
		c.log.Debug(fmt.Sprintf("Copying file from %s to %s", pair.Source, pair.Target))

		if err := c.FileManager.CopyFileWithInventory(pair.Source, pair.Target); err != nil {
			c.log.Error(fmt.Sprintf("Failed to copy file %s: %v", filepath.Base(pair.Source), err))
			continue
		}

		successCount++
	}

	if successCount != len(filesToInstall) {
		return fmt.Errorf("only %d/%d files copied successfully", successCount, len(filesToInstall))
	}

	// Update settings.json
	if err := c.SettingsManager.UpdateComponentVersion(c.Metadata.Name, c.Metadata.Version); err != nil {
		c.log.Error(fmt.Sprintf("Failed to update settings.json: %v", err))
	}

	c.log.Info(fmt.Sprintf("Commands component installed successfully with %d files", successCount))
	return nil
}

// Update installs the new version of commands
func (c *CommandsComponent) Update(installDir string, config map[string]interface{}) error {
	// Simply reinstall for now
	return c.Install(installDir, config)
}

// Uninstall removes the commands directory
func (c *CommandsComponent) Uninstall(installDir string, config map[string]interface{}) error {
	cmdDir := filepath.Join(installDir, ".claude", "commands")
	if err := os.RemoveAll(cmdDir); err != nil {
		return fmt.Errorf("failed to remove commands directory: %w", err)
	}
	return nil
}

// Validate checks if the commands component can be installed
func (c *CommandsComponent) Validate(installDir string) error {
	isValid, errors := c.ValidatePrerequisites(installDir)
	if !isValid {
		return fmt.Errorf("validation failed: %v", errors)
	}
	return nil
}

// IsInstalled checks if the commands directory exists
func (c *CommandsComponent) IsInstalled(installDir string) bool {
	// Check settings.json first
	if !c.BaseComponent.IsInstalled(installDir) {
		return false
	}

	// Also check if commands directory exists
	cmdDir := filepath.Join(installDir, ".claude", "commands")
	return c.FileManager.IsDirectory(cmdDir)
}

// GetInstalledVersion returns the installed version of the commands component
func (c *CommandsComponent) GetInstalledVersion(installDir string) string {
	return c.BaseComponent.GetInstalledVersion(installDir)
}

// GetSizeEstimate returns the estimated size for command files
func (c *CommandsComponent) GetSizeEstimate() int64 {
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
	return 512 * 1024 // 512KB default
}

// ValidateInstallation checks if component is correctly installed
func (c *CommandsComponent) ValidateInstallation(installDir string) (bool, []string) {
	return c.BaseComponent.ValidateInstallation(installDir)
}
