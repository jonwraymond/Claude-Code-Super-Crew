// Package metadata provides unified metadata management for Claude Code Super Crew
package metadata

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// UnifiedMetadata represents the comprehensive metadata for the entire installation
type UnifiedMetadata struct {
	Framework    FrameworkMetadata        `json:"framework"`
	Components   map[string]ComponentMeta `json:"components"`
	Documents    map[string]DocumentMeta  `json:"documents"`
	Features     map[string]FeatureMeta   `json:"features"`
	Installation InstallationMeta         `json:"installation"`
	Inventory    InventoryMeta            `json:"inventory"`
	Integrity    IntegrityMeta            `json:"integrity"`
}

// FrameworkMetadata contains overall framework information
type FrameworkMetadata struct {
	Version         string    `json:"version"`
	ReleaseDate     string    `json:"release_date"`
	UpdatedAt       time.Time `json:"updated_at"`
	PreviousVersion string    `json:"previous_version"`
	BuildHash       string    `json:"build_hash,omitempty"`
}

// ComponentMeta contains detailed component metadata
type ComponentMeta struct {
	Version         string    `json:"version"`
	UpdatedAt       time.Time `json:"updated_at"`
	PreviousVersion string    `json:"previous_version,omitempty"`
	Status          string    `json:"status"` // installed, missing, corrupted, outdated
	Dependencies    []string  `json:"dependencies,omitempty"`
	Size            int64     `json:"size,omitempty"`
	FileCount       int       `json:"file_count,omitempty"`
	Checksum        string    `json:"checksum,omitempty"`
}

// DocumentMeta tracks individual .md file versions
type DocumentMeta struct {
	Version         string    `json:"version"`
	PreviousVersion string    `json:"previous_version,omitempty"`
	UpdatedAt       time.Time `json:"updated_at"`
	Size            int64     `json:"size"`
	Checksum        string    `json:"checksum"`
	Component       string    `json:"component"` // which component owns this document
	Status          string    `json:"status"`    // present, missing, modified
}

// FeatureMeta tracks configurable features and flags
type FeatureMeta struct {
	Enabled     bool      `json:"enabled"`
	Version     string    `json:"version,omitempty"`
	UpdatedAt   time.Time `json:"updated_at"`
	Description string    `json:"description,omitempty"`
	Flags       []string  `json:"flags,omitempty"`
}

// InstallationMeta tracks installation-specific information
type InstallationMeta struct {
	InstallDir       string    `json:"install_dir"`
	InstalledAt      time.Time `json:"installed_at"`
	LastUpdated      time.Time `json:"last_updated"`
	InstallerVersion string    `json:"installer_version"`
	TotalSize        int64     `json:"total_size"`
	TotalFiles       int       `json:"total_files"`
}

// InventoryMeta tracks all files and directories created by crew
type InventoryMeta struct {
	CreatedFiles       []string  `json:"created_files"`       // List of all files created by crew
	CreatedDirectories []string  `json:"created_directories"` // List of all directories created by crew
	LastUpdated        time.Time `json:"last_updated"`
	TotalCreatedFiles  int       `json:"total_created_files"`
	TotalCreatedDirs   int       `json:"total_created_dirs"`
}

// FileIntegrityMeta tracks file integrity with original and current hashes
type FileIntegrityMeta struct {
	OriginalHash    string    `json:"original_hash"`    // Hash when first installed
	CurrentHash     string    `json:"current_hash"`     // Current hash of the file
	LastChecked     time.Time `json:"last_checked"`     // When integrity was last verified
	Status          string    `json:"status"`           // "clean", "modified", "missing", "corrupted"
	Component       string    `json:"component"`        // Which component owns this file
	FilePath        string    `json:"file_path"`        // Relative path to the file
	ModificationLog []string  `json:"modification_log"` // History of detected changes
}

// IntegrityMeta tracks file integrity across the entire installation
type IntegrityMeta struct {
	FileHashes     map[string]FileIntegrityMeta `json:"file_hashes"`     // File path -> integrity info
	LastScan       time.Time                    `json:"last_scan"`       // When integrity was last checked
	TotalFiles     int                          `json:"total_files"`     // Total files being tracked
	CleanFiles     int                          `json:"clean_files"`     // Files with matching hashes
	ModifiedFiles  int                          `json:"modified_files"`  // Files with hash mismatches
	MissingFiles   int                          `json:"missing_files"`   // Files that no longer exist
	CorruptedFiles int                          `json:"corrupted_files"` // Files that can't be read
	Status         string                       `json:"status"`          // Overall integrity status: "clean", "warning", "critical"
}

// MetadataManager handles unified metadata operations
//
// CANONICAL SCHEMA: This manager implements the unified metadata schema as the
// canonical standard for all crew-metadata.json operations. The schema includes:
//
// - Framework metadata with version tracking
// - Component metadata with size, file counts, and status
// - Document-level version tracking with checksums
// - Feature flags and activation management
// - Installation metadata with comprehensive statistics
//
// Version persistence is CRITICAL - document versions must survive refresh operations.
type MetadataManager struct {
	installDir   string
	metadataFile string
}

// NewMetadataManager creates a new metadata manager
func NewMetadataManager(installDir string) *MetadataManager {
	return &MetadataManager{
		installDir:   installDir,
		metadataFile: filepath.Join(installDir, ".crew", "config", "crew-metadata.json"),
	}
}

// LoadMetadata loads the unified metadata from disk
func (m *MetadataManager) LoadMetadata() (*UnifiedMetadata, error) {
	if _, err := os.Stat(m.metadataFile); os.IsNotExist(err) {
		return m.createEmptyMetadata(), nil
	}

	data, err := os.ReadFile(m.metadataFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}

	var metadata UnifiedMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	return &metadata, nil
}

// SaveMetadata saves the unified metadata to disk
func (m *MetadataManager) SaveMetadata(metadata *UnifiedMetadata) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(m.metadataFile), 0755); err != nil {
		return fmt.Errorf("failed to create metadata directory: %w", err)
	}

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(m.metadataFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	return nil
}

// RefreshMetadata comprehensively scans the installation and updates metadata
func (m *MetadataManager) RefreshMetadata() (*UnifiedMetadata, error) {
	metadata, err := m.LoadMetadata()
	if err != nil {
		return nil, err
	}

	// Update installation info
	metadata.Installation.LastUpdated = time.Now()
	metadata.Installation.InstallDir = m.installDir

	// Scan components
	if err := m.scanComponents(metadata); err != nil {
		return nil, fmt.Errorf("failed to scan components: %w", err)
	}

	// Scan documents
	if err := m.scanDocuments(metadata); err != nil {
		return nil, fmt.Errorf("failed to scan documents: %w", err)
	}

	// Update totals
	m.updateTotals(metadata)

	// Save updated metadata
	if err := m.SaveMetadata(metadata); err != nil {
		return nil, err
	}

	return metadata, nil
}

// scanComponents scans for installed components and updates their metadata
func (m *MetadataManager) scanComponents(metadata *UnifiedMetadata) error {
	if metadata.Components == nil {
		metadata.Components = make(map[string]ComponentMeta)
	}

	// Define expected components and their directories
	componentPaths := map[string]string{
		"core":     m.installDir,
		"commands": filepath.Join(m.installDir, "commands"),
		"agents":   filepath.Join(m.installDir, "agents"),
		"hooks":    filepath.Join(m.installDir, "hooks"),
		"mcp":      filepath.Join(m.installDir, ".crew", "mcp"),
	}

	for component, path := range componentPaths {
		meta := metadata.Components[component]
		meta.UpdatedAt = time.Now()

		if _, err := os.Stat(path); os.IsNotExist(err) {
			meta.Status = "missing"
			meta.Size = 0
			meta.FileCount = 0
		} else {
			meta.Status = "installed"
			size, count, err := m.calculateDirectorySize(path)
			if err != nil {
				meta.Status = "corrupted"
			} else {
				meta.Size = size
				meta.FileCount = count
			}
		}

		metadata.Components[component] = meta
	}

	return nil
}

// scanDocuments scans for .md files and tracks their versions
func (m *MetadataManager) scanDocuments(metadata *UnifiedMetadata) error {
	if metadata.Documents == nil {
		metadata.Documents = make(map[string]DocumentMeta)
	}

	// Define core framework documents
	coreDocuments := []string{
		"CLAUDE.md", "COMMANDS.md", "FLAGS.md", "MCP.md",
		"MODES.md", "ORCHESTRATOR.md", "PERSONAS.md",
		"PRINCIPLES.md", "RULES.md",
	}

	// Scan core documents
	for _, doc := range coreDocuments {
		docPath := filepath.Join(m.installDir, doc)
		var existingMeta *DocumentMeta
		if existing, exists := metadata.Documents[doc]; exists {
			existingMeta = &existing
		}
		meta := m.scanSingleDocument(docPath, "core", existingMeta)
		metadata.Documents[doc] = meta
	}

	// Scan agent documents
	agentsDir := filepath.Join(m.installDir, "agents")
	if _, err := os.Stat(agentsDir); err == nil {
		filepath.WalkDir(agentsDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(path, ".md") {
				return err
			}

			relPath, _ := filepath.Rel(m.installDir, path)
			var existingMeta *DocumentMeta
			if existing, exists := metadata.Documents[relPath]; exists {
				existingMeta = &existing
			}
			meta := m.scanSingleDocument(path, "agents", existingMeta)
			metadata.Documents[relPath] = meta
			return nil
		})
	}

	// Scan hooks documents
	hooksDir := filepath.Join(m.installDir, "hooks")
	if _, err := os.Stat(hooksDir); err == nil {
		filepath.WalkDir(hooksDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(path, ".md") {
				return err
			}

			relPath, _ := filepath.Rel(m.installDir, path)
			var existingMeta *DocumentMeta
			if existing, exists := metadata.Documents[relPath]; exists {
				existingMeta = &existing
			}
			meta := m.scanSingleDocument(path, "hooks", existingMeta)
			metadata.Documents[relPath] = meta
			return nil
		})
	}

	return nil
}

// scanSingleDocument scans a single document and returns its metadata
func (m *MetadataManager) scanSingleDocument(path, component string, existingMeta *DocumentMeta) DocumentMeta {
	meta := DocumentMeta{
		Component: component,
		UpdatedAt: time.Now(),
	}

	// Preserve existing version information if available - CRITICAL for version persistence
	if existingMeta != nil {
		meta.Version = existingMeta.Version
		meta.PreviousVersion = existingMeta.PreviousVersion
		// Only update timestamp if content changed (preserve version info)
		if existingMeta.Version != "" {
			meta.UpdatedAt = existingMeta.UpdatedAt
		}
	} else {
		meta.Version = "1.0.0" // Default version for new documents only
	}

	if stat, err := os.Stat(path); err == nil {
		meta.Status = "present"
		meta.Size = stat.Size()

		// Calculate checksum
		if checksum, err := m.calculateFileChecksum(path); err == nil {
			meta.Checksum = checksum
		}
	} else {
		meta.Status = "missing"
		meta.Size = 0
	}

	return meta
}

// calculateDirectorySize calculates the total size and file count of a directory
func (m *MetadataManager) calculateDirectorySize(dirPath string) (int64, int, error) {
	var totalSize int64
	var fileCount int

	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			if info, err := d.Info(); err == nil {
				totalSize += info.Size()
				fileCount++
			}
		}
		return nil
	})

	return totalSize, fileCount, err
}

// calculateFileChecksum calculates a simple checksum for a file
func (m *MetadataManager) calculateFileChecksum(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	// Calculate SHA-256 hash
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

// updateTotals updates the total size and file count in installation metadata
func (m *MetadataManager) updateTotals(metadata *UnifiedMetadata) {
	var totalSize int64
	var totalFiles int

	for _, comp := range metadata.Components {
		totalSize += comp.Size
		totalFiles += comp.FileCount
	}

	metadata.Installation.TotalSize = totalSize
	metadata.Installation.TotalFiles = totalFiles
}

// createEmptyMetadata creates a new empty metadata structure
func (m *MetadataManager) createEmptyMetadata() *UnifiedMetadata {
	now := time.Now()
	return &UnifiedMetadata{
		Framework: FrameworkMetadata{
			Version:     "1.0.0",
			ReleaseDate: now.Format("2006-01-02"),
			UpdatedAt:   now,
		},
		Components: make(map[string]ComponentMeta),
		Documents:  make(map[string]DocumentMeta),
		Features:   make(map[string]FeatureMeta),
		Installation: InstallationMeta{
			InstallDir:       m.installDir,
			InstalledAt:      now,
			LastUpdated:      now,
			InstallerVersion: "1.0.0",
		},
		Inventory: InventoryMeta{
			CreatedFiles:       []string{},
			CreatedDirectories: []string{},
			LastUpdated:        now,
			TotalCreatedFiles:  0,
			TotalCreatedDirs:   0,
		},
		Integrity: IntegrityMeta{
			FileHashes:     make(map[string]FileIntegrityMeta),
			LastScan:       now,
			TotalFiles:     0,
			CleanFiles:     0,
			ModifiedFiles:  0,
			MissingFiles:   0,
			CorruptedFiles: 0,
			Status:         "clean",
		},
	}
}

// GetComponentStatus returns detailed status for a specific component
func (m *MetadataManager) GetComponentStatus(componentName string) (*ComponentMeta, error) {
	metadata, err := m.LoadMetadata()
	if err != nil {
		return nil, err
	}

	if comp, exists := metadata.Components[componentName]; exists {
		return &comp, nil
	}

	return nil, fmt.Errorf("component %s not found", componentName)
}

// UpdateComponentVersion updates the version information for a component
func (m *MetadataManager) UpdateComponentVersion(componentName, version string) error {
	metadata, err := m.LoadMetadata()
	if err != nil {
		return err
	}

	comp := metadata.Components[componentName]
	comp.PreviousVersion = comp.Version
	comp.Version = version
	comp.UpdatedAt = time.Now()
	metadata.Components[componentName] = comp

	return m.SaveMetadata(metadata)
}

// SetFeatureFlag sets a feature flag value
func (m *MetadataManager) SetFeatureFlag(featureName string, enabled bool, description string) error {
	metadata, err := m.LoadMetadata()
	if err != nil {
		return err
	}

	if metadata.Features == nil {
		metadata.Features = make(map[string]FeatureMeta)
	}

	metadata.Features[featureName] = FeatureMeta{
		Enabled:     enabled,
		UpdatedAt:   time.Now(),
		Description: description,
	}

	return m.SaveMetadata(metadata)
}

// CheckInstallationExists checks if the installation exists by looking for metadata
func (m *MetadataManager) CheckInstallationExists() bool {
	_, err := os.Stat(m.metadataFile)
	return err == nil
}

// AddToInventory adds a file or directory to the creation inventory
func (m *MetadataManager) AddToInventory(path string, isDirectory bool) error {
	metadata, err := m.LoadMetadata()
	if err != nil {
		return err
	}

	// Convert to relative path from install directory
	relPath, err := filepath.Rel(m.installDir, path)
	if err != nil {
		relPath = path // Use absolute path if relative conversion fails
	}

	if isDirectory {
		// Check if directory is already in inventory
		for _, existingDir := range metadata.Inventory.CreatedDirectories {
			if existingDir == relPath {
				return nil // Already tracked
			}
		}
		metadata.Inventory.CreatedDirectories = append(metadata.Inventory.CreatedDirectories, relPath)
		metadata.Inventory.TotalCreatedDirs = len(metadata.Inventory.CreatedDirectories)
	} else {
		// Check if file is already in inventory
		for _, existingFile := range metadata.Inventory.CreatedFiles {
			if existingFile == relPath {
				return nil // Already tracked
			}
		}
		metadata.Inventory.CreatedFiles = append(metadata.Inventory.CreatedFiles, relPath)
		metadata.Inventory.TotalCreatedFiles = len(metadata.Inventory.CreatedFiles)
	}

	metadata.Inventory.LastUpdated = time.Now()
	return m.SaveMetadata(metadata)
}

// RemoveFromInventory removes a file or directory from the creation inventory
func (m *MetadataManager) RemoveFromInventory(path string, isDirectory bool) error {
	metadata, err := m.LoadMetadata()
	if err != nil {
		return err
	}

	// Convert to relative path from install directory
	relPath, err := filepath.Rel(m.installDir, path)
	if err != nil {
		relPath = path // Use absolute path if relative conversion fails
	}

	if isDirectory {
		// Remove directory from inventory
		for i, existingDir := range metadata.Inventory.CreatedDirectories {
			if existingDir == relPath {
				metadata.Inventory.CreatedDirectories = append(
					metadata.Inventory.CreatedDirectories[:i],
					metadata.Inventory.CreatedDirectories[i+1:]...)
				break
			}
		}
		metadata.Inventory.TotalCreatedDirs = len(metadata.Inventory.CreatedDirectories)
	} else {
		// Remove file from inventory
		for i, existingFile := range metadata.Inventory.CreatedFiles {
			if existingFile == relPath {
				metadata.Inventory.CreatedFiles = append(
					metadata.Inventory.CreatedFiles[:i],
					metadata.Inventory.CreatedFiles[i+1:]...)
				break
			}
		}
		metadata.Inventory.TotalCreatedFiles = len(metadata.Inventory.CreatedFiles)
	}

	metadata.Inventory.LastUpdated = time.Now()
	return m.SaveMetadata(metadata)
}

// GetInventory returns the current inventory of created files and directories
func (m *MetadataManager) GetInventory() (*InventoryMeta, error) {
	metadata, err := m.LoadMetadata()
	if err != nil {
		return nil, err
	}
	return &metadata.Inventory, nil
}

// CheckFileIntegrity verifies the integrity of all tracked files
func (m *MetadataManager) CheckFileIntegrity() (*IntegrityMeta, error) {
	metadata, err := m.LoadMetadata()
	if err != nil {
		return nil, err
	}

	// Initialize integrity metadata if not present
	if metadata.Integrity.FileHashes == nil {
		metadata.Integrity.FileHashes = make(map[string]FileIntegrityMeta)
	}

	// Check each tracked file
	cleanCount := 0
	modifiedCount := 0
	missingCount := 0
	corruptedCount := 0

	for filePath, integrity := range metadata.Integrity.FileHashes {
		currentStatus := m.checkSingleFileIntegrity(filePath, &integrity)
		metadata.Integrity.FileHashes[filePath] = integrity

		switch currentStatus {
		case "clean":
			cleanCount++
		case "modified":
			modifiedCount++
		case "missing":
			missingCount++
		case "corrupted":
			corruptedCount++
		}
	}

	// Update integrity summary
	metadata.Integrity.LastScan = time.Now()
	metadata.Integrity.TotalFiles = len(metadata.Integrity.FileHashes)
	metadata.Integrity.CleanFiles = cleanCount
	metadata.Integrity.ModifiedFiles = modifiedCount
	metadata.Integrity.MissingFiles = missingCount
	metadata.Integrity.CorruptedFiles = corruptedCount

	// Determine overall status
	if modifiedCount == 0 && missingCount == 0 && corruptedCount == 0 {
		metadata.Integrity.Status = "clean"
	} else if modifiedCount > 0 || missingCount > 0 {
		metadata.Integrity.Status = "warning"
	} else {
		metadata.Integrity.Status = "critical"
	}

	// Save updated metadata
	if err := m.SaveMetadata(metadata); err != nil {
		return nil, err
	}

	return &metadata.Integrity, nil
}

// checkSingleFileIntegrity checks the integrity of a single file
func (m *MetadataManager) checkSingleFileIntegrity(filePath string, integrity *FileIntegrityMeta) string {
	fullPath := filepath.Join(m.installDir, filePath)

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		integrity.Status = "missing"
		integrity.LastChecked = time.Now()
		integrity.ModificationLog = append(integrity.ModificationLog,
			fmt.Sprintf("%s: File not found", time.Now().Format("2006-01-02 15:04:05")))
		return "missing"
	}

	// Calculate current hash
	currentHash, err := m.calculateFileChecksum(fullPath)
	if err != nil {
		integrity.Status = "corrupted"
		integrity.LastChecked = time.Now()
		integrity.ModificationLog = append(integrity.ModificationLog,
			fmt.Sprintf("%s: Cannot read file - %v", time.Now().Format("2006-01-02 15:04:05"), err))
		return "corrupted"
	}

	// Update current hash and check against original
	integrity.CurrentHash = currentHash
	integrity.LastChecked = time.Now()

	if currentHash == integrity.OriginalHash {
		integrity.Status = "clean"
		return "clean"
	} else {
		integrity.Status = "modified"
		// Safely display hash with length check
		origHash := integrity.OriginalHash
		if len(origHash) > 8 {
			origHash = origHash[:8]
		}
		currHash := currentHash
		if len(currHash) > 8 {
			currHash = currHash[:8]
		}

		integrity.ModificationLog = append(integrity.ModificationLog,
			fmt.Sprintf("%s: Hash mismatch - Original: %s, Current: %s",
				time.Now().Format("2006-01-02 15:04:05"),
				origHash, currHash))
		return "modified"
	}
}

// AddFileToIntegrityTracking adds a file to integrity tracking
func (m *MetadataManager) AddFileToIntegrityTracking(filePath, component string) error {
	metadata, err := m.LoadMetadata()
	if err != nil {
		return err
	}

	// Initialize integrity metadata if not present
	if metadata.Integrity.FileHashes == nil {
		metadata.Integrity.FileHashes = make(map[string]FileIntegrityMeta)
	}

	// Calculate original hash
	fullPath := filepath.Join(m.installDir, filePath)
	originalHash, err := m.calculateFileChecksum(fullPath)
	if err != nil {
		return fmt.Errorf("failed to calculate hash for %s: %w", filePath, err)
	}

	// Create integrity record
	integrity := FileIntegrityMeta{
		OriginalHash:    originalHash,
		CurrentHash:     originalHash,
		LastChecked:     time.Now(),
		Status:          "clean",
		Component:       component,
		FilePath:        filePath,
		ModificationLog: []string{fmt.Sprintf("%s: File added to integrity tracking", time.Now().Format("2006-01-02 15:04:05"))},
	}

	metadata.Integrity.FileHashes[filePath] = integrity

	return m.SaveMetadata(metadata)
}

// RemoveFileFromIntegrityTracking removes a file from integrity tracking
func (m *MetadataManager) RemoveFileFromIntegrityTracking(filePath string) error {
	metadata, err := m.LoadMetadata()
	if err != nil {
		return err
	}

	if metadata.Integrity.FileHashes != nil {
		delete(metadata.Integrity.FileHashes, filePath)
	}

	return m.SaveMetadata(metadata)
}

// GetIntegrityStatus returns the current integrity status with visual indicators
func (m *MetadataManager) GetIntegrityStatus() (*IntegrityMeta, error) {
	metadata, err := m.LoadMetadata()
	if err != nil {
		return nil, err
	}
	return &metadata.Integrity, nil
}
