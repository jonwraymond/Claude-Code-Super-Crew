package cli

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jonwraymond/claude-code-super-crew/internal/core"
	"github.com/jonwraymond/claude-code-super-crew/internal/managers"
	"github.com/jonwraymond/claude-code-super-crew/internal/ui"
	"github.com/jonwraymond/claude-code-super-crew/internal/versioning"
	"github.com/jonwraymond/claude-code-super-crew/pkg/logger"
	"github.com/spf13/cobra"
)

// InstallFlags holds install command flags
type InstallFlags struct {
	Quick           bool
	Minimal         bool
	Profile         string
	Components      []string
	NoBackup        bool
	ListComponents  bool
	Diagnose        bool
	ClaudeMerge     bool
	ClaudeOverwrite bool
	ClaudeSkip      bool
}

var installFlags InstallFlags

// NewInstallCommand creates the install command
func NewInstallCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install Claude Code Super Crew framework components",
		Long: `Install Claude Code Super Crew Framework with various options and profiles.

Examples:
  crew install                          # Interactive installation
  crew install --quick --dry-run        # Quick installation (dry-run)
  crew install --profile developer      # Developer profile  
  crew install --components core mcp    # Specific components
  crew install --verbose --force        # Verbose with force mode
  crew install --claude-merge           # Merge existing CLAUDE.md
  crew install --claude-skip --yes      # Auto-install, keep existing CLAUDE.md`,
		RunE: runInstall,
	}

	// Register install flags
	cmd.Flags().BoolVar(&installFlags.Quick, "quick", false,
		"Quick installation with pre-selected components")
	cmd.Flags().BoolVar(&installFlags.Minimal, "minimal", false,
		"Minimal installation (core only)")
	cmd.Flags().StringVar(&installFlags.Profile, "profile", "",
		"Installation profile (quick, minimal, developer, etc.)")
	cmd.Flags().StringSliceVar(&installFlags.Components, "components", nil,
		"Specific components to install")
	cmd.Flags().BoolVar(&installFlags.NoBackup, "no-backup", false,
		"Skip backup creation")
	cmd.Flags().BoolVar(&installFlags.ListComponents, "list-components", false,
		"List available components and exit")
	cmd.Flags().BoolVar(&installFlags.Diagnose, "diagnose", false,
		"Run system diagnostics and show installation help")

	// CLAUDE.md handling flags
	cmd.Flags().BoolVar(&installFlags.ClaudeMerge, "claude-merge", false,
		"Merge existing CLAUDE.md with new version (preserves custom sections)")
	cmd.Flags().BoolVar(&installFlags.ClaudeOverwrite, "claude-overwrite", false,
		"Overwrite existing CLAUDE.md with new version")
	cmd.Flags().BoolVar(&installFlags.ClaudeSkip, "claude-skip", false,
		"Skip CLAUDE.md installation if it already exists")

	return cmd
}

func runInstall(cmd *cobra.Command, args []string) error {
	log := logger.GetLogger()
	gFlags := GetGlobalFlags()
	log.SetVerbose(gFlags.Verbose)
	log.SetQuiet(gFlags.Quiet)

	// Validate installation directory (skip in test mode)
	if !testMode {
		expectedHome := filepath.Join(os.Getenv("HOME"))
		if expectedHome == "" {
			expectedHome = filepath.Join(os.Getenv("USERPROFILE")) // Windows
		}
		actualDir, _ := filepath.Abs(gFlags.InstallDir)

		if !strings.HasPrefix(actualDir, expectedHome) {
			ui.DisplayError("Installation must be inside your user profile directory.")
			fmt.Printf("    Expected prefix: %s\n", expectedHome)
			fmt.Printf("    Provided path:   %s\n", actualDir)
			return fmt.Errorf("invalid installation directory")
		}
	}

	// Display header
	if !gFlags.Quiet {
		ui.DisplayHeader(
			"Claude Code Super Crew Installation v1.0",
			"Installing Claude Code Super Crew framework components",
		)
	}

	// Handle special modes
	if installFlags.ListComponents {
		return listAvailableComponents()
	}

	if installFlags.Diagnose {
		return runSystemDiagnostics()
	}

	// Initialize components
	log.Info("Initializing installation system...")

	// Get project root (parent of cmd directory)
	exe, _ := os.Executable()
	projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(exe)))

	registry := core.NewEnhancedComponentRegistry(filepath.Join(projectRoot, "setup", "components"))
	if err := registry.DiscoverComponents(); err != nil {
		return fmt.Errorf("failed to discover components: %w", err)
	}

	configManager, err := managers.NewConfigManager(filepath.Join(projectRoot, "config"), "")
	if err != nil {
		return fmt.Errorf("failed to create config manager: %w", err)
	}
	validator := core.NewValidator()

	// Validate configuration
	configErrors := configManager.ValidateConfigFiles()
	if len(configErrors) > 0 {
		log.Error("Configuration validation failed:")
		for _, err := range configErrors {
			log.Errorf("  - %s", err)
		}
		return fmt.Errorf("configuration validation failed")
	}

	// Get components to install
	components, err := getComponentsToInstall(installFlags, registry, configManager)
	if err != nil {
		return err
	}

	if len(components) == 0 {
		log.Error("No components selected for installation")
		return fmt.Errorf("no components selected")
	}

	// Validate components exist in registry
	availableComponents := registry.ListComponents()
	for _, component := range components {
		found := false
		for _, available := range availableComponents {
			if component == available {
				found = true
				break
			}
		}
		if !found {
			ui.DisplayError(fmt.Sprintf("Could not resolve dependencies: component %s not found in registry", component))
			return fmt.Errorf("invalid component: %s", component)
		}
	}

	// Check for conflicting flags
	if installFlags.Quick && installFlags.Minimal {
		ui.DisplayError("Conflicting flags: --quick and --minimal cannot be used together")
		return fmt.Errorf("conflicting flags: --quick and --minimal")
	}

	// Validate system requirements (skip in dry-run mode)
	if !gFlags.DryRun {
		requirements := configManager.GetRequirementsForComponents(components)
		if !validateSystemRequirements(validator, components, requirements) {
			if !gFlags.Force {
				log.Error("System requirements not met. Use --force to override.")
				return fmt.Errorf("system requirements not met")
			} else {
				log.Warn("System requirements not met, but continuing due to --force flag")
			}
		}
	}

	// Check for existing installation
	if _, err := os.Stat(gFlags.InstallDir); err == nil && !gFlags.Force {
		if !gFlags.DryRun {
			log.Warnf("Installation directory already exists: %s", gFlags.InstallDir)
			if !gFlags.Yes && !ui.Confirm("Continue and update existing installation?", false) {
				log.Info("Installation cancelled by user")
				return nil
			}
		}
	}

	// Display installation plan
	if !gFlags.Quiet {
		displayInstallationPlan(components, registry, gFlags.InstallDir)

		if !gFlags.DryRun {
			if !gFlags.Yes && !ui.Confirm("Proceed with installation?", true) {
				log.Info("Installation cancelled by user")
				return nil
			}
		}
	}

	// Perform installation
	success := performInstallation(components, installFlags, gFlags)

	if success {
		if !gFlags.Quiet {
			ui.DisplaySuccess("Claude Code Super Crew installation completed successfully!")

			if !gFlags.DryRun {
				fmt.Printf("\n%sNext steps:%s\n", ui.ColorCyan, ui.ColorReset)
				fmt.Println("1. Restart your Claude Code session")
				fmt.Printf("2. Framework files are now available in %s\n", gFlags.InstallDir)
				fmt.Println("3. Use Claude Code Super Crew commands and features in Claude Code")
			}
		}
		return nil
	} else {
		ui.DisplayError("Installation failed. Check logs for details.")
		return fmt.Errorf("installation failed")
	}
}

func listAvailableComponents() error {
	exe, _ := os.Executable()
	projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(exe)))

	registry := core.NewEnhancedComponentRegistry(filepath.Join(projectRoot, "setup", "components"))
	if err := registry.DiscoverComponents(); err != nil {
		return fmt.Errorf("failed to discover components: %w", err)
	}

	components := registry.ListComponents()
	if len(components) > 0 {
		fmt.Printf("\n%sAvailable Components:%s\n", ui.ColorCyan, ui.ColorReset)
		for _, name := range components {
			if metadata := registry.GetComponentMetadata(name); metadata != nil {
				fmt.Printf("  %s (%s) - %s\n", name, metadata.Category, metadata.Description)
			} else {
				fmt.Printf("  %s - Unknown component\n", name)
			}
		}
	} else {
		fmt.Println("No components found")
	}

	return nil
}

func runSystemDiagnostics() error {
	validator := core.NewValidator()
	diagnostics := validator.DiagnoseSystem()

	fmt.Printf("\n%s%sClaude Code Super Crew System Diagnostics%s\n", ui.ColorCyan, ui.ColorBright, ui.ColorReset)
	fmt.Println(strings.Repeat("=", 50))

	fmt.Printf("%sPlatform:%s %v\n", ui.ColorBlue, ui.ColorReset, diagnostics["platform"])

	fmt.Printf("\n%sSystem Checks:%s\n", ui.ColorBlue, ui.ColorReset)
	allPassed := true

	checks := diagnostics["checks"].(map[string]map[string]string)
	for checkName, checkInfo := range checks {
		status := checkInfo["status"]
		message := checkInfo["message"]

		if status == "pass" {
			fmt.Printf("  ✅ %s: %s\n", checkName, message)
		} else {
			fmt.Printf("  ❌ %s: %s\n", checkName, message)
			allPassed = false
		}
	}

	issues := diagnostics["issues"].([]string)
	if len(issues) > 0 {
		fmt.Printf("\n%sIssues Found:%s\n", ui.ColorYellow, ui.ColorReset)
		for _, issue := range issues {
			fmt.Printf("  ⚠️  %s\n", issue)
		}

		recommendations := diagnostics["recommendations"].([]string)
		fmt.Printf("\n%sRecommendations:%s\n", ui.ColorCyan, ui.ColorReset)
		for _, rec := range recommendations {
			fmt.Println(rec)
		}
	}

	if allPassed {
		fmt.Printf("\n%s✅ All system checks passed! Your system is ready for Claude Code Super Crew.%s\n", ui.ColorGreen, ui.ColorReset)
	} else {
		fmt.Printf("\n%s⚠️  Some issues found. Please address the recommendations above.%s\n", ui.ColorYellow, ui.ColorReset)
	}

	fmt.Printf("\n%sNext steps:%s\n", ui.ColorBlue, ui.ColorReset)
	if allPassed {
		fmt.Println("  1. Run 'crew install' to proceed with installation")
		fmt.Println("  2. Choose your preferred installation mode (quick, minimal, or custom)")
	} else {
		fmt.Println("  1. Install missing dependencies using the commands above")
		fmt.Println("  2. Restart your terminal after installing tools")
		fmt.Println("  3. Run 'crew install --diagnose' again to verify")
	}

	return nil
}

func getComponentsToInstall(flags InstallFlags, registry *core.EnhancedComponentRegistry, configManager *managers.ConfigManager) ([]string, error) {
	// Explicit components specified
	if len(flags.Components) > 0 {
		if contains(flags.Components, "all") {
			return []string{"core", "commands", "hooks", "mcp"}, nil
		}
		return flags.Components, nil
	}

	// Profile-based selection
	if flags.Profile != "" {
		// For now, use hardcoded profiles
		switch flags.Profile {
		case "quick":
			return []string{"core", "commands", "agents"}, nil
		case "minimal":
			return []string{"core"}, nil
		case "developer":
			return []string{"core", "commands", "hooks", "mcp"}, nil
		default:
			return nil, fmt.Errorf("unknown profile: %s", flags.Profile)
		}
	}

	// Quick installation
	if flags.Quick {
		return []string{"core", "commands", "hooks", "agents"}, nil
	}

	// Minimal installation
	if flags.Minimal {
		return []string{"core"}, nil
	}

	// Interactive selection - but respect --yes flag for automation
	gFlags := GetGlobalFlags()
	if gFlags.Yes {
		// If --yes is set, default to quick installation to avoid interactive prompts
		return []string{"core", "commands", "hooks", "agents"}, nil
	}

	return interactiveComponentSelection(registry)
}

func interactiveComponentSelection(registry *core.EnhancedComponentRegistry) ([]string, error) {
	log := logger.GetLogger()

	availableComponents := registry.ListComponents()
	if len(availableComponents) == 0 {
		log.Error("No components available for installation")
		return nil, fmt.Errorf("no components available")
	}

	// Create preset options
	presetOptions := []string{
		"Quick Installation (recommended components)",
		"Minimal Installation (core only)",
		"Custom Selection",
	}

	fmt.Printf("\n%sSuperCrewInstallation Options:%s\n", ui.ColorCyan, ui.ColorReset)
	menu := ui.NewMenu("Select installation type:", presetOptions, false)
	result, err := menu.Display()
	if err != nil {
		return nil, fmt.Errorf("menu selection failed: %w", err)
	}
	choice := result.(int)

	switch choice {
	case -1: // Cancelled
		return nil, fmt.Errorf("cancelled")
	case 0: // Quick
		return []string{"core", "commands", "hooks", "agents"}, nil
	case 1: // Minimal
		return []string{"core"}, nil
	case 2: // Custom
		// Create component menu with descriptions
		menuOptions := []string{}
		for _, name := range availableComponents {
			if metadata := registry.GetComponentMetadata(name); metadata != nil {
				menuOptions = append(menuOptions, fmt.Sprintf("%s (%s) - %s", name, metadata.Category, metadata.Description))
			} else {
				menuOptions = append(menuOptions, fmt.Sprintf("%s - Component description unavailable", name))
			}
		}

		fmt.Printf("\n%sAvailable Components:%s\n", ui.ColorCyan, ui.ColorReset)
		componentMenu := ui.NewMenu("Select components to install:", menuOptions, true)
		result, err := componentMenu.Display()
		if err != nil {
			return nil, fmt.Errorf("component selection failed: %w", err)
		}
		selections := result.([]int)

		if len(selections) == 0 {
			log.Warn("No components selected")
			return nil, fmt.Errorf("no components selected")
		}

		selected := []string{}
		for _, idx := range selections {
			selected = append(selected, availableComponents[idx])
		}
		return selected, nil
	}

	return nil, fmt.Errorf("invalid selection")
}

func validateSystemRequirements(validator *core.Validator, components []string, requirements map[string]map[string]string) bool {
	log := logger.GetLogger()

	log.Info("Validating system requirements...")

	success, errors := validator.ValidateComponentRequirements(components, requirements)

	if success {
		log.Success("All system requirements met")
		return true
	} else {
		log.Error("System requirements not met:")
		for _, err := range errors {
			log.Errorf("  - %s", err)
		}

		fmt.Printf("\n%s💡 Installation Help:%s\n", ui.ColorCyan, ui.ColorReset)
		fmt.Println("  Run 'crew install --diagnose' for detailed system diagnostics")
		fmt.Println("  and step-by-step installation instructions.")

		return false
	}
}

func displayInstallationPlan(components []string, registry *core.EnhancedComponentRegistry, installDir string) {
	fmt.Printf("\n%s%sInstallation Plan%s\n", ui.ColorCyan, ui.ColorBright, ui.ColorReset)
	fmt.Println(strings.Repeat("=", 50))

	// Resolve dependencies
	orderedComponents, err := registry.ResolveDependencies(components)
	if err != nil {
		logger.GetLogger().Errorf("Could not resolve dependencies: %v", err)
		orderedComponents = components
	}

	fmt.Printf("%sInstallation Directory:%s %s\n", ui.ColorBlue, ui.ColorReset, installDir)
	fmt.Printf("%sComponents to install:%s\n", ui.ColorBlue, ui.ColorReset)

	totalSize := int64(0)
	for i, name := range orderedComponents {
		if metadata := registry.GetComponentMetadata(name); metadata != nil {
			fmt.Printf("  %d. %s - %s\n", i+1, name, metadata.Description)

			// Get size estimate
			if comp, err := registry.GetComponentInstance(name, installDir); err == nil {
				size := comp.GetSizeEstimate()
				totalSize += size
			}
		} else {
			fmt.Printf("  %d. %s - Unknown component\n", i+1, name)
		}
	}

	if totalSize > 0 {
		fmt.Printf("\n%sEstimated size:%s %s\n", ui.ColorBlue, ui.ColorReset, ui.FormatSize(totalSize))
	}

	fmt.Println()
}

func performInstallation(components []string, flags InstallFlags, gFlags *GlobalFlags) bool {
	log := logger.GetLogger()

	// Get project root (where the binary is built from)
	exe, _ := os.Executable()
	projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(exe)))

	// Check if we're running from source (development mode)
	if _, err := os.Stat(filepath.Join(projectRoot, "SuperCrew")); os.IsNotExist(err) {
		// Try current working directory (development mode)
		if cwd, err := os.Getwd(); err == nil {
			if _, err := os.Stat(filepath.Join(cwd, "SuperCrew")); err == nil {
				projectRoot = cwd
			}
		}
	}

	superCrewSource := filepath.Join(projectRoot, "SuperCrew")
	if _, err := os.Stat(superCrewSource); os.IsNotExist(err) {
		log.Errorf("SuperCrew source directory not found at: %s", superCrewSource)
		log.Error("This usually means the binary was not built correctly or SuperCrew files are missing")
		return false
	}

	// Create ~/.claude directory (skip in dry-run mode)
	if !gFlags.DryRun {
		if err := os.MkdirAll(gFlags.InstallDir, 0755); err != nil {
			log.Errorf("Failed to create installation directory: %v", err)
			return false
		}
	}

	// Create backup if installation already exists
	if _, err := os.Stat(gFlags.InstallDir); err == nil && !gFlags.DryRun && !flags.NoBackup {
		log.Info("Creating backup of existing installation...")
		if err := createSimpleBackup(gFlags.InstallDir); err != nil {
			log.Warnf("Failed to create backup: %v", err)
		}
	}

	// Copy SuperCrew directory to ~/.claude/
	log.Info("Installing SuperCrew framework...")

	if gFlags.DryRun {
		log.Info("[DRY RUN] Would copy SuperCrew framework to ~/.claude/")
		return true
	}

	// Use component system for installation
	registry := core.NewEnhancedComponentRegistry(superCrewSource)
	if err := registry.DiscoverComponents(); err != nil {
		log.Errorf("Failed to discover components: %v", err)
		return false
	}

	success := true
	installed := []string{}

	// Resolve dependencies to get proper installation order
	resolvedComponents, err := registry.ResolveDependencies(components)
	if err != nil {
		log.Errorf("Failed to resolve component dependencies: %v", err)
		return false
	}

	log.Infof("Original components: %v", components)
	log.Infof("Resolved installation order: %v", resolvedComponents)

	// Install components using the component system in dependency order
	for _, componentName := range resolvedComponents {
		if !shouldInstallComponent(componentName, components) {
			continue
		}

		// Get component description
		descriptions := map[string]string{
			"commands": "Global slash commands",
			"core":     "Framework core files",
			"hooks":    "Git and development hooks",
			"agents":   "Agent templates and definitions",
		}
		description := descriptions[componentName]

		log.Infof("Installing %s (%s)...", componentName, description)

		if gFlags.DryRun {
			log.Infof("[DRY RUN] Would install %s to %s", componentName, gFlags.InstallDir)
			installed = append(installed, componentName)
			continue
		}

		// Create component instance
		component, err := registry.GetComponentInstance(componentName, gFlags.InstallDir)
		if err != nil {
			log.Errorf("Failed to create component: %s", componentName)
			success = false
			continue
		}

		// Install component with flags
		config := map[string]interface{}{
			"dry_run":          gFlags.DryRun,
			"claude_merge":     flags.ClaudeMerge,
			"claude_overwrite": flags.ClaudeOverwrite,
			"claude_skip":      flags.ClaudeSkip,
		}
		if err := component.Install(gFlags.InstallDir, config); err != nil {
			log.Errorf("Failed to install %s: %v", componentName, err)
			success = false
		} else {
			installed = append(installed, componentName)
			log.Successf("Installed %s successfully", componentName)
		}
	}

	// Show results
	if success && len(installed) > 0 {
		log.Successf("Installed framework components: %s", strings.Join(installed, ", "))

		if !gFlags.DryRun {
			// Initialize version manager and set version
			versionManager := versioning.NewVersionManager(gFlags.InstallDir)
			if err := versionManager.StandardizeAllVersions(); err != nil {
				log.Warnf("Failed to set version information: %v", err)
			} else {
				log.Infof("Framework version set to 1.0.0")
			}

			// Save installation metadata for update detection
			settingsManager := managers.NewSettingsManager(gFlags.InstallDir)
			installInfo := &managers.InstallationInfo{
				Version:          "1.0.0",
				InstalledAt:      time.Now().Format(time.RFC3339),
				LastUpdated:      time.Now().Format(time.RFC3339),
				Components:       make(map[string]string),
				InstallDir:       gFlags.InstallDir,
				InstallerVersion: "1.0.0",
			}

			// Add installed components to metadata
			for _, component := range installed {
				installInfo.Components[component] = "1.0.0"
			}

			if err := settingsManager.SaveInstallationInfo(installInfo); err != nil {
				log.Warnf("Failed to save installation metadata: %v", err)
			}

			log.Info("SuperCrew framework installed successfully!")

			// Install orchestrator-specialist agent
			if err := installOrchestratorAgent(log, gFlags.InstallDir, projectRoot); err != nil {
				log.Warnf("Failed to install orchestrator-specialist agent: %v", err)
			} else {
				log.Success("Orchestrator-specialist agent installed successfully")
			}

			// Metadata consistency is now handled automatically by unified system
			log.Info("Unified metadata system - no migration needed")

			// Install global binary (TODO: implement global binary installation)
			// if err := installGlobalBinary(log, gFlags); err != nil {
			//	log.Warnf("Failed to install global binary: %v", err)
			//	log.Info("You can run the binary from the project directory with ./crew")
			// } else {
			//	log.Info("Global binary installed successfully - 'crew' is now available system-wide")
			// }

			log.Info("Next: Use 'crew claude --install' in your projects to enable /crew: commands")
		} else {
			log.Info("[DRY RUN] SuperCrew framework installation completed (simulation)")
		}
	}

	return success
}

// shouldInstallComponent checks if a component should be installed based on the selected components
func shouldInstallComponent(component string, selectedComponents []string) bool {
	if len(selectedComponents) == 0 {
		return true // Install all if none specified
	}

	// If 'commands' is selected, ensure 'core' is also installed.
	if component == "core" {
		for _, selected := range selectedComponents {
			if selected == "commands" {
				return true
			}
		}
	}

	// Map component names to selection criteria
	componentMap := map[string][]string{
		"commands": {"core", "commands", "all"},
		"Commands": {"core", "commands", "all"},
		"core":     {"core", "all"},
		"Core":     {"core", "all"},
		"hooks":    {"hooks", "all"},
		"agents":   {"core", "agents", "all"},
	}

	allowedSelections, exists := componentMap[component]
	if !exists {
		return false
	}

	for _, selected := range selectedComponents {
		for _, allowed := range allowedSelections {
			if selected == allowed {
				return true
			}
		}
	}

	return false
}

// copyDirectoryRecursive copies a directory and all its contents
func copyDirectoryRecursive(src, dst string) error {
	// Resolve absolute paths to prevent infinite recursion
	absSrc, err := filepath.Abs(src)
	if err != nil {
		return fmt.Errorf("failed to resolve source path: %w", err)
	}
	absDst, err := filepath.Abs(dst)
	if err != nil {
		return fmt.Errorf("failed to resolve destination path: %w", err)
	}

	// Check if destination is a subdirectory of source
	if strings.HasPrefix(absDst, absSrc+string(filepath.Separator)) {
		return fmt.Errorf("cannot copy directory into itself")
	}

	// Get source directory info
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source: %w", err)
	}

	// Create destination directory
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to create destination: %w", err)
	}

	// Read directory contents
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	// Copy each entry
	for _, entry := range entries {
		// Skip the backups directory to prevent infinite recursion
		if entry.Name() == "backups" && entry.IsDir() {
			continue
		}

		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		// Additional check to prevent copying into subdirectory of source
		absDstPath, _ := filepath.Abs(dstPath)
		if strings.HasPrefix(absDstPath, absSrc+string(filepath.Separator)) {
			continue
		}

		if entry.IsDir() {
			if err := copyDirectoryRecursive(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFileSimple(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyDirectorySelective copies a directory with selective file overwrite behavior
func copyDirectorySelective(src, dst string, component string) error {
	// Resolve absolute paths to prevent infinite recursion
	absSrc, err := filepath.Abs(src)
	if err != nil {
		return fmt.Errorf("failed to resolve source path: %w", err)
	}
	absDst, err := filepath.Abs(dst)
	if err != nil {
		return fmt.Errorf("failed to resolve destination path: %w", err)
	}

	// Check if destination is a subdirectory of source
	if strings.HasPrefix(absDst, absSrc+string(filepath.Separator)) {
		return fmt.Errorf("cannot copy directory into itself")
	}

	// Get source directory info
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source: %w", err)
	}

	// Create destination directory
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to create destination: %w", err)
	}

	// Read directory contents
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	// Copy each entry
	for _, entry := range entries {
		// Skip the backups directory to prevent infinite recursion
		if entry.Name() == "backups" && entry.IsDir() {
			continue
		}

		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		// Additional check to prevent copying into subdirectory of source
		absDstPath, _ := filepath.Abs(dstPath)
		if strings.HasPrefix(absDstPath, absSrc+string(filepath.Separator)) {
			continue
		}

		if entry.IsDir() {
			if err := copyDirectorySelective(srcPath, dstPath, component); err != nil {
				return err
			}
		} else {
			if err := copyFileSelective(srcPath, dstPath, component); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFileSelective copies a file with selective overwrite behavior
func copyFileSelective(src, dst string, component string) error {
	// Check if destination file already exists
	if _, err := os.Stat(dst); err == nil {
		// File exists - check if we should overwrite it
		if !shouldOverwriteFile(dst, component) {
			// Skip existing user files
			logger.GetLogger().Debugf("Skipping existing file: %s", filepath.Base(dst))
			return nil
		}
		logger.GetLogger().Debugf("Updating SuperCrew-controlled file: %s", filepath.Base(dst))
	}

	return copyFileSimple(src, dst)
}

// shouldOverwriteFile determines if a file should be overwritten during installation
func shouldOverwriteFile(filePath, component string) bool {
	// Always overwrite SuperCrew-controlled components during updates
	superCrewComponents := []string{"commands", "agents", "hooks"}

	for _, superCrewComp := range superCrewComponents {
		if component == superCrewComp {
			// Check version information to determine if update is needed
			return shouldUpdateBasedOnVersion(component)
		}
	}

	// Never overwrite user files in other directories
	return false
}

// shouldUpdateBasedOnVersion checks if component should be updated based on version metadata
func shouldUpdateBasedOnVersion(component string) bool {
	log := logger.GetLogger()
	installDir := getGlobalInstallDir()
	if installDir == "" {
		log.Warn("Could not determine global install directory for version check.")
		return true // Default to allowing update
	}

	// Try to read the current metadata file
	metadataPath := filepath.Join(installDir, ".crew", "config", "crew-metadata.json")
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		// No metadata file exists, this is a fresh install - allow all updates
		log.Debug("No metadata file found, allowing all component updates")
		return true
	}

	// Read and parse metadata
	metadataContent, err := os.ReadFile(metadataPath)
	if err != nil {
		log.Warnf("Failed to read metadata file: %v", err)
		return true // Default to allowing updates if we can't read metadata
	}

	var metadata struct {
		Framework struct {
			Version string `json:"version"`
		} `json:"framework"`
		Components map[string]struct {
			Version string `json:"version"`
		} `json:"components"`
	}

	if err := json.Unmarshal(metadataContent, &metadata); err != nil {
		log.Warnf("Failed to parse metadata file: %v", err)
		return true // Default to allowing updates if we can't parse metadata
	}

	// Check if we have version info for this component
	if componentInfo, exists := metadata.Components[component]; exists {
		currentVersion := componentInfo.Version
		newVersion := "1.0.0" // Current framework version being installed

		// If versions are different, allow update
		if currentVersion != newVersion {
			log.Debugf("Component %s version change detected: %s -> %s", component, currentVersion, newVersion)
			return true
		}

		// Same version - skip update to preserve user modifications
		log.Debugf("Component %s already at version %s, preserving existing files", component, currentVersion)
		return false
	}

	// Component not found in metadata - allow update (new component)
	log.Debugf("Component %s not found in metadata, allowing update", component)
	return true
}

// copyFileSimple copies a single file (internal helper)
func copyFileSimple(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// createSimpleBackup creates a simple backup of the installation directory
func createSimpleBackup(installDir string) error {
	timestamp := time.Now().Format("20060102-150405")
	backupDir := filepath.Join(installDir, ".crew", "backups")
	backupName := fmt.Sprintf("crew-backup-%s.tar.gz", timestamp)
	finalBackupPath := filepath.Join(backupDir, backupName)

	// Create backups directory if it doesn't exist
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return err
	}

	// Create a temporary directory for the backup
	tempDir, err := os.MkdirTemp("", "crew-backup-temp-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary backup directory: %w", err)
	}
	defer os.RemoveAll(tempDir) // Clean up temp dir

	tempBackupPath := filepath.Join(tempDir, "backup")

	// Copy installation directory to temporary location, excluding backups
	if err := copyDirectorySelectiveBackup(installDir, tempBackupPath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Create tar.gz archive
	if err := createTarGzArchive(tempBackupPath, finalBackupPath); err != nil {
		return fmt.Errorf("failed to create backup archive: %w", err)
	}

	// Create metadata file
	metaPath := finalBackupPath + ".meta"
	if err := createBackupMetadata(installDir, metaPath); err != nil {
		// Don't fail on metadata creation error
		logger.GetLogger().Warn(fmt.Sprintf("Failed to create backup metadata: %v", err))
	}

	return nil
}

// createTarGzArchive creates a tar.gz archive from a directory
func createTarGzArchive(src, dst string) error {
	file, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer file.Close()

	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}

		// Update header name to be relative to src
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		header.Name = relPath

		// Write header
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		// Write file content if it's a regular file
		if info.Mode().IsRegular() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			if _, err := io.Copy(tarWriter, file); err != nil {
				return err
			}
		}

		return nil
	})
}

// createBackupMetadata creates a metadata file for the backup
func createBackupMetadata(installDir, metaPath string) error {
	// Get component versions
	components := make(map[string]string)

	// Use the metadata manager to get installed components
	metadataManager := managers.NewMetadataManager(installDir)
	installedComponents, err := metadataManager.GetInstalledComponents()
	if err == nil {
		for name, info := range installedComponents {
			if version, ok := info["version"].(string); ok {
				components[name] = version
			}
		}
	}

	// Get framework version
	versionManager := versioning.NewVersionManager(installDir)
	frameworkVersion, _ := versionManager.GetCurrentVersion()
	if frameworkVersion == "" {
		frameworkVersion = "1.0.0"
	}

	metadata := map[string]interface{}{
		"created":    time.Now().Format(time.RFC3339),
		"framework":  frameworkVersion,
		"components": components,
	}

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(metaPath, data, 0644)
}

// copyDirectorySelectiveBackup copies a directory excluding the backups folder
func copyDirectorySelectiveBackup(src, dst string) error {
	// Get source directory info
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source: %w", err)
	}

	// Create destination directory
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to create destination: %w", err)
	}

	// Read directory contents
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	// Copy each entry
	for _, entry := range entries {
		// Skip the backups directory entirely
		if entry.Name() == "backups" && entry.IsDir() {
			continue
		}

		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDirectorySelectiveBackup(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFileSimple(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// installCoreWithCLAUDEHandling installs Core component files directly into the destination
// with special handling for CLAUDE.md
func installCoreWithCLAUDEHandling(srcPath, dstPath string, gFlags *GlobalFlags) error {
	// Note: dstPath is now the root ~/.claude directory, not a Core subdirectory
	claudeMdSrc := filepath.Join(srcPath, "CLAUDE.md")
	claudeMdDst := filepath.Join(dstPath, "CLAUDE.md")

	// Check if CLAUDE.md exists in destination
	existingCLAUDE := false
	if _, err := os.Stat(claudeMdDst); err == nil {
		existingCLAUDE = true
	}

	// Handle CLAUDE.md based on flags or interactive choice
	if existingCLAUDE {
		var action int // 0=merge, 1=overwrite, 2=skip

		// Check command-line flags first
		flagCount := 0
		if installFlags.ClaudeMerge {
			action = 0
			flagCount++
		}
		if installFlags.ClaudeOverwrite {
			action = 1
			flagCount++
		}
		if installFlags.ClaudeSkip {
			action = 2
			flagCount++
		}

		// Validate that only one flag is set
		if flagCount > 1 {
			return fmt.Errorf("only one of --claude-merge, --claude-overwrite, or --claude-skip can be specified")
		}

		// If no flags set and not in auto mode, ask interactively
		if flagCount == 0 && !gFlags.Yes && !gFlags.Quiet {
			logger.GetLogger().Warn("Existing CLAUDE.md detected")

			options := []string{
				"Merge (preserve custom sections, update framework sections)",
				"Overwrite (replace with new version)",
				"Skip (keep existing file)",
			}

			choice, err := ui.PromptChoice("How would you like to handle the existing CLAUDE.md file?", options, 0)
			if err != nil {
				// Default to skip (safe) on error
				choice = 2
			}
			action = choice
		} else if flagCount == 0 && gFlags.Yes {
			// In auto mode (--yes) without explicit CLAUDE.md flags, default to skip (safe)
			action = 2
			logger.GetLogger().Info("Auto mode: preserving existing CLAUDE.md")
		}

		switch action {
		case 0:
			// Merge CLAUDE.md
			if err := mergeCLAUDEmd(claudeMdSrc, claudeMdDst); err != nil {
				return fmt.Errorf("failed to merge CLAUDE.md: %w", err)
			}
			logger.GetLogger().Success("Merged CLAUDE.md successfully")

			// Copy rest of Core contents excluding CLAUDE.md
			return copyCoreContents(srcPath, dstPath, []string{"CLAUDE.md"})

		case 1:
			// Overwrite - copy all Core contents
			logger.GetLogger().Info("Overwriting existing CLAUDE.md")
			return copyCoreContents(srcPath, dstPath, nil)

		case 2:
			// Skip CLAUDE.md
			logger.GetLogger().Info("Keeping existing CLAUDE.md")
			return copyCoreContents(srcPath, dstPath, []string{"CLAUDE.md"})
		}
	}

	// No existing CLAUDE.md or auto-mode, just copy everything
	return copyCoreContents(srcPath, dstPath, nil)
}

// copyCoreContents copies the contents of the Core directory directly into destination
func copyCoreContents(srcDir, dstDir string, excludeFiles []string) error {
	log := logger.GetLogger()

	// First, create the .crew directory structure for utilities
	log.Info("Creating .crew directory structure...")
	crewDirs := []string{
		filepath.Join(dstDir, ".crew"),
		filepath.Join(dstDir, ".crew", "logs"),
		filepath.Join(dstDir, ".crew", "workflows"),
		filepath.Join(dstDir, ".crew", "scripts"),
		filepath.Join(dstDir, ".crew", "config"),
		filepath.Join(dstDir, ".crew", "prompts"),
		filepath.Join(dstDir, ".crew", "completions"),
	}

	for _, dir := range crewDirs {
		log.Infof("Creating directory: %s", dir)
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Errorf("Failed to create directory %s: %v", dir, err)
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
		log.Infof("Successfully created directory: %s", dir)
	}

	// Read source directory contents
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return fmt.Errorf("failed to read source directory: %w", err)
	}

	// Copy each entry directly to destination
	for _, entry := range entries {
		// Skip excluded files
		if excludeFiles != nil && contains(excludeFiles, entry.Name()) {
			continue
		}

		srcPath := filepath.Join(srcDir, entry.Name())
		dstPath := filepath.Join(dstDir, entry.Name())

		if entry.IsDir() {
			// Recursively copy subdirectories
			if err := copyDirectoryRecursive(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy files
			if err := copyFileSimple(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// mergeCLAUDEmd merges the source CLAUDE.md with existing destination CLAUDE.md
func mergeCLAUDEmd(srcFile, dstFile string) error {
	// Read source CLAUDE.md
	srcContent, err := os.ReadFile(srcFile)
	if err != nil {
		return fmt.Errorf("failed to read source CLAUDE.md: %w", err)
	}

	// Read existing CLAUDE.md
	existingContent, err := os.ReadFile(dstFile)
	if err != nil {
		return fmt.Errorf("failed to read existing CLAUDE.md: %w", err)
	}

	// Parse sections from both files
	srcSections := parseCLAUDESections(string(srcContent))
	existingSections := parseCLAUDESections(string(existingContent))

	// Debug: log what sections we found
	logger.GetLogger().Debug(fmt.Sprintf("Source sections found: %d", len(srcSections)))
	logger.GetLogger().Debug(fmt.Sprintf("Existing sections found: %d", len(existingSections)))
	if footerContent, hasFooter := existingSections["__footer__"]; hasFooter {
		logger.GetLogger().Debug(fmt.Sprintf("Footer content found: %s", footerContent))
	}

	// Check versions
	srcVersion := extractVersion(string(srcContent))
	existingVersion := extractVersion(string(existingContent))
	if srcVersion != "" && existingVersion != "" {
		logger.GetLogger().Info(fmt.Sprintf("Version info - Source: %s, Existing: %s", srcVersion, existingVersion))
		if srcVersion == existingVersion {
			logger.GetLogger().Info("Versions match - preserving custom content only")
		}
	}

	// Merge strategy:
	// 1. Keep framework sections from source (they might have updates)
	// 2. Preserve custom sections from existing file
	// 3. Add any new sections from source

	mergedContent := buildMergedCLAUDE(srcSections, existingSections)

	// Write merged content
	if err := os.WriteFile(dstFile, []byte(mergedContent), 0644); err != nil {
		return fmt.Errorf("failed to write merged CLAUDE.md: %w", err)
	}

	return nil
}

// extractVersion extracts version from version block or content
func extractVersion(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "<sc-v") && strings.Contains(line, ">") {
			// Extract version between <sc-v and >
			start := strings.Index(line, "<sc-v") + 5
			end := strings.Index(line[start:], ">")
			if end > 0 {
				return line[start : start+end]
			}
		}
	}
	return ""
}

// parseCLAUDESections parses CLAUDE.md content into sections
func parseCLAUDESections(content string) map[string]string {
	sections := make(map[string]string)
	lines := strings.Split(content, "\n")

	currentSection := ""
	currentContent := []string{}

	// Track if we're still in the header or in a section
	inSection := false
	inVersionBlock := false
	versionContent := []string{}

	for i, line := range lines {
		// Check for version tags
		if strings.HasPrefix(line, "<sc-v") && strings.Contains(line, ">") {
			inVersionBlock = true
			versionContent = append(versionContent, line)
			continue
		}
		if inVersionBlock && strings.HasPrefix(line, "<sc-end-v") && strings.Contains(line, ">") {
			// Check if there's content after the closing tag on the same line
			endTagPos := strings.Index(line, ">")
			if endTagPos >= 0 && endTagPos < len(line)-1 {
				// There's content after the closing tag
				tagPart := line[:endTagPos+1]
				afterTagContent := line[endTagPos+1:]

				// Add only the tag part to version block
				versionContent = append(versionContent, tagPart)
				sections["__version_block__"] = strings.Join(versionContent, "\n")
				inVersionBlock = false
				versionContent = []string{}

				// Start collecting the content after the tag as footer content
				if strings.TrimSpace(afterTagContent) != "" {
					currentContent = []string{afterTagContent}
				}
			} else {
				// No content after closing tag, process normally
				versionContent = append(versionContent, line)
				sections["__version_block__"] = strings.Join(versionContent, "\n")
				inVersionBlock = false
				versionContent = []string{}
			}
			continue
		}
		if inVersionBlock {
			versionContent = append(versionContent, line)
			continue
		}

		// Check if line is a section header (starts with #)
		if strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "##") {
			// Save previous section
			if currentSection != "" {
				sections[currentSection] = strings.Join(currentContent, "\n")
			} else if !inSection && len(currentContent) > 0 {
				// This is content before any section (shouldn't happen with new format)
				sections["__header__"] = strings.Join(currentContent, "\n")
			}
			// Start new section
			currentSection = strings.TrimSpace(line)
			currentContent = []string{}
			inSection = true
		} else {
			// Only add non-empty lines or if it's not the first line
			if line != "" || i > 0 {
				currentContent = append(currentContent, line)
			}
		}
	}

	// Save last section or remaining content
	if currentSection != "" {
		sections[currentSection] = strings.Join(currentContent, "\n")
	} else if len(currentContent) > 0 && strings.TrimSpace(strings.Join(currentContent, "\n")) != "" {
		// Content after all sections
		sections["__footer__"] = strings.Join(currentContent, "\n")
	}

	return sections
}

// buildMergedCLAUDE builds merged CLAUDE.md content
func buildMergedCLAUDE(srcSections, existingSections map[string]string) string {
	var result []string

	// First, add the version block from source (which now contains the entire framework content)
	if versionBlock, hasVersion := srcSections["__version_block__"]; hasVersion {
		result = append(result, versionBlock)
	}

	// Then, add all custom sections from existing file (anything outside version block)
	for section, content := range existingSections {
		if section != "__header__" && section != "__footer__" && section != "__version_block__" &&
			section != "# Claude Code Super Crew Entry Point" { // Skip the old header since it's now in version block
			// This is a custom section, preserve it
			result = append(result, "")
			result = append(result, section)
			result = append(result, content)
		}
	}

	// Finally, add any footer content from existing file (custom content after all sections)
	if footerContent, exists := existingSections["__footer__"]; exists && strings.TrimSpace(footerContent) != "" {
		result = append(result, "")
		result = append(result, footerContent)
	}

	return strings.Join(result, "\n")
}

// installOrchestratorAgent installs the orchestrator-specialist.md agent file
func installOrchestratorAgent(log logger.Logger, installDir, projectRoot string) error {
	log.Info("Installing orchestrator-specialist agent...")
	
	// Define source and destination paths
	sourceFile := filepath.Join(projectRoot, "SuperCrew", "agents", "orchestrator-specialist.md")
	destDir := filepath.Join(installDir, "agents")
	destFile := filepath.Join(destDir, "orchestrator-specialist.md")
	
	// Check if source file exists
	if _, err := os.Stat(sourceFile); os.IsNotExist(err) {
		log.Warnf("Source file not found: %s", sourceFile)
		return fmt.Errorf("source file not found: %w", err)
	}
	
	// Ensure destination directory exists
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create agents directory: %w", err)
	}
	
	// Copy the file
	if err := copyFileSimple(sourceFile, destFile); err != nil {
		return fmt.Errorf("failed to copy agent file: %w", err)
	}
	
	log.Infof("Successfully installed agent file to: %s", destFile)
	return nil
}
