// Package managers provides helper utilities for file, settings, and security management.
// It implements functionality similar to the Python setup/managers package.
package managers

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/jonwraymond/claude-code-super-crew/internal/metadata"
)

// FileManager handles file operations for the installation system
type FileManager struct {
	metadataManager *metadata.MetadataManager
	installDir      string
}

// NewFileManager creates a new file manager instance
func NewFileManager() *FileManager {
	return &FileManager{}
}

// NewFileManagerWithMetadata creates a file manager with inventory tracking
func NewFileManagerWithMetadata(installDir string) *FileManager {
	return &FileManager{
		metadataManager: metadata.NewMetadataManager(installDir),
		installDir:      installDir,
	}
}

// SetMetadataManager sets the metadata manager for inventory tracking
func (fm *FileManager) SetMetadataManager(mm *metadata.MetadataManager) {
	fm.metadataManager = mm
}

// HasMetadataManager checks if metadata manager is available
func (fm *FileManager) HasMetadataManager() bool {
	return fm.metadataManager != nil
}

// AddToInventory adds a file or directory to the inventory tracking
func (fm *FileManager) AddToInventory(path string, isDirectory bool) error {
	if fm.metadataManager != nil {
		return fm.metadataManager.AddToInventory(path, isDirectory)
	}
	return nil
}

// CopyFileWithMerge copies a file with support for merging existing content
// This is used for files like CLAUDE.md that may contain user customizations
func (fm *FileManager) CopyFileWithMerge(src, dst string) error {
	// Check if destination file exists
	if fm.Exists(dst) {
		// For CLAUDE.md files, attempt to merge instead of overwrite
		if filepath.Base(dst) == "CLAUDE.md" {
			return fm.mergeClaudeFile(src, dst)
		}
	}

	// For other files, use normal copy with inventory tracking
	return fm.CopyFileWithInventory(src, dst)
}

// mergeClaudeFile merges framework sections into existing CLAUDE.md while preserving user content
func (fm *FileManager) mergeClaudeFile(src, dst string) error {
	// Read source (new framework content)
	srcContent, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	// Read destination (existing file with potential user content)
	dstContent, err := os.ReadFile(dst)
	if err != nil {
		return fmt.Errorf("failed to read destination file: %w", err)
	}

	// Merge the content
	mergedContent := fm.mergeClaudeContent(string(dstContent), string(srcContent))

	// Write merged content
	if err := os.WriteFile(dst, []byte(mergedContent), 0644); err != nil {
		return fmt.Errorf("failed to write merged file: %w", err)
	}

	// Track in inventory
	if fm.metadataManager != nil {
		if err := fm.metadataManager.AddToInventory(dst, false); err != nil {
			// Log error but don't fail the operation
			// TODO: Add logging here
		}
	}

	return nil
}

// mergeClaudeContent merges framework sections into existing CLAUDE.md content
func (fm *FileManager) mergeClaudeContent(existing, framework string) string {
	// Simple merge strategy: preserve user sections, update framework sections
	// This is a basic implementation - could be enhanced with more sophisticated parsing

	// For now, append framework content to existing content with a separator
	// In a more sophisticated implementation, we would:
	// 1. Parse both files into sections
	// 2. Identify framework vs user sections
	// 3. Update only framework sections
	// 4. Preserve all user sections

	separator := "\n\n<!-- FRAMEWORK CONTENT BELOW - DO NOT EDIT MANUALLY -->\n\n"
	return existing + separator + framework
}

// Exists checks if a file or directory exists
func (fm *FileManager) Exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// EnsureDirectory creates a directory and all parent directories
func (fm *FileManager) EnsureDirectory(path string) error {
	return os.MkdirAll(path, 0755)
}

// EnsureDirectoryWithInventory creates a directory and tracks it in inventory
func (fm *FileManager) EnsureDirectoryWithInventory(path string) error {
	existed := fm.Exists(path)

	// Create directory if it doesn't exist
	if !existed {
		if err := os.MkdirAll(path, 0755); err != nil {
			return err
		}
	}

	// Track in inventory if metadata manager is available (for both new and existing)
	// This ensures we track all directories we intend to manage
	if fm.metadataManager != nil {
		if err := fm.metadataManager.AddToInventory(path, true); err != nil {
			// Log error but don't fail installation
			// TODO: Add logging here
		}
	}

	return nil
}

// CopyFile copies a file from source to destination
func (fm *FileManager) CopyFile(src, dst string) error {
	// Open source file
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	// Get source file info
	sourceInfo, err := sourceFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat source file: %w", err)
	}

	// Create destination directory if needed
	destDir := filepath.Dir(dst)
	if err := fm.EnsureDirectory(destDir); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Create destination file
	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	// Copy file contents
	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	// Preserve file permissions
	if err := os.Chmod(dst, sourceInfo.Mode()); err != nil {
		return fmt.Errorf("failed to set file permissions: %w", err)
	}

	return nil
}

// CopyFileWithInventory copies a file from source to destination and tracks it
func (fm *FileManager) CopyFileWithInventory(src, dst string) error {
	// Copy the file
	if err := fm.CopyFile(src, dst); err != nil {
		return err
	}

	// Track in inventory if metadata manager is available
	if fm.metadataManager != nil {
		if err := fm.metadataManager.AddToInventory(dst, false); err != nil {
			// Log error but don't fail installation
			// TODO: Add logging here
		}

		// Add to integrity tracking
		relPath, err := filepath.Rel(fm.installDir, dst)
		if err != nil {
			relPath = dst // Use absolute path if relative conversion fails
		}

		// Determine component from path
		component := fm.determineComponentFromPath(relPath)

		if err := fm.metadataManager.AddFileToIntegrityTracking(relPath, component); err != nil {
			// Log error but don't fail the operation
			// TODO: Add proper logging here
		}
	}

	return nil
}

// determineComponentFromPath determines which component a file belongs to based on its path
func (fm *FileManager) determineComponentFromPath(relPath string) string {
	// Extract the first directory component
	parts := strings.Split(relPath, string(filepath.Separator))
	if len(parts) > 0 {
		switch parts[0] {
		case "hooks":
			return "hooks"
		case "agents":
			return "agents"
		case "commands":
			return "commands"
		case "mcp":
			return "mcp"
		default:
			// Check if it's a core file (no subdirectory)
			if len(parts) == 1 {
				return "core"
			}
		}
	}
	return "unknown"
}

// CopyDirectory recursively copies a directory
func (fm *FileManager) CopyDirectory(src, dst string) error {
	// Get source directory info
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source directory: %w", err)
	}

	// Create destination directory
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Read directory entries
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read source directory: %w", err)
	}

	// Copy each entry
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Recursively copy subdirectory
			if err := fm.CopyDirectory(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy file
			if err := fm.CopyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// RemoveFile safely removes a file
func (fm *FileManager) RemoveFile(path string) error {
	return os.Remove(path)
}

// RemoveDirectory safely removes a directory and its contents
func (fm *FileManager) RemoveDirectory(path string) error {
	return os.RemoveAll(path)
}

// FileExists checks if a file exists
func (fm *FileManager) FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsDirectory checks if a path is a directory
func (fm *FileManager) IsDirectory(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// IsFile checks if a path is a regular file
func (fm *FileManager) IsFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// GetFileSize returns the size of a file in bytes
func (fm *FileManager) GetFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// GetDirectorySize recursively calculates the size of a directory
func (fm *FileManager) GetDirectorySize(path string) (int64, error) {
	var size int64

	err := filepath.WalkDir(path, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return err
			}
			size += info.Size()
		}

		return nil
	})

	return size, err
}
