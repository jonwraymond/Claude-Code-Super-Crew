package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jonwraymond/claude-code-super-crew/internal/claude"
	"github.com/jonwraymond/claude-code-super-crew/internal/orchestrator"
	"github.com/jonwraymond/claude-code-super-crew/internal/ui"
	"github.com/jonwraymond/claude-code-super-crew/pkg/logger"
	"github.com/spf13/cobra"
)

// ClaudeFlags holds Claude Code integration command flags
type ClaudeFlags struct {
	Install     bool
	Uninstall   bool
	Status      bool
	Update      bool
	List        bool
	Test        string
	ClaudeDir   string
	CommandsDir string
	ProjectDir  string
	Shell       string
	Export      string
}

var claudeFlags ClaudeFlags

// NewClaudeCommand creates the Claude Code integration command
func NewClaudeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claude",
		Short: "Claude Code integration management",
		Long: `Manage Claude Code integration for supercrew /crew: commands.

This command manages project-level Claude Code integration for the current project,
enabling /crew: prefixed commands with tab completion and project-specific agents.

Examples:
  crew claude --install                    # Enable Claude integration for current project
  crew claude --status --verbose          # Check project integration status
  crew claude --list                      # List available /crew: commands
  crew claude --test /crew:analyze        # Test a specific command
  crew claude --export completions.json   # Export commands for external use
  crew claude --uninstall                 # Remove project integration`,
		RunE: runClaude,
	}

	// Main operations
	cmd.Flags().BoolVar(&claudeFlags.Install, "install", false,
		"Install Claude Code integration for current project")
	cmd.Flags().BoolVar(&claudeFlags.Uninstall, "uninstall", false,
		"Uninstall Claude Code integration for current project")
	cmd.Flags().BoolVar(&claudeFlags.Status, "status", false,
		"Show integration status")
	cmd.Flags().BoolVar(&claudeFlags.Update, "update", false,
		"Update existing integration")
	cmd.Flags().BoolVar(&claudeFlags.List, "list", false,
		"List available /crew: commands")
	cmd.Flags().StringVar(&claudeFlags.Test, "test", "",
		"Test a specific command (e.g., /crew:analyze)")
	cmd.Flags().StringVar(&claudeFlags.Export, "export", "",
		"Export commands to JSON file")

	// Configuration options
	cmd.Flags().StringVar(&claudeFlags.ClaudeDir, "claude-dir", "",
		"Claude Code configuration directory (default: project-dir/.claude)")
	cmd.Flags().StringVar(&claudeFlags.CommandsDir, "commands-dir", "",
		"Commands directory (default: ~/.claude/commands for global commands)")
	cmd.Flags().StringVar(&claudeFlags.ProjectDir, "project-dir", "",
		"Project directory for installation (default: current working directory)")
	cmd.Flags().StringVar(&claudeFlags.Shell, "shell", "",
		"Generate completion for specific shell (bash, zsh, fish)")

	// Mark operations as mutually exclusive
	cmd.MarkFlagsMutuallyExclusive("install", "uninstall", "status", "update", "list", "test", "export")

	return cmd
}

func runClaude(cmd *cobra.Command, args []string) error {
	log := logger.GetLogger()
	log.SetVerbose(globalFlags.Verbose)
	log.SetQuiet(globalFlags.Quiet)

	// Validate shell parameter if provided
	if claudeFlags.Shell != "" {
		validShells := []string{"bash", "zsh", "fish"}
		valid := false
		for _, validShell := range validShells {
			if claudeFlags.Shell == validShell {
				valid = true
				break
			}
		}
		if !valid {
			ui.DisplayError(fmt.Sprintf("Invalid shell: %s. Supported shells: bash, zsh, fish", claudeFlags.Shell))
			return fmt.Errorf("invalid shell: %s", claudeFlags.Shell)
		}
	}

	// Determine project directory
	if claudeFlags.ProjectDir == "" {
		// Default to current working directory for project-level installation
		pwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		claudeFlags.ProjectDir = pwd
	}

	// Set claude directory based on project directory
	if claudeFlags.ClaudeDir == "" {
		claudeFlags.ClaudeDir = filepath.Join(claudeFlags.ProjectDir, ".claude")
	}

	if claudeFlags.CommandsDir == "" {
		// Use global commands directory for command definitions (read-only)
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		claudeFlags.CommandsDir = filepath.Join(home, ".claude", "commands", "crew")
	}

	// Display header
	if !globalFlags.Quiet {
		ui.DisplayHeader(
			"Claude Code Super Crew Integration v1.0",
			"Manage Claude Code integration for /crew: commands",
		)
	}

	// Create integration manager
	integration, err := claude.NewClaudeIntegration(claudeFlags.CommandsDir, claudeFlags.ClaudeDir)
	if err != nil {
		// Check if it's because framework isn't installed
		if _, statErr := os.Stat(claudeFlags.CommandsDir); os.IsNotExist(statErr) {
			log.Error("SuperCrew framework not found. The global framework must be installed first.")
			log.Info("")
			log.Info("Run these commands:")
			log.Info("  1. crew install              # Install framework globally")
			log.Info("  2. crew claude --install     # Enable for this project")
			return fmt.Errorf("framework not installed")
		}
		return fmt.Errorf("failed to create integration manager: %w", err)
	}

	// Handle different operations
	switch {
	case claudeFlags.Install:
		return installClaudeIntegration(integration)

	case claudeFlags.Uninstall:
		return uninstallClaudeIntegration(integration)

	case claudeFlags.Status:
		return showClaudeStatus(integration)

	case claudeFlags.Update:
		return updateClaudeIntegration(integration)

	case claudeFlags.List:
		return listClaudeCommands(integration)

	case claudeFlags.Test != "":
		return testClaudeCommand(integration, claudeFlags.Test)

	case claudeFlags.Export != "":
		return exportClaudeCommands(integration, claudeFlags.Export)

	default:
		// Default to status if no operation specified
		return showClaudeStatus(integration)
	}
}

func installClaudeIntegration(integration *claude.ClaudeIntegration) error {
	log := logger.GetLogger()

	// Use the configured project directory
	projectDir := claudeFlags.ProjectDir

	log.Infof("Installing Claude Code integration for project: %s", projectDir)

	// Check if framework is installed globally
	if !isFrameworkInstalled() {
		log.Error("SuperCrew framework not found in ~/.claude/")
		log.Info("Please run 'crew install' first to install the global framework")
		log.Info("Then run 'crew claude --install' to enable this project")
		return fmt.Errorf("framework not installed - run 'crew install' first")
	}

	// Check if already installed for this project
	status, err := integration.CheckIntegration()
	if err != nil {
		return fmt.Errorf("failed to check integration status: %w", err)
	}

	if status.Installed && !globalFlags.Force {
		log.Warn("Claude Code integration already installed for this project")
		if !globalFlags.Yes && !ui.Confirm("Reinstall integration?", false) {
			log.Info("Installation cancelled")
			return nil
		}
	}

	// Install project-specific integration
	log.Info("Setting up project-level Claude Code integration...")

	// The claudeFlags.ClaudeDir is already set to project/.claude
	projectClaudeDir := claudeFlags.ClaudeDir
	projectAgentsDir := filepath.Join(projectClaudeDir, "agents")
	if err := os.MkdirAll(projectAgentsDir, 0755); err != nil {
		return fmt.Errorf("failed to create project agents directory: %w", err)
	}

	// Install orchestrator-specialist for this project
	orchestratorInstaller := orchestrator.NewOrchestratorInstaller(projectDir)
	if err := orchestratorInstaller.InstallOrValidate(); err != nil {
		return fmt.Errorf("failed to install orchestrator: %w", err)
	}

	// Create project marker file
	projectMarker := filepath.Join(projectClaudeDir, "project-config.json")
	projectConfig := map[string]interface{}{
		"version":         "1.0",
		"project_path":    projectDir,
		"global_commands": claudeFlags.CommandsDir,
		"created_at":      time.Now().Format(time.RFC3339),
		"type":            "project-integration",
	}

	configData, err := json.MarshalIndent(projectConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal project config: %w", err)
	}

	if err := os.WriteFile(projectMarker, configData, 0644); err != nil {
		return fmt.Errorf("failed to write project config: %w", err)
	}

	// Install Claude Code integration in PROJECT directory (not globally)
	log.Infof("Installing integration files to project directory: %s", projectClaudeDir)
	if err := integration.InstallIntegration(); err != nil {
		return fmt.Errorf("integration installation failed: %w", err)
	}

	if !globalFlags.Quiet {
		ui.DisplaySuccess("Project-level Claude Code integration installed!")

		fmt.Printf("\n%sNext steps:%s\n", ui.ColorCyan, ui.ColorReset)
		fmt.Println("1. Restart Claude Code to load the project commands")
		fmt.Println("2. Run '/crew:load' to analyze this project")
		fmt.Println("3. Use '/crew:help' to see available agents and commands")
		fmt.Printf("\nProject: %s\n", projectDir)
	}

	return nil
}

func uninstallClaudeIntegration(integration *claude.ClaudeIntegration) error {
	log := logger.GetLogger()

	// Use the configured project directory
	projectDir := claudeFlags.ProjectDir

	log.Infof("Uninstalling Claude Code integration for project: %s", projectDir)

	// Check if installed
	status, err := integration.CheckIntegration()
	if err != nil {
		return fmt.Errorf("failed to check integration status: %w", err)
	}

	if !status.Installed {
		log.Warn("Claude Code integration not installed for this project")
		return nil
	}

	// Confirm uninstallation
	if !globalFlags.Yes && !ui.Confirm("Remove Claude Code integration for this project?", false) {
		log.Info("Uninstallation cancelled")
		return nil
	}

	// Uninstall Claude Code integration
	log.Info("Removing project-level Claude Code integration...")
	if err := integration.UninstallIntegration(); err != nil {
		return fmt.Errorf("integration uninstallation failed: %w", err)
	}

	// Remove project-specific .claude directory
	projectClaudeDir := claudeFlags.ClaudeDir // This is the project's .claude directory
	if _, err := os.Stat(projectClaudeDir); err == nil {
		if err := os.RemoveAll(projectClaudeDir); err != nil {
			log.Warnf("Failed to remove %s: %v", projectClaudeDir, err)
		} else {
			log.Infof("Removed project directory: %s", projectClaudeDir)
		}
	}

	if !globalFlags.Quiet {
		ui.DisplaySuccess("Project-level Claude Code integration removed!")
		fmt.Printf("Project: %s\n", projectDir)
	}

	return nil
}

func showClaudeStatus(integration *claude.ClaudeIntegration) error {
	status, err := integration.CheckIntegration()
	if err != nil {
		return fmt.Errorf("failed to check integration status: %w", err)
	}

	fmt.Printf("\n%s%sSuper Crew Status%s\n", ui.ColorCyan, ui.ColorBright, ui.ColorReset)
	fmt.Println(strings.Repeat("=", 50))

	// Get current project directory
	projectDir := claudeFlags.ProjectDir

	// Check global framework installation
	globalInstalled := isFrameworkInstalled()
	if globalInstalled {
		fmt.Printf("%s✅ Global Framework Installed%s\n", ui.ColorGreen, ui.ColorReset)

		// Show command count
		installDir := getGlobalInstallDir()
		commandsDir := filepath.Join(installDir, "SuperCrew", "Commands")
		if entries, err := os.ReadDir(commandsDir); err == nil {
			cmdCount := 0
			for _, entry := range entries {
				if strings.HasSuffix(entry.Name(), ".md") {
					cmdCount++
				}
			}
			fmt.Printf("%sGlobal Commands:%s %d available\n", ui.ColorBlue, ui.ColorReset, cmdCount)
		}
	} else {
		fmt.Printf("%s❌ Global Framework Not Installed%s\n", ui.ColorRed, ui.ColorReset)
	}

	// Check project-level integration
	fmt.Printf("\n%sProject Status:%s %s\n", ui.ColorCyan, ui.ColorReset, projectDir)

	// Check for project .claude directory using configured project path
	projectClaudeDir := claudeFlags.ClaudeDir
	projectAgentsDir := filepath.Join(projectClaudeDir, "agents")
	projectIntegrated := false
	if _, err := os.Stat(projectAgentsDir); err == nil {
		projectIntegrated = true
		fmt.Printf("%s✅ Project Integration Active%s\n", ui.ColorGreen, ui.ColorReset)

		// Check for orchestrator
		orchestratorPath := filepath.Join(projectAgentsDir, "orchestrator-specialist.md")
		if _, err := os.Stat(orchestratorPath); err == nil {
			fmt.Printf("%s✅ Orchestrator Installed%s\n", ui.ColorGreen, ui.ColorReset)
		}

		// Count project specialists
		if entries, err := os.ReadDir(projectAgentsDir); err == nil {
			specialistCount := 0
			for _, entry := range entries {
				if strings.HasSuffix(entry.Name(), "-specialist.md") {
					specialistCount++
				}
			}
			fmt.Printf("%sProject Specialists:%s %d agents\n", ui.ColorBlue, ui.ColorReset, specialistCount)
		}

		// Show project integration path
		fmt.Printf("%sProject Path:%s %s\n", ui.ColorBlue, ui.ColorReset, projectClaudeDir)
	} else {
		fmt.Printf("%s❌ Project Not Integrated%s\n", ui.ColorRed, ui.ColorReset)
	}

	// Claude integration status
	if status.Installed {
		fmt.Printf("\n%s✅ Claude Code Integration Active%s", ui.ColorGreen, ui.ColorReset)
		if status.Version != "" {
			fmt.Printf(" (v%s)", status.Version)
		}
		fmt.Println()
		fmt.Printf("%sCommands Available:%s %d\n", ui.ColorBlue, ui.ColorReset, status.CommandCount)
	} else {
		fmt.Printf("\n%s❌ Claude Code Integration Not Active%s\n", ui.ColorRed, ui.ColorReset)
	}

	// Issues and recommendations
	if len(status.Issues) > 0 {
		fmt.Printf("\n%sIssues Found:%s\n", ui.ColorYellow, ui.ColorReset)
		for _, issue := range status.Issues {
			fmt.Printf("  ⚠️  %s\n", issue)
		}
	}

	// Recommendations
	fmt.Printf("\n%sRecommendations:%s\n", ui.ColorCyan, ui.ColorReset)
	if !globalInstalled {
		fmt.Println("  1. Run 'crew install' to install framework globally")
	}
	if globalInstalled && !projectIntegrated {
		fmt.Println("  2. Run 'crew claude --install' to enable this project")
	}
	if projectIntegrated && !status.Installed {
		fmt.Println("  3. Restart Claude Code to activate integration")
	}

	return nil
}

func updateClaudeIntegration(integration *claude.ClaudeIntegration) error {
	log := logger.GetLogger()

	// Check current status
	status, err := integration.CheckIntegration()
	if err != nil {
		return fmt.Errorf("failed to check integration status: %w", err)
	}

	if !status.Installed {
		log.Info("Integration not installed, performing fresh installation")
		return installClaudeIntegration(integration)
	}

	log.Info("Updating Claude Code integration...")

	// Update integration
	if err := integration.UpdateIntegration(); err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	if !globalFlags.Quiet {
		ui.DisplaySuccess("Claude Code integration updated successfully!")
	}

	return nil
}

func listClaudeCommands(integration *claude.ClaudeIntegration) error {
	commands := integration.ListAvailableCommands()

	fmt.Printf("\n%s%sAvailable /crew: Commands%s\n", ui.ColorCyan, ui.ColorBright, ui.ColorReset)
	fmt.Println(strings.Repeat("=", 60))

	if len(commands) == 0 {
		fmt.Printf("%sNo commands found%s\n", ui.ColorYellow, ui.ColorReset)
		return nil
	}

	// Create table data
	headers := []string{"Command", "Description"}
	var rows [][]string

	for _, cmd := range commands {
		name := fmt.Sprintf("/crew:%s", cmd.Name)
		description := cmd.Description
		if len(description) > 45 {
			description = description[:42] + "..."
		}
		rows = append(rows, []string{name, description})
	}

	ui.DisplayTable(headers, rows, "")

	fmt.Printf("\n%sTotal: %d commands%s\n", ui.ColorBlue, len(commands), ui.ColorReset)

	if globalFlags.Verbose {
		fmt.Printf("\n%sUsage:%s\n", ui.ColorCyan, ui.ColorReset)
		fmt.Println("  Type '/crew:' in Claude Code and press Tab to see completions")
		fmt.Println("  Use '/crew:help' for detailed command documentation")
	}

	return nil
}

func testClaudeCommand(integration *claude.ClaudeIntegration, commandLine string) error {
	log := logger.GetLogger()

	fmt.Printf("\n%sTesting Command: %s%s\n", ui.ColorCyan, commandLine, ui.ColorReset)
	fmt.Println(strings.Repeat("=", 50))

	// Check if this is a completion test (partial command)
	if strings.HasPrefix(commandLine, "/crew:") && !strings.Contains(commandLine, " ") {
		// Test completion for partial command
		completionProvider, err := claude.NewCompletionProvider(claudeFlags.CommandsDir)
		if err != nil {
			return fmt.Errorf("failed to create completion provider: %w", err)
		}

		result := completionProvider.GetCompletions(commandLine)

		fmt.Printf("%s[✓] Completion Test Results:%s\n", ui.ColorGreen, ui.ColorReset)
		fmt.Printf("Input: %s\n", result.Input)
		fmt.Printf("Type: %s\n", result.Type)
		fmt.Printf("Count: %d suggestions\n", result.Count)

		if len(result.Suggestions) > 0 {
			fmt.Printf("\n%sSuggestions:%s\n", ui.ColorCyan, ui.ColorReset)
			for _, suggestion := range result.Suggestions {
				fmt.Printf("  %s - %s\n", suggestion.Text, suggestion.Description)
				if suggestion.ArgumentHints != "" {
					fmt.Printf("    Args: %s\n", suggestion.ArgumentHints)
				}
			}
		}
		return nil
	}

	// Validate command format for execution
	if !strings.HasPrefix(commandLine, "/crew:") {
		return fmt.Errorf("invalid command format: %s (expected /crew:command)", commandLine)
	}

	// Execute command
	if globalFlags.DryRun {
		log.Info("[DRY RUN] Would execute command")
		return nil
	}

	err := integration.ExecuteSlashCommand(commandLine)
	if err != nil {
		ui.DisplayError(fmt.Sprintf("Command execution failed: %v", err))
		return err
	}

	ui.DisplaySuccess("Command executed successfully!")
	return nil
}

func exportClaudeCommands(integration *claude.ClaudeIntegration, outputFile string) error {
	log := logger.GetLogger()

	commands := integration.ListAvailableCommands()
	if len(commands) == 0 {
		return fmt.Errorf("no commands available to export")
	}

	// Export to JSON
	data, err := json.MarshalIndent(commands, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal commands: %w", err)
	}

	if err := os.WriteFile(outputFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write export file: %w", err)
	}

	log.Successf("Exported %d commands to %s", len(commands), outputFile)

	if globalFlags.Verbose {
		fmt.Printf("\n%sExported Commands:%s\n", ui.ColorBlue, ui.ColorReset)
		for _, cmd := range commands {
			fmt.Printf("  - /crew:%s\n", cmd.Name)
		}
	}

	return nil
}

// isFrameworkInstalled checks if the global SuperCrew framework is installed
func isFrameworkInstalled() bool {
	installDir := getGlobalInstallDir()
	// Check for official Claude commands directory
	commandsDir := filepath.Join(installDir, "commands")
	_, err := os.Stat(commandsDir)
	return err == nil
}

// getGlobalInstallDir returns the global installation directory
func getGlobalInstallDir() string {
	if globalFlags.InstallDir != "" {
		return globalFlags.InstallDir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude")
}
