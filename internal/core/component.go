// Package core provides the fundamental types and interfaces for the Claude Code Super Crew framework.
// It defines the component system that allows modular installation and management of framework features.
//
// The package includes:
//   - Component interface and base implementation
//   - Component registry for discovery and management
//   - System validator for requirement checking
//   - Individual component implementations (core, commands, hooks, MCP)
package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jonwraymond/claude-code-super-crew/internal/managers"
)

// ComponentMetadata holds metadata about a component
type ComponentMetadata struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Description  string            `json:"description"`
	Category     string            `json:"category"`
	Author       string            `json:"author,omitempty"`
	URL          string            `json:"url,omitempty"`
	Tags         []string          `json:"tags,omitempty"`
	Dependencies []string          `json:"dependencies,omitempty"`
	Conflicts    []string          `json:"conflicts,omitempty"`
	Requirements map[string]string `json:"requirements,omitempty"`
}

// ComponentFactory creates component instances
type ComponentFactory func(installDir, sourceDir string) Component

// Component interface that all components must implement
type Component interface {
	// GetMetadata returns component metadata
	GetMetadata() ComponentMetadata

	// Install installs the component
	Install(installDir string, config map[string]interface{}) error

	// Update updates the component
	Update(installDir string, config map[string]interface{}) error

	// Uninstall removes the component
	Uninstall(installDir string, config map[string]interface{}) error

	// Validate checks if component can be installed
	Validate(installDir string) error

	// GetSizeEstimate returns estimated size in bytes
	GetSizeEstimate() int64

	// IsInstalled checks if component is already installed
	IsInstalled(installDir string) bool

	// GetInstalledVersion returns currently installed version
	GetInstalledVersion(installDir string) string

	// GetFilesToInstall returns list of files to install
	GetFilesToInstall() []FilePair

	// ValidatePrerequisites checks if component can be installed
	ValidatePrerequisites(installDir string) (bool, []string)

	// ValidateInstallation checks if component is correctly installed
	ValidateInstallation(installDir string) (bool, []string)
}

// FilePair represents a source and destination file pair
type FilePair struct {
	Source string
	Target string
}

// BaseComponent provides common functionality for components
type BaseComponent struct {
	Metadata          ComponentMetadata
	InstallDir        string
	ComponentFiles    []string
	FileManager       *managers.FileManager
	SettingsManager   *managers.SettingsManager
	SecurityValidator *managers.SecurityValidator
}

// GetMetadata returns component metadata
func (b *BaseComponent) GetMetadata() ComponentMetadata {
	return b.Metadata
}

// GetInstallPath returns the full installation path for a file
func (b *BaseComponent) GetInstallPath(filename string) string {
	return filepath.Join(b.InstallDir, filename)
}

// GetSizeEstimate returns a default size estimate
func (b *BaseComponent) GetSizeEstimate() int64 {
	return 1024 * 1024 // 1MB default
}

// DiscoverFiles discovers files in a directory with given extension
func (b *BaseComponent) DiscoverFiles(directory, extension string, excludePatterns []string) ([]string, error) {
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		return nil, fmt.Errorf("directory not found: %s", directory)
	}

	files := []string{}
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		// Check extension
		if !strings.HasSuffix(strings.ToLower(name), strings.ToLower(extension)) {
			continue
		}

		// Check exclude patterns
		excluded := false
		for _, pattern := range excludePatterns {
			if name == pattern {
				excluded = true
				break
			}
		}

		if !excluded {
			files = append(files, name)
		}
	}

	return files, nil
}

// InitManagers initializes the helper managers
func (b *BaseComponent) InitManagers(installDir string) {
	if b.FileManager == nil {
		b.FileManager = managers.NewFileManagerWithMetadata(installDir)
	}
	if b.SettingsManager == nil {
		b.SettingsManager = managers.NewSettingsManager(installDir)
	}
	if b.SecurityValidator == nil {
		b.SecurityValidator = managers.NewSecurityValidator()
	}
}

// ValidatePrerequisites provides base validation logic
func (b *BaseComponent) ValidatePrerequisites(installDir string) (bool, []string) {
	var errors []string

	// Initialize managers if needed
	b.InitManagers(installDir)

	// Validate installation target
	targetDir := filepath.Join(installDir, ".claude")
	isValid, validationErrors := b.SecurityValidator.ValidateInstallationTarget(targetDir)
	if !isValid {
		errors = append(errors, validationErrors...)
	}

	// Check write permissions
	hasPerms, permErrors := b.SecurityValidator.CheckPermissions(installDir, []string{"write"})
	if !hasPerms {
		errors = append(errors, permErrors...)
	}

	return len(errors) == 0, errors
}

// GetInstalledVersion checks the metadata for installed version
func (b *BaseComponent) GetInstalledVersion(installDir string) string {
	b.InitManagers(installDir)

	// Try metadata first (new format)
	if version, err := b.SettingsManager.GetComponentVersionFromMetadata(b.Metadata.Name); err == nil && version != "" {
		return version
	}

	// Fall back to legacy installation.json format
	components, err := b.SettingsManager.GetInstalledComponents()
	if err != nil {
		return ""
	}

	if version, exists := components[b.Metadata.Name]; exists {
		return version
	}
	return ""
}

// IsInstalled checks if component is installed
func (b *BaseComponent) IsInstalled(installDir string) bool {
	return b.GetInstalledVersion(installDir) != ""
}

// ValidateInstallation checks if all component files exist
func (b *BaseComponent) ValidateInstallation(installDir string) (bool, []string) {
	var errors []string

	// Check if registered in metadata or settings
	if b.GetInstalledVersion(installDir) == "" {
		errors = append(errors, "Component not registered in metadata or settings")
	}

	// Check if all files exist
	for _, filePair := range b.GetFilesToInstall() {
		if !b.FileManager.FileExists(filePair.Target) {
			errors = append(errors, fmt.Sprintf("Missing file: %s", filePair.Target))
		}
	}

	return len(errors) == 0, errors
}

// GetFilesToInstall returns empty list - should be overridden by specific components
func (b *BaseComponent) GetFilesToInstall() []FilePair {
	return []FilePair{}
}
