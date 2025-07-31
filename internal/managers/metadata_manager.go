package managers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// MetadataManager handles complex metadata operations with dual file system support
// Implements the Python settings_manager.py functionality for Go
type MetadataManager struct {
	installDir   string
	settingsFile string
	metadataFile string
	backupDir    string
}

// NewMetadataManager creates a new metadata manager
func NewMetadataManager(installDir string) *MetadataManager {
	return &MetadataManager{
		installDir:   installDir,
		settingsFile: filepath.Join(installDir, "settings.json"),
		metadataFile: filepath.Join(installDir, ".crew", "config", "crew-metadata.json"),
		backupDir:    filepath.Join(installDir, ".crew", "backups", "settings"),
	}
}

// ComponentInfo represents component metadata in the registry
type ComponentInfo struct {
	Version     string                 `json:"version"`
	Category    string                 `json:"category,omitempty"`
	InstalledAt string                 `json:"installed_at"`
	Extra       map[string]interface{} `json:",inline"` // For additional fields
}

// FrameworkInfo represents framework metadata
type FrameworkInfo struct {
	Version   string `json:"version"`
	UpdatedAt string `json:"updated_at"`
}

// CrewMetadata represents the complete metadata structure
type CrewMetadata struct {
	Components map[string]ComponentInfo `json:"components,omitempty"`
	Framework  *FrameworkInfo           `json:"framework,omitempty"`
	MCP        map[string]interface{}   `json:"mcp,omitempty"`
	Extra      map[string]interface{}   `json:",inline"` // For additional top-level fields
}

// LoadSettings loads settings from settings.json
func (m *MetadataManager) LoadSettings() (map[string]interface{}, error) {
	if _, err := os.Stat(m.settingsFile); os.IsNotExist(err) {
		return make(map[string]interface{}), nil
	}

	data, err := os.ReadFile(m.settingsFile)
	if err != nil {
		return nil, fmt.Errorf("could not load settings from %s: %w", m.settingsFile, err)
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("could not parse settings from %s: %w", m.settingsFile, err)
	}

	return settings, nil
}

// SaveSettings saves settings to settings.json with optional backup
func (m *MetadataManager) SaveSettings(settings map[string]interface{}, createBackup bool) error {
	// Create backup if requested and file exists
	if createBackup {
		if _, err := os.Stat(m.settingsFile); err == nil {
			if err := m.createSettingsBackup(); err != nil {
				return fmt.Errorf("failed to create backup: %w", err)
			}
		}
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(m.settingsFile), 0755); err != nil {
		return err
	}

	// Save with pretty formatting
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(m.settingsFile, data, 0644); err != nil {
		return fmt.Errorf("could not save settings to %s: %w", m.settingsFile, err)
	}

	return nil
}

// LoadMetadata loads Claude Code Super Crew metadata from .crew-metadata.json
func (m *MetadataManager) LoadMetadata() (map[string]interface{}, error) {
	if _, err := os.Stat(m.metadataFile); os.IsNotExist(err) {
		return make(map[string]interface{}), nil
	}

	data, err := os.ReadFile(m.metadataFile)
	if err != nil {
		return nil, fmt.Errorf("could not load metadata from %s: %w", m.metadataFile, err)
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("could not parse metadata from %s: %w", m.metadataFile, err)
	}

	return metadata, nil
}

// SaveMetadata saves Claude Code Super Crew metadata to .crew-metadata.json
func (m *MetadataManager) SaveMetadata(metadata map[string]interface{}) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(m.metadataFile), 0755); err != nil {
		return err
	}

	// Save with pretty formatting
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(m.metadataFile, data, 0644); err != nil {
		return fmt.Errorf("could not save metadata to %s: %w", m.metadataFile, err)
	}

	return nil
}

// MergeMetadata performs deep merge of modifications into existing metadata
func (m *MetadataManager) MergeMetadata(modifications map[string]interface{}) (map[string]interface{}, error) {
	existing, err := m.LoadMetadata()
	if err != nil {
		return nil, err
	}

	return m.deepMerge(existing, modifications), nil
}

// UpdateMetadata updates metadata with modifications using deep merge
func (m *MetadataManager) UpdateMetadata(modifications map[string]interface{}) error {
	merged, err := m.MergeMetadata(modifications)
	if err != nil {
		return err
	}

	return m.SaveMetadata(merged)
}

// MigrateMetadata migrates crew-specific data from settings.json to metadata file
func (m *MetadataManager) MigrateMetadata() (bool, error) {
	settings, err := m.LoadSettings()
	if err != nil {
		return false, err
	}

	// Crew-specific fields to migrate
	crewFields := []string{"components", "framework", "crew", "mcp"}
	dataToMigrate := make(map[string]interface{})
	fieldsFound := false

	// Extract crew data
	for _, field := range crewFields {
		if value, exists := settings[field]; exists {
			dataToMigrate[field] = value
			fieldsFound = true
		}
	}

	if !fieldsFound {
		return false, nil
	}

	// Load existing metadata and merge
	existingMetadata, err := m.LoadMetadata()
	if err != nil {
		return false, err
	}

	mergedMetadata := m.deepMerge(existingMetadata, dataToMigrate)

	// Save to metadata file
	if err := m.SaveMetadata(mergedMetadata); err != nil {
		return false, err
	}

	// Remove crew fields from settings
	cleanSettings := make(map[string]interface{})
	for k, v := range settings {
		found := false
		for _, field := range crewFields {
			if k == field {
				found = true
				break
			}
		}
		if !found {
			cleanSettings[k] = v
		}
	}

	// Save cleaned settings
	if err := m.SaveSettings(cleanSettings, true); err != nil {
		return false, err
	}

	return true, nil
}

// MergeSettings performs deep merge of modifications into existing settings
func (m *MetadataManager) MergeSettings(modifications map[string]interface{}) (map[string]interface{}, error) {
	existing, err := m.LoadSettings()
	if err != nil {
		return nil, err
	}

	return m.deepMerge(existing, modifications), nil
}

// UpdateSettings updates settings with modifications using deep merge
func (m *MetadataManager) UpdateSettings(modifications map[string]interface{}, createBackup bool) error {
	merged, err := m.MergeSettings(modifications)
	if err != nil {
		return err
	}

	return m.SaveSettings(merged, createBackup)
}

// GetSetting gets setting value using dot-notation path
func (m *MetadataManager) GetSetting(keyPath string, defaultValue interface{}) (interface{}, error) {
	settings, err := m.LoadSettings()
	if err != nil {
		return defaultValue, err
	}

	return m.getNestedValue(settings, keyPath, defaultValue), nil
}

// SetSetting sets setting value using dot-notation path
func (m *MetadataManager) SetSetting(keyPath string, value interface{}, createBackup bool) error {
	// Build nested dict structure
	keys := strings.Split(keyPath, ".")
	modification := make(map[string]interface{})
	current := modification

	for i, key := range keys[:len(keys)-1] {
		current[key] = make(map[string]interface{})
		if i < len(keys)-2 {
			current = current[key].(map[string]interface{})
		}
	}

	// Set the final value
	if len(keys) == 1 {
		modification[keys[0]] = value
	} else {
		finalKey := keys[len(keys)-1]
		parent := current[keys[len(keys)-2]].(map[string]interface{})
		parent[finalKey] = value
	}

	return m.UpdateSettings(modification, createBackup)
}

// RemoveSetting removes setting using dot-notation path
func (m *MetadataManager) RemoveSetting(keyPath string, createBackup bool) (bool, error) {
	settings, err := m.LoadSettings()
	if err != nil {
		return false, err
	}

	keys := strings.Split(keyPath, ".")

	// Navigate to parent of target key
	current := settings
	for _, key := range keys[:len(keys)-1] {
		if next, exists := current[key]; exists {
			if nextMap, ok := next.(map[string]interface{}); ok {
				current = nextMap
			} else {
				return false, nil // Path doesn't exist
			}
		} else {
			return false, nil // Path doesn't exist
		}
	}

	// Remove the target key
	finalKey := keys[len(keys)-1]
	if _, exists := current[finalKey]; exists {
		delete(current, finalKey)
		if err := m.SaveSettings(settings, createBackup); err != nil {
			return false, err
		}
		return true, nil
	}

	return false, nil
}

// AddComponentRegistration adds component to registry in metadata
func (m *MetadataManager) AddComponentRegistration(componentName string, componentInfo map[string]interface{}) error {
	metadata, err := m.LoadMetadata()
	if err != nil {
		return err
	}

	// Ensure components map exists
	if metadata["components"] == nil {
		metadata["components"] = make(map[string]interface{})
	}

	components := metadata["components"].(map[string]interface{})

	// Add installed_at timestamp
	info := make(map[string]interface{})
	for k, v := range componentInfo {
		info[k] = v
	}
	info["installed_at"] = time.Now().Format(time.RFC3339)

	components[componentName] = info

	return m.SaveMetadata(metadata)
}

// RemoveComponentRegistration removes component from registry in metadata
func (m *MetadataManager) RemoveComponentRegistration(componentName string) (bool, error) {
	metadata, err := m.LoadMetadata()
	if err != nil {
		return false, err
	}

	if components, exists := metadata["components"]; exists {
		if componentsMap, ok := components.(map[string]interface{}); ok {
			if _, exists := componentsMap[componentName]; exists {
				delete(componentsMap, componentName)
				if err := m.SaveMetadata(metadata); err != nil {
					return false, err
				}
				return true, nil
			}
		}
	}

	return false, nil
}

// GetInstalledComponents gets all installed components from registry
func (m *MetadataManager) GetInstalledComponents() (map[string]map[string]interface{}, error) {
	metadata, err := m.LoadMetadata()
	if err != nil {
		return nil, err
	}

	result := make(map[string]map[string]interface{})

	if components, exists := metadata["components"]; exists {
		if componentsMap, ok := components.(map[string]interface{}); ok {
			for name, info := range componentsMap {
				if infoMap, ok := info.(map[string]interface{}); ok {
					result[name] = infoMap
				}
			}
		}
	}

	return result, nil
}

// IsComponentInstalled checks if component is registered as installed
func (m *MetadataManager) IsComponentInstalled(componentName string) (bool, error) {
	components, err := m.GetInstalledComponents()
	if err != nil {
		return false, err
	}

	_, exists := components[componentName]
	return exists, nil
}

// GetComponentVersion gets installed version of component
func (m *MetadataManager) GetComponentVersion(componentName string) (string, error) {
	components, err := m.GetInstalledComponents()
	if err != nil {
		return "", err
	}

	if componentInfo, exists := components[componentName]; exists {
		if version, exists := componentInfo["version"]; exists {
			if versionStr, ok := version.(string); ok {
				return versionStr, nil
			}
		}
	}

	return "", nil
}

// UpdateFrameworkVersion updates Claude Code Super Crew framework version in metadata
func (m *MetadataManager) UpdateFrameworkVersion(version string) error {
	metadata, err := m.LoadMetadata()
	if err != nil {
		return err
	}

	if metadata["framework"] == nil {
		metadata["framework"] = make(map[string]interface{})
	}

	framework := metadata["framework"].(map[string]interface{})
	framework["version"] = version
	framework["updated_at"] = time.Now().Format(time.RFC3339)

	return m.SaveMetadata(metadata)
}

// CheckInstallationExists checks if .crew-metadata.json exists
func (m *MetadataManager) CheckInstallationExists() bool {
	_, err := os.Stat(m.metadataFile)
	return err == nil
}

// CheckV2InstallationExists checks if settings.json exists (v2 format)
func (m *MetadataManager) CheckV2InstallationExists() bool {
	_, err := os.Stat(m.settingsFile)
	return err == nil
}

// GetMetadataSetting gets metadata value using dot-notation path
func (m *MetadataManager) GetMetadataSetting(keyPath string, defaultValue interface{}) (interface{}, error) {
	metadata, err := m.LoadMetadata()
	if err != nil {
		return defaultValue, err
	}

	return m.getNestedValue(metadata, keyPath, defaultValue), nil
}

// deepMerge performs deep merge of two maps
func (m *MetadataManager) deepMerge(base, overlay map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// Copy base
	for k, v := range base {
		if vMap, ok := v.(map[string]interface{}); ok {
			result[k] = m.copyMap(vMap)
		} else {
			result[k] = v
		}
	}

	// Merge overlay
	for k, v := range overlay {
		if existing, exists := result[k]; exists {
			if existingMap, ok := existing.(map[string]interface{}); ok {
				if vMap, ok := v.(map[string]interface{}); ok {
					result[k] = m.deepMerge(existingMap, vMap)
					continue
				}
			}
		}

		// Either key doesn't exist in base or values aren't both maps
		if vMap, ok := v.(map[string]interface{}); ok {
			result[k] = m.copyMap(vMap)
		} else {
			result[k] = v
		}
	}

	return result
}

// copyMap creates a deep copy of a map
func (m *MetadataManager) copyMap(original map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range original {
		if vMap, ok := v.(map[string]interface{}); ok {
			result[k] = m.copyMap(vMap)
		} else {
			result[k] = v
		}
	}
	return result
}

// getNestedValue gets value from nested map using dot-notation path
func (m *MetadataManager) getNestedValue(data map[string]interface{}, keyPath string, defaultValue interface{}) interface{} {
	keys := strings.Split(keyPath, ".")
	current := data

	for _, key := range keys {
		if value, exists := current[key]; exists {
			if nextMap, ok := value.(map[string]interface{}); ok {
				current = nextMap
			} else {
				// This is the final value
				return value
			}
		} else {
			return defaultValue
		}
	}

	return current
}

// createSettingsBackup creates timestamped backup of settings.json
func (m *MetadataManager) createSettingsBackup() error {
	if _, err := os.Stat(m.settingsFile); os.IsNotExist(err) {
		return fmt.Errorf("cannot backup non-existent settings file")
	}

	// Create backup directory
	if err := os.MkdirAll(m.backupDir, 0755); err != nil {
		return err
	}

	// Create timestamped backup
	timestamp := time.Now().Format("20060102_150405")
	backupFile := filepath.Join(m.backupDir, fmt.Sprintf("settings_%s.json", timestamp))

	// Copy file
	data, err := os.ReadFile(m.settingsFile)
	if err != nil {
		return err
	}

	if err := os.WriteFile(backupFile, data, 0644); err != nil {
		return err
	}

	// Cleanup old backups (keep only last 10)
	m.cleanupOldBackups(10)

	return nil
}

// cleanupOldBackups removes old backup files, keeping only the most recent
func (m *MetadataManager) cleanupOldBackups(keepCount int) {
	if _, err := os.Stat(m.backupDir); os.IsNotExist(err) {
		return
	}

	files, err := filepath.Glob(filepath.Join(m.backupDir, "settings_*.json"))
	if err != nil {
		return
	}

	if len(files) <= keepCount {
		return
	}

	// Get file info and sort by modification time
	type fileInfo struct {
		path    string
		modTime time.Time
	}

	var fileInfos []fileInfo
	for _, file := range files {
		if stat, err := os.Stat(file); err == nil {
			fileInfos = append(fileInfos, fileInfo{
				path:    file,
				modTime: stat.ModTime(),
			})
		}
	}

	// Sort by modification time (newest first)
	for i := 0; i < len(fileInfos)-1; i++ {
		for j := i + 1; j < len(fileInfos); j++ {
			if fileInfos[i].modTime.Before(fileInfos[j].modTime) {
				fileInfos[i], fileInfos[j] = fileInfos[j], fileInfos[i]
			}
		}
	}

	// Remove old backups
	for i := keepCount; i < len(fileInfos); i++ {
		os.Remove(fileInfos[i].path)
	}
}

// BackupInfo represents backup file information
type BackupInfo struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	Created  string `json:"created"`
	Modified string `json:"modified"`
}

// ListBackups lists available settings backups
func (m *MetadataManager) ListBackups() ([]BackupInfo, error) {
	if _, err := os.Stat(m.backupDir); os.IsNotExist(err) {
		return []BackupInfo{}, nil
	}

	files, err := filepath.Glob(filepath.Join(m.backupDir, "settings_*.json"))
	if err != nil {
		return nil, err
	}

	var backups []BackupInfo
	for _, file := range files {
		if stat, err := os.Stat(file); err == nil {
			backups = append(backups, BackupInfo{
				Name:     filepath.Base(file),
				Path:     file,
				Size:     stat.Size(),
				Created:  stat.ModTime().Format(time.RFC3339),
				Modified: stat.ModTime().Format(time.RFC3339),
			})
		}
	}

	// Sort by creation time, most recent first
	for i := 0; i < len(backups)-1; i++ {
		for j := i + 1; j < len(backups); j++ {
			if backups[i].Created < backups[j].Created {
				backups[i], backups[j] = backups[j], backups[i]
			}
		}
	}

	return backups, nil
}

// RestoreBackup restores settings from backup
func (m *MetadataManager) RestoreBackup(backupName string) error {
	backupFile := filepath.Join(m.backupDir, backupName)

	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
		return fmt.Errorf("backup file not found: %s", backupName)
	}

	// Validate backup file first
	data, err := os.ReadFile(backupFile)
	if err != nil {
		return err
	}

	var testData map[string]interface{}
	if err := json.Unmarshal(data, &testData); err != nil {
		return fmt.Errorf("invalid backup file: %w", err)
	}

	// Create backup of current settings
	if _, err := os.Stat(m.settingsFile); err == nil {
		if err := m.createSettingsBackup(); err != nil {
			return fmt.Errorf("failed to create backup before restore: %w", err)
		}
	}

	// Restore backup
	if err := os.WriteFile(m.settingsFile, data, 0644); err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	return nil
}
