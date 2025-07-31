package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// HooksComponent implements the hooks and automation component.
// It provides event-driven automation capabilities for Claude Code Super Crew,
// allowing users to run custom scripts on various events.
type HooksComponent struct {
	BaseComponent
}

// NewHooksComponent creates a new hooks component instance
func NewHooksComponent(installDir, sourceDir string) *HooksComponent {
	component := &HooksComponent{
		BaseComponent: BaseComponent{
			InstallDir: installDir,
			Metadata: ComponentMetadata{
				Name:         "hooks",
				Version:      HooksComponentVersion,
				Description:  "Event hooks and automation",
				Category:     "automation",
				Author:       "Claude Code Super Crew Team",
				Tags:         []string{"hooks", "automation", "events"},
				Dependencies: []string{"core"},
			},
		},
	}

	// Initialize managers for inventory tracking
	component.InitManagers(installDir)

	return component
}

// Install creates the hooks directory structure and installs hook templates
func (c *HooksComponent) Install(installDir string, config map[string]interface{}) error {
	// Check for dry-run mode
	dryRun := false
	if dryRunVal, exists := config["dry_run"]; exists {
		if dryRunBool, ok := dryRunVal.(bool); ok {
			dryRun = dryRunBool
		}
	}

	if dryRun {
		fmt.Printf("[DRY RUN] Would install hooks component files\n")
		return nil
	}
	// Create hooks directory structure with inventory tracking
	dirs := []string{
		filepath.Join(installDir, "hooks"),
		filepath.Join(installDir, "hooks", "templates"),
		filepath.Join(installDir, "hooks", "examples"),
	}

	for _, dir := range dirs {
		if err := c.FileManager.EnsureDirectoryWithInventory(dir); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Get project root - use current working directory instead of executable path
	projectRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	// Source hooks directory
	sourceHooksDir := filepath.Join(projectRoot, "SuperCrew", "hooks")

	// Copy hook scripts and documentation with inventory tracking
	entries, err := os.ReadDir(sourceHooksDir)
	if err != nil {
		return fmt.Errorf("failed to read source hooks directory: %w", err)
	}

	if len(entries) > 0 {
		for _, entry := range entries {
			// Copy .sh scripts and .md documentation files
			if strings.HasSuffix(entry.Name(), ".sh") || strings.HasSuffix(entry.Name(), ".md") {
				src := filepath.Join(sourceHooksDir, entry.Name())
				dst := filepath.Join(installDir, "hooks", entry.Name())

				// Use FileManager for proper inventory tracking
				if err := c.FileManager.CopyFileWithInventory(src, dst); err != nil {
					// Log error but continue with other files
					fmt.Printf("Warning: Could not copy hook file %s: %v\n", entry.Name(), err)
					continue
				}

				// Set executable permissions for .sh files only
				if strings.HasSuffix(entry.Name(), ".sh") {
					if err := os.Chmod(dst, 0755); err != nil {
						fmt.Printf("Warning: Could not set executable permissions on %s: %v\n", entry.Name(), err)
					}
				}
			}
		}
	}

	// Create example hook configuration with inventory tracking
	exampleConfig := `{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Write|Edit|MultiEdit",
        "hooks": [
          {
            "type": "command",
            "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/git-auto-commit.sh",
            "env": {
              "SUPERCREW_GIT_AUTO_COMMIT": "true"
            }
          }
        ]
      }
    ]
  }
}`

	examplePath := filepath.Join(installDir, "hooks", "examples", "hook-config.json")
	if err := os.WriteFile(examplePath, []byte(exampleConfig), 0644); err != nil {
		// Non-fatal error for example file
		fmt.Printf("Warning: Could not write example config: %v\n", err)
	} else {
		// Track the example file in inventory
		if c.FileManager.HasMetadataManager() {
			if err := c.FileManager.AddToInventory(examplePath, false); err != nil {
				fmt.Printf("Warning: Could not track example config in inventory: %v\n", err)
			}
		}
	}

	// Write version file with inventory tracking
	versionFile := filepath.Join(installDir, "hooks", ".version")
	if err := os.WriteFile(versionFile, []byte(c.Metadata.Version), 0644); err != nil {
		// Non-fatal error, just log it
		fmt.Printf("Warning: Could not write version file: %v\n", err)
	} else {
		// Track the version file in inventory
		if c.FileManager.HasMetadataManager() {
			if err := c.FileManager.AddToInventory(versionFile, false); err != nil {
				fmt.Printf("Warning: Could not track version file in inventory: %v\n", err)
			}
		}
	}

	return nil
}

// Update installs the new version of hooks
func (c *HooksComponent) Update(installDir string, config map[string]interface{}) error {
	// Preserve user hooks while updating templates
	// For now, just reinstall
	return c.Install(installDir, config)
}

// Uninstall removes the hooks directory
func (c *HooksComponent) Uninstall(installDir string, config map[string]interface{}) error {
	hooksDir := filepath.Join(installDir, "hooks")

	// Optionally preserve user-created hooks
	preserveUserHooks := false
	if preserve, ok := config["preserve_user_hooks"].(bool); ok {
		preserveUserHooks = preserve
	}

	if preserveUserHooks {
		// Only remove templates and examples
		templatesDir := filepath.Join(hooksDir, "templates")
		examplesDir := filepath.Join(hooksDir, "examples")
		os.RemoveAll(templatesDir)
		os.RemoveAll(examplesDir)
	} else {
		// Remove entire hooks directory
		if err := os.RemoveAll(hooksDir); err != nil {
			return fmt.Errorf("failed to remove hooks directory: %w", err)
		}
	}

	return nil
}

// Validate checks if the hooks component can be installed
func (c *HooksComponent) Validate(installDir string) error {
	// Check if core component is installed (dependency)
	coreMarker := filepath.Join(installDir, "CLAUDE.md")
	if _, err := os.Stat(coreMarker); os.IsNotExist(err) {
		return fmt.Errorf("core component must be installed first")
	}
	return nil
}

// IsInstalled checks if the hooks directory exists
func (c *HooksComponent) IsInstalled(installDir string) bool {
	hooksDir := filepath.Join(installDir, "hooks")
	info, err := os.Stat(hooksDir)
	return err == nil && info.IsDir()
}

// GetInstalledVersion returns the installed version of the hooks component
func (c *HooksComponent) GetInstalledVersion(installDir string) string {
	// Check for version file in hooks directory
	versionFile := filepath.Join(installDir, "hooks", ".version")
	if content, err := os.ReadFile(versionFile); err == nil {
		return strings.TrimSpace(string(content))
	}

	// Fallback: Check if hooks directory exists and assume current version
	if c.IsInstalled(installDir) {
		return c.Metadata.Version
	}

	return "unknown"
}

// GetSizeEstimate returns the estimated size for hook files
func (c *HooksComponent) GetSizeEstimate() int64 {
	// Approximately 256KB for hook templates and examples
	return 256 * 1024
}
