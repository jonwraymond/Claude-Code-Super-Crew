package backup

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jonwraymond/claude-code-super-crew/internal/managers"
	"github.com/jonwraymond/claude-code-super-crew/internal/versioning"
	"github.com/jonwraymond/claude-code-super-crew/pkg/logger"
)

// Options holds backup configuration
type Options struct {
	InstallDir    string
	BackupDir     string
	BackupName    string
	Compress      string // none, gzip, bzip2
	Verbose       bool
	DryRun        bool
	Overwrite     bool
	IncludeConfig bool
	IncludeLogs   bool
	Description   string
}

// BackupMetadata represents backup metadata
type BackupMetadata struct {
	BackupVersion    string            `json:"backup_version"`
	Created          string            `json:"created"`
	Timestamp        time.Time         `json:"timestamp"`
	InstallDir       string            `json:"install_dir"`
	Components       map[string]string `json:"components"`
	FrameworkVersion string            `json:"framework_version"`
	Size             int64             `json:"size"`
	Checksum         string            `json:"checksum"`
	BackupType       string            `json:"backup_type"`
	Description      string            `json:"description"`
}

// BackupInfo represents information about a backup
type BackupInfo struct {
	Path      string
	Exists    bool
	Size      int64
	Created   time.Time
	Metadata  *BackupMetadata
	FileCount int
	Error     error
}

// Manager handles backup operations
type Manager struct {
	opts        Options
	logger      logger.Logger
	fileManager *managers.FileManager
}

// NewManager creates a new backup manager
func NewManager(opts Options) *Manager {
	return &Manager{
		opts:        opts,
		logger:      logger.GetLogger(),
		fileManager: managers.NewFileManager(),
	}
}

// Create creates a new backup
func (m *Manager) Create() (string, error) {
	// Generate timestamp
	timestamp := time.Now().Format("20060102_150405")
	backupName := fmt.Sprintf("%s_%s", m.opts.BackupName, timestamp)
	
	if m.opts.Description != "" {
		// Sanitize description for filename
		desc := strings.ReplaceAll(m.opts.Description, " ", "_")
		desc = strings.ReplaceAll(desc, "/", "_")
		backupName = fmt.Sprintf("%s_%s_%s", m.opts.BackupName, timestamp, desc)
	}

	// Determine file extension based on compression
	var backupFile string
	var mode string
	switch m.opts.Compress {
	case "gzip":
		backupFile = filepath.Join(m.opts.BackupDir, backupName+".tar.gz")
		mode = "gz"
	case "bzip2":
		backupFile = filepath.Join(m.opts.BackupDir, backupName+".tar.bz2")
		mode = "bz2"
	default:
		backupFile = filepath.Join(m.opts.BackupDir, backupName+".tar")
		mode = "none"
	}

	if m.opts.Verbose {
		m.logger.Infof("Creating backup: %s", backupFile)
	}

	// Ensure backup directory exists
	if err := os.MkdirAll(m.opts.BackupDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Create metadata
	metadata := m.createBackupMetadata()

	// Create backup file
	file, err := os.Create(backupFile)
	if err != nil {
		return "", fmt.Errorf("failed to create backup file: %w", err)
	}
	defer file.Close()

	// Create tar writer with optional compression
	var tarWriter *tar.Writer
	var gzWriter *gzip.Writer
	if mode == "gz" {
		gzWriter = gzip.NewWriter(file)
		defer gzWriter.Close()
		tarWriter = tar.NewWriter(gzWriter)
	} else {
		tarWriter = tar.NewWriter(file)
	}
	defer tarWriter.Close()

	// Add metadata to archive
	metadataJSON, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal metadata: %w", err)
	}

	header := &tar.Header{
		Name:    "backup_metadata.json",
		Mode:    0644,
		Size:    int64(len(metadataJSON)),
		ModTime: time.Now(),
	}

	if err := tarWriter.WriteHeader(header); err != nil {
		return "", fmt.Errorf("failed to write metadata header: %w", err)
	}

	if _, err := tarWriter.Write(metadataJSON); err != nil {
		return "", fmt.Errorf("failed to write metadata: %w", err)
	}

	// Add installation directory contents
	filesAdded := 0
	err = filepath.Walk(m.opts.InstallDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files with errors
		}

		// Skip the backup file itself
		if path == backupFile {
			return nil
		}

		// Skip files based on options
		if !m.shouldIncludeFile(path, info) {
			return nil
		}

		// Create relative path
		relPath, err := filepath.Rel(m.opts.InstallDir, path)
		if err != nil {
			return nil
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return nil
		}
		header.Name = relPath

		// Write header
		if err := tarWriter.WriteHeader(header); err != nil {
			m.logger.Warnf("Could not add %s to backup: %v", path, err)
			return nil
		}

		// Write file content if not a directory
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return nil
			}
			defer file.Close()

			if _, err := io.Copy(tarWriter, file); err != nil {
				m.logger.Warnf("Could not copy %s to backup: %v", path, err)
				return nil
			}

			filesAdded++
			if filesAdded%10 == 0 && m.opts.Verbose {
				m.logger.Debugf("Added %d files to backup", filesAdded)
			}
		}

		return nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to create backup: %w", err)
	}

	m.logger.Infof("Files archived: %d", filesAdded)

	// Close writers to ensure all data is written
	if err := tarWriter.Close(); err != nil {
		return "", fmt.Errorf("failed to close tar writer: %w", err)
	}

	if gzWriter != nil {
		if err := gzWriter.Close(); err != nil {
			return "", fmt.Errorf("failed to close gzip writer: %w", err)
		}
	}

	if err := file.Close(); err != nil {
		return "", fmt.Errorf("failed to close backup file: %w", err)
	}

	// Calculate final size and checksum
	backupInfo, err := os.Stat(backupFile)
	if err != nil {
		return "", fmt.Errorf("failed to get backup file info: %w", err)
	}

	metadata.Size = backupInfo.Size()

	// Calculate checksum
	checksum, err := m.calculateChecksum(backupFile)
	if err != nil {
		m.logger.Warnf("Failed to calculate checksum: %v", err)
	} else {
		metadata.Checksum = checksum
	}

	// Update metadata in the file
	m.updateMetadataInFile(backupFile, metadata)

	// Save external metadata file
	metadataPath := backupFile + ".meta"
	if err := m.saveBackupMetadata(metadata, metadataPath); err != nil {
		m.logger.Warnf("Failed to save external metadata: %v", err)
	}

	return backupFile, nil
}

// Restore restores from a backup file
func (m *Manager) Restore(backupFile string) error {
	if m.opts.Verbose {
		m.logger.Infof("Restoring from %s", backupFile)
	}

	// Load backup metadata
	metadataPath := backupFile + ".meta"
	metadata, err := m.loadBackupMetadata(metadataPath)
	if err != nil {
		m.logger.Warnf("Could not load external metadata: %v", err)
		// Try to get metadata from the archive itself
		metadata = m.getMetadataFromArchive(backupFile)
	}

	// Verify backup if metadata available
	if metadata != nil && metadata.Checksum != "" {
		m.logger.Info("Verifying backup integrity...")
		if err := m.verifyBackup(backupFile, metadata); err != nil {
			return fmt.Errorf("backup verification failed: %w", err)
		}
		m.logger.Success("Backup verification passed")
	}

	// Open backup file
	file, err := os.Open(backupFile)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer file.Close()

	// Determine compression from file extension
	var tarReader *tar.Reader
	if filepath.Ext(backupFile) == ".gz" {
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			return fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzReader.Close()
		tarReader = tar.NewReader(gzReader)
	} else {
		tarReader = tar.NewReader(file)
	}

	// Extract files
	filesRestored := 0
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		// Skip metadata file
		if header.Name == "backup_metadata.json" {
			continue
		}

		targetPath := filepath.Join(m.opts.InstallDir, header.Name)

		// Security check: ensure path is within target directory
		if !strings.HasPrefix(targetPath, m.opts.InstallDir) {
			return fmt.Errorf("invalid path in backup: %s", header.Name)
		}

		// Check if file exists and overwrite flag
		if _, err := os.Stat(targetPath); err == nil && !m.opts.Overwrite {
			m.logger.Warnf("Skipping existing file: %s", targetPath)
			continue
		}

		// Create directory if needed
		if header.Typeflag == tar.TypeDir {
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				m.logger.Warnf("Could not create directory %s: %v", targetPath, err)
			}
			continue
		}

		// Create parent directory
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			m.logger.Warnf("Could not create parent directory for %s: %v", targetPath, err)
			continue
		}

		// Extract file
		outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
		if err != nil {
			m.logger.Warnf("Could not create file %s: %v", targetPath, err)
			continue
		}

		if _, err := io.Copy(outFile, tarReader); err != nil {
			outFile.Close()
			m.logger.Warnf("Could not extract file %s: %v", targetPath, err)
			continue
		}
		outFile.Close()

		// Set modification time
		if err := os.Chtimes(targetPath, header.ModTime, header.ModTime); err != nil {
			m.logger.Warnf("Failed to set modification time for %s: %v", targetPath, err)
		}

		filesRestored++
		if filesRestored%10 == 0 && m.opts.Verbose {
			m.logger.Debugf("Restored %d files", filesRestored)
		}
	}

	m.logger.Infof("Files restored: %d", filesRestored)

	return nil
}

// ListBackups returns a list of available backups
func (m *Manager) ListBackups() ([]BackupInfo, error) {
	backups := []BackupInfo{}

	// Check if backup directory exists
	if _, err := os.Stat(m.opts.BackupDir); os.IsNotExist(err) {
		return backups, nil
	}

	// Find all backup files
	err := filepath.Walk(m.opts.BackupDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files with errors
		}

		// Skip directories and metadata files
		if info.IsDir() || strings.HasSuffix(path, ".meta") {
			return nil
		}

		// Check if it's a backup file
		ext := filepath.Ext(path)
		if ext == ".tar" || ext == ".gz" || ext == ".bz2" {
			backupInfo := m.GetBackupInfo(path)
			backups = append(backups, backupInfo)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list backups: %w", err)
	}

	// Sort by creation date (newest first)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].Created.After(backups[j].Created)
	})

	return backups, nil
}

// GetBackupInfo gets information about a backup file
func (m *Manager) GetBackupInfo(backupPath string) BackupInfo {
	info := BackupInfo{
		Path:   backupPath,
		Exists: false,
	}

	// Check if file exists
	fileInfo, err := os.Stat(backupPath)
	if err != nil {
		info.Error = err
		return info
	}

	info.Exists = true
	info.Size = fileInfo.Size()
	info.Created = fileInfo.ModTime()

	// Try to load external metadata first
	metadataPath := backupPath + ".meta"
	metadata, err := m.loadBackupMetadata(metadataPath)
	if err == nil && metadata != nil {
		info.Metadata = metadata
		if metadata.Timestamp.IsZero() && metadata.Created != "" {
			// Parse created time for backward compatibility
			if t, err := time.Parse(time.RFC3339, metadata.Created); err == nil {
				metadata.Timestamp = t
			}
		}
	} else {
		// Try to read metadata from archive
		info.Metadata = m.getMetadataFromArchive(backupPath)
	}

	// Count files in archive
	info.FileCount = m.countFilesInArchive(backupPath)

	return info
}

// Cleanup removes old backups
func (m *Manager) Cleanup(keep int, olderThan int) (int, error) {
	backups, err := m.ListBackups()
	if err != nil {
		return 0, err
	}

	if len(backups) == 0 {
		return 0, nil
	}

	toRemove := []BackupInfo{}

	// Remove by age
	if olderThan > 0 {
		cutoffDate := time.Now().AddDate(0, 0, -olderThan)
		for _, backup := range backups {
			if backup.Created.Before(cutoffDate) {
				toRemove = append(toRemove, backup)
			}
		}
	}

	// Keep only N most recent
	if keep >= 0 && len(backups) > keep {
		// Backups are already sorted by date (newest first)
		for i := keep; i < len(backups); i++ {
			toRemove = append(toRemove, backups[i])
		}
	}

	// Remove duplicates
	seen := make(map[string]bool)
	uniqueToRemove := []BackupInfo{}
	for _, backup := range toRemove {
		if !seen[backup.Path] {
			seen[backup.Path] = true
			uniqueToRemove = append(uniqueToRemove, backup)
		}
	}

	// Remove backups
	removed := 0
	for _, backup := range uniqueToRemove {
		// Remove backup file
		if err := os.Remove(backup.Path); err != nil {
			m.logger.Warnf("Could not remove %s: %v", backup.Path, err)
		} else {
			m.logger.Infof("Removed backup: %s", filepath.Base(backup.Path))
			removed++
		}

		// Remove metadata file if exists
		metadataPath := backup.Path + ".meta"
		if err := os.Remove(metadataPath); err != nil && !os.IsNotExist(err) {
			m.logger.Warnf("Could not remove metadata %s: %v", metadataPath, err)
		}
	}

	return removed, nil
}

// Private helper methods

func (m *Manager) createBackupMetadata() *BackupMetadata {
	// Get version information
	versionManager := versioning.NewVersionManager(m.opts.InstallDir)
	frameworkVersion, _ := versionManager.GetCurrentVersion()
	if frameworkVersion == "" {
		frameworkVersion = "1.0.0"
	}

	metadata := &BackupMetadata{
		BackupVersion:    "1.0.0",
		Created:          time.Now().Format(time.RFC3339),
		Timestamp:        time.Now(),
		InstallDir:       m.opts.InstallDir,
		Components:       make(map[string]string),
		FrameworkVersion: frameworkVersion,
		BackupType:       "full",
		Description:      m.opts.Description,
	}

	// Get component versions from version manager
	components := []string{"core", "commands", "hooks", "mcp"}
	for _, comp := range components {
		if version, err := versionManager.GetComponentVersion(comp); err == nil {
			metadata.Components[comp] = version
		}
	}

	return metadata
}

func (m *Manager) shouldIncludeFile(path string, _ os.FileInfo) bool {
	// Skip hidden files and directories (unless specifically needed)
	name := filepath.Base(path)
	if strings.HasPrefix(name, ".") && name != ".claude" {
		return false
	}

	// Skip log files unless requested
	if !m.opts.IncludeLogs && (strings.HasSuffix(path, ".log") || strings.Contains(path, "/logs/")) {
		return false
	}

	// Skip temporary files
	if strings.HasSuffix(name, ".tmp") || strings.HasSuffix(name, ".temp") {
		return false
	}

	// Skip backup files
	if strings.HasSuffix(name, ".backup") || strings.HasSuffix(name, ".bak") {
		return false
	}

	// Skip the backups directory itself
	if strings.Contains(path, "/backups/") {
		return false
	}

	return true
}

func (m *Manager) calculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func (m *Manager) saveBackupMetadata(metadata *BackupMetadata, metadataPath string) error {
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(metadataPath, data, 0644)
}

func (m *Manager) loadBackupMetadata(metadataPath string) (*BackupMetadata, error) {
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, err
	}

	var metadata BackupMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, err
	}

	return &metadata, nil
}

func (m *Manager) verifyBackup(backupPath string, metadata *BackupMetadata) error {
	// Check if backup file exists
	if !m.fileManager.Exists(backupPath) {
		return fmt.Errorf("backup file does not exist: %s", backupPath)
	}

	// Verify file size if metadata available
	if metadata != nil && metadata.Size > 0 {
		info, err := os.Stat(backupPath)
		if err != nil {
			return fmt.Errorf("failed to get backup file info: %w", err)
		}

		if info.Size() != metadata.Size {
			return fmt.Errorf("backup file size mismatch: expected %d, got %d", metadata.Size, info.Size())
		}
	}

	// Verify checksum if available
	if metadata != nil && metadata.Checksum != "" {
		checksum, err := m.calculateChecksum(backupPath)
		if err != nil {
			return fmt.Errorf("failed to calculate checksum: %w", err)
		}

		if checksum != metadata.Checksum {
			return fmt.Errorf("backup checksum mismatch: expected %s, got %s", metadata.Checksum, checksum)
		}
	}

	// Try to read tar headers to verify archive integrity
	file, err := os.Open(backupPath)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer file.Close()

	var reader io.Reader = file

	if strings.HasSuffix(backupPath, ".gz") {
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			return fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzReader.Close()
		reader = gzReader
	}

	tarReader := tar.NewReader(reader)

	// Read a few headers to verify archive structure
	headerCount := 0
	for headerCount < 10 {
		_, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("backup archive is corrupted: %w", err)
		}
		headerCount++
	}

	return nil
}

func (m *Manager) getMetadataFromArchive(backupPath string) *BackupMetadata {
	file, err := os.Open(backupPath)
	if err != nil {
		return nil
	}
	defer file.Close()

	var reader io.Reader = file
	if strings.HasSuffix(backupPath, ".gz") {
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			return nil
		}
		defer gzReader.Close()
		reader = gzReader
	}

	tarReader := tar.NewReader(reader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		if header.Name == "backup_metadata.json" {
			var metadata BackupMetadata
			if err := json.NewDecoder(tarReader).Decode(&metadata); err == nil {
				return &metadata
			}
			break
		}
	}

	return nil
}

func (m *Manager) countFilesInArchive(backupPath string) int {
	file, err := os.Open(backupPath)
	if err != nil {
		return 0
	}
	defer file.Close()

	var reader io.Reader = file
	if strings.HasSuffix(backupPath, ".gz") {
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			return 0
		}
		defer gzReader.Close()
		reader = gzReader
	}

	tarReader := tar.NewReader(reader)
	fileCount := 0

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		if header.Typeflag == tar.TypeReg && header.Name != "backup_metadata.json" {
			fileCount++
		}
	}

	return fileCount
}

func (m *Manager) updateMetadataInFile(_ string, _ *BackupMetadata) error {
	// This would require re-writing the entire archive, which is complex
	// For now, we'll just rely on the external metadata file
	return nil
}