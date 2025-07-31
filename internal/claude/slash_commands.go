// Package claude provides Claude Code integration for SuperCrew slash commands.
// This enables /crew: prefixed commands with tab completion and automatic command discovery.
package claude

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jonwraymond/claude-code-super-crew/pkg/logger"
)

// SlashCommand represents a SuperCrew slash command for Claude Code integration
type SlashCommand struct {
	Name          string            `json:"name"`
	Description   string            `json:"description"`
	Usage         string            `json:"usage,omitempty"`
	Arguments     []CommandArgument `json:"arguments,omitempty"`
	AllowedTools  []string          `json:"allowed_tools,omitempty"`
	Category      string            `json:"category,omitempty"`
	Purpose       string            `json:"purpose,omitempty"`
	AutoActivates string            `json:"auto_activates,omitempty"`
	BestFor       string            `json:"best_for,omitempty"`
	Complexity    string            `json:"complexity,omitempty"`
	WaveEnabled   bool              `json:"wave_enabled,omitempty"`
}

// CommandArgument represents a command argument with its properties
type CommandArgument struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Required    bool     `json:"required"`
	Type        string   `json:"type"` // string, flag, choice
	Choices     []string `json:"choices,omitempty"`
}

// SlashCommandRegistry manages all available slash commands for tab completion
type SlashCommandRegistry struct {
	commands     map[string]*SlashCommand
	commandsPath string
	logger       logger.Logger
}

// NewSlashCommandRegistry creates a new slash command registry
func NewSlashCommandRegistry(commandsPath string) *SlashCommandRegistry {
	return &SlashCommandRegistry{
		commands:     make(map[string]*SlashCommand),
		commandsPath: commandsPath,
		logger:       logger.GetLogger(),
	}
}

// LoadCommands discovers and loads all available slash commands from the SuperCrew/Commands directory
func (r *SlashCommandRegistry) LoadCommands() error {
	if _, err := os.Stat(r.commandsPath); os.IsNotExist(err) {
		return fmt.Errorf("commands directory not found: %s", r.commandsPath)
	}

	entries, err := os.ReadDir(r.commandsPath)
	if err != nil {
		return fmt.Errorf("failed to read commands directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		commandName := strings.TrimSuffix(entry.Name(), ".md")
		commandPath := filepath.Join(r.commandsPath, entry.Name())

		if err := r.loadCommand(commandName, commandPath); err != nil {
			r.logger.Warnf("Failed to load command %s: %v", commandName, err)
			continue
		}
	}

	r.logger.Infof("Loaded %d slash commands", len(r.commands))
	return nil
}

// loadCommand parses a command markdown file and extracts command metadata
func (r *SlashCommandRegistry) loadCommand(name, filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read command file: %w", err)
	}

	command := &SlashCommand{
		Name: name,
	}

	lines := strings.Split(string(content), "\n")
	var inFrontMatter bool
	var frontMatterLines []string

	// Parse frontmatter for metadata
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "---" {
			if !inFrontMatter {
				inFrontMatter = true
				continue
			} else {
				break
			}
		}

		if inFrontMatter {
			frontMatterLines = append(frontMatterLines, line)
			continue
		}

		// Extract description from first header
		if strings.HasPrefix(line, "# /crew:") && command.Description == "" {
			parts := strings.SplitN(line, " - ", 2)
			if len(parts) == 2 {
				command.Description = parts[1]
			}
		}

		// Extract usage from usage section
		if strings.HasPrefix(line, "```") && strings.Contains(strings.ToLower(line), "crew:") {
			continue
		}
		if strings.HasPrefix(line, "/crew:") && command.Usage == "" {
			command.Usage = line
		}
	}

	// Parse frontmatter for SuperCrew metadata
	for _, fmLine := range frontMatterLines {
		if strings.HasPrefix(fmLine, "allowed-tools:") {
			toolsStr := strings.TrimPrefix(fmLine, "allowed-tools:")
			toolsStr = strings.Trim(toolsStr, " []")
			if toolsStr != "" {
				tools := strings.Split(toolsStr, ",")
				for i, tool := range tools {
					tools[i] = strings.Trim(tool, " \"")
				}
				command.AllowedTools = tools
			}
		}
		if strings.HasPrefix(fmLine, "description:") {
			desc := strings.TrimPrefix(fmLine, "description:")
			desc = strings.Trim(desc, " \"")
			if desc != "" && command.Description == "" {
				command.Description = desc
			}
		}
		if strings.HasPrefix(fmLine, "category:") {
			category := strings.TrimPrefix(fmLine, "category:")
			command.Category = strings.Trim(category, " \"")
		}
		if strings.HasPrefix(fmLine, "purpose:") {
			purpose := strings.TrimPrefix(fmLine, "purpose:")
			command.Purpose = strings.Trim(purpose, " \"")
		}
		if strings.HasPrefix(fmLine, "auto-activates:") {
			autoActivates := strings.TrimPrefix(fmLine, "auto-activates:")
			command.AutoActivates = strings.Trim(autoActivates, " \"")
		}
		if strings.HasPrefix(fmLine, "best-for:") {
			bestFor := strings.TrimPrefix(fmLine, "best-for:")
			command.BestFor = strings.Trim(bestFor, " \"")
		}
		if strings.HasPrefix(fmLine, "complexity:") {
			complexity := strings.TrimPrefix(fmLine, "complexity:")
			command.Complexity = strings.Trim(complexity, " \"")
		}
		if strings.HasPrefix(fmLine, "wave-enabled:") {
			waveStr := strings.TrimPrefix(fmLine, "wave-enabled:")
			waveStr = strings.Trim(waveStr, " \"")
			command.WaveEnabled = waveStr == "true"
		}
	}

	// Parse arguments from usage pattern
	if command.Usage != "" {
		command.Arguments = r.parseArguments(command.Usage)
	}

	// Set default description if none found
	if command.Description == "" {
		command.Description = fmt.Sprintf("SuperCrew %s command", name)
	}

	r.commands[name] = command
	return nil
}

// parseArguments extracts argument information from usage string following SuperCrew patterns
func (r *SlashCommandRegistry) parseArguments(usage string) []CommandArgument {
	var args []CommandArgument

	// Enhanced parsing for SuperCrew command patterns
	parts := strings.Fields(usage)
	for i, part := range parts {
		if i == 0 { // Skip command name
			continue
		}

		arg := CommandArgument{
			Type: "string",
		}

		if strings.HasPrefix(part, "[") && strings.HasSuffix(part, "]") {
			// Optional argument like [target] or [--flag]
			cleaned := strings.Trim(part, "[]")
			arg.Name = cleaned
			arg.Required = false

			if strings.HasPrefix(cleaned, "--") {
				arg.Type = "flag"
				arg.Description = fmt.Sprintf("Optional flag: %s", cleaned)
			} else {
				arg.Description = fmt.Sprintf("Optional %s parameter", cleaned)
			}
		} else if strings.HasPrefix(part, "--") {
			// Flag argument like --focus, --type
			arg.Name = part
			arg.Type = "flag"
			arg.Required = false

			// Add common flag choices based on SuperCrew documentation
			switch part {
			case "--focus":
				arg.Choices = []string{"quality", "security", "performance", "architecture"}
				arg.Description = "Focus area for analysis or improvement"
			case "--type":
				arg.Choices = []string{"component", "api", "service", "feature", "architecture"}
				arg.Description = "Type of implementation or design"
			case "--scope":
				arg.Choices = []string{"file", "module", "project", "system"}
				arg.Description = "Scope of operation"
			default:
				arg.Description = fmt.Sprintf("Flag: %s", part)
			}
		} else {
			// Required positional argument
			arg.Name = part
			arg.Required = true
			arg.Description = fmt.Sprintf("Required %s parameter", part)
		}

		args = append(args, arg)
	}

	return args
}

// GetCommand returns a command by name
func (r *SlashCommandRegistry) GetCommand(name string) (*SlashCommand, bool) {
	cmd, exists := r.commands[name]
	return cmd, exists
}

// ListCommands returns all available commands sorted by name
func (r *SlashCommandRegistry) ListCommands() []*SlashCommand {
	var commands []*SlashCommand
	var names []string

	for name := range r.commands {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		commands = append(commands, r.commands[name])
	}

	return commands
}

// GetCompletions returns command completions for tab completion
func (r *SlashCommandRegistry) GetCompletions(prefix string) []string {
	var completions []string

	for name := range r.commands {
		if strings.HasPrefix(name, prefix) {
			completions = append(completions, fmt.Sprintf("/crew:%s", name))
		}
	}

	sort.Strings(completions)
	return completions
}

// ExecuteCommand handles execution of a slash command
func (r *SlashCommandRegistry) ExecuteCommand(commandLine string) error {
	parts := strings.Fields(commandLine)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	// Extract command name from /crew:commandname format
	fullCommand := parts[0]
	if !strings.HasPrefix(fullCommand, "/crew:") {
		return fmt.Errorf("invalid command format: %s", fullCommand)
	}

	commandName := strings.TrimPrefix(fullCommand, "/crew:")
	command, exists := r.GetCommand(commandName)
	if !exists {
		return fmt.Errorf("unknown command: %s", commandName)
	}

	r.logger.Infof("Executing slash command: %s", commandName)

	// For now, delegate to the existing CLI system
	// This can be enhanced to parse arguments and call specific handlers
	args := parts[1:] // Remove command name
	return r.executeCommandHandler(commandName, args, command)
}

// executeCommandHandler executes the actual command logic
func (r *SlashCommandRegistry) executeCommandHandler(name string, args []string, command *SlashCommand) error {
	// Map slash commands to CLI operations or direct handlers
	switch name {
	case "analyze":
		return r.handleAnalyzeCommand(args)
	case "build":
		return r.handleBuildCommand(args)
	case "install":
		return r.handleInstallCommand(args)
	case "cleanup":
		return r.handleCleanupCommand(args)
	case "document":
		return r.handleDocumentCommand(args)
	case "improve":
		return r.handleImproveCommand(args)
	case "test":
		return r.handleTestCommand(args)
	case "load":
		return r.handleLoadCommand(args)
	case "implement":
		return r.handleImplementCommand(args)
	default:
		return r.handleGenericCommand(name, args, command)
	}
}

// Command handlers - these can be expanded with full implementations

func (r *SlashCommandRegistry) handleAnalyzeCommand(args []string) error {
	r.logger.Info("Executing analyze command via slash interface")
	
	// Generate the Task tool invocation for Claude Code
	prompt := fmt.Sprintf("You are the orchestrator-specialist for Claude Code Super Crew. Execute /crew:analyze by chaining agents: 1) analyzer-persona for comprehensive analysis, 2) go-backend-specialist for Go-specific insights, 3) scribe-persona for clear documentation. Follow the mandatory sub-agent chaining framework. Args: %s", strings.Join(args, " "))
	
	fmt.Printf("\nExecuting /crew:analyze...\n")
	fmt.Printf("Task(description='Analyze project code and architecture', prompt='%s', subagent_type='orchestrator-specialist')\n", prompt)
	
	return nil
}

func (r *SlashCommandRegistry) handleBuildCommand(args []string) error {
	r.logger.Info("Executing build command via slash interface")
	
	// Generate the Task tool invocation for Claude Code
	prompt := fmt.Sprintf("You are the orchestrator-specialist for Claude Code Super Crew. Execute /crew:build by chaining agents: 1) analyzer-persona for build requirements analysis, 2) go-backend-specialist for Go build implementation, 3) qa-persona for build validation. Use the project's Makefile and Go module system. Args: %s", strings.Join(args, " "))
	
	fmt.Printf("\nExecuting /crew:build...\n")
	fmt.Printf("Task(description='Build and compile Go project', prompt='%s', subagent_type='orchestrator-specialist')\n", prompt)
	
	return nil
}

func (r *SlashCommandRegistry) handleInstallCommand(args []string) error {
	r.logger.Info("Executing install command via slash interface")
	// Note: Install integration requires refactoring to avoid circular import with cli package
	return fmt.Errorf("install command implementation pending - circular import resolution needed")
}

func (r *SlashCommandRegistry) handleCleanupCommand(args []string) error {
	r.logger.Info("Executing cleanup command via slash interface")
	// Note: Cleanup functionality requires orchestrator-agent coordination
	return fmt.Errorf("cleanup command implementation pending - use orchestrator-agent routing")
}

func (r *SlashCommandRegistry) handleDocumentCommand(args []string) error {
	r.logger.Info("Executing document command via slash interface")
	// Note: Documentation generation awaiting scribe-persona integration
	return fmt.Errorf("document command implementation pending - use Task tool with scribe-persona")
}

func (r *SlashCommandRegistry) handleImproveCommand(args []string) error {
	r.logger.Info("Executing improve command via slash interface")
	// Note: Code improvement integration awaiting refactorer-persona coordination
	return fmt.Errorf("improve command implementation pending - use Task tool with refactorer-persona")
}

func (r *SlashCommandRegistry) handleTestCommand(args []string) error {
	r.logger.Info("Executing test command via slash interface")
	// Note: Testing framework integration awaiting qa-persona coordination
	return fmt.Errorf("test command implementation pending - use Task tool with qa-persona")
}

func (r *SlashCommandRegistry) handleLoadCommand(args []string) error {
	r.logger.Info("🎭 [Orchestrator]: Routing /crew:load to global orchestrator agent")

	// Get current working directory
	workingDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Parse arguments to check if a different target directory was specified
	targetDir := workingDir
	for i, arg := range args {
		if i == 0 && !strings.HasPrefix(arg, "--") {
			// First non-flag argument is the target directory
			targetDir = arg
			break
		}
	}

	// Generate orchestrator metaprompt for project analysis and setup
	metaprompt := r.generateLoadCommandMetaprompt(targetDir)

	// Route to the global orchestrator via Task tool
	r.logger.Info("🎯 Delegating to global orchestrator agent...")
	
	// The actual routing happens through Claude's Task tool
	// This message instructs Claude to use the orchestrator agent
	fmt.Printf("\n🎭 ORCHESTRATOR METAPROMPT:\n")
	fmt.Printf("Task(subagent_type='orchestrator-agent', prompt='%s')\n", metaprompt)
	
	r.logger.Success("🎭 [Orchestrator]: Global orchestrator activated for project analysis")

	return nil
}

func (r *SlashCommandRegistry) handleImplementCommand(args []string) error {
	r.logger.Info("Executing implement command via slash interface")
	
	// Generate the Task tool invocation for Claude Code
	prompt := fmt.Sprintf("You are the orchestrator-specialist for Claude Code Super Crew. Execute /crew:implement by chaining agents based on the feature type: 1) analyzer-persona for requirements analysis, 2) appropriate specialist (go-backend-specialist for Go code, frontend-persona for UI), 3) qa-persona for testing. Args: %s", strings.Join(args, " "))
	
	fmt.Printf("\nExecuting /crew:implement...\n")
	fmt.Printf("Task(description='Implement feature or component', prompt='%s', subagent_type='orchestrator-specialist')\n", prompt)
	
	return nil
}

// generateLoadCommandMetaprompt creates the metaprompt for /crew:load orchestrator delegation
func (r *SlashCommandRegistry) generateLoadCommandMetaprompt(targetDir string) string {
	return fmt.Sprintf(`You are the global orchestrator agent from ~/.claude/agents/orchestrator.agent.md. 

TASK: Execute the /crew:load command for project analysis and setup.

PROJECT DIRECTORY: %s

INSTRUCTIONS:
1. **Project Analysis Phase**:
   - Use Glob to discover all source files and project structure
   - Count files by extension to determine primary languages
   - Use Read to examine key configuration files (go.mod, package.json, etc.)
   - Use Grep to identify architectural patterns and frameworks
   - Analyze project complexity and development patterns

2. **Local Orchestrator Creation**:
   - Create/update .claude/agents/orchestrator-specialist.md based on the global template
   - Customize it with project-specific routing rules and context
   - Include project analysis findings and recommended workflows

3. **Specialist Generation** (only if patterns justify it):
   - Generate project-specific specialists for REPEATED complex patterns
   - Examples: go-specialist (if 20+ Go files), api-specialist (if 10+ API endpoints)
   - Each specialist should be in .claude/agents/[name]-specialist.md

4. **Shadow Command Creation** (if beneficial):
   - Create enhanced versions of global commands in .claude/commands/shadows/
   - Only for commands that benefit from project-specific customization

5. **Project Configuration**:
   - Update .claude/project-config.json with analysis results
   - Include detected languages, frameworks, and architectural patterns

6. **Completion Report**:
   - Summarize what was created/updated
   - List available agents and their specialties
   - Recommend next steps for the user

CONTEXT: This is the initial project setup. Be intelligent about what specialists and shadow commands are actually needed based on real patterns in the codebase, not just assumptions.

Execute this comprehensive project analysis and setup now.`, targetDir)
}

func (r *SlashCommandRegistry) handleGenericCommand(name string, args []string, command *SlashCommand) error {
	// Route to orchestrator for agent delegation
	r.logger.Info(fmt.Sprintf("🎯 [Orchestrator]: Analyzing command /%s for agent routing", name))

	// Check if this is an agent-specific command
	agentSuggestion := r.suggestAgentForCommand(name, args)
	if agentSuggestion != "" {
		r.logger.Info(agentSuggestion)
		return nil
	}

	// Generic handler for commands without specific implementations
	r.logger.Infof("Command %s not yet implemented. Consider using an appropriate agent.", name)
	return nil
}

// suggestAgentForCommand analyzes command complexity and routes appropriately
func (r *SlashCommandRegistry) suggestAgentForCommand(name string, args []string) string {
	// Check for explicit orchestration commands
	switch name {
	case "orchestrate", "chain", "workflow", "multimodal":
		return r.handleOrchestratorCommand(name, args)
	case "agent-help", "help":
		return r.showAvailableAgents()
	}

	// Analyze complexity
	complexity := r.analyzeCommandComplexity(name, args)

	if complexity.RequiresOrchestration {
		return fmt.Sprintf(`🎯 [Orchestrator]: This task appears complex and may benefit from orchestration.
Complexity indicators: %s
Domains detected: %s
Recommended: /crew:orchestrate "%s %s"
Or continue with single agent routing if you prefer.`,
			strings.Join(complexity.Indicators, ", "),
			strings.Join(complexity.Domains, ", "),
			name, strings.Join(args, " "))
	}

	// Route simple commands to appropriate personas
	return r.routeSimpleCommand(name, args)
}

// analyzeCommandComplexity determines if orchestration is beneficial
func (r *SlashCommandRegistry) analyzeCommandComplexity(name string, args []string) CommandComplexity {
	complexity := CommandComplexity{
		Score:      0.0,
		Indicators: []string{},
		Domains:    []string{},
	}

	// Check for multi-domain indicators
	fullCommand := name + " " + strings.Join(args, " ")

	// Multi-step indicators
	if strings.Contains(fullCommand, " and ") || strings.Contains(fullCommand, " then ") {
		complexity.Score += 0.3
		complexity.Indicators = append(complexity.Indicators, "multi-step")
	}

	// Cross-domain keywords
	crossDomainKeywords := []string{"full", "complete", "entire", "comprehensive", "integrate", "secure", "test", "deploy"}
	for _, keyword := range crossDomainKeywords {
		if strings.Contains(strings.ToLower(fullCommand), keyword) {
			complexity.Score += 0.2
			complexity.Indicators = append(complexity.Indicators, keyword)
		}
	}

	// Domain detection
	if strings.Contains(fullCommand, "api") || strings.Contains(fullCommand, "endpoint") {
		complexity.Domains = append(complexity.Domains, "api")
	}
	if strings.Contains(fullCommand, "test") || strings.Contains(fullCommand, "validate") {
		complexity.Domains = append(complexity.Domains, "testing")
	}
	if strings.Contains(fullCommand, "secure") || strings.Contains(fullCommand, "auth") {
		complexity.Domains = append(complexity.Domains, "security")
	}
	if strings.Contains(fullCommand, "deploy") || strings.Contains(fullCommand, "build") {
		complexity.Domains = append(complexity.Domains, "deployment")
	}
	if strings.Contains(fullCommand, "design") || strings.Contains(fullCommand, "architect") {
		complexity.Domains = append(complexity.Domains, "architecture")
	}

	// Multiple domains = higher complexity
	if len(complexity.Domains) >= 2 {
		complexity.Score += 0.3
		complexity.Indicators = append(complexity.Indicators, "cross-domain")
	}

	complexity.RequiresOrchestration = complexity.Score >= 0.7 || len(complexity.Domains) >= 3

	return complexity
}

// routeSimpleCommand handles single-domain routing
func (r *SlashCommandRegistry) routeSimpleCommand(name string, args []string) string {
	// Check if project has specialists (only check, don't hardcode)
	agentsDir := ".claude/agents"
	specialists := r.listProjectSpecialists(agentsDir)

	// Route based on command to personas (no hardcoded specialists)
	switch name {
	case "go", "backend", "api", "endpoint", "rest":
		// Check if a relevant specialist exists
		for _, specialist := range specialists {
			if strings.Contains(specialist, "backend") || strings.Contains(specialist, name) {
				return fmt.Sprintf("🎯 [Orchestrator]: Found local %s. Use: Task(subagent_type='%s', prompt='%s')",
					specialist, specialist, strings.Join(append([]string{name}, args...), " "))
			}
		}
		// Default to persona
		return "🎯 [Orchestrator]: Using backend-persona. Use: Task(subagent_type='backend-persona', prompt='Handle " + name + " task')"

	case "test", "validate", "qa":
		return "🎯 [Orchestrator]: Using qa-persona. Use: Task(subagent_type='qa-persona', prompt='Test " + strings.Join(args, " ") + "')"

	case "analyze", "investigate", "debug":
		return "🎯 [Orchestrator]: Using analyzer-persona. Use: Task(subagent_type='analyzer-persona', prompt='Analyze " + strings.Join(args, " ") + "')"

	case "document", "docs", "readme":
		return "🎯 [Orchestrator]: Using scribe-persona. Use: Task(subagent_type='scribe-persona', prompt='Document " + strings.Join(args, " ") + "')"

	case "design", "architecture":
		return "🎯 [Orchestrator]: Using architect-persona. Use: Task(subagent_type='architect-persona', prompt='Design " + strings.Join(args, " ") + "')"

	case "optimize", "profile", "performance":
		return "🎯 [Orchestrator]: Using performance-persona. Use: Task(subagent_type='performance-persona', prompt='Optimize " + strings.Join(args, " ") + "')"

	case "secure", "audit", "vulnerability":
		return "🎯 [Orchestrator]: Using security-persona. Use: Task(subagent_type='security-persona', prompt='Security " + name + " for " + strings.Join(args, " ") + "')"

	default:
		return fmt.Sprintf(`🎯 [Orchestrator]: Analyzing request...
Task: %s %s
Available: %d specialists, 11 personas
Recommendation: Try a persona directly or /crew:orchestrate for complex coordination.
Use /crew:help to see available agents.`, name, strings.Join(args, " "), len(specialists))
	}
}

// handleOrchestratorCommand processes orchestration-specific commands
func (r *SlashCommandRegistry) handleOrchestratorCommand(name string, args []string) string {
	switch name {
	case "orchestrate":
		return "🎯 [Orchestrator]: Ready to coordinate complex workflow. Use: Task(subagent_type='orchestrator-specialist', prompt='Orchestrate: " + strings.Join(args, " ") + "')"

	case "chain":
		return "🎯 [Orchestrator]: Setting up agent chain. Use: Task(subagent_type='orchestrator-specialist', prompt='Chain workflow: " + strings.Join(args, " ") + "')"

	case "workflow":
		return "🎯 [Orchestrator]: Designing multi-agent workflow. Use: Task(subagent_type='orchestrator-specialist', prompt='Design workflow: " + strings.Join(args, " ") + "')"

	case "multimodal":
		return "🎯 [Orchestrator]: Coordinating cross-domain task. Use: Task(subagent_type='orchestrator-specialist', prompt='Multimodal task: " + strings.Join(args, " ") + "')"

	default:
		return "🎯 [Orchestrator]: Ready to coordinate. Specify your complex task."
	}
}

// Helper types
type CommandComplexity struct {
	Score                 float64
	RequiresOrchestration bool
	Indicators            []string
	Domains               []string
}

// listProjectSpecialists dynamically lists specialists in project
func (r *SlashCommandRegistry) listProjectSpecialists(dir string) []string {
	specialists := []string{}

	// Check if directory exists
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return specialists
	}

	// List .md files
	entries, err := os.ReadDir(dir)
	if err != nil {
		return specialists
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), "-specialist.md") {
			// Don't include orchestrator-specialist as it's always present
			if entry.Name() != "orchestrator-specialist.md" {
				specialists = append(specialists, strings.TrimSuffix(entry.Name(), ".md"))
			}
		}
	}

	return specialists
}

// showAvailableAgents displays all available agents with their specialties
func (r *SlashCommandRegistry) showAvailableAgents() string {
	var output strings.Builder
	output.WriteString("🎯 [Orchestrator]: Available agents for this project:\n\n")

	// Check for project specialists
	agentsDir := ".claude/agents"
	specialists := r.listProjectSpecialists(agentsDir)

	// Show project-level orchestrator (always present)
	output.WriteString("PROJECT-LEVEL AGENT (Always installed):\n")
	output.WriteString("🎯 orchestrator-specialist - Project orchestration and intelligent routing\n")
	output.WriteString("   • Location: .claude/agents/orchestrator-specialist.md\n")
	output.WriteString("   • Commands: /crew:orchestrate, /crew:chain, /crew:analyze\n")
	output.WriteString("   • Purpose: Analyzes complexity, routes to agents, suggests specialists\n\n")
	
	// Show other project specialists if any exist
	if len(specialists) > 0 {
		output.WriteString("ADDITIONAL PROJECT SPECIALISTS (Created on-demand):\n")
		for _, specialist := range specialists {
			// Try to infer emoji based on name
			emoji := "📋"
			if strings.Contains(specialist, "go") || strings.Contains(specialist, "backend") {
				emoji = "⚙️"
			} else if strings.Contains(specialist, "api") {
				emoji = "🔌"
			} else if strings.Contains(specialist, "cli") {
				emoji = "💻"
			} else if strings.Contains(specialist, "install") {
				emoji = "📦"
			}
			output.WriteString(fmt.Sprintf("%s %s - Project-specific expertise\n", emoji, specialist))
		}
		output.WriteString("\n")
	} else {
		output.WriteString("ADDITIONAL PROJECT SPECIALISTS: None yet\n")
		output.WriteString("   • Created by Claude when patterns emerge (5+ files with pattern)\n")
		output.WriteString("   • Based on project-analysis.json findings\n")
		output.WriteString("   • Only for complex, repeated patterns personas can't handle\n\n")
	}

	// Always show global personas
	output.WriteString("GLOBAL PERSONAS (Always available):\n")
	output.WriteString("🏗️ architect-persona - System design, architecture\n")
	output.WriteString("🎨 frontend-persona - UI/UX, accessibility\n")
	output.WriteString("⚙️ backend-persona - Server reliability, APIs\n")
	output.WriteString("🛡️ security-persona - Threat modeling, audits\n")
	output.WriteString("🔍 analyzer-persona - Debugging, root cause\n")
	output.WriteString("⚡ performance-persona - Optimization, speed\n")
	output.WriteString("🎯 qa-persona - Testing, quality\n")
	output.WriteString("🔧 refactorer-persona - Code cleanup\n")
	output.WriteString("🚀 devops-persona - CI/CD, deployment\n")
	output.WriteString("📚 mentor-persona - Teaching, guidance\n")
	output.WriteString("✍️ scribe-persona - Documentation\n\n")

	output.WriteString("USAGE GUIDANCE:\n")
	output.WriteString("• Simple tasks → Use personas directly\n")
	output.WriteString("• Complex/multi-step → /crew:orchestrate \"your task\"\n")
	output.WriteString("• Explicit chaining → /crew:chain \"step1 → step2 → step3\"\n")
	output.WriteString("• Check project needs → /crew:analyze\n\n")

	output.WriteString("Use: Task(subagent_type='agent-name', prompt='your task')")

	return output.String()
}

// GenerateCompletionScript generates a shell completion script for /crew: commands
func (r *SlashCommandRegistry) GenerateCompletionScript(shell string) (string, error) {
	commands := r.ListCommands()

	switch shell {
	case "bash":
		return r.generateBashCompletion(commands), nil
	case "zsh":
		return r.generateZshCompletion(commands), nil
	case "fish":
		return r.generateFishCompletion(commands), nil
	default:
		return "", fmt.Errorf("unsupported shell: %s", shell)
	}
}

func (r *SlashCommandRegistry) generateBashCompletion(commands []*SlashCommand) string {
	var script strings.Builder

	script.WriteString("# SuperCrew bash completion\n")
	script.WriteString("_crew_complete() {\n")
	script.WriteString("    local cur prev opts\n")
	script.WriteString("    COMPREPLY=()\n")
	script.WriteString("    cur=\"${COMP_WORDS[COMP_CWORD]}\"\n")
	script.WriteString("    prev=\"${COMP_WORDS[COMP_CWORD-1]}\"\n")

	script.WriteString("    if [[ ${cur} == /crew:* ]]; then\n")
	script.WriteString("        opts=\"")
	for _, cmd := range commands {
		script.WriteString(fmt.Sprintf("/crew:%s ", cmd.Name))
	}
	script.WriteString("\"\n")
	script.WriteString("        COMPREPLY=( $(compgen -W \"${opts}\" -- ${cur}) )\n")
	script.WriteString("        return 0\n")
	script.WriteString("    fi\n")
	script.WriteString("}\n")
	script.WriteString("complete -F _crew_complete claude\n")

	return script.String()
}

func (r *SlashCommandRegistry) generateZshCompletion(commands []*SlashCommand) string {
	var script strings.Builder

	script.WriteString("# SuperCrew zsh completion\n")
	script.WriteString("#compdef claude\n")

	script.WriteString("_crew_commands() {\n")
	script.WriteString("    local commands=(")
	for _, cmd := range commands {
		script.WriteString(fmt.Sprintf("\n        '/crew:%s:%s'", cmd.Name, cmd.Description))
	}
	script.WriteString("\n    )\n")
	script.WriteString("    _describe 'commands' commands\n")
	script.WriteString("}\n")

	script.WriteString("_claude() {\n")
	script.WriteString("    if [[ $words[CURRENT] == /crew:* ]]; then\n")
	script.WriteString("        _crew_commands\n")
	script.WriteString("    fi\n")
	script.WriteString("}\n")

	return script.String()
}

func (r *SlashCommandRegistry) generateFishCompletion(commands []*SlashCommand) string {
	var script strings.Builder

	script.WriteString("# SuperCrew fish completion\n")
	for _, cmd := range commands {
		script.WriteString(fmt.Sprintf("complete -c claude -x -a '/crew:%s' -d '%s'\n", cmd.Name, cmd.Description))
	}

	return script.String()
}

// ExportCommandsJSON exports all commands as JSON for Claude Code integration
func (r *SlashCommandRegistry) ExportCommandsJSON() ([]byte, error) {
	commands := r.ListCommands()
	return json.MarshalIndent(commands, "", "  ")
}
