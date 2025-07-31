// Package claude provides Claude Code integration utilities for SuperCrew.
// This enables seamless integration with Claude Code's command system, tab completion,
// and orchestrator-agent routing through /crew: slash commands.
package claude

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jonwraymond/claude-code-super-crew/pkg/logger"
)

// ClaudeIntegration manages integration with Claude Code
type ClaudeIntegration struct {
	registry     *SlashCommandRegistry
	claudeDir    string
	configFile   string
	pathResolver *PathResolver
	logger       logger.Logger
}

// IntegrationConfig represents the Claude Code integration configuration
type IntegrationConfig struct {
	Version        string          `json:"version"`
	Commands       []*SlashCommand `json:"commands"`
	CompletionPath string          `json:"completion_path,omitempty"`
	Metadata       IntegrationMeta `json:"metadata"`
}

// IntegrationMeta contains integration metadata
type IntegrationMeta struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Author       string `json:"author"`
	Homepage     string `json:"homepage"`
	CommandCount int    `json:"command_count"`
}

// NewClaudeIntegration creates a new Claude Code integration manager
func NewClaudeIntegration(commandsPath, claudeDir string) (*ClaudeIntegration, error) {
	registry := NewSlashCommandRegistry(commandsPath)
	if err := registry.LoadCommands(); err != nil {
		return nil, fmt.Errorf("failed to load commands: %w", err)
	}

	pathResolver := NewPathResolver(claudeDir)
	
	integration := &ClaudeIntegration{
		registry:     registry,
		claudeDir:    claudeDir,
		configFile:   pathResolver.GetMainConfigFile(),
		pathResolver: pathResolver,
		logger:       logger.GetLogger(),
	}

	return integration, nil
}

// InstallIntegration installs SuperCrew integration into Claude Code
func (ci *ClaudeIntegration) InstallIntegration() error {
	ci.logger.Info("Installing SuperCrew integration for Claude Code...")

	// Ensure Claude directory exists
	if err := os.MkdirAll(ci.claudeDir, 0755); err != nil {
		return fmt.Errorf("failed to create Claude directory: %w", err)
	}

	// Perform directory restructuring if needed
	if err := ci.ensureDirectoryStructure(); err != nil {
		ci.logger.Warnf("Directory restructuring failed: %v", err)
		// Continue with installation even if restructuring fails
	}

	// Generate integration configuration
	config := ci.generateIntegrationConfig()

	// Write configuration file
	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(ci.configFile, configData, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// Install completion scripts
	if err := ci.installCompletionScripts(); err != nil {
		return fmt.Errorf("failed to install completion scripts: %w", err)
	}

	ci.logger.Successf("SuperCrew integration installed successfully")
	ci.logger.Infof("Configuration file: %s", ci.configFile)
	ci.logger.Infof("Commands available: %d", len(config.Commands))

	return nil
}

// generateIntegrationConfig creates the integration configuration
func (ci *ClaudeIntegration) generateIntegrationConfig() *IntegrationConfig {
	commands := ci.registry.ListCommands()

	return &IntegrationConfig{
		Version:  "1.0.0",
		Commands: commands,
		Metadata: IntegrationMeta{
			Name:         "Claude Code Super Crew",
			Description:  "SuperCrew framework integration for Claude Code with /crew: commands",
			Author:       "SuperCrew Framework",
			Homepage:     "https://github.com/supercrew/claude-code-super-crew",
			CommandCount: len(commands),
		},
	}
}

// installCompletionScripts installs shell completion scripts
func (ci *ClaudeIntegration) installCompletionScripts() error {
	completionDir := ci.pathResolver.GetCompletionsDir()
	if err := os.MkdirAll(completionDir, 0755); err != nil {
		return fmt.Errorf("failed to create completion directory: %w", err)
	}

	shells := []string{"bash", "zsh", "fish"}
	for _, shell := range shells {
		script, err := ci.registry.GenerateCompletionScript(shell)
		if err != nil {
			ci.logger.Warnf("Failed to generate %s completion: %v", shell, err)
			continue
		}

		scriptPath := filepath.Join(completionDir, fmt.Sprintf("supercrew.%s", shell))
		if err := os.WriteFile(scriptPath, []byte(script), 0644); err != nil {
			ci.logger.Warnf("Failed to write %s completion script: %v", shell, err)
			continue
		}

		ci.logger.Infof("Installed %s completion: %s", shell, scriptPath)
	}

	return nil
}

// UpdateIntegration updates the Claude Code integration
func (ci *ClaudeIntegration) UpdateIntegration() error {
	ci.logger.Info("Updating SuperCrew integration...")

	// Perform directory restructuring if needed
	if err := ci.ensureDirectoryStructure(); err != nil {
		ci.logger.Warnf("Directory restructuring failed: %v", err)
		// Continue with update even if restructuring fails
	}

	// Reload commands
	if err := ci.registry.LoadCommands(); err != nil {
		return fmt.Errorf("failed to reload commands: %w", err)
	}

	// Reinstall integration
	return ci.InstallIntegration()
}

// UninstallIntegration removes SuperCrew integration from Claude Code
func (ci *ClaudeIntegration) UninstallIntegration() error {
	ci.logger.Info("Uninstalling SuperCrew integration...")

	// Remove configuration file
	if err := os.Remove(ci.configFile); err != nil && !os.IsNotExist(err) {
		ci.logger.Warnf("Failed to remove config file: %v", err)
	}

	// Remove completion scripts
	completionDir := ci.pathResolver.GetCompletionsDir()
	if err := os.RemoveAll(completionDir); err != nil && !os.IsNotExist(err) {
		ci.logger.Warnf("Failed to remove completion scripts: %v", err)
	}

	ci.logger.Success("SuperCrew integration uninstalled")
	return nil
}

// CheckIntegration verifies the integration status
func (ci *ClaudeIntegration) CheckIntegration() (*IntegrationStatus, error) {
	status := &IntegrationStatus{
		Installed: false,
	}

	// Check if config file exists
	if _, err := os.Stat(ci.configFile); err == nil {
		status.Installed = true
		status.ConfigPath = ci.configFile

		// Read and validate config
		configData, err := os.ReadFile(ci.configFile)
		if err != nil {
			status.Issues = append(status.Issues, fmt.Sprintf("Failed to read config: %v", err))
		} else {
			var config IntegrationConfig
			if err := json.Unmarshal(configData, &config); err != nil {
				status.Issues = append(status.Issues, fmt.Sprintf("Invalid config format: %v", err))
			} else {
				status.Version = config.Version
				status.CommandCount = config.Metadata.CommandCount
			}
		}
	}

	// Check completion scripts
	completionDir := ci.pathResolver.GetCompletionsDir()
	if _, err := os.Stat(completionDir); err == nil {
		status.CompletionInstalled = true
		status.CompletionPath = completionDir

		// Count completion files
		entries, err := os.ReadDir(completionDir)
		if err == nil {
			for _, entry := range entries {
				if !entry.IsDir() && filepath.Ext(entry.Name()) != "" {
					status.CompletionScripts = append(status.CompletionScripts, entry.Name())
				}
			}
		}
	}

	return status, nil
}

// IntegrationStatus represents the current integration status
type IntegrationStatus struct {
	Installed           bool     `json:"installed"`
	Version             string   `json:"version,omitempty"`
	ConfigPath          string   `json:"config_path,omitempty"`
	CommandCount        int      `json:"command_count,omitempty"`
	CompletionInstalled bool     `json:"completion_installed"`
	CompletionPath      string   `json:"completion_path,omitempty"`
	CompletionScripts   []string `json:"completion_scripts,omitempty"`
	Issues              []string `json:"issues,omitempty"`
}

// ExecuteSlashCommand handles execution of /crew: commands
func (ci *ClaudeIntegration) ExecuteSlashCommand(commandLine string) error {
	return ci.registry.ExecuteCommand(commandLine)
}

// GetCommandCompletions returns completions for partial command input
func (ci *ClaudeIntegration) GetCommandCompletions(partial string) []string {
	return ci.registry.GetCompletions(partial)
}

// ListAvailableCommands returns all available commands with descriptions
func (ci *ClaudeIntegration) ListAvailableCommands() []*SlashCommand {
	return ci.registry.ListCommands()
}

// ensureDirectoryStructure ensures the proper .crew/ directory structure
func (ci *ClaudeIntegration) ensureDirectoryStructure() error {
	// Check if restructuring is needed
	crewDir := filepath.Join(ci.claudeDir, ".crew")
	
	// If .crew directory doesn't exist, check if old structure exists and needs migration
	if _, err := os.Stat(crewDir); os.IsNotExist(err) {
		ci.logger.Info("Setting up .crew directory structure...")
		
		// Check if any utility directories exist in old locations
		restructurer := NewDirectoryRestructurer(ci.claudeDir)
		
		// Check for directories that need to be moved
		needsRestructure := false
		for _, dirName := range DirectoriesToMove {
			oldPath := filepath.Join(ci.claudeDir, dirName)
			if _, err := os.Stat(oldPath); err == nil {
				needsRestructure = true
				ci.logger.Debugf("Found directory that needs restructuring: %s", dirName)
				break
			}
		}
		
		if needsRestructure {
			ci.logger.Info("Migrating utility directories to .crew/ structure...")
			if err := restructurer.RestructureDirectories(); err != nil {
				return fmt.Errorf("failed to restructure directories: %w", err)
			}
		} else {
			// Just create the .crew directory structure
			if err := ci.pathResolver.EnsureCrewDirectories(); err != nil {
				return fmt.Errorf("failed to create .crew directories: %w", err)
			}
		}
	}
	
	// Ensure all core directories exist
	if err := ci.pathResolver.EnsureCoreDirectories(); err != nil {
		return fmt.Errorf("failed to create core directories: %w", err)
	}
	
	return nil
}
