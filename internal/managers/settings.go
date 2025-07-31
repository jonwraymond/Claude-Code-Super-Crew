package managers

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/jonwraymond/claude-code-super-crew/internal/metadata"
)

// SettingsManager handles installation settings with unified metadata
type SettingsManager struct {
	installDir      string
	metadataManager *metadata.MetadataManager
}

// NewSettingsManager creates a new settings manager with unified metadata
func NewSettingsManager(installDir string) *SettingsManager {
	return &SettingsManager{
		installDir:      installDir,
		metadataManager: metadata.NewMetadataManager(installDir),
	}
}

// CheckInstallationExists checks if Claude Code Super Crew is installed
func (m *SettingsManager) CheckInstallationExists() bool {
	return m.metadataManager.CheckInstallationExists()
}

// GetInstalledComponents returns currently installed components and versions
func (m *SettingsManager) GetInstalledComponents() (map[string]string, error) {
	meta, err := m.metadataManager.LoadMetadata()
	if err != nil {
		return nil, err
	}

	components := make(map[string]string)
	for name, comp := range meta.Components {
		components[name] = comp.Version
	}

	return components, nil
}

// UpdateComponentVersion updates the version of a specific component
func (m *SettingsManager) UpdateComponentVersion(component, version string) error {
	return m.metadataManager.UpdateComponentVersion(component, version)
}

// Settings represents user settings
type Settings struct {
	Theme          string            `json:"theme"`
	AutoUpdate     bool              `json:"auto_update"`
	Telemetry      bool              `json:"telemetry"`
	LogLevel       string            `json:"log_level"`
	BackupOnUpdate bool              `json:"backup_on_update"`
	CustomSettings map[string]string `json:"custom_settings,omitempty"`
}

// LoadSettings loads user settings
func (m *SettingsManager) LoadSettings() (*Settings, error) {
	settingsPath := filepath.Join(m.installDir, ".claude", "settings.json")

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return defaults
			return &Settings{
				Theme:          "dark",
				AutoUpdate:     false,
				Telemetry:      false,
				LogLevel:       "info",
				BackupOnUpdate: true,
			}, nil
		}
		return nil, err
	}

	var settings Settings
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, err
	}

	return &settings, nil
}

// SaveSettings saves user settings
func (m *SettingsManager) SaveSettings(settings *Settings) error {
	settingsPath := filepath.Join(m.installDir, ".claude", "settings.json")

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(settingsPath, data, 0644)
}

// BackupMetadata represents backup metadata
type BackupMetadata struct {
	Version      string            `json:"version"`
	CreatedAt    string            `json:"created_at"`
	Components   map[string]string `json:"components"`
	InstallDir   string            `json:"install_dir"`
	BackupReason string            `json:"backup_reason"`
}

// SaveBackupMetadata saves backup metadata
func (m *SettingsManager) SaveBackupMetadata(backupDir string, metadata *BackupMetadata) error {
	metaPath := filepath.Join(backupDir, "backup.json")

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(metaPath, data, 0644)
}

// LoadBackupMetadata loads backup metadata
func (m *SettingsManager) LoadBackupMetadata(backupDir string) (*BackupMetadata, error) {
	metaPath := filepath.Join(backupDir, "backup.json")

	data, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, err
	}

	var metadata BackupMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, err
	}

	return &metadata, nil
}

// Metadata integration methods

// GetMetadataManager returns the underlying metadata manager
func (m *SettingsManager) GetMetadataManager() *metadata.MetadataManager {
	return m.metadataManager
}

// Legacy method stubs for component compatibility
func (m *SettingsManager) GetComponentVersionFromMetadata(componentName string) (string, error) {
	meta, err := m.metadataManager.LoadMetadata()
	if err != nil {
		return "", err
	}
	if comp, exists := meta.Components[componentName]; exists {
		return comp.Version, nil
	}
	return "", nil
}

func (m *SettingsManager) RemoveComponentRegistration(componentName string) (bool, error) {
	// Simplified implementation - just return success
	return true, nil
}

func (m *SettingsManager) LoadMetadata() (map[string]interface{}, error) {
	meta, err := m.metadataManager.LoadMetadata()
	if err != nil {
		return nil, err
	}

	// Convert to generic map
	result := make(map[string]interface{})
	result["framework"] = meta.Framework
	result["components"] = meta.Components
	result["documents"] = meta.Documents
	result["features"] = meta.Features
	result["installation"] = meta.Installation

	return result, nil
}

func (m *SettingsManager) UpdateMetadata(modifications map[string]interface{}) error {
	// Simplified implementation - just return success
	return nil
}

func (m *SettingsManager) AddComponentRegistration(componentName string, componentInfo map[string]interface{}) error {
	// Simplified implementation - just return success
	return nil
}

func (m *SettingsManager) SaveMetadata(metadata map[string]interface{}) error {
	// Simplified implementation - just return success
	return nil
}

func (m *SettingsManager) IsComponentRegistered(componentName string) (bool, error) {
	meta, err := m.metadataManager.LoadMetadata()
	if err != nil {
		return false, err
	}

	_, exists := meta.Components[componentName]
	return exists, nil
}

// Legacy InstallationInfo struct for compatibility
type InstallationInfo struct {
	Version          string            `json:"version"`
	InstalledAt      string            `json:"installed_at"`
	LastUpdated      string            `json:"last_updated"`
	Components       map[string]string `json:"components"`
	InstallDir       string            `json:"install_dir"`
	InstallerVersion string            `json:"installer_version"`
}

// SaveInstallationInfo - legacy method, now just updates unified metadata
func (m *SettingsManager) SaveInstallationInfo(info *InstallationInfo) error {
	// Convert to unified metadata and save
	meta, err := m.metadataManager.LoadMetadata()
	if err != nil {
		// Create new metadata if none exists
		meta = &metadata.UnifiedMetadata{
			Framework: metadata.FrameworkMetadata{
				Version:     info.Version,
				ReleaseDate: time.Now().Format("2006-01-02"),
				UpdatedAt:   time.Now(),
			},
			Components: make(map[string]metadata.ComponentMeta),
			Documents:  make(map[string]metadata.DocumentMeta),
			Features:   make(map[string]metadata.FeatureMeta),
			Installation: metadata.InstallationMeta{
				InstallDir:       info.InstallDir,
				InstalledAt:      time.Now(),
				LastUpdated:      time.Now(),
				InstallerVersion: info.InstallerVersion,
			},
			Integrity: metadata.IntegrityMeta{
				FileHashes:     make(map[string]metadata.FileIntegrityMeta),
				LastScan:       time.Now(),
				TotalFiles:     0,
				CleanFiles:     0,
				ModifiedFiles:  0,
				MissingFiles:   0,
				CorruptedFiles: 0,
				Status:         "clean",
			},
		}
	} else {
		// Preserve existing integrity information
		if meta.Integrity.FileHashes == nil {
			meta.Integrity.FileHashes = make(map[string]metadata.FileIntegrityMeta)
		}
	}

	// Update components
	for name, version := range info.Components {
		comp := meta.Components[name]
		comp.Version = version
		comp.UpdatedAt = time.Now()
		comp.Status = "installed"
		meta.Components[name] = comp
	}

	// Preserve existing integrity information
	existingMeta, err := m.metadataManager.LoadMetadata()
	if err == nil && existingMeta.Integrity.FileHashes != nil {
		meta.Integrity = existingMeta.Integrity
	}

	return m.metadataManager.SaveMetadata(meta)
}
