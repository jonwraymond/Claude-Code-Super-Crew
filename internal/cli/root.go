package cli

import (
	"fmt"

	"github.com/jonwraymond/claude-code-super-crew/internal/ui"
	"github.com/spf13/cobra"
)

// GlobalFlags holds all global command flags
type GlobalFlags struct {
	Verbose    bool
	Quiet      bool
	InstallDir string
	DryRun     bool
	Force      bool
	Yes        bool
}

var globalFlags GlobalFlags

// testMode is set to true during tests to bypass certain validations
var testMode = false

// SetTestMode sets the test mode for the CLI
func SetTestMode(enabled bool) {
	testMode = enabled
}

// NewRootCommand creates the root command
func NewRootCommand(version string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "crew",
		Short: "Claude Code Super Crew Framework Management Hub",
		Long: `Claude Code Super Crew Framework Management Hub - Unified CLI
		
The Claude Code Super Crew CLI provides operations to install, update, manage, 
and configure the Super Crew framework for Claude AI.`,
		Version: version,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Check for conflicting flags
			if globalFlags.Verbose && globalFlags.Quiet {
				return fmt.Errorf("conflicting flags: --verbose and --quiet cannot be used together")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			if !globalFlags.Quiet {
				ui.DisplayHeader("Claude Code Super Crew v"+version, "Unified CLI for all operations")
				fmt.Printf("%sAvailable operations:%s\n", ui.ColorCyan, ui.ColorReset)
				fmt.Printf("  %-12s %s\n", "install", "Install SuperCrew framework globally")
				fmt.Printf("  %-12s %s\n", "status", "Show detailed component and feature status")
				fmt.Printf("  %-12s %s\n", "claude", "Manage project-level Claude Code integration")
				fmt.Printf("  %-12s %s\n", "update", "Update existing Claude Code Super Crew installation")
				fmt.Printf("  %-12s %s\n", "update-document", "Update document version with pipeline propagation")
				fmt.Printf("  %-12s %s\n", "uninstall", "Remove Claude Code Super Crew installation")
				fmt.Printf("  %-12s %s\n", "backup", "Backup and restore operations")
				fmt.Printf("\n%sQuick Start:%s\n", ui.ColorGreen, ui.ColorReset)
				fmt.Printf("  1. crew install              # Install framework globally (once)\n")
				fmt.Printf("  2. crew claude --install     # Enable for current project\n")
			}
		},
	}

	// Add global flags
	rootCmd.PersistentFlags().BoolVarP(&globalFlags.Verbose, "verbose", "v", false, "Enable verbose logging")
	rootCmd.PersistentFlags().BoolVarP(&globalFlags.Quiet, "quiet", "q", false, "Suppress all output except errors")
	rootCmd.PersistentFlags().StringVar(&globalFlags.InstallDir, "install-dir", expandPath("~/.claude"), "Target installation directory")
	rootCmd.PersistentFlags().BoolVar(&globalFlags.DryRun, "dry-run", false, "Simulate operation without making changes")
	rootCmd.PersistentFlags().BoolVar(&globalFlags.Force, "force", false, "Force execution, skipping checks")
	rootCmd.PersistentFlags().BoolVarP(&globalFlags.Yes, "yes", "y", false, "Automatically answer yes to all prompts")

	// Add subcommands
	rootCmd.AddCommand(NewInstallCommand())
	rootCmd.AddCommand(NewStatusCommand())
	rootCmd.AddCommand(NewUpdateCommand())
	rootCmd.AddCommand(NewUpdateDocumentCommand())
	rootCmd.AddCommand(NewUninstallCommand())
	rootCmd.AddCommand(NewBackupCommand())
	rootCmd.AddCommand(NewClaudeCommand())
	rootCmd.AddCommand(NewHooksCommand())
	rootCmd.AddCommand(NewVersionCommand())
	rootCmd.AddCommand(NewIntegrityCommand())

	return rootCmd
}

// GetGlobalFlags returns the global flags
func GetGlobalFlags() *GlobalFlags {
	return &globalFlags
}
