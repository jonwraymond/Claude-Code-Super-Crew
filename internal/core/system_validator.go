package core

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// ValidationResult represents the result of a system validation check
type ValidationResult struct {
	Success bool
	Message string
}

// SystemRequirements defines system requirements for components
type SystemRequirements struct {
	Python      *PythonRequirement `json:"python,omitempty"`
	Node        *NodeRequirement   `json:"node,omitempty"`
	DiskSpaceMB int                `json:"disk_space_mb"`
	Tools       map[string]*ToolRequirement `json:"external_tools,omitempty"`
}

// PythonRequirement defines Python version requirements
type PythonRequirement struct {
	MinVersion string `json:"min_version"`
	MaxVersion string `json:"max_version,omitempty"`
}

// NodeRequirement defines Node.js version requirements  
type NodeRequirement struct {
	MinVersion   string   `json:"min_version"`
	MaxVersion   string   `json:"max_version,omitempty"`
	RequiredFor  []string `json:"required_for,omitempty"`
}

// ToolRequirement defines external tool requirements
type ToolRequirement struct {
	Command     string   `json:"command"`
	MinVersion  string   `json:"min_version,omitempty"`
	RequiredFor []string `json:"required_for,omitempty"`
	Optional    bool     `json:"optional,omitempty"`
}

// SystemValidator provides system requirement validation
type SystemValidator struct {
	validationCache map[string]*ValidationResult
}

// NewSystemValidator creates a new system validator
func NewSystemValidator() *SystemValidator {
	return &SystemValidator{
		validationCache: make(map[string]*ValidationResult),
	}
}

// CheckPython validates Python installation and version
func (v *SystemValidator) CheckPython(minVersion, maxVersion string) *ValidationResult {
	cacheKey := fmt.Sprintf("python_%s_%s", minVersion, maxVersion)
	if cached, exists := v.validationCache[cacheKey]; exists {
		return cached
	}

	result := &ValidationResult{}

	// Try different Python commands
	pythonCmds := []string{"python3", "python"}
	var currentVersion string
	var err error

	for _, cmd := range pythonCmds {
		currentVersion, err = v.getPythonVersion(cmd)
		if err == nil {
			break
		}
	}

	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Python not found - required for installation: %v", err)
		v.validationCache[cacheKey] = result
		return result
	}

	// Check minimum version
	if minVersion != "" {
		if valid, err := v.compareVersions(currentVersion, minVersion); err != nil || !valid {
			result.Success = false
			result.Message = fmt.Sprintf("Python %s+ required, found %s", minVersion, currentVersion)
			v.validationCache[cacheKey] = result
			return result
		}
	}

	// Check maximum version
	if maxVersion != "" {
		if valid, err := v.compareVersions(maxVersion, currentVersion); err != nil || !valid {
			result.Success = false
			result.Message = fmt.Sprintf("Python version %s exceeds maximum supported %s", currentVersion, maxVersion)
			v.validationCache[cacheKey] = result
			return result
		}
	}

	result.Success = true
	result.Message = fmt.Sprintf("Python %s meets requirements", currentVersion)
	v.validationCache[cacheKey] = result
	return result
}

// CheckNode validates Node.js installation and version
func (v *SystemValidator) CheckNode(minVersion, maxVersion string) *ValidationResult {
	cacheKey := fmt.Sprintf("node_%s_%s", minVersion, maxVersion)
	if cached, exists := v.validationCache[cacheKey]; exists {
		return cached
	}

	result := &ValidationResult{}

	cmd := exec.Command("node", "--version")
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", "node --version")
	}

	output, err := cmd.Output()
	if err != nil {
		result.Success = false
		result.Message = "Node.js not found in PATH - required for MCP servers"
		v.validationCache[cacheKey] = result
		return result
	}

	versionOutput := strings.TrimSpace(string(output))
	currentVersion := strings.TrimPrefix(versionOutput, "v")

	// Check minimum version
	if minVersion != "" {
		if valid, err := v.compareVersions(currentVersion, minVersion); err != nil || !valid {
			result.Success = false
			result.Message = fmt.Sprintf("Node.js %s+ required, found %s", minVersion, currentVersion)
			v.validationCache[cacheKey] = result
			return result
		}
	}

	// Check maximum version
	if maxVersion != "" {
		if valid, err := v.compareVersions(maxVersion, currentVersion); err != nil || !valid {
			result.Success = false
			result.Message = fmt.Sprintf("Node.js version %s exceeds maximum supported %s", currentVersion, maxVersion)
			v.validationCache[cacheKey] = result
			return result
		}
	}

	result.Success = true
	result.Message = fmt.Sprintf("Node.js %s meets requirements", currentVersion)
	v.validationCache[cacheKey] = result
	return result
}

// CheckClaudeCLI validates Claude CLI installation
func (v *SystemValidator) CheckClaudeCLI(minVersion string) *ValidationResult {
	cacheKey := fmt.Sprintf("claude_cli_%s", minVersion)
	if cached, exists := v.validationCache[cacheKey]; exists {
		return cached
	}

	result := &ValidationResult{}

	cmd := exec.Command("claude", "--version")
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", "claude --version")
	}

	output, err := cmd.Output()
	if err != nil {
		result.Success = false
		result.Message = "Claude CLI not found in PATH - required for MCP server management"
		v.validationCache[cacheKey] = result
		return result
	}

	versionOutput := strings.TrimSpace(string(output))
	
	// Extract version using regex
	versionRegex := regexp.MustCompile(`(\d+\.\d+\.\d+)`)
	matches := versionRegex.FindStringSubmatch(versionOutput)
	
	if len(matches) == 0 {
		result.Success = true
		result.Message = "Claude CLI found (version format unknown)"
		v.validationCache[cacheKey] = result
		return result
	}

	currentVersion := matches[1]

	// Check minimum version if specified
	if minVersion != "" {
		if valid, err := v.compareVersions(currentVersion, minVersion); err != nil || !valid {
			result.Success = false
			result.Message = fmt.Sprintf("Claude CLI %s+ required, found %s", minVersion, currentVersion)
			v.validationCache[cacheKey] = result
			return result
		}
	}

	result.Success = true
	result.Message = fmt.Sprintf("Claude CLI %s found", currentVersion)
	v.validationCache[cacheKey] = result
	return result
}

// CheckExternalTool validates external tool availability and version
func (v *SystemValidator) CheckExternalTool(toolName, command, minVersion string) *ValidationResult {
	cacheKey := fmt.Sprintf("tool_%s_%s_%s", toolName, command, minVersion)
	if cached, exists := v.validationCache[cacheKey]; exists {
		return cached
	}

	result := &ValidationResult{}

	cmdParts := strings.Fields(command)
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		args := append([]string{"/C"}, cmdParts...)
		cmd = exec.Command("cmd", args...)
	} else {
		cmd = exec.Command(cmdParts[0], cmdParts[1:]...)
	}

	// Set timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd = exec.CommandContext(ctx, cmdParts[0], cmdParts[1:]...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("%s not found or command failed", toolName)
		v.validationCache[cacheKey] = result
		return result
	}

	// Extract version if minVersion specified
	if minVersion != "" {
		versionOutput := string(output)
		versionRegex := regexp.MustCompile(`(\d+\.\d+(?:\.\d+)?)`)
		matches := versionRegex.FindStringSubmatch(versionOutput)

		if len(matches) > 0 {
			currentVersion := matches[1]

			if valid, err := v.compareVersions(currentVersion, minVersion); err != nil || !valid {
				result.Success = false
				result.Message = fmt.Sprintf("%s %s+ required, found %s", toolName, minVersion, currentVersion)
				v.validationCache[cacheKey] = result
				return result
			}

			result.Success = true
			result.Message = fmt.Sprintf("%s %s found", toolName, currentVersion)
		} else {
			result.Success = true
			result.Message = fmt.Sprintf("%s found (version unknown)", toolName)
		}
	} else {
		result.Success = true
		result.Message = fmt.Sprintf("%s found", toolName)
	}

	v.validationCache[cacheKey] = result
	return result
}

// ValidateSystemRequirements validates all system requirements
func (v *SystemValidator) ValidateSystemRequirements(requirements *SystemRequirements) (bool, []string) {
	var errors []string

	// Check Python requirements
	if requirements.Python != nil {
		result := v.CheckPython(requirements.Python.MinVersion, requirements.Python.MaxVersion)
		if !result.Success {
			errors = append(errors, fmt.Sprintf("Python: %s", result.Message))
		}
	}

	// Check Node.js requirements
	if requirements.Node != nil {
		result := v.CheckNode(requirements.Node.MinVersion, requirements.Node.MaxVersion)
		if !result.Success {
			errors = append(errors, fmt.Sprintf("Node.js: %s", result.Message))
		}
	}

	// Check external tools
	if requirements.Tools != nil {
		for toolName, toolReq := range requirements.Tools {
			result := v.CheckExternalTool(toolName, toolReq.Command, toolReq.MinVersion)
			if !result.Success && !toolReq.Optional {
				errors = append(errors, fmt.Sprintf("%s: %s", toolName, result.Message))
			}
		}
	}

	return len(errors) == 0, errors
}

// ValidateComponentRequirements validates requirements for specific components
func (v *SystemValidator) ValidateComponentRequirements(componentNames []string, allRequirements *SystemRequirements) (bool, []string) {
	// Start with base requirements
	baseRequirements := &SystemRequirements{
		Python:      allRequirements.Python,
		DiskSpaceMB: allRequirements.DiskSpaceMB,
		Tools:       make(map[string]*ToolRequirement),
	}

	// Check if any component needs Node.js
	nodeComponents := []string{"mcp"} // Components that need Node.js
	nodeRequired := false
	for _, component := range componentNames {
		for _, nodeComp := range nodeComponents {
			if component == nodeComp {
				nodeRequired = true
				break
			}
		}
		if nodeRequired {
			break
		}
	}

	if nodeRequired && allRequirements.Node != nil {
		baseRequirements.Node = allRequirements.Node
	}

	// Add external tools needed by components
	if allRequirements.Tools != nil {
		for toolName, toolReq := range allRequirements.Tools {
			// Check if any of our components need this tool
			for _, component := range componentNames {
				for _, requiredFor := range toolReq.RequiredFor {
					if component == requiredFor {
						baseRequirements.Tools[toolName] = toolReq
						break
					}
				}
			}
		}
	}

	// Validate consolidated requirements
	return v.ValidateSystemRequirements(baseRequirements)
}

// GetSystemInfo returns comprehensive system information
func (v *SystemValidator) GetSystemInfo() map[string]interface{} {
	info := map[string]interface{}{
		"platform":         runtime.GOOS,
		"architecture":     runtime.GOARCH,
		"go_version":       runtime.Version(),
	}

	// Add Node.js info if available
	nodeResult := v.CheckNode("16.0", "")
	info["node_available"] = nodeResult.Success
	if nodeResult.Success {
		info["node_message"] = nodeResult.Message
	}

	// Add Claude CLI info if available
	claudeResult := v.CheckClaudeCLI("")
	info["claude_cli_available"] = claudeResult.Success
	if claudeResult.Success {
		info["claude_cli_message"] = claudeResult.Message
	}

	// Add Python info if available
	pythonResult := v.CheckPython("3.8", "")
	info["python_available"] = pythonResult.Success
	if pythonResult.Success {
		info["python_message"] = pythonResult.Message
	}

	return info
}

// DiagnoseSystem performs comprehensive system diagnostics
func (v *SystemValidator) DiagnoseSystem() map[string]interface{} {
	diagnostics := map[string]interface{}{
		"platform": runtime.GOOS,
		"checks":   make(map[string]interface{}),
		"issues":   []string{},
		"recommendations": []string{},
	}

	checks := diagnostics["checks"].(map[string]interface{})
	var issues []string
	var recommendations []string

	// Check Python
	pythonResult := v.CheckPython("3.8", "")
	checks["python"] = map[string]interface{}{
		"status":  boolToStatus(pythonResult.Success),
		"message": pythonResult.Message,
	}
	if !pythonResult.Success {
		issues = append(issues, "Python version issue")
		recommendations = append(recommendations, v.getInstallationHelp("python"))
	}

	// Check Node.js
	nodeResult := v.CheckNode("16.0", "")
	checks["node"] = map[string]interface{}{
		"status":  boolToStatus(nodeResult.Success),
		"message": nodeResult.Message,
	}
	if !nodeResult.Success {
		issues = append(issues, "Node.js not found or version issue")
		recommendations = append(recommendations, v.getInstallationHelp("node"))
	}

	// Check Claude CLI
	claudeResult := v.CheckClaudeCLI("")
	checks["claude_cli"] = map[string]interface{}{
		"status":  boolToStatus(claudeResult.Success),
		"message": claudeResult.Message,
	}
	if !claudeResult.Success {
		issues = append(issues, "Claude CLI not found")
		recommendations = append(recommendations, v.getInstallationHelp("claude_cli"))
	}

	diagnostics["issues"] = issues
	diagnostics["recommendations"] = recommendations

	return diagnostics
}

// ClearCache clears the validation cache
func (v *SystemValidator) ClearCache() {
	v.validationCache = make(map[string]*ValidationResult)
}

// Helper functions

func (v *SystemValidator) getPythonVersion(pythonCmd string) (string, error) {
	// Validate python command name for security
	if strings.ContainsAny(pythonCmd, "|&;()<>{}$`\\\"'") {
		return "", fmt.Errorf("invalid python command: contains unsafe characters")
	}
	
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		// Use safe command execution on Windows
		cmd = exec.Command("cmd", "/C", pythonCmd, "--version")
	} else {
		cmd = exec.Command(pythonCmd, "--version")
	}

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	versionOutput := strings.TrimSpace(string(output))
	// Python version output format: "Python 3.9.7"
	parts := strings.Fields(versionOutput)
	if len(parts) >= 2 && parts[0] == "Python" {
		return parts[1], nil
	}

	return "", fmt.Errorf("unexpected Python version output: %s", versionOutput)
}

func (v *SystemValidator) compareVersions(version1, version2 string) (bool, error) {
	// Simple version comparison: version1 >= version2
	v1Parts, err := parseVersion(version1)
	if err != nil {
		return false, err
	}

	v2Parts, err := parseVersion(version2)
	if err != nil {
		return false, err
	}

	// Pad shorter version with zeros
	maxLen := len(v1Parts)
	if len(v2Parts) > maxLen {
		maxLen = len(v2Parts)
	}

	for len(v1Parts) < maxLen {
		v1Parts = append(v1Parts, 0)
	}
	for len(v2Parts) < maxLen {
		v2Parts = append(v2Parts, 0)
	}

	// Compare each part
	for i := 0; i < maxLen; i++ {
		if v1Parts[i] > v2Parts[i] {
			return true, nil
		} else if v1Parts[i] < v2Parts[i] {
			return false, nil
		}
	}

	return true, nil // Equal versions
}

func parseVersion(version string) ([]int, error) {
	parts := strings.Split(version, ".")
	var intParts []int

	for _, part := range parts {
		intPart, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid version format: %s", version)
		}
		intParts = append(intParts, intPart)
	}

	return intParts, nil
}

func (v *SystemValidator) getInstallationHelp(tool string) string {
	helpMessages := map[string]string{
		"python": "Install Python 3.8+ from https://python.org/downloads",
		"node":   "Install Node.js 16+ from https://nodejs.org/en/download/",
		"claude_cli": "Install Claude CLI from https://github.com/anthropics/claude-code",
	}

	if help, exists := helpMessages[tool]; exists {
		return help
	}

	return fmt.Sprintf("Please install %s manually", tool)
}

func boolToStatus(success bool) string {
	if success {
		return "pass"
	}
	return "fail"
}

