package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jonwraymond/claude-code-super-crew/internal/managers"
	"github.com/jonwraymond/claude-code-super-crew/internal/ui"
	"github.com/jonwraymond/claude-code-super-crew/pkg/logger"
	"github.com/spf13/cobra"
)

// UninstallFlags holds uninstall command flags
type UninstallFlags struct {
	Components   []string
	Complete     bool
	KeepBackups  bool
	KeepLogs     bool
	KeepSettings bool
	NoConfirm    bool
}

var uninstallFlags UninstallFlags

// NewUninstallCommand creates the uninstall command
func NewUninstallCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Remove Claude Code Super Crew framework installation",
		Long: `Uninstall Claude Code Super Crew Framework components.

Examples:
  crew uninstall                    # Interactive uninstall
  crew uninstall --components core  # Remove specific components
  crew uninstall --complete --force # Complete removal (forced)
  crew uninstall --keep-backups     # Keep backup files`,
		RunE: runUninstall,
	}

	// Uninstall mode options
	cmd.Flags().StringSliceVar(&uninstallFlags.Components, "components", nil,
		"Specific components to uninstall")
	cmd.Flags().BoolVar(&uninstallFlags.Complete, "complete", false,
		"Complete uninstall (remove all files and directories)")

	// Data preservation options
	cmd.Flags().BoolVar(&uninstallFlags.KeepBackups, "keep-backups", false,
		"Keep backup files during uninstall")
	cmd.Flags().BoolVar(&uninstallFlags.KeepLogs, "keep-logs", false,
		"Keep log files during uninstall")
	cmd.Flags().BoolVar(&uninstallFlags.KeepSettings, "keep-settings", false,
		"Keep user settings during uninstall")

	// Safety options
	cmd.Flags().BoolVar(&uninstallFlags.NoConfirm, "no-confirm", false,
		"Skip confirmation prompts (use with caution)")

	return cmd
}

func runUninstall(cmd *cobra.Command, args []string) error {
	log := logger.GetLogger()
	log.SetVerbose(globalFlags.Verbose)
	log.SetQuiet(globalFlags.Quiet)

	// Validate installation directory (skip in test mode)
	if !testMode {
		expectedHome := filepath.Join(os.Getenv("HOME"))
		if expectedHome == "" {
			expectedHome = filepath.Join(os.Getenv("USERPROFILE")) // Windows
		}
		actualDir, _ := filepath.Abs(globalFlags.InstallDir)

		if !strings.HasPrefix(actualDir, expectedHome) {
			ui.DisplayError("Installation must be inside your user profile directory.")
			fmt.Printf("    Expected prefix: %s\n", expectedHome)
			fmt.Printf("    Provided path:   %s\n", actualDir)
			return fmt.Errorf("invalid installation directory")
		}
	}

	// Display header
	if !globalFlags.Quiet {
		ui.DisplayHeader(
			"Claude Code Super Crew Uninstall v1.0",
			"Removing Claude Code Super Crew framework components",
		)
	}

	// Get installation information
	info := getInstallationInfo(globalFlags.InstallDir)

	// Display current installation
	if !globalFlags.Quiet {
		displayUninstallInfo(info)
	}

	// Check if Claude Code Super Crew is installed
	if !info["exists"].(bool) {
		log.Warnf("No Claude Code Super Crew installation found in %s", globalFlags.InstallDir)
		return nil
	}

	// Get components to uninstall
	installedComponents := info["components"].(map[string]string)
	components, err := getComponentsToUninstall(uninstallFlags, installedComponents)
	if err != nil {
		return err
	}

	if len(components) == 0 {
		log.Info("No components selected for uninstall")
		return nil
	}

	// Display uninstall plan
	if !globalFlags.Quiet {
		displayUninstallPlan(components, uninstallFlags, info)
	}

	// Confirmation (skip for dry-run)
	if !uninstallFlags.NoConfirm && !globalFlags.Yes && !globalFlags.DryRun {
		var warningMsg string
		if uninstallFlags.Complete {
			warningMsg = "This will completely remove Claude Code Super Crew. Continue?"
		} else {
			warningMsg = fmt.Sprintf("This will remove %d component(s). Continue?", len(components))
		}

		if !ui.Confirm(warningMsg, false) {
			log.Info("Uninstall cancelled by user")
			return nil
		}
	}

	// Create backup if not dry run and not keeping backups
	if !globalFlags.DryRun && !uninstallFlags.KeepBackups {
		createUninstallBackup(globalFlags.InstallDir, components)
	}

	// Perform uninstall
	success := performUninstall(components, uninstallFlags, info)

	if success {
		if !globalFlags.Quiet {
			ui.DisplaySuccess("Claude Code Super Crew uninstall completed successfully!")

			if !globalFlags.DryRun {
				fmt.Printf("\n%sUninstall complete:%s\n", ui.ColorCyan, ui.ColorReset)
				fmt.Printf("Claude Code Super Crew has been removed from %s\n", globalFlags.InstallDir)
				if !uninstallFlags.Complete {
					fmt.Println("You can reinstall anytime using 'crew install'")
				}
			}
		}
		return nil
	} else {
		ui.DisplayError("Uninstall completed with some failures. Check logs for details.")
		return fmt.Errorf("uninstall failed")
	}
}

func getInstallationInfo(installDir string) map[string]interface{} {
	info := map[string]interface{}{
		"install_dir": installDir,
		"exists":      false,
		"components":  make(map[string]string),
		"directories": []string{},
		"files":       []string{},
		"total_size":  int64(0),
	}

	// Check if installation exists
	if _, err := os.Stat(installDir); os.IsNotExist(err) {
		return info
	}

	info["exists"] = true

	// Get installed components
	settingsManager := managers.NewSettingsManager(installDir)
	if components, err := settingsManager.GetInstalledComponents(); err == nil {
		info["components"] = components
	}

	// Scan installation directory
	var files []string
	var dirs []string
	var totalSize int64

	err := filepath.Walk(installDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue walking
		}

		if info.IsDir() {
			dirs = append(dirs, path)
		} else {
			files = append(files, path)
			totalSize += info.Size()
		}

		return nil
	})

	if err == nil {
		info["files"] = files
		info["directories"] = dirs
		info["total_size"] = totalSize
	}

	return info
}

func displayUninstallInfo(info map[string]interface{}) {
	fmt.Printf("\n%s%sCurrent Installation%s\n", ui.ColorCyan, ui.ColorBright, ui.ColorReset)
	fmt.Println(strings.Repeat("=", 50))

	if !info["exists"].(bool) {
		fmt.Printf("%sNo Claude Code Super Crew installation found%s\n", ui.ColorYellow, ui.ColorReset)
		return
	}

	fmt.Printf("%sInstallation Directory:%s %s\n", ui.ColorBlue, ui.ColorReset, info["install_dir"])

	components := info["components"].(map[string]string)
	if len(components) > 0 {
		fmt.Printf("%sInstalled Components:%s\n", ui.ColorBlue, ui.ColorReset)
		for component, version := range components {
			fmt.Printf("  %s: v%s\n", component, version)
		}
	}

	files := info["files"].([]string)
	dirs := info["directories"].([]string)
	totalSize := info["total_size"].(int64)

	fmt.Printf("%sFiles:%s %d\n", ui.ColorBlue, ui.ColorReset, len(files))
	fmt.Printf("%sDirectories:%s %d\n", ui.ColorBlue, ui.ColorReset, len(dirs))

	if totalSize > 0 {
		fmt.Printf("%sTotal Size:%s %s\n", ui.ColorBlue, ui.ColorReset, ui.FormatSize(totalSize))
	}

	fmt.Println()
}

func getComponentsToUninstall(flags UninstallFlags, installed map[string]string) ([]string, error) {
	log := logger.GetLogger()

	// Complete uninstall
	if flags.Complete {
		components := []string{}
		for name := range installed {
			components = append(components, name)
		}
		return components, nil
	}

	// Explicit components specified
	if len(flags.Components) > 0 {
		// Validate that specified components are installed
		invalidComponents := []string{}
		for _, c := range flags.Components {
			if _, ok := installed[c]; !ok {
				invalidComponents = append(invalidComponents, c)
			}
		}
		if len(invalidComponents) > 0 {
			log.Errorf("Components not installed: %v", invalidComponents)
			return nil, fmt.Errorf("components not installed")
		}
		return flags.Components, nil
	}

	// Interactive selection
	return interactiveUninstallSelection(installed)
}

func interactiveUninstallSelection(installed map[string]string) ([]string, error) {
	if len(installed) == 0 {
		return []string{}, nil
	}

	fmt.Printf("\n%sUninstall Options:%s\n", ui.ColorCyan, ui.ColorReset)

	// Create menu options
	presetOptions := []string{
		"Complete Uninstall (remove everything)",
		"Remove Specific Components",
		"Cancel Uninstall",
	}

	menu := ui.NewMenu("Select uninstall option:", presetOptions, false)
	result, err := menu.Display()
	if err != nil {
		// If menu fails (e.g., no stdin in test environment), default to complete uninstall
		components := []string{}
		for name := range installed {
			components = append(components, name)
		}
		return components, nil
	}
	choice := result.(int)

	switch choice {
	case -1, 2: // Cancelled
		return nil, fmt.Errorf("cancelled")
	case 0: // Complete uninstall
		components := []string{}
		for name := range installed {
			components = append(components, name)
		}
		return components, nil
	case 1: // Select specific components
		componentOptions := []string{}
		componentNames := []string{}

		for component, version := range installed {
			componentOptions = append(componentOptions, fmt.Sprintf("%s (v%s)", component, version))
			componentNames = append(componentNames, component)
		}

		componentMenu := ui.NewMenu("Select components to uninstall:", componentOptions, true)
		result, err := componentMenu.Display()
		if err != nil {
			return nil, fmt.Errorf("component selection failed: %w", err)
		}
		selections := result.([]int)

		if len(selections) == 0 {
			return nil, fmt.Errorf("no components selected")
		}

		selected := []string{}
		for _, i := range selections {
			selected = append(selected, componentNames[i])
		}
		return selected, nil
	}

	return nil, fmt.Errorf("invalid selection")
}

func displayUninstallPlan(components []string, flags UninstallFlags, info map[string]interface{}) {
	fmt.Printf("\n%s%sUninstall Plan%s\n", ui.ColorCyan, ui.ColorBright, ui.ColorReset)
	fmt.Println(strings.Repeat("=", 50))

	fmt.Printf("%sInstallation Directory:%s %s\n", ui.ColorBlue, ui.ColorReset, info["install_dir"])

	installedComponents := info["components"].(map[string]string)
	if len(components) > 0 {
		fmt.Printf("%sComponents to remove:%s\n", ui.ColorBlue, ui.ColorReset)
		for i, componentName := range components {
			version := installedComponents[componentName]
			fmt.Printf("  %d. %s (v%s)\n", i+1, componentName, version)
		}
	}

	// Show what will be preserved
	preserved := []string{}
	if flags.KeepBackups {
		preserved = append(preserved, "backup files")
	}
	if flags.KeepLogs {
		preserved = append(preserved, "log files")
	}
	if flags.KeepSettings {
		preserved = append(preserved, "user settings")
	}

	if len(preserved) > 0 {
		fmt.Printf("%sWill preserve:%s %s\n", ui.ColorGreen, ui.ColorReset, strings.Join(preserved, ", "))
	}

	if flags.Complete {
		fmt.Printf("%sWARNING: Complete uninstall will remove all Claude Code Super Crew files%s\n", ui.ColorRed, ui.ColorReset)
	}

	fmt.Println()
}

func createUninstallBackup(installDir string, components []string) {
	log := logger.GetLogger()

	// Simple backup creation
	backupDir := filepath.Join(installDir, ".crew", "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		log.Warnf("Could not create backup directory: %v", err)
		return
	}

	// Note: Backup implementation deferred - use 'crew backup --create' before uninstall for manual backup
	// Automated backup would require additional configuration to avoid duplicating backup system
	log.Info("Creating uninstall backup placeholder...")
	log.Warn("Automated backup not implemented - use 'crew backup --create' before uninstalling")
}

func performUninstall(components []string, flags UninstallFlags, info map[string]interface{}) bool {
	log := logger.GetLogger()

	// Setup progress tracking
	progress := ui.NewProgressBar(len(components), 50, "Uninstalling: ", "")

	// Uninstall components using simplified approach
	log.Infof("Uninstalling %d components...", len(components))

	success := true
	installDir := globalFlags.InstallDir

	for i, component := range components {
		progress.Update(i+1, fmt.Sprintf("Uninstalling %s", component))

		if globalFlags.DryRun {
			log.Infof("[DRY RUN] Would remove component: %s", component)
			continue
		}

		// Remove component based on known structure
		componentPath := filepath.Join(installDir, component)
		if component == "Core" {
			// Core files are installed directly in the root, not in a subdirectory
			// Only remove files that we explicitly installed and tracked
			coreFiles := []string{
				"CLAUDE.md", "COMMANDS.md", "FLAGS.md", "MCP.md",
				"MODES.md", "ORCHESTRATOR.md", "PERSONAS.md",
				"PRINCIPLES.md", "RULES.md",
			}

			// Load metadata to check which files we actually installed
			settingsManager := managers.NewSettingsManager(installDir)
			metadataManager := settingsManager.GetMetadataManager()
			metadata, _ := metadataManager.LoadMetadata()

			removedCount := 0
			for _, file := range coreFiles {
				filePath := filepath.Join(installDir, file)

				// Check if file exists and was tracked by our framework
				shouldRemove := false
				if _, err := os.Stat(filePath); err == nil {
					if metadata != nil && metadata.Documents != nil {
						// Check if this file is tracked in our documents
						if docMeta, exists := metadata.Documents[file]; exists && docMeta.Status == "present" {
							shouldRemove = true
						}
					} else {
						// Fallback: only remove if it's a known framework file
						shouldRemove = isKnownFrameworkFile(file, "core")
					}
				}

				if shouldRemove {
					if err := os.Remove(filePath); err != nil {
						log.Warnf("Failed to remove %s: %v", file, err)
					} else {
						log.Infof("Removed tracked framework file: %s", file)
						removedCount++
					}
				} else {
					log.Infof("Preserved user file: %s", file)
				}
			}

			if removedCount > 0 {
				log.Infof("Removed %d tracked framework files from core component", removedCount)
			} else {
				log.Infof("No tracked framework files found in core component")
			}
		} else {
			// Other components are in subdirectories - use selective removal
			if _, err := os.Stat(componentPath); err == nil {
				if err := removeCrewFilesFromDirectory(componentPath); err != nil {
					log.Errorf("Failed to selectively remove component %s: %v", component, err)
					success = false
				} else {
					log.Infof("Selectively removed component: %s", component)
				}
			}
		}
	}

	progress.Finish("Uninstall complete")

	// Handle complete uninstall cleanup
	if flags.Complete && !globalFlags.DryRun {
		cleanupInstallationDirectory(globalFlags.InstallDir, flags)
	}

	return success
}

func cleanupInstallationDirectory(installDir string, flags UninstallFlags) {
	log := logger.GetLogger()

	// Use selective removal based on metadata tracking instead of removing entire directory
	if flags.Complete {
		selectiveRemoveTrackedFiles(installDir, flags)
		return
	}

	// Selective removal based on preservation flags (for partial uninstalls)
	itemsToRemove := []string{}

	// Determine what to remove
	if !flags.KeepBackups {
		itemsToRemove = append(itemsToRemove, filepath.Join(installDir, ".crew", "backups"))
	}
	if !flags.KeepLogs {
		itemsToRemove = append(itemsToRemove, filepath.Join(installDir, "logs"))
		itemsToRemove = append(itemsToRemove, filepath.Join(installDir, ".crew", "logs"))
	}
	if !flags.KeepSettings {
		itemsToRemove = append(itemsToRemove, filepath.Join(installDir, "settings.json"))
		itemsToRemove = append(itemsToRemove, filepath.Join(installDir, ".crew", "config"))
	}

	// Remove selected items (only crew-created items)
	for _, item := range itemsToRemove {
		if _, err := os.Stat(item); err == nil {
			if err := os.RemoveAll(item); err != nil {
				log.Warnf("Could not remove %s: %v", item, err)
			} else {
				log.Infof("Removed %s", item)
			}
		}
	}

	// Remove .crew directory only if it's empty of user files
	cleanupCrewDirectory(installDir)
}

// selectiveRemoveTrackedFiles removes only files and directories that were created by crew using inventory
func selectiveRemoveTrackedFiles(installDir string, flags UninstallFlags) {
	log := logger.GetLogger()

	// Load metadata to get inventory of created files
	settingsManager := managers.NewSettingsManager(installDir)
	metadataManager := settingsManager.GetMetadataManager()
	metadata, err := metadataManager.LoadMetadata()
	if err != nil {
		log.Warnf("Could not load metadata for selective removal: %v", err)
		// Fallback to pattern-based cleanup if metadata unavailable
		fallbackPatternBasedRemoval(installDir, flags)
		return
	}

	// Use inventory if available, otherwise fall back to pattern matching
	if len(metadata.Inventory.CreatedFiles) > 0 || len(metadata.Inventory.CreatedDirectories) > 0 {
		log.Infof("Using inventory-based removal (%d files, %d directories tracked)",
			len(metadata.Inventory.CreatedFiles), len(metadata.Inventory.CreatedDirectories))

		// Remove tracked files from inventory
		for _, relPath := range metadata.Inventory.CreatedFiles {
			fullPath := filepath.Join(installDir, relPath)
			if _, err := os.Stat(fullPath); err == nil {
				if err := os.Remove(fullPath); err != nil {
					log.Warnf("Could not remove tracked file %s: %v", relPath, err)
				} else {
					log.Infof("Removed tracked file: %s", relPath)
				}
			}
		}

		// Remove tracked directories from inventory (in reverse order to handle nested directories)
		for i := len(metadata.Inventory.CreatedDirectories) - 1; i >= 0; i-- {
			relPath := metadata.Inventory.CreatedDirectories[i]
			fullPath := filepath.Join(installDir, relPath)
			if _, err := os.Stat(fullPath); err == nil {
				// Check if directory is empty or only contains user files
				if canSafelyRemoveDirectory(fullPath) {
					if err := os.RemoveAll(fullPath); err != nil {
						log.Warnf("Could not remove tracked directory %s: %v", relPath, err)
					} else {
						log.Infof("Removed tracked directory: %s", relPath)
					}
				} else {
					log.Infof("Preserved directory %s (contains user files)", relPath)
				}
			}
		}
	} else {
		log.Infof("No inventory found, falling back to pattern-based removal")
		fallbackPatternBasedRemoval(installDir, flags)
	}

	// Handle preservation flags for crew-managed files
	if !flags.KeepSettings {
		settingsPath := filepath.Join(installDir, "settings.json")
		if _, err := os.Stat(settingsPath); err == nil {
			if err := os.Remove(settingsPath); err != nil {
				log.Warnf("Could not remove settings.json: %v", err)
			} else {
				log.Infof("Removed settings.json")
			}
		}
	}

	// Clean up .crew directory structure
	cleanupCrewDirectory(installDir)

	// Remove any empty directories we created (but preserve user content)
	cleanupEmptyDirectories(installDir)
}

// canSafelyRemoveDirectory checks if a directory can be safely removed
func canSafelyRemoveDirectory(dirPath string) bool {
	log := logger.GetLogger()

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		log.Warnf("Could not read directory %s: %v", dirPath, err)
		return false
	}

	// If directory is empty, it's safe to remove
	if len(entries) == 0 {
		return true
	}

	// For now, if directory has any content, preserve it
	// This is conservative but safer - we could enhance this later
	// to check if all contents are also in the inventory
	return false
}

// fallbackPatternBasedRemoval provides fallback removal when inventory is not available
func fallbackPatternBasedRemoval(installDir string, flags UninstallFlags) {
	log := logger.GetLogger()

	log.Infof("Using fallback pattern-based removal")

	// Load metadata for documents and components
	settingsManager := managers.NewSettingsManager(installDir)
	metadataManager := settingsManager.GetMetadataManager()
	metadata, err := metadataManager.LoadMetadata()
	if err != nil {
		log.Warnf("Could not load metadata, skipping pattern-based removal: %v", err)
		return
	}

	// Remove tracked documents (files we know we created)
	if metadata.Documents != nil {
		for docPath, docMeta := range metadata.Documents {
			if docMeta.Status == "present" {
				fullPath := filepath.Join(installDir, docPath)
				if _, err := os.Stat(fullPath); err == nil {
					if err := os.Remove(fullPath); err != nil {
						log.Warnf("Could not remove tracked file %s: %v", docPath, err)
					} else {
						log.Infof("Removed tracked file: %s", docPath)
					}
				}
			}
		}
	}

	// Remove component directories using pattern matching
	if metadata.Components != nil {
		for componentName, componentMeta := range metadata.Components {
			if componentMeta.Status == "installed" {
				// Skip Core component as its files are tracked individually via Documents
				if componentName == "Core" {
					continue
				}

				componentPath := filepath.Join(installDir, componentName)
				if _, err := os.Stat(componentPath); err == nil {
					// Use pattern-based removal for component directories
					if err := removeCrewFilesFromDirectory(componentPath); err != nil {
						log.Warnf("Could not remove files from component directory %s: %v", componentName, err)
					}
				}
			}
		}
	}
}

// removeCrewOwnedDirectory recursively removes only crew-created files from a directory
func removeCrewOwnedDirectory(dirPath string) error {
	log := logger.GetLogger()

	// Check if directory exists
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return nil // Already gone
	}

	// For safety, only process directories that match known crew patterns
	dirName := filepath.Base(dirPath)
	crewOwnedDirs := []string{"agents", "commands", "hooks", "mcp", "core"}

	isCrewOwned := false
	for _, ownedDir := range crewOwnedDirs {
		if dirName == ownedDir {
			isCrewOwned = true
			break
		}
	}

	if !isCrewOwned {
		log.Warnf("Skipping removal of non-crew directory: %s", dirPath)
		return nil
	}

	// Instead of removing entire directory, selectively remove only crew files
	return removeCrewFilesFromDirectory(dirPath)
}

// removeCrewFilesFromDirectory removes only crew-created files from a directory
// This is a conservative approach that preserves user-created content
func removeCrewFilesFromDirectory(dirPath string) error {
	log := logger.GetLogger()

	// Get list of crew-created files based on known patterns
	crewFilePatterns := getCrewFilePatterns(filepath.Base(dirPath))

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", dirPath, err)
	}

	removedCount := 0
	userFileCount := 0

	for _, entry := range entries {
		entryPath := filepath.Join(dirPath, entry.Name())

		if entry.IsDir() {
			// For subdirectories, check if they are crew-created
			if isCrewCreatedSubdirectory(entry.Name(), filepath.Base(dirPath)) {
				if err := os.RemoveAll(entryPath); err != nil {
					log.Warnf("Failed to remove crew subdirectory %s: %v", entryPath, err)
				} else {
					log.Infof("Removed crew subdirectory: %s", entryPath)
					removedCount++
				}
			} else {
				log.Infof("Preserved user subdirectory: %s", entryPath)
				userFileCount++
			}
		} else {
			// For files, check if they match crew patterns or are known framework files
			isCrewFile := isCrewCreatedFile(entry.Name(), crewFilePatterns)
			isFrameworkFile := isKnownFrameworkFile(entry.Name(), filepath.Base(dirPath))

			log.Debugf("File %s: isCrewFile=%v, isFrameworkFile=%v, patterns=%v, component=%s",
				entry.Name(), isCrewFile, isFrameworkFile, crewFilePatterns, filepath.Base(dirPath))

			if isCrewFile || isFrameworkFile {
				if err := os.Remove(entryPath); err != nil {
					log.Warnf("Failed to remove crew file %s: %v", entryPath, err)
				} else {
					log.Infof("Removed crew file: %s", entryPath)
					removedCount++
				}
			} else {
				log.Infof("Preserved user file: %s", entryPath)
				userFileCount++
			}
		}
	}

	// NEVER remove the directory itself - it may contain user data
	// Only log what we found
	if removedCount > 0 {
		if userFileCount > 0 {
			log.Infof("Removed %d crew files, preserved %d user files in directory: %s",
				removedCount, userFileCount, dirPath)
		} else {
			log.Infof("Removed %d crew files from directory: %s", removedCount, dirPath)
		}
	} else {
		log.Infof("No crew files found in directory: %s", dirPath)
	}

	return nil
}

// getCrewFilePatterns returns file patterns that crew creates for each component
// This is used as a fallback when inventory tracking is not available
func getCrewFilePatterns(componentName string) []string {
	switch componentName {
	case "agents":
		return []string{
			"*-persona.md",
			"orchestrator.agent.md",
			"second-opinion-generator.md",
			".version",
		}
	case "commands":
		return []string{
			"analyze.md",
			"build.md",
			"cleanup.md",
			"design.md",
			"document.md",
			"estimate.md",
			"explain.md",
			"git.md",
			"implement.md",
			"improve.md",
			"index.md",
			"load.md",
			"spawn.md",
			"task.md",
			"test.md",
			"troubleshoot.md",
			"workflow.md",
			".version",
		}
	case "hooks":
		return []string{
			"backup-before-change.sh",
			"git-auto-commit.sh",
			"lint-on-save.sh",
			"security-scan.sh",
			"test-on-change.sh",
			"README.md",
			".version",
		}
	case "mcp":
		return []string{
			"*.json",
			"*.yaml",
			"*.yml",
		}
	case "core":
		return []string{
			"CLAUDE.md",
			"COMMANDS.md",
			"FLAGS.md",
			"MCP.md",
			"MODES.md",
			"ORCHESTRATOR.md",
			"PERSONAS.md",
			"PRINCIPLES.md",
			"RULES.md",
		}
	default:
		return []string{}
	}
}

// isCrewCreatedFile checks if a file matches crew creation patterns
func isCrewCreatedFile(filename string, patterns []string) bool {
	for _, pattern := range patterns {
		if matched, _ := filepath.Match(pattern, filename); matched {
			return true
		}
	}
	return false
}

// isKnownFrameworkFile checks if a file is a known framework file that should be removed
// even if it's not tracked in the inventory or doesn't match patterns
func isKnownFrameworkFile(filename, componentName string) bool {
	switch componentName {
	case "hooks":
		// Known framework files for hooks component - ONLY files we actually install
		knownFiles := []string{
			"README.md",
			".version",
			"backup-before-change.sh",
			"git-auto-commit.sh",
			"lint-on-save.sh",
			"security-scan.sh",
			"test-on-change.sh",
		}
		for _, knownFile := range knownFiles {
			if filename == knownFile {
				return true
			}
		}
	case "agents":
		// Known framework files for agents component
		knownFiles := []string{
			".version",
			"analyzer-persona.md",
			"architect-persona.md",
			"backend-persona.md",
			"devops-persona.md",
			"frontend-persona.md",
			"mentor-persona.md",
			"orchestrator.agent.md",
			"performance-persona.md",
			"qa-persona.md",
			"refactorer-persona.md",
			"scribe-persona.md",
			"second-opinion-generator.md",
			"security-persona.md",
		}
		for _, knownFile := range knownFiles {
			if filename == knownFile {
				return true
			}
		}
	case "commands":
		// Known framework files for commands component
		knownFiles := []string{
			".version",
			"analyze.md",
			"build.md",
			"cleanup.md",
			"design.md",
			"document.md",
			"estimate.md",
			"explain.md",
			"git.md",
			"implement.md",
			"improve.md",
			"index.md",
			"load.md",
			"spawn.md",
			"task.md",
			"test.md",
			"troubleshoot.md",
			"workflow.md",
		}
		for _, knownFile := range knownFiles {
			if filename == knownFile {
				return true
			}
		}
	case "core":
		// Known framework files for core component
		knownFiles := []string{
			"CLAUDE.md", "COMMANDS.md", "FLAGS.md", "MCP.md",
			"MODES.md", "ORCHESTRATOR.md", "PERSONAS.md",
			"PRINCIPLES.md", "RULES.md",
		}
		for _, knownFile := range knownFiles {
			if filename == knownFile {
				return true
			}
		}
	}
	return false
}

// isCrewCreatedSubdirectory checks if a subdirectory was created by crew
func isCrewCreatedSubdirectory(subdirName, componentName string) bool {
	switch componentName {
	case "agents":
		// Only the "templates" subdirectory is crew-created in agents
		return subdirName == "templates"
	case "commands":
		// All subdirectories in commands are crew-created
		return true
	case "hooks":
		// Common hook subdirectories
		return subdirName == "pre-commit" || subdirName == "post-commit" ||
			subdirName == "pre-push" || subdirName == "post-push"
	case "mcp":
		// MCP server configurations
		return subdirName == "servers" || subdirName == "config"
	case "core":
		// Core typically doesn't have subdirectories
		return false
	default:
		return false
	}
}

// cleanupCrewDirectory handles .crew directory cleanup
func cleanupCrewDirectory(installDir string) {
	log := logger.GetLogger()
	crewDir := filepath.Join(installDir, ".crew")

	if _, err := os.Stat(crewDir); os.IsNotExist(err) {
		return // Already gone
	}

	// Remove known crew subdirectories
	crewSubdirs := []string{"config", "backups", "logs", "workflows", "scripts", "prompts", "completions"}
	for _, subdir := range crewSubdirs {
		subdirPath := filepath.Join(crewDir, subdir)
		if _, err := os.Stat(subdirPath); err == nil {
			if err := os.RemoveAll(subdirPath); err != nil {
				log.Warnf("Could not remove .crew subdirectory %s: %v", subdir, err)
			} else {
				log.Infof("Removed .crew subdirectory: %s", subdir)
			}
		}
	}

	// Try to remove .crew directory itself if it's empty
	if isEmpty, err := isDirEmpty(crewDir); err == nil && isEmpty {
		if err := os.Remove(crewDir); err != nil {
			log.Warnf("Could not remove empty .crew directory: %v", err)
		} else {
			log.Infof("Removed empty .crew directory")
		}
	} else if err != nil {
		log.Warnf("Could not check if .crew directory is empty: %v", err)
	} else {
		log.Infof("Preserved .crew directory (contains user files)")
	}
}

// cleanupEmptyDirectories removes any empty directories that were created by crew
func cleanupEmptyDirectories(installDir string) {
	log := logger.GetLogger()

	// Only try to remove the install directory itself if it's completely empty
	if isEmpty, err := isDirEmpty(installDir); err == nil && isEmpty {
		if err := os.Remove(installDir); err != nil {
			log.Warnf("Could not remove empty installation directory: %v", err)
		} else {
			log.Infof("Removed empty installation directory: %s", installDir)
		}
	}
}

// isDirEmpty checks if a directory is empty
func isDirEmpty(dirPath string) (bool, error) {
	f, err := os.Open(dirPath)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == nil {
		return false, nil // Directory has at least one entry
	}
	if os.IsNotExist(err) || err.Error() == "EOF" {
		return true, nil // Directory is empty
	}
	return false, err
}
