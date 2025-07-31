// Package cli provides command-line interface for SuperCrew
package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/jonwraymond/claude-code-super-crew/internal/hooks"
	"github.com/jonwraymond/claude-code-super-crew/pkg/logger"
	"github.com/spf13/cobra"
)

// NewHooksCommand creates the hooks command
func NewHooksCommand() *cobra.Command {
	var (
		enableHook  string
		disableHook string
		listHooks   bool
		installHooksOnly bool
	)

	cmd := &cobra.Command{
		Use:   "hooks",
		Short: "Manage Claude Code hooks for automation",
		Long: `Manage Claude Code hooks that run automatically on various events.

Hooks allow you to automate tasks like:
- Auto-committing changes to git
- Running linters on file save
- Running tests on code changes
- Scanning for security vulnerabilities
- Creating backups before modifications`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runHooksInteractive(cmd, args, enableHook, disableHook, listHooks, installHooksOnly)
		},
	}
	
	cmd.Flags().StringVar(&enableHook, "enable", "", "Enable a specific hook")
	cmd.Flags().StringVar(&disableHook, "disable", "", "Disable a specific hook")
	cmd.Flags().BoolVar(&listHooks, "list", false, "List all available hooks")
	cmd.Flags().BoolVar(&installHooksOnly, "install-only", false, "Only install hook scripts without configuration")

	return cmd
}

func runHooksInteractive(cmd *cobra.Command, args []string, enableHook, disableHook string, listHooks, installHooksOnly bool) error {
	lg := logger.GetLogger()
	
	// Get project root
	projectRoot, err := findProjectRoot()
	if err != nil {
		// Use SuperCrew installation directory as fallback
		homeDir, _ := os.UserHomeDir()
		projectRoot = homeDir
	}

	// Create hook manager
	hm := hooks.NewHookManager(projectRoot)
	
	// Discover available hooks
	if err := hm.DiscoverHooks(); err != nil {
		return fmt.Errorf("failed to discover hooks: %w", err)
	}

	// Handle flags
	if listHooks {
		return listAvailableHooks(hm)
	}

	if enableHook != "" {
		return hm.EnableHook(enableHook)
	}

	if disableHook != "" {
		return hm.DisableHook(disableHook)
	}

	if installHooksOnly {
		homeDir, _ := os.UserHomeDir()
		targetDir := fmt.Sprintf("%s/.claude", homeDir)
		return hm.InstallHooks(targetDir)
	}

	// Interactive mode
	return runInteractiveHookManager(hm, lg)
}

func runInteractiveHookManager(hm *hooks.HookManager, lg logger.Logger) error {
	for {
		// Show main menu
		action := ""
		prompt := &survey.Select{
			Message: "What would you like to do?",
			Options: []string{
				"View available hooks",
				"Enable hooks",
				"Disable hooks",
				"Configure hook settings",
				"Configure hooks in settings",
				"Exit",
			},
		}
		
		if err := survey.AskOne(prompt, &action); err != nil {
			return err
		}

		switch action {
		case "View available hooks":
			if err := listAvailableHooks(hm); err != nil {
				lg.Error(err.Error())
			}
			
		case "Enable hooks":
			if err := enableHooksInteractive(hm, lg); err != nil {
				lg.Error(err.Error())
			}
			
		case "Disable hooks":
			if err := disableHooksInteractive(hm, lg); err != nil {
				lg.Error(err.Error())
			}
			
		case "Configure hook settings":
			if err := configureHooksInteractive(hm, lg); err != nil {
				lg.Error(err.Error())
			}
			
		case "Configure hooks in settings":
			homeDir, _ := os.UserHomeDir()
			targetDir := fmt.Sprintf("%s/.claude", homeDir)
			if err := hm.InstallHooks(targetDir); err != nil {
				lg.Error(err.Error())
			} else {
				lg.Success("Hooks installed successfully!")
			}
			
		case "Exit":
			return nil
		}
		
		fmt.Println() // Add spacing
	}
}

func listAvailableHooks(hm *hooks.HookManager) error {
	hooks := hm.ListHooks()
	
	if len(hooks) == 0 {
		fmt.Println("No hooks available")
		return nil
	}

	// Create table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	
	// Header
	fmt.Fprintln(w, "NAME\tSTATUS\tTYPE\tDESCRIPTION")
	fmt.Fprintln(w, "----\t------\t----\t-----------")
	
	// Hooks
	for _, hook := range hooks {
		status := color.RedString("disabled")
		if hook.Enabled {
			status = color.GreenString("enabled")
		}
		
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", 
			hook.Name, 
			status,
			hook.Type,
			hook.Description,
		)
	}
	
	w.Flush()
	return nil
}

func enableHooksInteractive(hm *hooks.HookManager, lg logger.Logger) error {
	hooks := hm.ListHooks()
	
	// Filter disabled hooks
	var disabledHooks []string
	for _, hook := range hooks {
		if !hook.Enabled {
			disabledHooks = append(disabledHooks, 
				fmt.Sprintf("%s - %s", hook.Name, hook.Description))
		}
	}
	
	if len(disabledHooks) == 0 {
		lg.Info("All hooks are already enabled")
		return nil
	}
	
	// Multi-select prompt
	selected := []string{}
	prompt := &survey.MultiSelect{
		Message: "Select hooks to enable:",
		Options: disabledHooks,
	}
	
	if err := survey.AskOne(prompt, &selected); err != nil {
		return err
	}
	
	// Enable selected hooks
	for _, selection := range selected {
		hookName := strings.Split(selection, " - ")[0]
		if err := hm.EnableHook(hookName); err != nil {
			lg.Errorf("Failed to enable %s: %v", hookName, err)
		}
	}
	
	return nil
}

func disableHooksInteractive(hm *hooks.HookManager, lg logger.Logger) error {
	hooks := hm.ListHooks()
	
	// Filter enabled hooks
	var enabledHooks []string
	for _, hook := range hooks {
		if hook.Enabled {
			enabledHooks = append(enabledHooks, 
				fmt.Sprintf("%s - %s", hook.Name, hook.Description))
		}
	}
	
	if len(enabledHooks) == 0 {
		lg.Info("No hooks are enabled")
		return nil
	}
	
	// Multi-select prompt
	selected := []string{}
	prompt := &survey.MultiSelect{
		Message: "Select hooks to disable:",
		Options: enabledHooks,
	}
	
	if err := survey.AskOne(prompt, &selected); err != nil {
		return err
	}
	
	// Disable selected hooks
	for _, selection := range selected {
		hookName := strings.Split(selection, " - ")[0]
		if err := hm.DisableHook(hookName); err != nil {
			lg.Errorf("Failed to disable %s: %v", hookName, err)
		}
	}
	
	return nil
}

func configureHooksInteractive(hm *hooks.HookManager, lg logger.Logger) error {
	hooks := hm.ListHooks()
	
	// Build hook options
	var hookOptions []string
	for _, hook := range hooks {
		hookOptions = append(hookOptions, 
			fmt.Sprintf("%s - %s", hook.Name, hook.Description))
	}
	
	// Select hook to configure
	var selected string
	prompt := &survey.Select{
		Message: "Select hook to configure:",
		Options: hookOptions,
	}
	
	if err := survey.AskOne(prompt, &selected); err != nil {
		return err
	}
	
	hookName := strings.Split(selected, " - ")[0]
	hook, err := hm.GetHookInfo(hookName)
	if err != nil {
		return err
	}
	
	// Show current configuration
	fmt.Println("\nCurrent configuration:")
	for key, value := range hook.Config {
		fmt.Printf("  %s = %s\n", key, value)
	}
	fmt.Println()
	
	// Configure each setting
	newConfig := make(map[string]string)
	for key, currentValue := range hook.Config {
		var newValue string
		prompt := &survey.Input{
			Message: fmt.Sprintf("%s:", key),
			Default: currentValue,
		}
		
		if err := survey.AskOne(prompt, &newValue); err != nil {
			return err
		}
		
		newConfig[key] = newValue
	}
	
	// Apply configuration
	if err := hm.ConfigureHook(hookName, newConfig); err != nil {
		return err
	}
	
	lg.Success("Configuration updated!")
	return nil
}

// findProjectRoot attempts to find the project root directory
func findProjectRoot() (string, error) {
	// Start from current directory
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Look for .git directory or go.mod file
	for {
		// Check for .git directory
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}

		// Check for go.mod file
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("project root not found")
}