// Package claude provides tab completion mechanisms for Claude Code integration.
// This enables dynamic command discovery and intelligent completion suggestions.
package claude

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jonwraymond/claude-code-super-crew/pkg/logger"
)

// CompletionProvider handles tab completion for /crew: commands
type CompletionProvider struct {
	registry *SlashCommandRegistry
	logger   logger.Logger
}

// NewCompletionProvider creates a new completion provider
func NewCompletionProvider(commandsPath string) (*CompletionProvider, error) {
	registry := NewSlashCommandRegistry(commandsPath)
	if err := registry.LoadCommands(); err != nil {
		return nil, fmt.Errorf("failed to load commands: %w", err)
	}

	return &CompletionProvider{
		registry: registry,
		logger:   logger.GetLogger(),
	}, nil
}

// GetCompletions returns completion suggestions for a given input
func (cp *CompletionProvider) GetCompletions(input string) *CompletionResult {
	result := &CompletionResult{
		Input: input,
	}

	// Handle different completion scenarios
	switch {
	case input == "":
		// No input - show all /crew: commands
		result.Suggestions = cp.getAllCommandCompletions()
		result.Type = "commands"

	case input == "/":
		// Starting slash - show crew prefix
		result.Suggestions = []CompletionSuggestion{
			{Text: "/crew:", Description: "SuperCrew framework commands"},
		}
		result.Type = "prefix"

	case input == "/crew":
		// Partial crew - complete to /crew:
		result.Suggestions = []CompletionSuggestion{
			{Text: "/crew:", Description: "SuperCrew framework commands"},
		}
		result.Type = "prefix"

	case input == "/crew:":
		// Complete prefix - show all commands
		result.Suggestions = cp.getAllCommandCompletions()
		result.Type = "commands"

	case strings.HasPrefix(input, "/crew:"):
		// Partial command - find matches
		partial := strings.TrimPrefix(input, "/crew:")
		result.Suggestions = cp.getPartialCommandCompletions(partial)
		result.Type = "commands"

	default:
		// No match
		result.Suggestions = []CompletionSuggestion{}
		result.Type = "none"
	}

	result.Count = len(result.Suggestions)
	return result
}

// getAllCommandCompletions returns all available /crew: commands
func (cp *CompletionProvider) getAllCommandCompletions() []CompletionSuggestion {
	var suggestions []CompletionSuggestion
	commands := cp.registry.ListCommands()

	for _, cmd := range commands {
		// Use command metadata for better categorization
		category := cmd.Category
		if category == "" {
			category = cp.categorizeCommand(cmd.Name)
		}

		description := cmd.Description
		if cmd.Purpose != "" {
			description = cmd.Purpose
		}

		suggestion := CompletionSuggestion{
			Text:        fmt.Sprintf("/crew:%s", cmd.Name),
			Description: description,
			Usage:       cmd.Usage,
			Category:    category,
		}

		// Add argument hints if available
		if len(cmd.Arguments) > 0 {
			var argHints []string
			for _, arg := range cmd.Arguments {
				if arg.Required {
					argHints = append(argHints, fmt.Sprintf("<%s>", arg.Name))
				} else {
					argHints = append(argHints, fmt.Sprintf("[%s]", arg.Name))
				}
			}
			suggestion.ArgumentHints = strings.Join(argHints, " ")
		}

		suggestions = append(suggestions, suggestion)
	}

	// Sort by category, then by name
	sort.Slice(suggestions, func(i, j int) bool {
		if suggestions[i].Category != suggestions[j].Category {
			return suggestions[i].Category < suggestions[j].Category
		}
		return suggestions[i].Text < suggestions[j].Text
	})

	return suggestions
}

// getPartialCommandCompletions returns completions for partial command input
func (cp *CompletionProvider) getPartialCommandCompletions(partial string) []CompletionSuggestion {
	var suggestions []CompletionSuggestion
	commands := cp.registry.ListCommands()

	for _, cmd := range commands {
		if strings.HasPrefix(cmd.Name, partial) {
			suggestion := CompletionSuggestion{
				Text:        fmt.Sprintf("/crew:%s", cmd.Name),
				Description: cmd.Description,
				Usage:       cmd.Usage,
				Category:    cp.categorizeCommand(cmd.Name),
			}

			// Add argument hints
			if len(cmd.Arguments) > 0 {
				var argHints []string
				for _, arg := range cmd.Arguments {
					if arg.Required {
						argHints = append(argHints, fmt.Sprintf("<%s>", arg.Name))
					} else {
						argHints = append(argHints, fmt.Sprintf("[%s]", arg.Name))
					}
				}
				suggestion.ArgumentHints = strings.Join(argHints, " ")
			}

			suggestions = append(suggestions, suggestion)
		}
	}

	// Sort by relevance (exact prefix match first, then alphabetical)
	sort.Slice(suggestions, func(i, j int) bool {
		iName := strings.TrimPrefix(suggestions[i].Text, "/crew:")
		jName := strings.TrimPrefix(suggestions[j].Text, "/crew:")

		// Exact prefix match gets priority
		if strings.HasPrefix(iName, partial) && !strings.HasPrefix(jName, partial) {
			return true
		}
		if !strings.HasPrefix(iName, partial) && strings.HasPrefix(jName, partial) {
			return false
		}

		return iName < jName
	})

	return suggestions
}

// categorizeCommand assigns a category to commands for better organization following SuperCrew patterns
func (cp *CompletionProvider) categorizeCommand(name string) string {
	// Categorize commands based on SuperCrew documentation structure
	categories := map[string]string{
		// Development commands
		"build":     "development",
		"implement": "development",
		"design":    "development",
		"workflow":  "development",

		// Analysis commands
		"analyze":      "analysis",
		"troubleshoot": "analysis",
		"explain":      "analysis",
		"index":        "analysis",

		// Quality commands
		"improve": "quality",
		"cleanup": "quality",
		"test":    "quality",

		// Utilities commands
		"document": "utilities",
		"git":      "utilities",
		"load":     "utilities",
		"estimate": "utilities",
		"task":     "utilities",
		"spawn":    "utilities",
	}

	if category, exists := categories[name]; exists {
		return category
	}
	return "general"
}

// CompletionResult represents the result of a completion request
type CompletionResult struct {
	Input       string                 `json:"input"`
	Type        string                 `json:"type"`
	Count       int                    `json:"count"`
	Suggestions []CompletionSuggestion `json:"suggestions"`
}

// CompletionSuggestion represents a single completion suggestion
type CompletionSuggestion struct {
	Text          string `json:"text"`
	Description   string `json:"description"`
	Usage         string `json:"usage,omitempty"`
	ArgumentHints string `json:"argument_hints,omitempty"`
	Category      string `json:"category"`
}

// GenerateClaudeDesktopConfig generates Claude Desktop configuration for MCP integration
func (cp *CompletionProvider) GenerateClaudeDesktopConfig(configPath string) error {
	// Create Claude Desktop MCP server configuration
	config := map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"supercrew": map[string]interface{}{
				"command": "crew",
				"args":    []string{"claude", "--mcp-server"},
				"env": map[string]string{
					"SUPERCREW_COMMANDS_DIR": filepath.Join(filepath.Dir(configPath), "..", "SuperCrew", "Commands"),
				},
			},
		},
	}

	// Write configuration file
	// Note: JSON marshaling implementation deferred pending MCP integration requirements
	_ = config // Placeholder for future JSON marshaling
	cp.logger.Infof("Claude Desktop MCP configuration would be written to: %s", configPath)
	return nil
}

// StartMCPServer starts an MCP server for Claude Desktop integration
func (cp *CompletionProvider) StartMCPServer() error {
	cp.logger.Info("Starting MCP server for Claude Desktop integration")

	// MCP protocol implementation framework - awaiting Claude Desktop integration spec
	// Framework provides interface for future MCP server implementation

	return fmt.Errorf("MCP server implementation pending Claude Desktop integration spec")
}

// ValidateCommands validates that all commands are properly configured
func (cp *CompletionProvider) ValidateCommands() *ValidationReport {
	report := &ValidationReport{
		TotalCommands: len(cp.registry.commands),
		ValidCommands: 0,
		Issues:        []ValidationIssue{},
	}

	for name, cmd := range cp.registry.commands {
		// Check for required fields
		if cmd.Description == "" {
			report.Issues = append(report.Issues, ValidationIssue{
				Command:  name,
				Severity: "warning",
				Message:  "Missing description",
			})
		}

		// Check for valid usage pattern
		if cmd.Usage == "" {
			report.Issues = append(report.Issues, ValidationIssue{
				Command:  name,
				Severity: "info",
				Message:  "Missing usage pattern",
			})
		}

		// Check for allowed tools
		if len(cmd.AllowedTools) == 0 {
			report.Issues = append(report.Issues, ValidationIssue{
				Command:  name,
				Severity: "info",
				Message:  "No allowed tools specified",
			})
		}

		// Count commands with no critical issues as valid
		hasErrors := false
		for _, issue := range report.Issues {
			if issue.Command == name && issue.Severity == "error" {
				hasErrors = true
				break
			}
		}
		if !hasErrors {
			report.ValidCommands++
		}
	}

	return report
}

// ValidationReport contains command validation results
type ValidationReport struct {
	TotalCommands int               `json:"total_commands"`
	ValidCommands int               `json:"valid_commands"`
	Issues        []ValidationIssue `json:"issues"`
}

// ValidationIssue represents a command validation issue
type ValidationIssue struct {
	Command  string `json:"command"`
	Severity string `json:"severity"` // error, warning, info
	Message  string `json:"message"`
}

// ExportCompletions exports completion data for external tools
func (cp *CompletionProvider) ExportCompletions(format, outputPath string) error {
	switch format {
	case "json":
		return cp.exportJSON(outputPath)
	case "yaml":
		return cp.exportYAML(outputPath)
	case "shell":
		return cp.exportShellScript(outputPath)
	default:
		return fmt.Errorf("unsupported export format: %s", format)
	}
}

func (cp *CompletionProvider) exportJSON(outputPath string) error {
	completions := cp.getAllCommandCompletions()
	// JSON marshaling would happen here
	cp.logger.Infof("Would export %d completions to JSON: %s", len(completions), outputPath)
	return nil
}

func (cp *CompletionProvider) exportYAML(outputPath string) error {
	completions := cp.getAllCommandCompletions()
	// YAML marshaling would happen here
	cp.logger.Infof("Would export %d completions to YAML: %s", len(completions), outputPath)
	return nil
}

func (cp *CompletionProvider) exportShellScript(outputPath string) error {
	// Generate shell completion script
	script, err := cp.registry.GenerateCompletionScript("bash")
	if err != nil {
		return err
	}

	if err := os.WriteFile(outputPath, []byte(script), 0644); err != nil {
		return fmt.Errorf("failed to write shell script: %w", err)
	}

	cp.logger.Successf("Exported shell completion script to: %s", outputPath)
	return nil
}
