package versioning

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jonwraymond/claude-code-super-crew/internal/core"
)

// VersionManager handles version tracking and updates for Claude Code Super Crew
type VersionManager struct {
	installDir     string
	metadataFile   string
}

// NewVersionManager creates a new version manager
func NewVersionManager(installDir string) *VersionManager {
	return &VersionManager{
		installDir:     installDir,
		metadataFile:   filepath.Join(installDir, ".crew", "config", "crew-metadata.json"),
	}
}

// VersionInfo represents version information for a component or framework
type VersionInfo struct {
	Version      string    `json:"version"`
	ReleaseDate  string    `json:"release_date,omitempty"`
	UpdatedAt    time.Time `json:"updated_at"`
	PreviousVersion string `json:"previous_version,omitempty"`
}

// InstallationMetadata represents the complete installation metadata
type InstallationMetadata struct {
	Framework  VersionInfo              `json:"framework"`
	Components map[string]VersionInfo   `json:"components"`
	Settings   map[string]interface{}   `json:"settings,omitempty"`
}

// GetCurrentVersion returns the current framework version from metadata
func (vm *VersionManager) GetCurrentVersion() (string, error) {
	// Use only metadata file for version tracking
	metadata, err := vm.LoadMetadata()
	if err != nil {
		return "", err
	}
	
	return metadata.Framework.Version, nil
}

// SetVersion updates the framework version in metadata only
func (vm *VersionManager) SetVersion(version string) error {
	// Validate version format
	if !vm.IsValidVersion(version) {
		return fmt.Errorf("invalid version format: %s", version)
	}
	
	// Update metadata only - no longer using VERSION file
	metadata, err := vm.LoadMetadata()
	if err != nil {
		// Create new metadata if it doesn't exist
		metadata = &InstallationMetadata{
			Components: make(map[string]VersionInfo),
		}
	}
	
	// Store previous version
	previousVersion := metadata.Framework.Version
	
	metadata.Framework = VersionInfo{
		Version:         version,
		UpdatedAt:       time.Now(),
		PreviousVersion: previousVersion,
		ReleaseDate:     time.Now().Format("2006-01-02"),
	}
	
	return vm.SaveMetadata(metadata)
}

// GetComponentVersion returns the version of a specific component
func (vm *VersionManager) GetComponentVersion(component string) (string, error) {
	// Use constants from core package
	switch component {
	case "core":
		return core.CoreComponentVersion, nil
	case "commands":
		return core.CommandsComponentVersion, nil
	case "hooks":
		return core.HooksComponentVersion, nil
	case "mcp":
		return core.MCPComponentVersion, nil
	case "agents":
		return core.AgentsComponentVersion, nil
	}
	
	// Check metadata for custom components
	metadata, err := vm.LoadMetadata()
	if err != nil {
		return "", err
	}
	
	if info, exists := metadata.Components[component]; exists {
		return info.Version, nil
	}
	
	return "", fmt.Errorf("component not found: %s", component)
}

// SetComponentVersion updates the version of a specific component
func (vm *VersionManager) SetComponentVersion(component, version string) error {
	if !vm.IsValidVersion(version) {
		return fmt.Errorf("invalid version format: %s", version)
	}
	
	metadata, err := vm.LoadMetadata()
	if err != nil {
		metadata = &InstallationMetadata{
			Components: make(map[string]VersionInfo),
		}
	}
	
	// Store previous version if exists
	var previousVersion string
	if existing, ok := metadata.Components[component]; ok {
		previousVersion = existing.Version
	}
	
	metadata.Components[component] = VersionInfo{
		Version:         version,
		UpdatedAt:       time.Now(),
		PreviousVersion: previousVersion,
	}
	
	return vm.SaveMetadata(metadata)
}

// CheckForUpdates compares current version with a target version
func (vm *VersionManager) CheckForUpdates(targetVersion string) (bool, error) {
	currentVersion, err := vm.GetCurrentVersion()
	if err != nil {
		return false, err
	}
	
	comparison := vm.CompareVersions(currentVersion, targetVersion)
	return comparison < 0, nil
}

// CompareVersions compares two semantic versions
// Returns: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func (vm *VersionManager) CompareVersions(v1, v2 string) int {
	parts1 := vm.parseVersion(v1)
	parts2 := vm.parseVersion(v2)
	
	for i := 0; i < 3; i++ {
		if parts1[i] < parts2[i] {
			return -1
		}
		if parts1[i] > parts2[i] {
			return 1
		}
	}
	
	return 0
}

// IsValidVersion checks if a version string follows semantic versioning
func (vm *VersionManager) IsValidVersion(version string) bool {
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return false
	}
	
	for _, part := range parts {
		if _, err := strconv.Atoi(part); err != nil {
			return false
		}
	}
	
	return true
}

// LoadMetadata loads the installation metadata
func (vm *VersionManager) LoadMetadata() (*InstallationMetadata, error) {
	data, err := os.ReadFile(vm.metadataFile)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty metadata if file doesn't exist
			return &InstallationMetadata{
				Framework: VersionInfo{
					Version: core.FrameworkVersion,
				},
				Components: make(map[string]VersionInfo),
			}, nil
		}
		return nil, err
	}
	
	var metadata InstallationMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, err
	}
	
	// Ensure components map exists
	if metadata.Components == nil {
		metadata.Components = make(map[string]VersionInfo)
	}
	
	return &metadata, nil
}

// SaveMetadata saves the installation metadata
func (vm *VersionManager) SaveMetadata(metadata *InstallationMetadata) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(vm.metadataFile), 0755); err != nil {
		return err
	}
	
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(vm.metadataFile, data, 0644)
}

// GetVersionHistory returns the version history from metadata
func (vm *VersionManager) GetVersionHistory() ([]string, error) {
	metadata, err := vm.LoadMetadata()
	if err != nil {
		return nil, err
	}
	
	history := []string{metadata.Framework.Version}
	if metadata.Framework.PreviousVersion != "" {
		history = append(history, metadata.Framework.PreviousVersion)
	}
	
	return history, nil
}

// StandardizeAllVersions ensures all components use version 1.0.0
func (vm *VersionManager) StandardizeAllVersions() error {
	standardVersion := "1.0.0"
	
	// Set framework version
	if err := vm.SetVersion(standardVersion); err != nil {
		return err
	}
	
	// Set all component versions
	components := []string{"core", "commands", "hooks", "mcp"}
	for _, component := range components {
		if err := vm.SetComponentVersion(component, standardVersion); err != nil {
			return err
		}
	}
	
	return nil
}

// parseVersion parses a semantic version string into major, minor, patch
func (vm *VersionManager) parseVersion(version string) [3]int {
	var result [3]int
	parts := strings.Split(version, ".")
	
	for i := 0; i < 3 && i < len(parts); i++ {
		val, _ := strconv.Atoi(parts[i])
		result[i] = val
	}
	
	return result
}

// UpdateInfo represents information about an available update
type UpdateInfo struct {
	CurrentVersion   string `json:"current_version"`
	AvailableVersion string `json:"available_version"`
	UpdateAvailable  bool   `json:"update_available"`
	ReleaseNotes     string `json:"release_notes,omitempty"`
}

// CheckUpdateStatus checks if updates are available based on metadata
func (vm *VersionManager) CheckUpdateStatus(availableVersion string) (*UpdateInfo, error) {
	currentVersion, err := vm.GetCurrentVersion()
	if err != nil {
		return nil, err
	}
	
	needsUpdate, err := vm.CheckForUpdates(availableVersion)
	if err != nil {
		return nil, err
	}
	
	return &UpdateInfo{
		CurrentVersion:   currentVersion,
		AvailableVersion: availableVersion,
		UpdateAvailable:  needsUpdate,
	}, nil
}