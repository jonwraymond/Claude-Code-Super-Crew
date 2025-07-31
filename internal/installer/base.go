package installer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/jonwraymond/claude-code-super-crew/pkg/backup"
	"github.com/jonwraymond/claude-code-super-crew/internal/core"
	"github.com/jonwraymond/claude-code-super-crew/internal/managers"
	"github.com/jonwraymond/claude-code-super-crew/pkg/logger"
)

// Installer handles installation operations
type Installer struct {
	installDir          string
	dryRun              bool
	components          map[string]core.Component
	installedComponents []string
	updatedComponents   []string
	failedComponents    []string
	backupPath          string
	settingsManager     *managers.SettingsManager
	logger              logger.Logger
}

// NewInstaller creates a new installer
func NewInstaller(installDir string, dryRun bool) *Installer {
	return &Installer{
		installDir:          installDir,
		dryRun:              dryRun,
		components:          make(map[string]core.Component),
		installedComponents: []string{},
		updatedComponents:   []string{},
		failedComponents:    []string{},
		settingsManager:     managers.NewSettingsManager(installDir),
		logger:              logger.GetLogger(),
	}
}

// RegisterComponents registers components with the installer
func (i *Installer) RegisterComponents(components []core.Component) {
	for _, comp := range components {
		metadata := comp.GetMetadata()
		i.components[metadata.Name] = comp
	}
}

// InstallComponents installs the specified components
func (i *Installer) InstallComponents(componentNames []string, config map[string]interface{}) bool {
	success := true

	// Check if installation already exists and create backup
	if _, err := os.Stat(i.installDir); err == nil && !i.dryRun {
		// Installation exists, create backup regardless of config
		i.logger.Info("Existing installation detected, creating backup...")
		if err := i.createBackup("pre-install"); err != nil {
			i.logger.Warnf("Failed to create pre-install backup: %v", err)
		}
	} else if backup, ok := config["backup"].(bool); ok && backup && !i.dryRun {
		// Create backup if explicitly requested
		if err := i.createBackup("pre-install"); err != nil {
			i.logger.Warnf("Failed to create backup: %v", err)
		}
	}

	// Install each component
	for _, name := range componentNames {
		comp, ok := i.components[name]
		if !ok {
			i.logger.Errorf("Component %s not found", name)
			i.failedComponents = append(i.failedComponents, name)
			success = false
			continue
		}

		i.logger.Infof("Installing %s...", name)

		if i.dryRun {
			i.logger.Infof("[DRY RUN] Would install %s", name)
			i.installedComponents = append(i.installedComponents, name)
			continue
		}

		// Validate component
		if err := comp.Validate(i.installDir); err != nil {
			i.logger.Errorf("Validation failed for %s: %v", name, err)
			i.failedComponents = append(i.failedComponents, name)
			success = false
			continue
		}

		// Install component
		if err := comp.Install(i.installDir, config); err != nil {
			i.logger.Errorf("Installation failed for %s: %v", name, err)
			i.failedComponents = append(i.failedComponents, name)
			success = false
			continue
		}

		// Update settings
		metadata := comp.GetMetadata()
		if err := i.settingsManager.UpdateComponentVersion(name, metadata.Version); err != nil {
			i.logger.Warnf("Failed to update settings for %s: %v", name, err)
		}

		i.installedComponents = append(i.installedComponents, name)
		i.logger.Successf("Installed %s successfully", name)
	}

	// Run post-installation validation
	if !i.dryRun && len(i.installedComponents) > 0 {
		i.runPostInstallValidation()
	}

	return success
}

// runPostInstallValidation runs validation for all installed components
func (i *Installer) runPostInstallValidation() {
	i.logger.Info("Running post-installation validation...")

	allValid := true
	for _, name := range i.installedComponents {
		comp, ok := i.components[name]
		if !ok {
			continue
		}

		// Check if component implements ValidateInstallation
		type validator interface {
			ValidateInstallation(installDir string) (bool, []string)
		}

		if v, ok := comp.(validator); ok {
			isValid, errors := v.ValidateInstallation(i.installDir)
			if isValid {
				i.logger.Successf("✓ %s: Valid", name)
			} else {
				i.logger.Errorf("✗ %s: Invalid", name)
				for _, err := range errors {
					i.logger.Errorf("  - %s", err)
				}
				allValid = false
			}
		}
	}

	if allValid {
		i.logger.Success("All components validated successfully!")
	} else {
		i.logger.Warn("Some components failed validation. Check errors above.")
	}
}

// UpdateComponents updates the specified components
func (i *Installer) UpdateComponents(componentNames []string, config map[string]interface{}) bool {
	success := true

	// Create backup if requested
	if backup, ok := config["backup"].(bool); ok && backup && !i.dryRun {
		if err := i.createBackup("pre-update"); err != nil {
			i.logger.Warnf("Failed to create backup: %v", err)
		}
	}

	// Update each component
	for _, name := range componentNames {
		comp, ok := i.components[name]
		if !ok {
			i.logger.Errorf("Component %s not found", name)
			i.failedComponents = append(i.failedComponents, name)
			success = false
			continue
		}

		i.logger.Infof("Updating %s...", name)

		if i.dryRun {
			i.logger.Infof("[DRY RUN] Would update %s", name)
			i.updatedComponents = append(i.updatedComponents, name)
			continue
		}

		// Update component
		if err := comp.Update(i.installDir, config); err != nil {
			i.logger.Errorf("Update failed for %s: %v", name, err)
			i.failedComponents = append(i.failedComponents, name)
			success = false
			continue
		}

		// Update settings
		metadata := comp.GetMetadata()
		if err := i.settingsManager.UpdateComponentVersion(name, metadata.Version); err != nil {
			i.logger.Warnf("Failed to update settings for %s: %v", name, err)
		}

		i.updatedComponents = append(i.updatedComponents, name)
		i.logger.Successf("Updated %s successfully", name)
	}

	return success
}

// UninstallComponents uninstalls the specified components
func (i *Installer) UninstallComponents(componentNames []string, config map[string]interface{}) bool {
	success := true

	// Create backup if requested
	if backup, ok := config["backup"].(bool); ok && backup && !i.dryRun {
		if err := i.createBackup("pre-uninstall"); err != nil {
			i.logger.Warnf("Failed to create backup: %v", err)
		}
	}

	// Uninstall in reverse order
	for idx := len(componentNames) - 1; idx >= 0; idx-- {
		name := componentNames[idx]
		comp, ok := i.components[name]
		if !ok {
			i.logger.Errorf("Component %s not found", name)
			i.failedComponents = append(i.failedComponents, name)
			success = false
			continue
		}

		i.logger.Infof("Uninstalling %s...", name)

		if i.dryRun {
			i.logger.Infof("[DRY RUN] Would uninstall %s", name)
			continue
		}

		// Uninstall component
		if err := comp.Uninstall(i.installDir, config); err != nil {
			i.logger.Errorf("Uninstall failed for %s: %v", name, err)
			i.failedComponents = append(i.failedComponents, name)
			success = false
			continue
		}

		i.logger.Successf("Uninstalled %s successfully", name)
	}

	return success
}

// GetInstallationSummary returns a summary of the installation
func (i *Installer) GetInstallationSummary() map[string]interface{} {
	return map[string]interface{}{
		"installed":   i.installedComponents,
		"failed":      i.failedComponents,
		"backup_path": i.backupPath,
	}
}

// GetUpdateSummary returns a summary of the update
func (i *Installer) GetUpdateSummary() map[string]interface{} {
	return map[string]interface{}{
		"updated":     i.updatedComponents,
		"failed":      i.failedComponents,
		"backup_path": i.backupPath,
	}
}

// createBackup creates a backup of the installation
func (i *Installer) createBackup(reason string) error {
	backupsDir := filepath.Join(i.installDir, ".crew", "backups")

	// Create backup manager
	mgr := backup.NewManager(backup.Options{
		InstallDir:    i.installDir,
		BackupDir:     backupsDir,
		BackupName:    "crew_backup",
		Compress:      "gzip",
		Verbose:       false,
		DryRun:        false,
		IncludeConfig: true,
		IncludeLogs:   false,
		Description:   reason,
	})

	// Create the backup
	backupPath, err := mgr.Create()
	if err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	i.backupPath = backupPath
	i.logger.Infof("Created backup at %s", backupPath)
	return nil
}

// copyDir copies a directory recursively
func copyDir(src, dst string) error {
	// Get source directory info
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Create destination directory
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	// Read directory contents
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	// Copy each entry
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Recursively copy directory
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy file
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a single file
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Get source file info
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	// Create destination file
	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Copy content
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	return nil
}

// createTarball creates a tar.gz archive from the source directory
func createTarball(sourceDir, targetFile string) error {
	// Use the backup package to create the tarball
	mgr := backup.NewManager(backup.Options{
		InstallDir: sourceDir,
		BackupDir:  filepath.Dir(targetFile),
		BackupName: "temp",
		Compress:   "gzip",
		Verbose:    false,
	})

	// Create the backup
	backupPath, err := mgr.Create()
	if err != nil {
		return fmt.Errorf("failed to create tarball: %w", err)
	}

	// Rename to target file
	if err := os.Rename(backupPath, targetFile); err != nil {
		return fmt.Errorf("failed to rename backup: %w", err)
	}

	return nil
}
