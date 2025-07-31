// Package claude provides comprehensive help system for SuperCrew commands.
// This enables intelligent help discovery, context-aware assistance, and command guidance
// for the /crew: slash command system integrated with Claude Code.
package claude

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jonwraymond/claude-code-super-crew/pkg/logger"
)

// HelpProvider provides context-aware help for supercrew commands
type HelpProvider struct {
	registry *SlashCommandRegistry
	logger   logger.Logger
}

// NewHelpProvider creates a new help provider
func NewHelpProvider(commandsPath string) (*HelpProvider, error) {
	registry := NewSlashCommandRegistry(commandsPath)
	if err := registry.LoadCommands(); err != nil {
		return nil, fmt.Errorf("failed to load commands: %w", err)
	}

	return &HelpProvider{
		registry: registry,
		logger:   logger.GetLogger(),
	}, nil
}

// SuperCrewHelp contains comprehensive help information
type SuperCrewHelp struct {
	Overview    string                   `json:"overview"`
	QuickStart  []string                 `json:"quick_start"`
	Categories  map[string][]CommandHelp `json:"categories"`
	CommonUsage []UsageExample           `json:"common_usage"`
	Tips        []string                 `json:"tips"`
}

// CommandHelp provides detailed help for individual commands
type CommandHelp struct {
	Name          string         `json:"name"`
	Purpose       string         `json:"purpose"`
	AutoActivates string         `json:"auto_activates"`
	BestFor       string         `json:"best_for"`
	Usage         string         `json:"usage"`
	Examples      []UsageExample `json:"examples"`
	Flags         []FlagHelp     `json:"flags"`
}

// UsageExample shows practical command usage
type UsageExample struct {
	Command     string `json:"command"`
	Description string `json:"description"`
	Context     string `json:"context"`
}

// FlagHelp provides flag documentation
type FlagHelp struct {
	Flag        string   `json:"flag"`
	Description string   `json:"description"`
	Values      []string `json:"values,omitempty"`
}

// GetComprehensiveHelp returns complete SuperCrew help system
func (hp *HelpProvider) GetComprehensiveHelp() *SuperCrewHelp {
	return &SuperCrewHelp{
		Overview:    hp.getOverview(),
		QuickStart:  hp.getQuickStart(),
		Categories:  hp.getCategorizedCommands(),
		CommonUsage: hp.getCommonUsageExamples(),
		Tips:        hp.getTips(),
	}
}

// getOverview provides SuperCrew framework overview
func (hp *HelpProvider) getOverview() string {
	return `Claude Code Super Crew makes Claude Code smarter for development work. 
Instead of generic responses, you get specialized help from different experts 
(security, performance, frontend, etc.) who know their stuff.

🎯 The Simple Truth: You don't need to learn all the commands, flags, and personas. 
Just start using it! Claude Code Super Crew has an intelligent routing system that 
tries to figure out what you need automatically.

Learning emerges during use - you'll naturally discover what works without 
studying manuals first.`
}

// getQuickStart provides essential commands to get users started
func (hp *HelpProvider) getQuickStart() []string {
	return []string{
		"/crew:help                    # See what's available",
		"/crew:analyze README.md       # Analyze your project intelligently",
		"/crew:workflow feature-prd.md # Generate implementation workflow from PRD",
		"/crew:implement user-auth     # Create features and components",
		"/crew:build                   # Smart build with auto-optimization",
		"/crew:improve messy-file.js   # Clean up code automatically",
	}
}

// getCategorizedCommands organizes commands by category with detailed help
func (hp *HelpProvider) getCategorizedCommands() map[string][]CommandHelp {
	categories := make(map[string][]CommandHelp)
	commands := hp.registry.ListCommands()

	for _, cmd := range commands {
		category := hp.categorizeCommand(cmd.Name)
		help := hp.getCommandHelp(cmd)
		categories[category] = append(categories[category], help)
	}

	// Sort commands within each category
	for category := range categories {
		sort.Slice(categories[category], func(i, j int) bool {
			return categories[category][i].Name < categories[category][j].Name
		})
	}

	return categories
}

// getCommandHelp provides detailed help for a specific command
func (hp *HelpProvider) getCommandHelp(cmd *SlashCommand) CommandHelp {
	help := CommandHelp{
		Name:     cmd.Name,
		Purpose:  cmd.Description,
		Usage:    cmd.Usage,
		Examples: hp.getCommandExamples(cmd.Name),
		Flags:    hp.getCommandFlags(cmd.Name),
	}

	// Add SuperCrew-specific context
	switch cmd.Name {
	case "analyze":
		help.AutoActivates = "Security/performance experts based on code patterns"
		help.BestFor = "Finding issues, understanding codebases, security analysis"
	case "implement":
		help.AutoActivates = "Domain-specific experts (frontend, backend, security)"
		help.BestFor = "Creating features, components, APIs, services"
	case "build":
		help.AutoActivates = "Frontend/backend specialists based on project type"
		help.BestFor = "Compilation, bundling, deployment preparation"
	case "improve":
		help.AutoActivates = "Quality experts, refactoring specialists"
		help.BestFor = "Code cleanup, refactoring, optimization, quality fixes"
	case "workflow":
		help.AutoActivates = "Architecture experts, project management"
		help.BestFor = "Creating implementation plans from PRDs, project planning"
	case "troubleshoot":
		help.AutoActivates = "Debug specialists, domain experts"
		help.BestFor = "Debugging, issue investigation, root cause analysis"
	case "test":
		help.AutoActivates = "QA experts, testing specialists"
		help.BestFor = "Running tests, coverage analysis, quality assurance"
	case "document":
		help.AutoActivates = "Writing specialists, technical documentation experts"
		help.BestFor = "README files, code comments, API documentation"
	case "git":
		help.AutoActivates = "DevOps specialists, version control experts"
		help.BestFor = "Smart commits, branch management, release workflows"
	case "design":
		help.AutoActivates = "Architecture experts, system designers"
		help.BestFor = "System architecture, API design, component planning"
	default:
		help.AutoActivates = "Context-aware expert selection"
		help.BestFor = "Specialized " + cmd.Name + " operations"
	}

	return help
}

// getCommandExamples provides practical usage examples for each command
func (hp *HelpProvider) getCommandExamples(commandName string) []UsageExample {
	examples := make(map[string][]UsageExample)

	examples["analyze"] = []UsageExample{
		{"/crew:analyze src/", "Analyze entire source directory", "Code review, quality assessment"},
		{"/crew:analyze auth/ --focus security", "Security-focused analysis", "Security audit, vulnerability assessment"},
		{"/crew:analyze --focus performance api/", "Performance analysis", "Bottleneck identification, optimization"},
	}

	examples["implement"] = []UsageExample{
		{"/crew:implement user-auth --type feature", "Implement authentication feature", "Feature development"},
		{"/crew:implement dashboard --type component", "Create dashboard component", "UI development"},
		{"/crew:implement payment-api --type api", "Build payment API", "Backend development"},
	}

	examples["build"] = []UsageExample{
		{"/crew:build", "Smart project build", "General project compilation"},
		{"/crew:build frontend/ --optimize", "Optimized frontend build", "Production builds"},
		{"/crew:build --target production", "Production build", "Deployment preparation"},
	}

	examples["improve"] = []UsageExample{
		{"/crew:improve messy-code.js", "Clean up messy code", "Code refactoring"},
		{"/crew:improve --focus performance", "Performance improvements", "Optimization"},
		{"/crew:improve --safe-mode legacy-code/", "Safe legacy code improvements", "Legacy modernization"},
	}

	examples["workflow"] = []UsageExample{
		{"/crew:workflow feature-prd.md", "Generate workflow from PRD", "Project planning"},
		{"/crew:workflow --type feature user-stories/", "Feature workflow planning", "Development planning"},
	}

	examples["troubleshoot"] = []UsageExample{
		{"/crew:troubleshoot \"login fails randomly\"", "Debug authentication issues", "Problem investigation"},
		{"/crew:troubleshoot --focus performance slow-api", "Performance troubleshooting", "Performance debugging"},
	}

	examples["test"] = []UsageExample{
		{"/crew:test", "Run comprehensive tests", "Quality assurance"},
		{"/crew:test --coverage", "Test with coverage analysis", "Code coverage assessment"},
		{"/crew:test --type e2e", "End-to-end testing", "Integration testing"},
	}

	examples["document"] = []UsageExample{
		{"/crew:document README --type guide", "Create user guide", "Documentation"},
		{"/crew:document api/ --type api", "Generate API documentation", "API documentation"},
	}

	examples["git"] = []UsageExample{
		{"/crew:git --smart-commit", "Intelligent commit creation", "Version control"},
		{"/crew:git --release", "Release management", "Release workflow"},
	}

	examples["design"] = []UsageExample{
		{"/crew:design user-service --type api", "API design", "Architecture planning"},
		{"/crew:design --type architecture system/", "System architecture design", "System design"},
	}

	if cmdExamples, exists := examples[commandName]; exists {
		return cmdExamples
	}

	return []UsageExample{
		{fmt.Sprintf("/crew:%s [args]", commandName), fmt.Sprintf("Basic %s usage", commandName), "General usage"},
	}
}

// getCommandFlags provides flag documentation for commands
func (hp *HelpProvider) getCommandFlags(commandName string) []FlagHelp {
	commonFlags := []FlagHelp{
		{"--focus", "Focus area for operation", []string{"quality", "security", "performance", "architecture"}},
		{"--type", "Type of implementation or analysis", []string{"component", "api", "service", "feature", "architecture"}},
		{"--scope", "Scope of operation", []string{"file", "module", "project", "system"}},
		{"--safe-mode", "Use conservative, safe approach", nil},
		{"--validate", "Enable validation and safety checks", nil},
		{"--think", "Enable deeper analysis and reasoning", nil},
	}

	// Command-specific flags
	specificFlags := make(map[string][]FlagHelp)

	specificFlags["analyze"] = []FlagHelp{
		{"--deep", "Comprehensive deep analysis", nil},
		{"--summary", "Provide high-level summary", nil},
	}

	specificFlags["implement"] = []FlagHelp{
		{"--framework", "Target framework", []string{"react", "vue", "express", "fastapi"}},
		{"--with-tests", "Include test implementation", nil},
	}

	specificFlags["build"] = []FlagHelp{
		{"--optimize", "Enable build optimizations", nil},
		{"--target", "Build target", []string{"development", "production", "testing"}},
	}

	specificFlags["improve"] = []FlagHelp{
		{"--preview", "Show changes before applying", nil},
		{"--loop", "Enable iterative improvement", nil},
	}

	specificFlags["test"] = []FlagHelp{
		{"--coverage", "Include coverage analysis", nil},
		{"--benchmark", "Include performance benchmarks", nil},
	}

	// Combine common and specific flags
	flags := commonFlags
	if specific, exists := specificFlags[commandName]; exists {
		flags = append(flags, specific...)
	}

	return flags
}

// getCommonUsageExamples provides practical workflow examples
func (hp *HelpProvider) getCommonUsageExamples() []UsageExample {
	return []UsageExample{
		{"/crew:analyze project/ --focus security", "Security audit workflow", "Security assessment"},
		{"/crew:implement auth --type feature --with-tests", "Feature development with testing", "Full-stack development"},
		{"/crew:improve legacy-code/ --safe-mode --preview", "Safe legacy modernization", "Legacy maintenance"},
		{"/crew:workflow feature-prd.md && /crew:implement", "PRD to implementation workflow", "Project execution"},
		{"/crew:troubleshoot \"slow API\" --focus performance", "Performance debugging", "Performance optimization"},
		{"/crew:build --optimize --target production", "Production deployment", "DevOps workflow"},
		{"/crew:test --coverage && /crew:document --type api", "Quality assurance with docs", "Quality workflow"},
		{"/crew:design system --type architecture", "System architecture planning", "Architecture design"},
	}
}

// getTips provides practical usage tips
func (hp *HelpProvider) getTips() []string {
	return []string{
		"🎯 Start simple: Begin with basic commands like /crew:analyze and /crew:implement",
		"🤖 Trust auto-activation: SuperCrew usually picks the right experts automatically",
		"🔒 Use --safe-mode for critical code: Enables conservative changes with validation",
		"🧠 Add --think for complex problems: Enables deeper analysis and reasoning",
		"📊 Try --focus flags: Direct attention to specific aspects (security, performance, etc.)",
		"🔄 Use --loop for iterative improvement: Great for code cleanup and optimization",
		"📖 Use --preview first: See what changes would be made before applying them",
		"⚡ Start with /crew:workflow for new features: Creates structured implementation plans",
		"🎭 Let personas auto-activate: Security expert for auth, performance expert for optimization",
		"📚 Commands work together: /crew:analyze → /crew:improve → /crew:test is a common flow",
	}
}

// categorizeCommand categorizes commands for help organization
func (hp *HelpProvider) categorizeCommand(name string) string {
	categories := map[string]string{
		// Development commands
		"build":     "Development",
		"implement": "Development",
		"design":    "Development",
		"workflow":  "Development",

		// Analysis commands
		"analyze":      "Analysis",
		"troubleshoot": "Analysis",
		"explain":      "Analysis",
		"index":        "Analysis",

		// Quality commands
		"improve": "Quality",
		"cleanup": "Quality",
		"test":    "Quality",

		// Utilities commands
		"document": "Utilities",
		"git":      "Utilities",
		"load":     "Utilities",
		"estimate": "Utilities",
		"task":     "Utilities",
		"spawn":    "Utilities",
	}

	if category, exists := categories[name]; exists {
		return category
	}
	return "General"
}

// GetCommandHelp returns help for a specific command
func (hp *HelpProvider) GetCommandHelp(commandName string) (*CommandHelp, error) {
	cmd, exists := hp.registry.GetCommand(commandName)
	if !exists {
		return nil, fmt.Errorf("command not found: %s", commandName)
	}

	help := hp.getCommandHelp(cmd)
	return &help, nil
}

// SearchCommands finds commands matching a search query
func (hp *HelpProvider) SearchCommands(query string) []CommandHelp {
	var matches []CommandHelp
	commands := hp.registry.ListCommands()

	query = strings.ToLower(query)

	for _, cmd := range commands {
		// Search in command name, description, and usage
		searchText := strings.ToLower(fmt.Sprintf("%s %s %s",
			cmd.Name, cmd.Description, cmd.Usage))

		if strings.Contains(searchText, query) {
			matches = append(matches, hp.getCommandHelp(cmd))
		}
	}

	return matches
}

// GetSuggestedCommands returns command suggestions based on user context
func (hp *HelpProvider) GetSuggestedCommands(context string) []CommandHelp {
	var suggestions []CommandHelp
	context = strings.ToLower(context)

	// Context-based suggestions
	contextMappings := map[string][]string{
		"security":    {"analyze", "implement", "troubleshoot", "test"},
		"performance": {"analyze", "improve", "troubleshoot", "test"},
		"frontend":    {"implement", "build", "improve", "design"},
		"backend":     {"implement", "build", "analyze", "test"},
		"api":         {"design", "implement", "test", "document"},
		"bug":         {"troubleshoot", "analyze", "improve", "test"},
		"feature":     {"workflow", "implement", "design", "test"},
		"refactor":    {"analyze", "improve", "test", "document"},
		"deploy":      {"build", "test", "git", "document"},
	}

	for contextKey, commands := range contextMappings {
		if strings.Contains(context, contextKey) {
			for _, cmdName := range commands {
				if cmd, exists := hp.registry.GetCommand(cmdName); exists {
					suggestions = append(suggestions, hp.getCommandHelp(cmd))
				}
			}
			break // Use first matching context
		}
	}

	// If no context match, return most common commands
	if len(suggestions) == 0 {
		commonCommands := []string{"analyze", "implement", "improve", "build"}
		for _, cmdName := range commonCommands {
			if cmd, exists := hp.registry.GetCommand(cmdName); exists {
				suggestions = append(suggestions, hp.getCommandHelp(cmd))
			}
		}
	}

	return suggestions
}
