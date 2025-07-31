package orchestrator

import (
	"fmt"
	"os/exec"
	"strings"
)

// CLITool represents an external CLI tool that can be used by agents
type CLITool struct {
	Name             string
	Command          string
	Description      string
	Installation     string
	CheckCommand     string
	UseCases         []string
	BenefitingAgents []string
	Examples         []CLIExample
}

// CLIExample represents a usage example for a CLI tool
type CLIExample struct {
	Description string
	Command     string
	UseCase     string
}

// CLIToolsEnhancer manages CLI tool integrations
type CLIToolsEnhancer struct {
	tools       map[string]CLITool
	projectRoot string
}

// NewCLIToolsEnhancer creates a new CLI tools enhancer
func NewCLIToolsEnhancer(projectRoot string) *CLIToolsEnhancer {
	return &CLIToolsEnhancer{
		tools:       initializeCLITools(),
		projectRoot: projectRoot,
	}
}

// initializeCLITools sets up available CLI tools
func initializeCLITools() map[string]CLITool {
	return map[string]CLITool{
		"code2prompt": {
			Name:         "code2prompt",
			Command:      "code2prompt",
			Description:  "Convert entire codebases into structured LLM prompts with smart filtering and templating",
			Installation: "cargo install code2prompt OR brew install code2prompt",
			CheckCommand: "code2prompt --version",
			UseCases: []string{
				"Creating comprehensive code context for AI analysis",
				"Generating second opinion prompts for external AI tools",
				"Facilitating agent-to-agent communication with full context",
				"Preparing code documentation packages",
			},
			BenefitingAgents: []string{
				"second-opinion-generator",
				"orchestrator-specialist",
				"analyzer-persona",
				"mentor-persona",
			},
			Examples: []CLIExample{
				{
					Description: "Create comprehensive context for analysis",
					Command:     `code2prompt --include "**/*.{go,ts,js}" --exclude "**/node_modules/**" --max-tokens 50000`,
					UseCase:     "second-opinion",
				},
				{
					Description: "Generate focused module context",
					Command:     `code2prompt --include "internal/api/**/*.go" --output context.md`,
					UseCase:     "module-analysis",
				},
				{
					Description: "Create prompt with custom template",
					Command:     `code2prompt --template custom.hbs --include "src/**/*" --exclude "**/*.test.*"`,
					UseCase:     "custom-prompt",
				},
			},
		},
		"ast-grep": {
			Name:         "ast-grep",
			Command:      "ast-grep",
			Description:  "Semantic code search and transformation using AST patterns",
			Installation: "cargo install ast-grep OR npm install -g @ast-grep/cli",
			CheckCommand: "ast-grep --version",
			UseCases: []string{
				"Finding code patterns semantically",
				"Refactoring code with AST-aware transformations",
				"Identifying anti-patterns and code smells",
				"Creating custom linting rules",
				"Precise code modifications at symbol level",
			},
			BenefitingAgents: []string{
				"refactorer-persona",
				"analyzer-persona",
				"security-persona",
				"qa-persona",
				"backend-persona",
				"frontend-persona",
			},
			Examples: []CLIExample{
				{
					Description: "Find all error handling patterns",
					Command:     `ast-grep run -p 'if err != nil { return $$ }' --lang go`,
					UseCase:     "error-analysis",
				},
				{
					Description: "Find React hooks usage",
					Command:     `ast-grep run -p 'use$HOOK($$$)' --lang typescript`,
					UseCase:     "react-patterns",
				},
				{
					Description: "Rewrite deprecated API calls",
					Command:     `ast-grep run -p 'oldAPI($ARG)' -r 'newAPI($ARG)' --lang javascript`,
					UseCase:     "api-migration",
				},
				{
					Description: "Find security vulnerabilities",
					Command:     `ast-grep run -p 'eval($STR)' --lang javascript`,
					UseCase:     "security-scan",
				},
			},
		},
	}
}

// CheckToolAvailability checks if a CLI tool is installed
func (e *CLIToolsEnhancer) CheckToolAvailability(toolName string) (bool, string) {
	tool, exists := e.tools[toolName]
	if !exists {
		return false, "Tool not recognized"
	}

	// Safely execute the check command by parsing and validating arguments
	// Split the command to prevent shell injection
	parts := strings.Fields(tool.CheckCommand)
	if len(parts) == 0 {
		return false, "Invalid check command"
	}
	
	// Validate command is safe (no shell metacharacters)
	for _, part := range parts {
		if strings.ContainsAny(part, "|&;()<>{}$`\\\"'") {
			return false, "Check command contains unsafe characters"
		}
	}
	
	// Execute command safely without shell interpretation
	cmd := exec.Command(parts[0], parts[1:]...)
	if err := cmd.Run(); err != nil {
		return false, fmt.Sprintf("Tool not installed. Install with: %s", tool.Installation)
	}

	return true, "Tool is available"
}

// GenerateToolsSection generates documentation for enabled CLI tools
func (e *CLIToolsEnhancer) GenerateToolsSection(enabledTools []string) string {
	var sections []string

	sections = append(sections, "## 🛠️ CLI Tool Integrations\n")
	sections = append(sections, "This project has the following CLI tools available for enhanced code analysis:")
	sections = append(sections, "")

	for _, toolName := range enabledTools {
		if tool, exists := e.tools[toolName]; exists {
			sections = append(sections, e.formatTool(tool))
		}
	}

	// Add integration patterns
	sections = append(sections, e.generateIntegrationPatterns(enabledTools))

	return strings.Join(sections, "\n")
}

// formatTool formats a single CLI tool documentation
func (e *CLIToolsEnhancer) formatTool(tool CLITool) string {
	var parts []string

	parts = append(parts, fmt.Sprintf("### %s", tool.Name))
	parts = append(parts, "")
	parts = append(parts, fmt.Sprintf("**Description**: %s", tool.Description))
	parts = append(parts, "")
	parts = append(parts, fmt.Sprintf("**Installation**: `%s`", tool.Installation))
	parts = append(parts, "")

	parts = append(parts, "**Use Cases**:")
	for _, useCase := range tool.UseCases {
		parts = append(parts, fmt.Sprintf("- %s", useCase))
	}
	parts = append(parts, "")

	parts = append(parts, "**Benefiting Agents**:")
	for _, agent := range tool.BenefitingAgents {
		parts = append(parts, fmt.Sprintf("- `%s`", agent))
	}
	parts = append(parts, "")

	parts = append(parts, "**Examples**:")
	for _, example := range tool.Examples {
		parts = append(parts, fmt.Sprintf("\n%s:", example.Description))
		parts = append(parts, "```bash")
		parts = append(parts, example.Command)
		parts = append(parts, "```")
	}
	parts = append(parts, "")

	return strings.Join(parts, "\n")
}

// generateIntegrationPatterns creates usage patterns for tool combinations
func (e *CLIToolsEnhancer) generateIntegrationPatterns(enabledTools []string) string {
	var patterns []string

	patterns = append(patterns, "## 📋 Tool Integration Patterns\n")

	// Check which tools are enabled
	hasCode2prompt := false
	hasAstGrep := false

	for _, tool := range enabledTools {
		if tool == "code2prompt" {
			hasCode2prompt = true
		}
		if tool == "ast-grep" {
			hasAstGrep = true
		}
	}

	if hasCode2prompt && hasAstGrep {
		patterns = append(patterns, "### Combined Analysis Workflow")
		patterns = append(patterns, "```bash")
		patterns = append(patterns, "# 1. Use ast-grep to find specific patterns")
		patterns = append(patterns, `ast-grep run -p 'function $FUNC($$$) { $$$ }' --lang go > patterns.txt`)
		patterns = append(patterns, "")
		patterns = append(patterns, "# 2. Use code2prompt to create comprehensive context")
		patterns = append(patterns, `code2prompt --include "**/*.go" --exclude "**/vendor/**" --max-tokens 30000`)
		patterns = append(patterns, "")
		patterns = append(patterns, "# 3. Combine for targeted analysis")
		patterns = append(patterns, `echo "Found patterns:" && cat patterns.txt`)
		patterns = append(patterns, "```")
	}

	if hasCode2prompt {
		patterns = append(patterns, "\n### Second Opinion Generation")
		patterns = append(patterns, "```bash")
		patterns = append(patterns, "# Generate comprehensive prompt for external AI review")
		patterns = append(patterns, `code2prompt --include "src/**/*" --exclude "**/*.test.*" --output .claude/prompts/review.md`)
		patterns = append(patterns, "```")
	}

	if hasAstGrep {
		patterns = append(patterns, "\n### Semantic Refactoring")
		patterns = append(patterns, "```bash")
		patterns = append(patterns, "# Find and fix deprecated patterns")
		patterns = append(patterns, `ast-grep scan --rule deprecated-patterns.yml --update-all`)
		patterns = append(patterns, "```")
	}

	patterns = append(patterns, "")
	return strings.Join(patterns, "\n")
}

// CreateAgentToolsConfig generates tool configuration for specific agents
func (e *CLIToolsEnhancer) CreateAgentToolsConfig(agentType string, enabledTools []string) string {
	var config []string

	config = append(config, "## CLI Tools Configuration\n")

	// Map agent types to recommended tools
	toolRecommendations := map[string][]string{
		"second-opinion-generator": {"code2prompt", "ast-grep"},
		"refactorer-persona":       {"ast-grep"},
		"analyzer-persona":         {"ast-grep", "code2prompt"},
		"security-persona":         {"ast-grep"},
		"orchestrator-specialist":  {"code2prompt"},
	}

	if recommended, exists := toolRecommendations[agentType]; exists {
		config = append(config, fmt.Sprintf("### Recommended Tools for %s", agentType))
		config = append(config, "")

		for _, toolName := range recommended {
			// Check if tool is enabled
			isEnabled := false
			for _, enabled := range enabledTools {
				if enabled == toolName {
					isEnabled = true
					break
				}
			}

			if isEnabled {
				if tool, exists := e.tools[toolName]; exists {
					config = append(config, fmt.Sprintf("#### %s ✅", tool.Name))
					config = append(config, fmt.Sprintf("- Status: Enabled"))
					config = append(config, fmt.Sprintf("- Primary Use: %s", tool.UseCases[0]))

					// Add agent-specific example
					for _, example := range tool.Examples {
						if strings.Contains(agentType, "second-opinion") && example.UseCase == "second-opinion" {
							config = append(config, fmt.Sprintf("- Example: `%s`", example.Command))
							break
						}
						if strings.Contains(agentType, "security") && example.UseCase == "security-scan" {
							config = append(config, fmt.Sprintf("- Example: `%s`", example.Command))
							break
						}
					}
					config = append(config, "")
				}
			} else {
				config = append(config, fmt.Sprintf("#### %s ❌", toolName))
				config = append(config, fmt.Sprintf("- Status: Not enabled"))
				config = append(config, fmt.Sprintf("- Enable with: --tools=%s", toolName))
				config = append(config, "")
			}
		}
	}

	return strings.Join(config, "\n")
}

// InstallationGuide generates installation instructions for missing tools
func (e *CLIToolsEnhancer) InstallationGuide(requiredTools []string) string {
	var guide []string

	guide = append(guide, "## 📦 Tool Installation Guide\n")

	missingTools := []string{}
	for _, toolName := range requiredTools {
		if available, _ := e.CheckToolAvailability(toolName); !available {
			missingTools = append(missingTools, toolName)
		}
	}

	if len(missingTools) == 0 {
		guide = append(guide, "✅ All required tools are installed!")
		return strings.Join(guide, "\n")
	}

	guide = append(guide, "The following tools need to be installed:")
	guide = append(guide, "")

	for _, toolName := range missingTools {
		if tool, exists := e.tools[toolName]; exists {
			guide = append(guide, fmt.Sprintf("### %s", tool.Name))
			guide = append(guide, "```bash")
			guide = append(guide, fmt.Sprintf("# Option 1: %s", strings.Split(tool.Installation, " OR ")[0]))
			if strings.Contains(tool.Installation, " OR ") {
				guide = append(guide, fmt.Sprintf("# Option 2: %s", strings.Split(tool.Installation, " OR ")[1]))
			}
			guide = append(guide, "```")
			guide = append(guide, "")
		}
	}

	return strings.Join(guide, "\n")
}

// EnhanceProjectCLAUDEWithTools adds CLI tools section to CLAUDE.md
func (e *CLIToolsEnhancer) EnhanceProjectCLAUDEWithTools(existingContent string, enabledTools []string) string {
	toolsSection := e.GenerateToolsSection(enabledTools)

	// If existing content has CLI tools section, replace it
	if strings.Contains(existingContent, "## 🛠️ CLI Tool Integrations") {
		// Find and replace existing CLI tools section
		startIdx := strings.Index(existingContent, "## 🛠️ CLI Tool Integrations")
		endIdx := strings.Index(existingContent[startIdx:], "\n## ")
		if endIdx == -1 {
			// CLI tools section is at the end
			return existingContent[:startIdx] + toolsSection
		}
		endIdx += startIdx
		return existingContent[:startIdx] + toolsSection + existingContent[endIdx:]
	}

	// Otherwise, append CLI tools section
	return existingContent + "\n\n" + toolsSection
}
