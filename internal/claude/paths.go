// Package claude provides path utilities for the new .crew/ directory structure
package claude

import (
	"os"
	"path/filepath"
)

// PathResolver handles path resolution for the new .crew/ directory structure
type PathResolver struct {
	claudeDir string
	crewDir   string
}

// NewPathResolver creates a new path resolver
func NewPathResolver(claudeDir string) *PathResolver {
	if claudeDir == "" {
		homeDir, _ := os.UserHomeDir()
		claudeDir = filepath.Join(homeDir, ".claude")
	}
	
	return &PathResolver{
		claudeDir: claudeDir,
		crewDir:   filepath.Join(claudeDir, ".crew"),
	}
}

// GetClaudeDir returns the main .claude directory
func (pr *PathResolver) GetClaudeDir() string {
	return pr.claudeDir
}

// GetCrewDir returns the .crew utility directory
func (pr *PathResolver) GetCrewDir() string {
	return pr.crewDir
}

// Utility directory paths (under .crew/)

// GetBackupsDir returns the backups directory path
func (pr *PathResolver) GetBackupsDir() string {
	return filepath.Join(pr.crewDir, "backups")
}

// GetLogsDir returns the logs directory path
func (pr *PathResolver) GetLogsDir() string {
	return filepath.Join(pr.crewDir, "logs")
}

// GetConfigDir returns the config directory path
func (pr *PathResolver) GetConfigDir() string {
	return filepath.Join(pr.crewDir, "config")
}

// GetCompletionsDir returns the completions directory path
func (pr *PathResolver) GetCompletionsDir() string {
	return filepath.Join(pr.crewDir, "completions")
}

// GetScriptsDir returns the scripts directory path
func (pr *PathResolver) GetScriptsDir() string {
	return filepath.Join(pr.crewDir, "scripts")
}

// GetWorkflowsDir returns the workflows directory path
func (pr *PathResolver) GetWorkflowsDir() string {
	return filepath.Join(pr.crewDir, "workflows")
}

// GetPromptsDir returns the prompts directory path
func (pr *PathResolver) GetPromptsDir() string {
	return filepath.Join(pr.crewDir, "prompts")
}

// GetCrewPath returns the path for any directory under .crew/
func (pr *PathResolver) GetCrewPath(dirName string) string {
	return filepath.Join(pr.crewDir, dirName)
}

// Core SuperCrew directories (remain in main .claude/)

// GetCommandsDir returns the commands directory path
func (pr *PathResolver) GetCommandsDir() string {
	return filepath.Join(pr.claudeDir, "commands")
}

// GetHooksDir returns the hooks directory path
func (pr *PathResolver) GetHooksDir() string {
	return filepath.Join(pr.claudeDir, "hooks")
}

// GetAgentsDir returns the agents directory path
func (pr *PathResolver) GetAgentsDir() string {
	return filepath.Join(pr.claudeDir, "agents")
}

// Framework files (remain in main .claude/)

// GetFrameworkFiles returns paths to core framework files
func (pr *PathResolver) GetFrameworkFiles() map[string]string {
	return map[string]string{
		"CLAUDE.md":       filepath.Join(pr.claudeDir, "CLAUDE.md"),
		"COMMANDS.md":     filepath.Join(pr.claudeDir, "COMMANDS.md"),
		"FLAGS.md":        filepath.Join(pr.claudeDir, "FLAGS.md"),
		"PRINCIPLES.md":   filepath.Join(pr.claudeDir, "PRINCIPLES.md"),
		"RULES.md":        filepath.Join(pr.claudeDir, "RULES.md"),
		"MCP.md":          filepath.Join(pr.claudeDir, "MCP.md"),
		"PERSONAS.md":     filepath.Join(pr.claudeDir, "PERSONAS.md"),
		"ORCHESTRATOR.md": filepath.Join(pr.claudeDir, "ORCHESTRATOR.md"),
		"MODES.md":        filepath.Join(pr.claudeDir, "MODES.md"),
	}
}

// GetMainConfigFile returns the main supercrew-commands.json config file path
func (pr *PathResolver) GetMainConfigFile() string {
	return filepath.Join(pr.claudeDir, "supercrew-commands.json")
}

// GetInstallationMetadata returns the installation metadata file path
func (pr *PathResolver) GetInstallationMetadata() string {
	return filepath.Join(pr.crewDir, "config", "crew-metadata.json")
}

// GetUserSettings returns the user settings file path
func (pr *PathResolver) GetUserSettings() string {
	return filepath.Join(pr.crewDir, "config", "settings.json")
}

// Convenience methods for common operations

// EnsureCrewDirectories creates all .crew utility directories
func (pr *PathResolver) EnsureCrewDirectories() error {
	dirs := []string{
		pr.GetBackupsDir(),
		pr.GetLogsDir(),
		pr.GetConfigDir(),
		pr.GetCompletionsDir(),
		pr.GetScriptsDir(),
		pr.GetWorkflowsDir(),
		pr.GetPromptsDir(),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}

// EnsureCoreDirectories creates all core SuperCrew directories
func (pr *PathResolver) EnsureCoreDirectories() error {
	dirs := []string{
		pr.GetCommandsDir(),
		pr.GetHooksDir(),
		pr.GetAgentsDir(),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}

// GetAllDirectoryPaths returns all directory paths with their purposes
func (pr *PathResolver) GetAllDirectoryPaths() map[string]string {
	return map[string]string{
		// Utility directories under .crew/
		"backups":     pr.GetBackupsDir(),
		"logs":        pr.GetLogsDir(),
		"config":      pr.GetConfigDir(),
		"completions": pr.GetCompletionsDir(),
		"scripts":     pr.GetScriptsDir(),
		"workflows":   pr.GetWorkflowsDir(),
		"prompts":     pr.GetPromptsDir(),
		
		// Core SuperCrew directories
		"commands": pr.GetCommandsDir(),
		"hooks":    pr.GetHooksDir(),
		"agents":   pr.GetAgentsDir(),
	}
}