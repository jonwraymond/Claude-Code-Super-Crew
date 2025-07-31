package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jonwraymond/claude-code-super-crew/internal/core"
	"github.com/jonwraymond/claude-code-super-crew/internal/installer"
	"github.com/jonwraymond/claude-code-super-crew/internal/managers"
	"github.com/jonwraymond/claude-code-super-crew/internal/ui"
	"github.com/jonwraymond/claude-code-super-crew/internal/versioning"
	"github.com/jonwraymond/claude-code-super-crew/pkg/logger"
	"github.com/spf13/cobra"
)

// UpdateFlags holds update command flags
type UpdateFlags struct {
	Check      bool
	Components []string
	Backup     bool
	NoBackup   bool
	Reinstall  bool
}

var updateFlags UpdateFlags

// NewUpdateCommand creates the update command
func NewUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update existing Claude Code Super Crew installation",
		Long: `Update Claude Code Super Crew Framework components to latest versions.

Examples:
  crew update                       # Interactive update
  crew update --check --verbose     # Check for updates (verbose)
  crew update --components core mcp # Update specific components
  crew update --backup --force      # Create backup before update (forced)`,
		RunE: runUpdate,
	}

	// Update mode options
	cmd.Flags().BoolVar(&updateFlags.Check, "check", false,
		"Check for available updates without installing")
	cmd.Flags().StringSliceVar(&updateFlags.Components, "components", nil,
		"Specific components to update")

	// Backup options
	cmd.Flags().BoolVar(&updateFlags.Backup, "backup", false,
		"Create backup before update")
	cmd.Flags().BoolVar(&updateFlags.NoBackup, "no-backup", false,
		"Skip backup creation")

	// Update options
	cmd.Flags().BoolVar(&updateFlags.Reinstall, "reinstall", false,
		"Reinstall components even if versions match")

	return cmd
}

func runUpdate(cmd *cobra.Command, args []string) error {
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
			"Claude Code Super Crew Update v1.0",
			"Updating Claude Code Super Crew framework components",
		)
	}

	// Check if Claude Code Super Crew is installed
	settingsManager := managers.NewSettingsManager(globalFlags.InstallDir)
	if !settingsManager.CheckInstallationExists() {
		log.Errorf("Claude Code Super Crew installation not found in %s", globalFlags.InstallDir)
		log.Info("Use 'crew install' to install Claude Code Super Crew first")
		return fmt.Errorf("no installation found")
	}

	// Initialize components
	log.Info("Checking for available updates...")

	exe, _ := os.Executable()
	projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(exe)))

	registry := core.NewEnhancedComponentRegistry(filepath.Join(projectRoot, "setup", "components"))
	if err := registry.DiscoverComponents(); err != nil {
		return fmt.Errorf("failed to discover components: %w", err)
	}

	// Get installed components
	installedComponents, err := settingsManager.GetInstalledComponents()
	if err != nil || len(installedComponents) == 0 {
		log.Error("Could not determine installed components")
		return fmt.Errorf("could not determine installed components")
	}

	// Check for available updates
	availableUpdates := getAvailableUpdates(installedComponents, registry)

	// Display update check results
	if !globalFlags.Quiet {
		displayUpdateCheck(installedComponents, availableUpdates)
	}

	// If only checking for updates, exit here
	if updateFlags.Check {
		return nil
	}

	// Get components to update
	components, err := getComponentsToUpdate(updateFlags, installedComponents, availableUpdates)
	if err != nil {
		return err
	}

	if len(components) == 0 {
		log.Info("No components selected for update")
		return nil
	}

	// Display update plan
	if !globalFlags.Quiet {
		displayUpdatePlan(components, availableUpdates, installedComponents, globalFlags.InstallDir)

		if !globalFlags.DryRun {
			if !globalFlags.Yes && !ui.Confirm("Proceed with update?", true) {
				log.Info("Update cancelled by user")
				return nil
			}
		}
	}

	// Perform update
	success := performUpdate(components, updateFlags)

	if success {
		if !globalFlags.Quiet {
			ui.DisplaySuccess("Claude Code Super Crew update completed successfully!")

			if !globalFlags.DryRun {
				fmt.Printf("\n%sNext steps:%s\n", ui.ColorCyan, ui.ColorReset)
				fmt.Println("1. Restart your Claude Code session")
				fmt.Println("2. Updated components are now available")
				fmt.Println("3. Check for any breaking changes in documentation")
			}
		}
		return nil
	} else {
		ui.DisplayError("Update failed. Check logs for details.")
		return fmt.Errorf("update failed")
	}
}

func getAvailableUpdates(installed map[string]string, registry *core.EnhancedComponentRegistry) map[string]map[string]string {
	updates := make(map[string]map[string]string)
	versionManager := versioning.NewVersionManager(globalFlags.InstallDir)

	for componentName, currentVersion := range installed {
		metadata := registry.GetComponentMetadata(componentName)
		if metadata != nil {
			availableVersion := metadata.Version
			// Use version manager to properly compare versions
			comparison := versionManager.CompareVersions(currentVersion, availableVersion)
			if comparison < 0 { // current version is older
				updates[componentName] = map[string]string{
					"current":     currentVersion,
					"available":   availableVersion,
					"description": metadata.Description,
				}
			}
		}
	}

	return updates
}

func displayUpdateCheck(installed map[string]string, updates map[string]map[string]string) {
	fmt.Printf("\n%s%sUpdate Check Results%s\n", ui.ColorCyan, ui.ColorBright, ui.ColorReset)
	fmt.Println(strings.Repeat("=", 50))

	if len(installed) == 0 {
		fmt.Printf("%sNo Claude Code Super Crew installation found%s\n", ui.ColorYellow, ui.ColorReset)
		return
	}

	fmt.Printf("%sCurrently installed components:%s\n", ui.ColorBlue, ui.ColorReset)
	for component, version := range installed {
		fmt.Printf("  %s: v%s\n", component, version)
	}

	if len(updates) > 0 {
		fmt.Printf("\n%sAvailable updates:%s\n", ui.ColorGreen, ui.ColorReset)
		for component, info := range updates {
			fmt.Printf("  %s: v%s → v%s\n", component, info["current"], info["available"])
			fmt.Printf("    %s\n", info["description"])
		}
	} else {
		fmt.Printf("\n%sAll components are up to date%s\n", ui.ColorGreen, ui.ColorReset)
	}

	fmt.Println()
}

func getComponentsToUpdate(flags UpdateFlags, installed map[string]string, updates map[string]map[string]string) ([]string, error) {
	log := logger.GetLogger()

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

	// If no updates available and not forcing reinstall
	if len(updates) == 0 && !flags.Reinstall {
		log.Info("No updates available")
		return []string{}, nil
	}

	// Interactive selection
	if len(updates) > 0 {
		return interactiveUpdateSelection(updates, installed)
	} else if flags.Reinstall {
		// Reinstall all components
		components := []string{}
		for name := range installed {
			components = append(components, name)
		}
		return components, nil
	}

	return []string{}, nil
}

func interactiveUpdateSelection(updates map[string]map[string]string, installed map[string]string) ([]string, error) {
	if len(updates) == 0 {
		return []string{}, nil
	}

	fmt.Printf("\n%sAvailable Updates:%s\n", ui.ColorCyan, ui.ColorReset)

	// Create menu options
	updateOptions := []string{}
	componentNames := []string{}

	for component, info := range updates {
		updateOptions = append(updateOptions, fmt.Sprintf("%s: v%s → v%s", component, info["current"], info["available"]))
		componentNames = append(componentNames, component)
	}

	// Add bulk options
	presetOptions := []string{
		"Update All Components",
		"Select Individual Components",
		"Cancel Update",
	}

	menu := ui.NewMenu("Select update option:", presetOptions, false)
	result, err := menu.Display()
	if err != nil {
		return nil, fmt.Errorf("menu selection failed: %w", err)
	}
	choice := result.(int)

	switch choice {
	case -1, 2: // Cancelled
		return nil, fmt.Errorf("cancelled")
	case 0: // Update all
		return componentNames, nil
	case 1: // Select individual
		componentMenu := ui.NewMenu("Select components to update:", updateOptions, true)
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

func displayUpdatePlan(components []string, updates map[string]map[string]string, installed map[string]string, installDir string) {
	fmt.Printf("\n%s%sUpdate Plan%s\n", ui.ColorCyan, ui.ColorBright, ui.ColorReset)
	fmt.Println(strings.Repeat("=", 50))

	fmt.Printf("%sInstallation Directory:%s %s\n", ui.ColorBlue, ui.ColorReset, installDir)
	fmt.Printf("%sComponents to update:%s\n", ui.ColorBlue, ui.ColorReset)

	for i, componentName := range components {
		if info, ok := updates[componentName]; ok {
			fmt.Printf("  %d. %s: v%s → v%s\n", i+1, componentName, info["current"], info["available"])
		} else {
			currentVersion := installed[componentName]
			fmt.Printf("  %d. %s: v%s (reinstall)\n", i+1, componentName, currentVersion)
		}
	}

	fmt.Println()
}

func performUpdate(components []string, flags UpdateFlags) bool {
	log := logger.GetLogger()

	exe, _ := os.Executable()
	projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(exe)))

	// Create installer
	inst := installer.NewInstaller(globalFlags.InstallDir, globalFlags.DryRun)

	// Create component registry
	registry := core.NewEnhancedComponentRegistry(filepath.Join(projectRoot, "setup", "components"))
	if err := registry.DiscoverComponents(); err != nil {
		log.Errorf("Failed to discover components: %v", err)
		return false
	}

	// Create component instances
	componentInstances, err := registry.CreateComponentInstances(components, globalFlags.InstallDir)
	if err != nil {
		log.Errorf("Failed to create component instances: %v", err)
		return false
	}

	if len(componentInstances) == 0 {
		log.Error("No valid component instances created")
		return false
	}

	// Register components with installer
	compList := []core.Component{}
	for _, comp := range componentInstances {
		compList = append(compList, comp)
	}
	inst.RegisterComponents(compList)

	// Setup progress tracking
	progress := ui.NewProgressBar(len(components), 50, "Updating: ", "")

	// Update components
	log.Infof("Updating %d components...", len(components))

	// Determine backup strategy
	backup := flags.Backup || (!flags.NoBackup && !globalFlags.DryRun)

	config := map[string]interface{}{
		"force":       globalFlags.Force,
		"backup":      backup,
		"dry_run":     globalFlags.DryRun,
		"update_mode": true,
	}

	success := inst.UpdateComponents(components, config)

	// Update progress
	summary := inst.GetUpdateSummary()
	updated := summary["updated"].([]string)
	failed := summary["failed"].([]string)

	for i, name := range components {
		if contains(updated, name) {
			progress.Update(i+1, fmt.Sprintf("Updated %s", name))
		} else {
			progress.Update(i+1, fmt.Sprintf("Failed %s", name))
		}
	}

	progress.Finish("Update complete")

	// Show results
	if success {
		if len(updated) > 0 {
			log.Infof("Updated components: %s", strings.Join(updated, ", "))
		}

		if backupPath, ok := summary["backup_path"].(string); ok && backupPath != "" {
			log.Infof("Backup created: %s", backupPath)
		}
	} else {
		if len(failed) > 0 {
			log.Errorf("Failed components: %s", strings.Join(failed, ", "))
		}
	}

	return success
}
