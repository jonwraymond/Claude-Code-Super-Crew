// Package claude provides directory restructuring utilities for moving install-related files to .crew/
package claude

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jonwraymond/claude-code-super-crew/pkg/logger"
)

// DirectoryRestructurer handles the migration of utility directories to .crew/
type DirectoryRestructurer struct {
	claudeDir string
	crewDir   string
	logger    logger.Logger
}

// NewDirectoryRestructurer creates a new directory restructurer
func NewDirectoryRestructurer(claudeDir string) *DirectoryRestructurer {
	return &DirectoryRestructurer{
		claudeDir: claudeDir,
		crewDir:   filepath.Join(claudeDir, ".crew"),
		logger:    logger.GetLogger(),
	}
}

// DirectoriesToMove defines the utility directories that should be moved to .crew/
// Note: statsig and shell-snapshots belong to Claude Code itself, not SuperCrew
var DirectoriesToMove = []string{
	"backups",
	"logs", 
	"config",
	"completions",
	"scripts",
	"workflows",
	"prompts",
}

// RestructureDirectories moves utility directories to .crew/ while keeping SuperCrew core directories in place
func (dr *DirectoryRestructurer) RestructureDirectories() error {
	dr.logger.Info("Starting directory restructuring to move utility files to .crew/")

	// Create the .crew directory
	if err := os.MkdirAll(dr.crewDir, 0755); err != nil {
		return fmt.Errorf("failed to create .crew directory: %w", err)
	}
	dr.logger.Infof("Created .crew directory: %s", dr.crewDir)

	// Move each utility directory
	movedCount := 0
	for _, dirName := range DirectoriesToMove {
		oldPath := filepath.Join(dr.claudeDir, dirName)
		newPath := filepath.Join(dr.crewDir, dirName)

		if err := dr.moveDirectory(oldPath, newPath); err != nil {
			dr.logger.Warnf("Failed to move %s: %v", dirName, err)
		} else {
			movedCount++
		}
	}

	dr.logger.Successf("Directory restructuring completed. Moved %d utility directories to .crew/", movedCount)
	return nil
}

// moveDirectory moves a directory from old path to new path
func (dr *DirectoryRestructurer) moveDirectory(oldPath, newPath string) error {
	// Check if source directory exists
	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		dr.logger.Debugf("Directory %s doesn't exist, skipping", oldPath)
		return nil
	}

	// Check if destination already exists
	if _, err := os.Stat(newPath); err == nil {
		// Destination exists - merge contents
		dr.logger.Debugf("Destination %s already exists, merging contents", newPath)
		
		// Move contents of old directory to new directory
		entries, err := os.ReadDir(oldPath)
		if err != nil {
			return fmt.Errorf("failed to read source directory: %w", err)
		}
		
		for _, entry := range entries {
			srcPath := filepath.Join(oldPath, entry.Name())
			dstPath := filepath.Join(newPath, entry.Name())
			
			// Check if destination file/dir already exists
			if _, err := os.Stat(dstPath); err == nil {
				dr.logger.Warnf("Skipping %s - already exists in destination", entry.Name())
				continue
			}
			
			// Move the item
			if err := os.Rename(srcPath, dstPath); err != nil {
				return fmt.Errorf("failed to move %s: %w", entry.Name(), err)
			}
		}
		
		// Remove old directory if empty
		if err := os.Remove(oldPath); err != nil {
			// Directory might not be empty, that's ok
			dr.logger.Debugf("Could not remove old directory %s: %v", oldPath, err)
		}
		
		dr.logger.Infof("Merged %s → %s", filepath.Base(oldPath), newPath)
		return nil
	}

	// Create parent directory for new path
	if err := os.MkdirAll(filepath.Dir(newPath), 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	// Move the directory
	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("failed to move directory from %s to %s: %w", oldPath, newPath, err)
	}

	dr.logger.Infof("Moved %s → %s", filepath.Base(oldPath), newPath)
	return nil
}

// GetCrewPath returns the path for a utility directory under .crew/
func (dr *DirectoryRestructurer) GetCrewPath(dirName string) string {
	return filepath.Join(dr.crewDir, dirName)
}

// GetUtilityPaths returns the new paths for all utility directories
func (dr *DirectoryRestructurer) GetUtilityPaths() map[string]string {
	paths := make(map[string]string)
	for _, dirName := range DirectoriesToMove {
		paths[dirName] = dr.GetCrewPath(dirName)
	}
	return paths
}

// ValidateRestructure validates that the restructuring was successful
func (dr *DirectoryRestructurer) ValidateRestructure() []string {
	issues := []string{}

	// Check that .crew directory exists
	if _, err := os.Stat(dr.crewDir); os.IsNotExist(err) {
		issues = append(issues, ".crew directory not found")
		return issues
	}

	// Check that utility directories are in .crew/
	for _, dirName := range DirectoriesToMove {
		crewPath := dr.GetCrewPath(dirName)
		oldPath := filepath.Join(dr.claudeDir, dirName)

		// Check if directory exists in .crew/ (if it was supposed to be moved)
		if _, err := os.Stat(oldPath); err == nil {
			// Old directory still exists, check if it was copied to .crew/
			if _, err := os.Stat(crewPath); os.IsNotExist(err) {
				issues = append(issues, fmt.Sprintf("%s was not moved to .crew/", dirName))
			}
		}
	}

	return issues
}

// CreateCrewDirectoryStructure creates the full .crew directory structure
func (dr *DirectoryRestructurer) CreateCrewDirectoryStructure() error {
	dr.logger.Info("Creating .crew directory structure")

	// Create .crew directory
	if err := os.MkdirAll(dr.crewDir, 0755); err != nil {
		return fmt.Errorf("failed to create .crew directory: %w", err)
	}

	// Create all utility subdirectories
	for _, dirName := range DirectoriesToMove {
		dirPath := dr.GetCrewPath(dirName)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return fmt.Errorf("failed to create %s directory: %w", dirName, err)
		}
		dr.logger.Debugf("Created directory: %s", dirPath)
	}

	dr.logger.Success("Created .crew directory structure with all utility directories")
	return nil
}