package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jonwraymond/claude-code-super-crew/pkg/logger"
)

// AgentsComponent implements the agents framework component.
// It manages persona subagent files and other agent-related templates
// that are installed in the agents directory for use by Claude Code.
type AgentsComponent struct {
	BaseComponent
	sourceDir string
	log       logger.Logger
}

// NewAgentsComponent creates a new agents component instance
func NewAgentsComponent(installDir, sourceDir string) *AgentsComponent {
	c := &AgentsComponent{
		BaseComponent: BaseComponent{
			InstallDir: installDir,
			Metadata: ComponentMetadata{
				Name:        "agents",
				Version:     AgentsComponentVersion,
				Description: "Claude Code Super Crew persona subagent files and templates",
				Category:    "agents",
				Author:      "Claude Code Super Crew Team",
				Tags:        []string{"personas", "subagents", "templates"},
				Dependencies: []string{"core"}, // Agents depend on core being installed
			},
		},
		sourceDir: sourceDir,
		log:       logger.GetLogger(),
	}

	// Initialize managers
	c.InitManagers(installDir)

	// Discover agent files
	if sourceDir != "" {
		// Look for .md files in the agents directory
		if files, err := c.DiscoverFiles(sourceDir, ".md", []string{}); err == nil {
			c.ComponentFiles = files
			c.log.Debug(fmt.Sprintf("Discovered %d agent files: %v", len(files), files))
		} else {
			c.log.Error(fmt.Sprintf("Failed to discover agent files: %v", err))
		}
	}

	return c
}

// GetFilesToInstall returns list of agent files to install
func (c *AgentsComponent) GetFilesToInstall() []FilePair {
	pairs := make([]FilePair, 0, len(c.ComponentFiles))

	for _, file := range c.ComponentFiles {
		source := filepath.Join(c.sourceDir, file)
		target := filepath.Join(c.InstallDir, "agents", file)
		pairs = append(pairs, FilePair{
			Source: source,
			Target: target,
		})
	}

	return pairs
}

// ValidatePrerequisites checks if the agents component can be installed
func (c *AgentsComponent) ValidatePrerequisites(installDir string) (bool, []string) {
	var errors []string

	// Call base validation
	isValid, baseErrors := c.BaseComponent.ValidatePrerequisites(installDir)
	if !isValid {
		errors = append(errors, baseErrors...)
	}

	// Check that core component is installed (dependency)
	coreMarker := filepath.Join(installDir, "CLAUDE.md")
	if !c.FileManager.FileExists(coreMarker) {
		errors = append(errors, "Core component must be installed before agents component")
	}

	// Check source directory exists
	if c.sourceDir == "" {
		errors = append(errors, "Agents source directory not specified")
	} else if !c.FileManager.IsDirectory(c.sourceDir) {
		errors = append(errors, fmt.Sprintf("Agents source directory not found: %s", c.sourceDir))
	}

	// Check agents directory can be created
	agentsDir := filepath.Join(installDir, "agents")
	if err := c.FileManager.EnsureDirectory(agentsDir); err != nil {
		errors = append(errors, fmt.Sprintf("Cannot create agents directory: %v", err))
	}

	// Validate all agent files for security
	filesToInstall := c.GetFilesToInstall()
	fileNames := make([]string, len(filesToInstall))
	for i, pair := range filesToInstall {
		fileNames[i] = filepath.Base(pair.Source)
	}

	targetDir := filepath.Join(installDir, "agents")
	isSecure, securityErrors := c.SecurityValidator.ValidateComponentFiles(fileNames, c.sourceDir, targetDir)
	if !isSecure {
		errors = append(errors, securityErrors...)
	}

	return len(errors) == 0, errors
}

// Install creates the agents directory and installs agent files
func (c *AgentsComponent) Install(installDir string, config map[string]interface{}) error {
	c.log.Info("=== AGENTS COMPONENT INSTALL METHOD CALLED ===")
	c.log.Info(fmt.Sprintf("Installing agents component version %s", c.Metadata.Version))

	// Validate prerequisites
	isValid, errors := c.ValidatePrerequisites(installDir)
	if !isValid {
		for _, err := range errors {
			c.log.Error(fmt.Sprintf("Validation error: %s", err))
		}
		return fmt.Errorf("prerequisites validation failed")
	}

	// Ensure agents directory exists
	agentsDir := filepath.Join(installDir, "agents")
	if err := c.FileManager.EnsureDirectory(agentsDir); err != nil {
		return fmt.Errorf("failed to create agents directory: %w", err)
	}

	// Copy all agent files
	filesToInstall := c.GetFilesToInstall()
	successCount := 0

	for _, pair := range filesToInstall {
		c.log.Debug(fmt.Sprintf("Installing agent file from %s to %s", pair.Source, pair.Target))

		if err := c.FileManager.CopyFile(pair.Source, pair.Target); err != nil {
			c.log.Error(fmt.Sprintf("Failed to install agent file %s: %v", filepath.Base(pair.Source), err))
			continue
		}

		// Set appropriate permissions for agent files
		if err := os.Chmod(pair.Target, 0644); err != nil {
			c.log.Warn(fmt.Sprintf("Failed to set permissions on agent file %s: %v", filepath.Base(pair.Target), err))
		}

		successCount++
		c.log.Debug(fmt.Sprintf("Successfully installed agent file %s", filepath.Base(pair.Source)))
	}

	if successCount != len(filesToInstall) {
		return fmt.Errorf("only %d/%d agent files installed successfully", successCount, len(filesToInstall))
	}

	// Update settings.json with agents component version
	if err := c.SettingsManager.UpdateComponentVersion(c.Metadata.Name, c.Metadata.Version); err != nil {
		c.log.Error(fmt.Sprintf("Failed to update settings.json for agents component: %v", err))
		// Don't fail installation for settings update failure
	}

	c.log.Info(fmt.Sprintf("Agents component installed successfully with %d files", successCount))
	return nil
}

// Update backs up existing agent files and installs the new version
func (c *AgentsComponent) Update(installDir string, config map[string]interface{}) error {
	c.log.Info(fmt.Sprintf("Updating agents component from version %s to %s", 
		c.GetInstalledVersion(installDir), c.Metadata.Version))

	// Create backup of existing agent files
	agentsDir := filepath.Join(installDir, "agents")
	if c.FileManager.IsDirectory(agentsDir) {
		backupDir := filepath.Join(installDir, ".crew", "backups", fmt.Sprintf("agents-backup-%s", c.GetInstalledVersion(installDir)))
		if err := c.FileManager.EnsureDirectory(backupDir); err == nil {
			// Copy existing agent files to backup
			if entries, err := os.ReadDir(agentsDir); err == nil {
				for _, entry := range entries {
					if strings.HasSuffix(entry.Name(), ".md") {
						srcPath := filepath.Join(agentsDir, entry.Name())
						dstPath := filepath.Join(backupDir, entry.Name())
						c.FileManager.CopyFile(srcPath, dstPath)
					}
				}
				c.log.Info(fmt.Sprintf("Created backup of existing agent files in %s", backupDir))
			}
		}
	}

	// Perform installation (will overwrite existing files)
	return c.Install(installDir, config)
}

// Uninstall removes agent files but preserves user-created agent files
func (c *AgentsComponent) Uninstall(installDir string, config map[string]interface{}) error {
	c.log.Info("Uninstalling agents component")

	agentsDir := filepath.Join(installDir, "agents")
	if !c.FileManager.IsDirectory(agentsDir) {
		c.log.Info("Agents directory not found, nothing to uninstall")
		return nil
	}

	// List of standard agent files to remove (preserve user-created agents)
	standardAgentFiles := []string{
		"architect-persona.md",
		"frontend-persona.md", 
		"backend-persona.md",
		"security-persona.md",
		"performance-persona.md",
		"analyzer-persona.md",
		"qa-persona.md",
		"refactorer-persona.md",
		"devops-persona.md",
		"mentor-persona.md",
		"scribe-persona.md",
		"orchestrator-agent.md",
		"second-opinion-generator.md",
	}

	// Only remove standard agent files, preserve user-created ones
	removedCount := 0
	for _, fileName := range standardAgentFiles {
		filePath := filepath.Join(agentsDir, fileName)
		if c.FileManager.FileExists(filePath) {
			if err := os.Remove(filePath); err != nil {
				c.log.Warn(fmt.Sprintf("Failed to remove agent file %s: %v", fileName, err))
			} else {
				removedCount++
				c.log.Debug(fmt.Sprintf("Removed agent file %s", fileName))
			}
		}
	}

	// Update settings.json to remove agents component
	if _, err := c.SettingsManager.RemoveComponentRegistration(c.Metadata.Name); err != nil {
		c.log.Warn(fmt.Sprintf("Failed to update settings.json during uninstall: %v", err))
	}

	c.log.Info(fmt.Sprintf("Agents component uninstalled, removed %d standard agent files", removedCount))
	return nil
}

// Validate ensures the agents component is properly configured
func (c *AgentsComponent) Validate(installDir string) error {
	isValid, errors := c.ValidatePrerequisites(installDir)
	if !isValid {
		return fmt.Errorf("validation failed: %v", errors)
	}
	return nil
}

// IsInstalled checks if the agents component is installed
func (c *AgentsComponent) IsInstalled(installDir string) bool {
	// Check settings.json first
	if !c.BaseComponent.IsInstalled(installDir) {
		return false
	}

	// Check if agents directory exists and has at least one persona file
	agentsDir := filepath.Join(installDir, "agents")
	if !c.FileManager.IsDirectory(agentsDir) {
		return false
	}

	// Look for at least one persona file as a marker
	personaMarkers := []string{
		"architect-persona.md",
		"frontend-persona.md",
		"orchestrator-agent.md",
	}

	for _, marker := range personaMarkers {
		if c.FileManager.FileExists(filepath.Join(agentsDir, marker)) {
			return true
		}
	}

	return false
}

// GetInstalledVersion reads the version from the installed component
func (c *AgentsComponent) GetInstalledVersion(installDir string) string {
	return c.BaseComponent.GetInstalledVersion(installDir)
}

// GetSizeEstimate returns the estimated installation size for agent files
func (c *AgentsComponent) GetSizeEstimate() int64 {
	var totalSize int64

	for _, pair := range c.GetFilesToInstall() {
		if size, err := c.FileManager.GetFileSize(pair.Source); err == nil {
			totalSize += size
		}
	}

	// Return calculated size or default if no files found
	if totalSize > 0 {
		return totalSize
	}
	return 500 * 1024 // 500KB default for agent files
}

// ValidateInstallation checks if agents component is correctly installed
func (c *AgentsComponent) ValidateInstallation(installDir string) (bool, []string) {
	isValid, errors := c.BaseComponent.ValidateInstallation(installDir)

	// Check that agents directory exists
	agentsDir := filepath.Join(installDir, "agents")
	if !c.FileManager.IsDirectory(agentsDir) {
		errors = append(errors, "agents directory not found")
		isValid = false
	}

	// Check for essential persona files
	essentialPersonas := []string{
		"architect-persona.md",
		"frontend-persona.md", 
		"backend-persona.md",
		"security-persona.md",
	}

	missingPersonas := []string{}
	for _, persona := range essentialPersonas {
		personaPath := filepath.Join(agentsDir, persona)
		if !c.FileManager.FileExists(personaPath) {
			missingPersonas = append(missingPersonas, persona)
		}
	}

	if len(missingPersonas) > 0 {
		errors = append(errors, fmt.Sprintf("Missing essential persona files: %s", strings.Join(missingPersonas, ", ")))
		isValid = false
	}

	return isValid, errors
}

// GetAgentCount returns the number of installed agent files
func (c *AgentsComponent) GetAgentCount(installDir string) int {
	agentsDir := filepath.Join(installDir, "agents")
	if !c.FileManager.IsDirectory(agentsDir) {
		return 0
	}

	count := 0
	if entries, err := os.ReadDir(agentsDir); err == nil {
		for _, entry := range entries {
			if strings.HasSuffix(entry.Name(), ".md") && !entry.IsDir() {
				count++
			}
		}
	}

	return count
}

// ListInstalledAgents returns a list of installed agent files
func (c *AgentsComponent) ListInstalledAgents(installDir string) []string {
	agentsDir := filepath.Join(installDir, "agents")
	if !c.FileManager.IsDirectory(agentsDir) {
		return []string{}
	}

	agents := []string{}
	if entries, err := os.ReadDir(agentsDir); err == nil {
		for _, entry := range entries {
			if strings.HasSuffix(entry.Name(), ".md") && !entry.IsDir() {
				agents = append(agents, entry.Name())
			}
		}
	}

	return agents
}

// ValidateVersionCompatibility checks if the current agents version is compatible with the framework
func (c *AgentsComponent) ValidateVersionCompatibility(installDir string) error {
	// Get current installed version
	installedVersion := c.GetInstalledVersion(installDir)
	if installedVersion == "" {
		return fmt.Errorf("agents component not installed")
	}

	// Get framework version from metadata
	if settingsManager := c.SettingsManager; settingsManager != nil {
		if metadata, err := settingsManager.LoadMetadata(); err == nil {
			if framework, ok := metadata["framework"].(map[string]interface{}); ok {
				if version, ok := framework["version"].(string); ok && version != "" {
					// Check compatibility (for now, require exact match for major version)
					if !c.isVersionCompatible(installedVersion, version) {
						return fmt.Errorf("agents version %s is not compatible with framework version %s", 
							installedVersion, version)
					}
				}
			}
		} else {
			c.log.Warn("Framework version not found in metadata")
			return nil // Allow operation to continue
		}
	}

	return nil
}

// isVersionCompatible checks if two versions are compatible
func (c *AgentsComponent) isVersionCompatible(agentsVersion, frameworkVersion string) bool {
	// For now, use simple major version compatibility
	// In production, this could be more sophisticated
	agentsMajor := strings.Split(agentsVersion, ".")[0]
	frameworkMajor := strings.Split(frameworkVersion, ".")[0]
	
	return agentsMajor == frameworkMajor
}

// RequiresUpdate checks if the agents component needs to be updated
func (c *AgentsComponent) RequiresUpdate(installDir string) (bool, string, error) {
	installedVersion := c.GetInstalledVersion(installDir)
	if installedVersion == "" {
		return true, c.Metadata.Version, nil // Not installed, needs installation
	}

	// Compare with available version
	if installedVersion != c.Metadata.Version {
		return true, c.Metadata.Version, nil
	}

	// Check if any essential agent files are missing
	isValid, errors := c.ValidateInstallation(installDir)
	if !isValid {
		c.log.Debug(fmt.Sprintf("Agents component validation failed: %v", errors))
		return true, c.Metadata.Version, fmt.Errorf("agents component incomplete: %s", strings.Join(errors, ", "))
	}

	return false, installedVersion, nil
}

// GetUpdateStrategy returns the recommended update strategy for the agents component
func (c *AgentsComponent) GetUpdateStrategy(installDir string) string {
	installedVersion := c.GetInstalledVersion(installDir)
	
	if installedVersion == "" {
		return "install" // Fresh installation
	}
	
	// Check if this is a major version change
	installedMajor := strings.Split(installedVersion, ".")[0]
	newMajor := strings.Split(c.Metadata.Version, ".")[0]
	
	if installedMajor != newMajor {
		return "upgrade" // Major version upgrade
	}
	
	return "update" // Minor/patch update
}

// PrepareForUpdate prepares the agents component for update
func (c *AgentsComponent) PrepareForUpdate(installDir string) error {
	// Validate current installation
	if err := c.ValidateVersionCompatibility(installDir); err != nil {
		c.log.Warn(fmt.Sprintf("Version compatibility warning: %v", err))
	}

	// Ensure backup directory exists
	backupDir := filepath.Join(installDir, ".crew", "backups")
	if err := c.FileManager.EnsureDirectory(backupDir); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Check disk space (estimate needed space)
	estimatedSize := c.GetSizeEstimate()
	// Add 20% buffer for temporary files during update
	requiredSpace := estimatedSize + (estimatedSize / 5)
	
	// Simple check - ensure we have at least the estimated space
	// In production, this would be more sophisticated
	if availableSpace, err := c.getAvailableDiskSpace(installDir); err == nil {
		if availableSpace < requiredSpace {
			return fmt.Errorf("insufficient disk space: need %d bytes, have %d bytes", 
				requiredSpace, availableSpace)
		}
	}

	return nil
}

// getAvailableDiskSpace estimates available disk space (simplified implementation)
func (c *AgentsComponent) getAvailableDiskSpace(path string) (int64, error) {
	// This is a simplified implementation
	// In production, you'd use syscall to get actual disk space
	return 100 * 1024 * 1024, nil // Assume 100MB available for now
}