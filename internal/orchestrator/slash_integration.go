package orchestrator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jonwraymond/claude-code-super-crew/pkg/logger"
)

// SlashCommandOrchestrator handles orchestration for slash commands
type SlashCommandOrchestrator struct {
	logger logger.Logger
}

// NewSlashCommandOrchestrator creates a new slash command orchestrator
func NewSlashCommandOrchestrator() *SlashCommandOrchestrator {
	return &SlashCommandOrchestrator{
		logger: logger.GetLogger(),
	}
}

// HandleLoadCommand handles the /crew:load command with orchestrator integration
func (o *SlashCommandOrchestrator) HandleLoadCommand(args []string, workingDir string) error {
	o.logger.Info("Orchestrator: Loading project with intelligent agent support...")

	// If no working directory provided, use current directory
	if workingDir == "" {
		var err error
		workingDir, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
	}

	// Check if this is a project directory (has code files)
	if !o.isProjectDirectory(workingDir) {
		o.logger.Debug("Not a project directory, skipping agent generation")
		return nil
	}

	// Use the new LoadCommandHandler
	loadHandler := NewLoadCommandHandler(workingDir)
	
	// Parse MCP and tool flags from args
	mcpFlags := o.parseMCPFlags(args)
	toolFlags := o.parseToolFlags(args)
	
	// Execute with enhancements if any flags provided
	if len(mcpFlags) > 0 || len(toolFlags) > 0 {
		if len(mcpFlags) > 0 {
			o.logger.Info(fmt.Sprintf("Enabling MCP integrations: %s", strings.Join(mcpFlags, ", ")))
		}
		if len(toolFlags) > 0 {
			o.logger.Info(fmt.Sprintf("Enabling CLI tools: %s", strings.Join(toolFlags, ", ")))
		}
		return loadHandler.ExecuteWithEnhancements(mcpFlags, toolFlags)
	}
	
	// Otherwise, execute standard load
	return loadHandler.Execute()
}

// parseMCPFlags extracts MCP server flags from command arguments
func (o *SlashCommandOrchestrator) parseMCPFlags(args []string) []string {
	var mcpFlags []string
	
	// Official MCP servers in our framework
	officialServers := map[string]bool{
		"context7":    true,
		"sequential":  true,
		"magic":       true,
		"playwright":  true,
		"serena":      true,
	}
	
	for _, arg := range args {
		// Check for --mcp=server1,server2 format
		if strings.HasPrefix(arg, "--mcp=") {
			servers := strings.Split(strings.TrimPrefix(arg, "--mcp="), ",")
			for _, server := range servers {
				server = strings.TrimSpace(strings.ToLower(server))
				if officialServers[server] {
					mcpFlags = append(mcpFlags, server)
				}
			}
		}
		
		// Check for individual --context7, --sequential, etc.
		for server := range officialServers {
			if arg == "--"+server {
				mcpFlags = append(mcpFlags, server)
			}
		}
		
		// Check for --all-mcp flag
		if arg == "--all-mcp" {
			// Add all official servers
			for server := range officialServers {
				mcpFlags = append(mcpFlags, server)
			}
			break // No need to check more flags
		}
	}
	
	// Remove duplicates
	seen := make(map[string]bool)
	unique := []string{}
	for _, server := range mcpFlags {
		if !seen[server] {
			seen[server] = true
			unique = append(unique, server)
		}
	}
	
	return unique
}

// parseToolFlags extracts CLI tool flags from command arguments
func (o *SlashCommandOrchestrator) parseToolFlags(args []string) []string {
	var toolFlags []string
	
	// Official CLI tools in our framework
	officialTools := map[string]bool{
		"code2prompt": true,
		"ast-grep":    true,
	}
	
	for _, arg := range args {
		// Check for --tools=tool1,tool2 format
		if strings.HasPrefix(arg, "--tools=") {
			tools := strings.Split(strings.TrimPrefix(arg, "--tools="), ",")
			for _, tool := range tools {
				tool = strings.TrimSpace(strings.ToLower(tool))
				if officialTools[tool] {
					toolFlags = append(toolFlags, tool)
				}
			}
		}
		
		// Check for individual --code2prompt, --ast-grep, etc.
		for tool := range officialTools {
			if arg == "--"+tool {
				toolFlags = append(toolFlags, tool)
			}
		}
		
		// Check for --all-tools flag
		if arg == "--all-tools" {
			// Add all official tools
			for tool := range officialTools {
				toolFlags = append(toolFlags, tool)
			}
			break // No need to check more flags
		}
	}
	
	// Remove duplicates
	seen := make(map[string]bool)
	unique := []string{}
	for _, tool := range toolFlags {
		if !seen[tool] {
			seen[tool] = true
			unique = append(unique, tool)
		}
	}
	
	return unique
}

// isProjectDirectory checks if the directory contains code files
func (o *SlashCommandOrchestrator) isProjectDirectory(dir string) bool {
	// Check for common project indicators
	indicators := []string{
		"go.mod", "package.json", "requirements.txt", "Cargo.toml",
		"pom.xml", "build.gradle", "composer.json", "Gemfile",
		".git", "Makefile", "CMakeLists.txt", "setup.py",
	}

	for _, indicator := range indicators {
		if _, err := os.Stat(filepath.Join(dir, indicator)); err == nil {
			return true
		}
	}

	// Check for source code directories
	sourceDirs := []string{"src", "lib", "app", "pkg", "internal", "cmd"}
	for _, srcDir := range sourceDirs {
		if _, err := os.Stat(filepath.Join(dir, srcDir)); err == nil {
			return true
		}
	}

	// Check for any code files in the root
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}

	codeExtensions := []string{
		".go", ".js", ".ts", ".jsx", ".tsx", ".py", ".java", ".rs",
		".cpp", ".c", ".h", ".cs", ".rb", ".php", ".swift", ".kt",
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		for _, ext := range codeExtensions {
			if strings.HasSuffix(name, ext) {
				return true
			}
		}
	}

	return false
}

// GenerateProjectAgents is a convenience function that can be called from other parts of the system
func GenerateProjectAgents(projectPath string) error {
	orchestrator := NewSlashCommandOrchestrator()
	return orchestrator.HandleLoadCommand([]string{}, projectPath)
}
