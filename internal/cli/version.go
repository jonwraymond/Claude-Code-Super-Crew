package cli

import (
	"fmt"
	"strings"

	"github.com/jonwraymond/claude-code-super-crew/internal/ui"
	"github.com/jonwraymond/claude-code-super-crew/internal/versioning"
	"github.com/spf13/cobra"
)

// VersionFlags holds version command flags
type VersionFlags struct {
	Components bool
	All        bool
	Check      bool
}

var versionFlags VersionFlags

// NewVersionCommand creates the version command
func NewVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Display Claude Code Super Crew version information",
		Long: `Display version information for Claude Code Super Crew framework and components.

Examples:
  crew version                  # Show framework version
  crew version --components     # Show all component versions
  crew version --all           # Show detailed version information
  crew version --check         # Check if updates are available`,
		RunE: runVersion,
	}

	cmd.Flags().BoolVar(&versionFlags.Components, "components", false,
		"Show individual component versions")
	cmd.Flags().BoolVar(&versionFlags.All, "all", false,
		"Show all version information")
	cmd.Flags().BoolVar(&versionFlags.Check, "check", false,
		"Check for available updates")

	return cmd
}

func runVersion(cmd *cobra.Command, args []string) error {
	versionManager := versioning.NewVersionManager(globalFlags.InstallDir)

	// Get current framework version
	currentVersion, err := versionManager.GetCurrentVersion()
	if err != nil {
		return fmt.Errorf("failed to get version: %w", err)
	}

	if !versionFlags.Components && !versionFlags.All && !versionFlags.Check {
		// Simple version display
		fmt.Printf("Claude Code Super Crew v%s\n", currentVersion)
		return nil
	}

	// Display header
	if !globalFlags.Quiet {
		ui.DisplayHeader(
			fmt.Sprintf("Claude Code Super Crew v%s", currentVersion),
			"Version Information",
		)
	}

	// Show component versions if requested
	if versionFlags.Components || versionFlags.All {
		fmt.Printf("\n%sComponent Versions:%s\n", ui.ColorCyan, ui.ColorReset)
		fmt.Println(strings.Repeat("-", 30))
		
		components := []string{"core", "commands", "hooks", "mcp"}
		for _, comp := range components {
			version, err := versionManager.GetComponentVersion(comp)
			if err != nil {
				fmt.Printf("  %-10s: %s\n", comp, "unknown")
			} else {
				fmt.Printf("  %-10s: v%s\n", comp, version)
			}
		}
	}

	// Show detailed information if requested
	if versionFlags.All {
		metadata, err := versionManager.LoadMetadata()
		if err == nil {
			fmt.Printf("\n%sInstallation Details:%s\n", ui.ColorCyan, ui.ColorReset)
			fmt.Println(strings.Repeat("-", 30))
			
			if metadata.Framework.UpdatedAt.IsZero() {
				fmt.Printf("  Installed:   Unknown\n")
			} else {
				fmt.Printf("  Installed:   %s\n", metadata.Framework.UpdatedAt.Format("2006-01-02 15:04:05"))
			}
			
			if metadata.Framework.PreviousVersion != "" {
				fmt.Printf("  Previous:    v%s\n", metadata.Framework.PreviousVersion)
			}
			
			fmt.Printf("  Location:    %s\n", globalFlags.InstallDir)
		}
		
		// Show version history
		history, err := versionManager.GetVersionHistory()
		if err == nil && len(history) > 1 {
			fmt.Printf("\n%sVersion History:%s\n", ui.ColorCyan, ui.ColorReset)
			fmt.Println(strings.Repeat("-", 30))
			for i, ver := range history {
				if i == 0 {
					fmt.Printf("  v%s (current)\n", ver)
				} else {
					fmt.Printf("  v%s\n", ver)
				}
			}
		}
	}

	// Check for updates if requested
	if versionFlags.Check {
		fmt.Printf("\n%sChecking for updates...%s\n", ui.ColorCyan, ui.ColorReset)
		
		// For now, we'll check against a fixed target version
		// In production, this would check against a remote repository
		targetVersion := "1.0.1" // This would come from a remote source
		
		updateInfo, err := versionManager.CheckUpdateStatus(targetVersion)
		if err != nil {
			return fmt.Errorf("failed to check for updates: %w", err)
		}
		
		if updateInfo.UpdateAvailable {
			fmt.Printf("\n%sUpdate available:%s\n", ui.ColorGreen, ui.ColorReset)
			fmt.Printf("  Current:   v%s\n", updateInfo.CurrentVersion)
			fmt.Printf("  Available: v%s\n", updateInfo.AvailableVersion)
			fmt.Printf("\nRun 'crew update' to update to the latest version.\n")
		} else {
			fmt.Printf("\n%sYou are running the latest version.%s\n", ui.ColorGreen, ui.ColorReset)
		}
	}

	return nil
}