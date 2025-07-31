package core

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// MCPServerInfo represents configuration for an MCP server
type MCPServerInfo struct {
	Name              string `json:"name"`
	Description       string `json:"description"`
	NPMPackage        string `json:"npm_package"`
	Required          bool   `json:"required"`
	APIKeyEnv         string `json:"api_key_env,omitempty"`
	APIKeyDescription string `json:"api_key_description,omitempty"`
}

// MCPComponent implements MCP servers integration
type MCPComponent struct {
	BaseComponent
	MCPServers map[string]MCPServerInfo
}

// NewMCPComponent creates a new MCP component
func NewMCPComponent() *MCPComponent {
	comp := &MCPComponent{
		BaseComponent: BaseComponent{
			Metadata: ComponentMetadata{
				Name:         "mcp",
				Version:      MCPComponentVersion,
				Description:  "MCP server integration (Context7, Sequential, Magic, Playwright)",
				Category:     "integration",
				Author:       "Claude Code Super Crew Team",
				Tags:         []string{"mcp", "servers", "integration", "claude-desktop"},
				Dependencies: []string{"core"},
				Requirements: map[string]string{
					"node":   ">=18.0.0",
					"npm":    ">=8.0.0",
					"claude": ">=1.0.0",
				},
			},
		},
		MCPServers: map[string]MCPServerInfo{
			"sequential-thinking": {
				Name:        "sequential-thinking",
				Description: "Multi-step problem solving and systematic analysis",
				NPMPackage:  "@modelcontextprotocol/server-sequential-thinking",
				Required:    true,
			},
			"context7": {
				Name:        "context7",
				Description: "Official library documentation and code examples",
				NPMPackage:  "@upstash/context7-mcp",
				Required:    true,
			},
			"magic": {
				Name:              "magic",
				Description:       "Modern UI component generation and design systems",
				NPMPackage:        "@21st-dev/magic",
				Required:          false,
				APIKeyEnv:         "TWENTYFIRST_API_KEY",
				APIKeyDescription: "21st.dev API key for UI component generation",
			},
			"playwright": {
				Name:        "playwright",
				Description: "Cross-browser E2E testing and automation",
				NPMPackage:  "@playwright/mcp@latest",
				Required:    false,
			},
		},
	}
	return comp
}

// Validate checks prerequisites for MCP component
func (c *MCPComponent) Validate(installDir string) error {
	isValid, errors := c.ValidatePrerequisites(installDir)
	if !isValid {
		return fmt.Errorf("MCP prerequisites not met: %s", strings.Join(errors, "; "))
	}
	return nil
}

// ValidatePrerequisites checks MCP-specific prerequisites
func (c *MCPComponent) ValidatePrerequisites(installDir string) (bool, []string) {
	var errors []string

	// Check base prerequisites first
	baseValid, baseErrors := c.BaseComponent.ValidatePrerequisites(installDir)
	if !baseValid {
		errors = append(errors, baseErrors...)
	}

	// Check if Node.js is available
	if err := c.checkNodeJS(); err != nil {
		errors = append(errors, err.Error())
	}

	// Check if Claude CLI is available
	if err := c.checkClaudeCLI(); err != nil {
		errors = append(errors, err.Error())
	}

	// Check if npm is available
	if err := c.checkNPM(); err != nil {
		errors = append(errors, err.Error())
	}

	return len(errors) == 0, errors
}

// checkNodeJS verifies Node.js installation and version
func (c *MCPComponent) checkNodeJS() error {
	cmd := exec.Command("node", "--version")
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", "node --version")
	}

	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("node.js not found - required for MCP servers")
	}

	version := strings.TrimSpace(string(output))
	// Check version (require 18+)
	if len(version) > 1 && version[0] == 'v' {
		versionParts := strings.Split(version[1:], ".")
		if len(versionParts) > 0 {
			if major, err := strconv.Atoi(versionParts[0]); err == nil {
				if major < 18 {
					return fmt.Errorf("Node.js version %s found, but version 18+ required", version)
				}
			}
		}
	}

	return nil
}

// checkClaudeCLI verifies Claude CLI installation
func (c *MCPComponent) checkClaudeCLI() error {
	cmd := exec.Command("claude", "--version")
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", "claude --version")
	}

	_, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Claude CLI not found - required for MCP server management")
	}

	return nil
}

// checkNPM verifies npm installation
func (c *MCPComponent) checkNPM() error {
	cmd := exec.Command("npm", "--version")
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", "npm --version")
	}

	_, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("npm not found - required for MCP server installation")
	}

	return nil
}

// Install installs the MCP component and all servers
func (c *MCPComponent) Install(installDir string, config map[string]interface{}) error {
	c.InitManagers(installDir)

	// Validate prerequisites
	if err := c.Validate(installDir); err != nil {
		return err
	}

	// Install each MCP server
	installedCount := 0
	var failedServers []string

	for serverName, serverInfo := range c.MCPServers {
		if err := c.installMCPServer(serverInfo, config); err != nil {
			failedServers = append(failedServers, serverName)
			if serverInfo.Required {
				return fmt.Errorf("required MCP server %s failed to install: %w", serverName, err)
			}
		} else {
			installedCount++
		}
	}

	// Verify installation if not dry run
	if dryRun, ok := config["dry_run"].(bool); !ok || !dryRun {
		if err := c.verifyMCPInstallation(); err != nil {
			return fmt.Errorf("MCP server verification failed: %w", err)
		}
	}

	// Register component in metadata
	if err := c.postInstall(); err != nil {
		return fmt.Errorf("failed to register component: %w", err)
	}

	if len(failedServers) > 0 {
		return fmt.Errorf("some MCP servers failed to install: %v", failedServers)
	}

	return nil
}

// installMCPServer installs a single MCP server
func (c *MCPComponent) installMCPServer(serverInfo MCPServerInfo, config map[string]interface{}) error {
	// Check if already installed
	if installed, err := c.checkMCPServerInstalled(serverInfo.Name); err == nil && installed {
		return nil // Already installed
	}

	// Handle API key requirements
	if serverInfo.APIKeyEnv != "" {
		if dryRun, ok := config["dry_run"].(bool); !ok || !dryRun {
			if os.Getenv(serverInfo.APIKeyEnv) == "" {
				// Log warning but continue
				fmt.Printf("Warning: API key %s not found in environment. Server may not function properly.\n", serverInfo.APIKeyEnv)
			}
		}
	}

	// Check for dry run
	if dryRun, ok := config["dry_run"].(bool); ok && dryRun {
		fmt.Printf("Would install MCP server (user scope): claude mcp add -s user %s npx -y %s\n",
			serverInfo.Name, serverInfo.NPMPackage)
		return nil
	}

	// Install using Claude CLI
	cmd := exec.Command("claude", "mcp", "add", "-s", "user", "--",
		serverInfo.Name, "npx", "-y", serverInfo.NPMPackage)
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C",
			fmt.Sprintf("claude mcp add -s user -- %s npx -y %s",
				serverInfo.Name, serverInfo.NPMPackage))
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install MCP server %s: %w\nOutput: %s",
			serverInfo.Name, err, string(output))
	}

	return nil
}

// checkMCPServerInstalled checks if an MCP server is installed
func (c *MCPComponent) checkMCPServerInstalled(serverName string) (bool, error) {
	cmd := exec.Command("claude", "mcp", "list")
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", "claude mcp list")
	}

	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	return strings.Contains(strings.ToLower(string(output)), strings.ToLower(serverName)), nil
}

// verifyMCPInstallation verifies that MCP servers are properly installed
func (c *MCPComponent) verifyMCPInstallation() error {
	cmd := exec.Command("claude", "mcp", "list")
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", "claude mcp list")
	}

	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("could not verify MCP server installation: %w", err)
	}

	outputStr := strings.ToLower(string(output))
	for serverName, serverInfo := range c.MCPServers {
		if serverInfo.Required && !strings.Contains(outputStr, strings.ToLower(serverName)) {
			return fmt.Errorf("required MCP server not found: %s", serverName)
		}
	}

	return nil
}

// postInstall registers the component in metadata
func (c *MCPComponent) postInstall() error {
	// Get metadata modifications
	modifications := c.getMetadataModifications()

	// Update metadata
	if err := c.SettingsManager.UpdateMetadata(modifications); err != nil {
		return err
	}

	// Add component registration
	componentInfo := map[string]interface{}{
		"version":       MCPComponentVersion,
		"category":      c.Metadata.Category,
		"servers_count": len(c.MCPServers),
	}

	return c.SettingsManager.AddComponentRegistration(c.Metadata.Name, componentInfo)
}

// getMetadataModifications returns metadata modifications for MCP component
func (c *MCPComponent) getMetadataModifications() map[string]interface{} {
	serverNames := make([]string, 0, len(c.MCPServers))
	for name := range c.MCPServers {
		serverNames = append(serverNames, name)
	}

	return map[string]interface{}{
		"components": map[string]interface{}{
			"mcp": map[string]interface{}{
				"version":       MCPComponentVersion,
				"installed":     true,
				"servers_count": len(c.MCPServers),
			},
		},
		"mcp": map[string]interface{}{
			"enabled":     true,
			"servers":     serverNames,
			"auto_update": false,
		},
	}
}

// Update updates the MCP component
func (c *MCPComponent) Update(installDir string, config map[string]interface{}) error {
	c.InitManagers(installDir)

	// Check current version
	currentVersion := c.GetInstalledVersion(installDir)
	if currentVersion == MCPComponentVersion {
		return nil // Already up to date
	}

	// For MCP servers, update means reinstall to get latest versions
	var failedServers []string
	updatedCount := 0

	for serverName, serverInfo := range c.MCPServers {
		// Uninstall old version
		if installed, err := c.checkMCPServerInstalled(serverName); err == nil && installed {
			if err := c.uninstallMCPServer(serverName); err != nil {
				failedServers = append(failedServers, serverName)
				continue
			}
		}

		// Install new version
		if err := c.installMCPServer(serverInfo, config); err != nil {
			failedServers = append(failedServers, serverName)
		} else {
			updatedCount++
		}
	}

	// Update metadata
	if err := c.postInstall(); err != nil {
		return fmt.Errorf("failed to update metadata: %w", err)
	}

	if len(failedServers) > 0 {
		return fmt.Errorf("some MCP servers failed to update: %v", failedServers)
	}

	return nil
}

// Uninstall removes the MCP component and all servers
func (c *MCPComponent) Uninstall(installDir string, config map[string]interface{}) error {
	c.InitManagers(installDir)

	// Uninstall each MCP server
	for serverName := range c.MCPServers {
		if err := c.uninstallMCPServer(serverName); err != nil {
			// Log error but continue with other servers
			fmt.Printf("Warning: failed to uninstall MCP server %s: %v\n", serverName, err)
		}
	}

	// Remove component registration
	if removed, err := c.SettingsManager.RemoveComponentRegistration(c.Metadata.Name); err != nil {
		return err
	} else if removed {
		// Also remove MCP configuration from metadata
		metadata, err := c.SettingsManager.LoadMetadata()
		if err == nil {
			if _, exists := metadata["mcp"]; exists {
				delete(metadata, "mcp")
				c.SettingsManager.SaveMetadata(metadata)
			}
		}
	}

	return nil
}

// uninstallMCPServer removes a single MCP server
func (c *MCPComponent) uninstallMCPServer(serverName string) error {
	// Check if installed
	if installed, err := c.checkMCPServerInstalled(serverName); err != nil || !installed {
		return nil // Not installed or can't check
	}

	cmd := exec.Command("claude", "mcp", "remove", serverName)
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", fmt.Sprintf("claude mcp remove %s", serverName))
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to uninstall MCP server %s: %w\nOutput: %s",
			serverName, err, string(output))
	}

	return nil
}

// GetSizeEstimate returns estimated installation size
func (c *MCPComponent) GetSizeEstimate() int64 {
	// MCP servers are installed via npm, estimate based on typical sizes
	return 50 * 1024 * 1024 // ~50MB for all servers combined
}

// GetFilesToInstall returns files to install (none for MCP component)
func (c *MCPComponent) GetFilesToInstall() []FilePair {
	return []FilePair{} // MCP servers are installed via Claude CLI, not file copying
}

// ValidateInstallation validates MCP component installation
func (c *MCPComponent) ValidateInstallation(installDir string) (bool, []string) {
	var errors []string

	c.InitManagers(installDir)

	// Check metadata registration
	if registered, err := c.SettingsManager.IsComponentRegistered(c.Metadata.Name); err != nil || !registered {
		errors = append(errors, "MCP component not registered in metadata")
		return false, errors
	}

	// Check version matches
	installedVersion := c.GetInstalledVersion(installDir)
	if installedVersion != MCPComponentVersion {
		errors = append(errors, fmt.Sprintf("Version mismatch: installed %s, expected %s",
			installedVersion, MCPComponentVersion))
	}

	// Check if Claude CLI is available and can list servers
	if err := c.verifyMCPInstallation(); err != nil {
		errors = append(errors, fmt.Sprintf("MCP server verification failed: %v", err))
	}

	return len(errors) == 0, errors
}

// GetInstallationSummary returns installation summary
func (c *MCPComponent) GetInstallationSummary() map[string]interface{} {
	serverNames := make([]string, 0, len(c.MCPServers))
	for name := range c.MCPServers {
		serverNames = append(serverNames, name)
	}

	return map[string]interface{}{
		"component":      c.Metadata.Name,
		"version":        MCPComponentVersion,
		"servers_count":  len(c.MCPServers),
		"mcp_servers":    serverNames,
		"estimated_size": c.GetSizeEstimate(),
		"dependencies":   []string{"core"},
		"required_tools": []string{"node", "npm", "claude"},
	}
}
