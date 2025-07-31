package orchestrator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LoadCommandHandler integrates with /crew:load command
type LoadCommandHandler struct {
	ProjectRoot   string
	Installer     *OrchestratorInstaller
	MCPEnhancer   *MCPEnhancer
	ToolsEnhancer *CLIToolsEnhancer
}

// NewLoadCommandHandler creates a new handler
func NewLoadCommandHandler(projectRoot string) *LoadCommandHandler {
	return &LoadCommandHandler{
		ProjectRoot:   projectRoot,
		Installer:     NewOrchestratorInstaller(projectRoot),
		MCPEnhancer:   NewMCPEnhancer(),
		ToolsEnhancer: NewCLIToolsEnhancer(projectRoot),
	}
}

// Execute handles the /crew:load command - simplified for Claude analysis
func (lch *LoadCommandHandler) Execute() error {
	fmt.Println("🎯 Claude Code Super Crew - Project Load")
	fmt.Println("═══════════════════════════════════════")

	// Step 1: Check if local orchestrator exists, prompt creation if needed
	fmt.Println("\n📋 Step 1: Checking for orchestrator-specialist...")
	orchestratorPath := filepath.Join(lch.ProjectRoot, ".claude", "agents", "orchestrator-specialist.md")
	if _, err := os.Stat(orchestratorPath); os.IsNotExist(err) {
		// Orchestrator doesn't exist, prompt Claude to create from global template
		if err := lch.promptOrchestratorCreation(); err != nil {
			return fmt.Errorf("prompting orchestrator creation: %w", err)
		}
		// Don't continue until Claude creates it
		return nil
	} else {
		fmt.Println("✅ Local orchestrator-specialist already exists")
	}

	// Step 2: Create analysis template for Claude
	fmt.Println("\n📝 Step 2: Creating project analysis template...")
	template := lch.Installer.CreateAnalysisTemplate()

	// Step 3: Save template for Claude to fill out
	if err := lch.saveAnalysisTemplate(template); err != nil {
		return fmt.Errorf("failed to save analysis template: %w", err)
	}

	// Step 4: Display instructions for Claude
	lch.instructClaudeToAnalyze()

	// Step 5: Prompt Claude to determine what to enable
	lch.promptForIntelligentEnhancements()

	return nil
}

// saveAnalysisTemplate saves the empty template for Claude
func (lch *LoadCommandHandler) saveAnalysisTemplate(template string) error {
	agentsDir := filepath.Join(lch.ProjectRoot, ".claude", "agents")
	analysisPath := filepath.Join(agentsDir, "project-analysis.json")

	if err := os.WriteFile(analysisPath, []byte(template), 0644); err != nil {
		return fmt.Errorf("failed to write analysis template: %w", err)
	}

	fmt.Println("✅ Analysis template created at .claude/agents/project-analysis.json")
	return nil
}

// instructClaudeToAnalyze provides clear instructions for Claude
func (lch *LoadCommandHandler) instructClaudeToAnalyze() {
	fmt.Println("\n🤖 Instructions for Claude Code:")
	fmt.Println("════════════════════════════════")
	fmt.Println()
	fmt.Println("The orchestrator-specialist has been installed!")
	fmt.Println()
	fmt.Println("📋 Next Steps:")
	fmt.Println("1. Claude, please analyze this project and fill out the project-analysis.json")
	fmt.Println("2. Use your tools to explore:")
	fmt.Println("   - Language distribution (Glob for file extensions)")
	fmt.Println("   - Framework detection (Read key files like go.mod, package.json)")
	fmt.Println("   - Architectural patterns (Grep for patterns like 'handler', 'api', 'cli')")
	fmt.Println("   - Complexity assessment (LS for structure, file counts)")
	fmt.Println()
	fmt.Println("3. Update project-analysis.json with your findings:")
	fmt.Println("   - Replace 'TO_BE_ANALYZED' placeholders")
	fmt.Println("   - Add discovered languages, frameworks, patterns")
	fmt.Println("   - Assess complexity and orchestration benefit")
	fmt.Println("   - Recommend specialists only if truly needed")
	fmt.Println()
	fmt.Println("4. The orchestrator will use your analysis to:")
	fmt.Println("   - Route commands intelligently")
	fmt.Println("   - Suggest specialist creation when patterns emerge")
	fmt.Println("   - Coordinate multi-agent workflows")
	fmt.Println()
	fmt.Println("🎯 Remember: Only recommend specialists for repeated, complex patterns!")
}

// ClaudeAnalysisInstructions returns structured instructions for Claude
func (lch *LoadCommandHandler) ClaudeAnalysisInstructions() string {
	return `
CLAUDE ANALYSIS WORKFLOW:

1. LANGUAGE ANALYSIS:
   - Use Glob to find all source files
   - Count files by extension
   - Identify primary language (most files)
   - Note any language-specific patterns

2. FRAMEWORK DETECTION:
   - Check for dependency files (go.mod, package.json, etc.)
   - Read configuration files
   - Identify build tools and package managers
   - Note framework versions if available

3. ARCHITECTURAL PATTERNS:
   - Use Grep to find common patterns:
     * API: "handler", "endpoint", "route", "controller"
     * CLI: "command", "flag", "args", "cli"
     * Web: "server", "http", "template"
     * Testing: "_test", "spec", "mock"
   - Assess confidence based on file count and consistency

4. COMPLEXITY ASSESSMENT:
   - Count total files and directories
   - Measure directory depth
   - Count distinct technical domains
   - Evaluate orchestration benefit

5. SPECIALIST RECOMMENDATIONS:
   - Only recommend if:
     * Pattern appears in 5+ files
     * Domain is complex enough to warrant specialization
     * Generic personas insufficient
   - Include clear trigger conditions

6. UPDATE project-analysis.json:
   - Replace all placeholders with findings
   - Be conservative with recommendations
   - Focus on actual needs, not possibilities
`
}

// ExecuteWithEnhancements handles the /crew:load command with MCP and tools integration
func (lch *LoadCommandHandler) ExecuteWithEnhancements(mcpFlags []string, toolFlags []string) error {
	// Run standard load process first
	if err := lch.Execute(); err != nil {
		return err
	}

	// If MCP flags provided, enhance CLAUDE.md
	if len(mcpFlags) > 0 {
		fmt.Println("\n🔧 Step 5: Enhancing project with MCP integrations...")
		if err := lch.enhanceProjectCLAUDE(mcpFlags); err != nil {
			return fmt.Errorf("failed to enhance CLAUDE.md: %w", err)
		}
	}

	// If tool flags provided, enhance with CLI tools
	if len(toolFlags) > 0 {
		fmt.Println("\n🛠️ Step 6: Enhancing project with CLI tool integrations...")
		if err := lch.enhanceProjectWithTools(toolFlags); err != nil {
			return fmt.Errorf("failed to enhance with CLI tools: %w", err)
		}
	}

	return nil
}

// ExecuteWithMCP is kept for backward compatibility
func (lch *LoadCommandHandler) ExecuteWithMCP(mcpFlags []string) error {
	return lch.ExecuteWithEnhancements(mcpFlags, nil)
}

// enhanceProjectCLAUDE adds MCP integrations to project CLAUDE.md
func (lch *LoadCommandHandler) enhanceProjectCLAUDE(enabledServers []string) error {
	// CLAUDE.md goes in project root, not .claude/
	claudeFile := filepath.Join(lch.ProjectRoot, "CLAUDE.md")
	
	// Still ensure .claude directory exists for other files
	claudeDir := filepath.Join(lch.ProjectRoot, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return fmt.Errorf("failed to create .claude directory: %w", err)
	}

	// Read existing CLAUDE.md or create default
	var existingContent string
	if data, err := os.ReadFile(claudeFile); err == nil {
		existingContent = string(data)
	} else {
		// Create basic CLAUDE.md
		existingContent = lch.createDefaultCLAUDE()
	}

	// Detect project type (simplified - you can enhance this)
	projectType := lch.detectProjectType()

	// Enhance with MCP content
	enhancedContent := lch.MCPEnhancer.EnhanceProjectCLAUDE(existingContent, enabledServers, projectType)

	// Write enhanced content
	if err := os.WriteFile(claudeFile, []byte(enhancedContent), 0644); err != nil {
		return fmt.Errorf("failed to write CLAUDE.md: %w", err)
	}

	fmt.Printf("✅ Enhanced CLAUDE.md with MCP integrations: %s\n", strings.Join(enabledServers, ", "))
	fmt.Println("📄 File created/updated at CLAUDE.md (project root)")

	return nil
}

// createDefaultCLAUDE creates a basic CLAUDE.md template
func (lch *LoadCommandHandler) createDefaultCLAUDE() string {
	projectName := filepath.Base(lch.ProjectRoot)
	return fmt.Sprintf(`# %s Project Configuration

## 🎯 Project Overview
This project uses Claude Code Super Crew for intelligent code assistance and orchestration.

## 📋 Project Conventions
- Follow existing code patterns and style
- Maintain test coverage
- Document significant changes

## 🤖 Orchestrator Integration
The project includes an orchestrator-specialist at .claude/agents/orchestrator-specialist.md that:
- Routes commands intelligently based on complexity
- Coordinates multi-agent workflows when needed
- Suggests specialist creation for repeated patterns

`, projectName)
}

// detectProjectType attempts to detect the project type
func (lch *LoadCommandHandler) detectProjectType() string {
	// Check for various project indicators
	if _, err := os.Stat(filepath.Join(lch.ProjectRoot, "go.mod")); err == nil {
		return "go"
	}
	if _, err := os.Stat(filepath.Join(lch.ProjectRoot, "package.json")); err == nil {
		// Read package.json to determine if React/Vue/Angular
		if data, err := os.ReadFile(filepath.Join(lch.ProjectRoot, "package.json")); err == nil {
			content := string(data)
			if strings.Contains(content, "react") {
				return "react"
			}
			if strings.Contains(content, "vue") {
				return "vue"
			}
			if strings.Contains(content, "angular") {
				return "angular"
			}
		}
		return "frontend"
	}
	if _, err := os.Stat(filepath.Join(lch.ProjectRoot, "Cargo.toml")); err == nil {
		return "rust"
	}

	// Default to generic
	return "fullstack"
}

// enhanceProjectWithTools adds CLI tools to project CLAUDE.md
func (lch *LoadCommandHandler) enhanceProjectWithTools(enabledTools []string) error {
	// CLAUDE.md goes in project root, not .claude/
	claudeFile := filepath.Join(lch.ProjectRoot, "CLAUDE.md")

	// Read existing CLAUDE.md
	existingContent, err := os.ReadFile(claudeFile)
	if err != nil {
		return fmt.Errorf("failed to read CLAUDE.md: %w", err)
	}

	// Check tool availability and provide installation guide
	missingTools := []string{}
	for _, tool := range enabledTools {
		if available, _ := lch.ToolsEnhancer.CheckToolAvailability(tool); !available {
			missingTools = append(missingTools, tool)
		}
	}

	if len(missingTools) > 0 {
		fmt.Println("\n⚠️  Some tools are not installed:")
		fmt.Println(lch.ToolsEnhancer.InstallationGuide(missingTools))
	}

	// Enhance with CLI tools content
	enhancedContent := lch.ToolsEnhancer.EnhanceProjectCLAUDEWithTools(string(existingContent), enabledTools)

	// Write enhanced content
	if err := os.WriteFile(claudeFile, []byte(enhancedContent), 0644); err != nil {
		return fmt.Errorf("failed to write CLAUDE.md: %w", err)
	}

	fmt.Printf("✅ Enhanced CLAUDE.md with CLI tools: %s\n", strings.Join(enabledTools, ", "))

	// Create second-opinion-generator if code2prompt is enabled
	if lch.shouldCreateSecondOpinionAgent(enabledTools) {
		fmt.Println("\n📋 Creating second-opinion-generator agent...")
		if err := lch.createSecondOpinionAgent(); err != nil {
			fmt.Printf("⚠️  Failed to create second-opinion agent: %v\n", err)
		} else {
			fmt.Println("✅ Created .claude/agents/second-opinion-generator.md")
		}
	}

	return nil
}

// shouldCreateSecondOpinionAgent checks if we should create the second opinion agent
func (lch *LoadCommandHandler) shouldCreateSecondOpinionAgent(enabledTools []string) bool {
	for _, tool := range enabledTools {
		if tool == "code2prompt" {
			return true
		}
	}
	return false
}

// createSecondOpinionAgent creates the second-opinion-generator agent
func (lch *LoadCommandHandler) createSecondOpinionAgent() error {
	agentsDir := filepath.Join(lch.ProjectRoot, ".claude", "agents")
	agentFile := filepath.Join(agentsDir, "second-opinion-generator.md")

	// Check if agent already exists
	if _, err := os.Stat(agentFile); err == nil {
		fmt.Println("ℹ️  second-opinion-generator.md already exists, skipping creation")
		return nil
	}

	// Read the template from SuperCrew/Agents
	templatePath := filepath.Join(lch.ProjectRoot, "SuperCrew", "Agents", "second-opinion-generator.md")
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		// If template not found, use a basic version
		templateContent = []byte(lch.getBasicSecondOpinionTemplate())
	}

	// Write to project agents directory
	if err := os.WriteFile(agentFile, templateContent, 0644); err != nil {
		return fmt.Errorf("failed to write agent file: %w", err)
	}

	return nil
}

// getBasicSecondOpinionTemplate returns a basic template if the full one isn't available
func (lch *LoadCommandHandler) getBasicSecondOpinionTemplate() string {
	return `---
name: second-opinion-generator
description: Expert at generating comprehensive second-opinion prompt packages for sharing with other AI tools. Creates self-contained Markdown files with full context and actionable instructions.
tools: Read, Write, Bash, Grep, Glob, TodoWrite, Task
---

You are a specialized agent for creating comprehensive second-opinion prompt packages that can be shared with other AI tools.

## Core Responsibilities
1. Analyze the current context and problem
2. Gather relevant code using code2prompt and ast-grep
3. Create structured, self-contained prompts
4. Save to .claude/prompts/[category]/
5. Route issues to appropriate specialists

## Workflow
1. Use code2prompt for comprehensive context:
   ` + "`code2prompt --include \"**/*.{go,js,ts}\" --exclude \"**/node_modules/**\"`" + `

2. Use ast-grep for semantic patterns:
   ` + "`ast-grep run -p 'pattern' --lang language`" + `

3. Create structured markdown following the template format
4. Save to appropriate category directory
5. Trigger orchestration for identified issues

See full documentation in SuperCrew/Agents/second-opinion-generator.md
`
}

// promptForIntelligentEnhancements prompts Claude to analyze and determine what to enable
// promptOrchestratorCreation prompts Claude to create local orchestrator from global template
func (lch *LoadCommandHandler) promptOrchestratorCreation() error {
	fmt.Println("\n🎯 Orchestrator Creation Required")
	fmt.Println("═══════════════════════════════════════")
	fmt.Println()
	fmt.Println("The local orchestrator is the brain of your project's agent system!")
	fmt.Println("It needs to be created from the global template to understand YOUR project.")
	fmt.Println()
	fmt.Println("📋 Claude, please:")
	fmt.Println()
	fmt.Println("1. **Read** the global orchestrator template:")
	fmt.Println("   Path: templates/agents/orchestrator.template.md")
	fmt.Println()
	fmt.Println("2. **Analyze** this project to understand:")
	fmt.Println("   - Primary programming language(s)")
	fmt.Println("   - Frameworks and libraries in use")
	fmt.Println("   - Project structure and patterns")
	fmt.Println("   - Common workflows and pain points")
	fmt.Println()
	fmt.Println("3. **Create** a customized orchestrator at:")
	fmt.Println("   Path: .claude/agents/orchestrator-specialist.md")
	fmt.Println()
	fmt.Println("4. **Include** these essential features:")
	fmt.Println("   - Project-specific routing rules")
	fmt.Println("   - Common workflow patterns")
	fmt.Println("   - Double/triple check completion verification")
	fmt.Println("   - Support for local slash commands")
	fmt.Println("   - Dynamic specialist recommendations")
	fmt.Println()
	fmt.Println("5. **Customize** for this project:")
	fmt.Println("   - Add project context section")
	fmt.Println("   - Define routing patterns based on your analysis")
	fmt.Println("   - List any existing specialists")
	fmt.Println("   - Suggest new specialists if patterns warrant")
	fmt.Println("   - IMPORTANT: Recommend language-specific specialists:")
	fmt.Println("     • Go project → suggest go-specialist")
	fmt.Println("     • Python project → suggest python-specialist")
	fmt.Println("     • JavaScript/TypeScript → suggest js-specialist or ts-specialist")
	fmt.Println("     • Java project → suggest java-specialist")
	fmt.Println("     • Multiple languages → suggest specialists for each primary language")
	fmt.Println()
	fmt.Println("💡 Remember: The orchestrator should feel custom-built for THIS project!")
	fmt.Println("It's the intelligent router that makes multi-agent workflows seamless.")
	fmt.Println()
	fmt.Println("After creating the orchestrator, run /crew:load again to continue setup.")
	return nil
}

func (lch *LoadCommandHandler) promptForIntelligentEnhancements() {
	fmt.Println("\n🤖 Claude Code Super Crew - Intelligent Project Enhancement")
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("## 🎯 Your Mission")
	fmt.Println("Claude, you have access to powerful subagent capabilities! Please analyze this")
	fmt.Println("project and create the best combination of agents to improve development speed")
	fmt.Println("and code quality. Be creative and think about what would truly help!")
	fmt.Println()
	fmt.Println("## 📋 Available Resources")
	fmt.Println()
	fmt.Println("### MCP Servers (Enable what helps):")
	fmt.Println("- **Context7**: Documentation lookup for any library")
	fmt.Println("- **Sequential**: Complex multi-step analysis") 
	fmt.Println("- **Magic**: UI component generation")
	fmt.Println("- **Playwright**: Browser testing automation")
	fmt.Println("- **Serena**: Semantic code understanding")
	fmt.Println()
	fmt.Println("### CLI Tools (Enable if useful):")
	fmt.Println("- **code2prompt**: Generate comprehensive code context")
	fmt.Println("- **ast-grep**: Semantic pattern matching")
	fmt.Println()
	fmt.Println("### Framework Agents (Always available):")
	fmt.Println("- **orchestrator-specialist**: Already installed for routing")
	fmt.Println("- **second-opinion-generator**: Created when code2prompt enabled")
	fmt.Println()
	fmt.Println("## 🚀 Your Task")
	fmt.Println()
	fmt.Println("1. **Analyze** the project deeply - understand its purpose and patterns")
	fmt.Println("2. **Identify** pain points and repetitive tasks that slow development")
	fmt.Println("3. **Design** specialist agents that would genuinely help, such as:")
	fmt.Println("   - Language-specific specialists based on what you find:")
	fmt.Println("     • Go project → go-specialist (patterns, idioms, performance)")
	fmt.Println("     • Python project → python-specialist (pythonic patterns, type hints)")
	fmt.Println("     • JavaScript/TypeScript → js-specialist or ts-specialist")
	fmt.Println("     • Multiple languages → polyglot-specialist for cross-language work")
	fmt.Println("   - Pattern-specific agents (e.g., API endpoint creator)")
	fmt.Println("   - Domain experts (e.g., error handling specialist)")
	fmt.Println("   - Workflow automators (e.g., test generator)")
	fmt.Println("   - Integration helpers (e.g., third-party API integrator)")
	fmt.Println()
	fmt.Println("4. **Create** the agents with clear, actionable capabilities")
	fmt.Println("5. **Enable** appropriate MCP servers and tools")
	fmt.Println("6. **Document** everything in CLAUDE.md")
	fmt.Println()
	fmt.Println("## 💡 Guidelines")
	fmt.Println()
	fmt.Println("- Be **creative** - don't just create generic agents")
	fmt.Println("- Be **specific** - agents should solve real problems in this codebase")
	fmt.Println("- Be **practical** - focus on what would actually speed up development")
	fmt.Println("- Use the **Task tool** to delegate complex agent creation if needed")
	fmt.Println("- Consider **subagent chains** for multi-step workflows")
	fmt.Println()
	fmt.Println("## 📚 Reference")
	fmt.Println("Check the Claude Code documentation for subagent best practices.")
	fmt.Println("Look for patterns in the codebase that could benefit from automation.")
	fmt.Println()
	fmt.Println("## 🛠️ Pro Tip: Use Available Tools!")
	fmt.Println()
	fmt.Println("The orchestrator-specialist can use tools to help you:")
	fmt.Println("- **code2prompt**: Generate full project context for better analysis")
	fmt.Println("- **ast-grep**: Find specific patterns across the codebase")
	fmt.Println("- **Serena**: Understand code structure semantically")
	fmt.Println()
	fmt.Println("Example: Use code2prompt to create a comprehensive overview:")
	fmt.Println("  code2prompt --include \"**/*.go\" --exclude \"vendor/**\" --max-tokens 30000")
	fmt.Println()
	fmt.Println("🎯 Remember: The goal is to make this project easier and faster to develop!")
	fmt.Println("Create agents that developers would actually want to use every day.")
	fmt.Println()
	fmt.Println("Please proceed with your analysis and create amazing agents!")
}

// ExecuteAutoEnhance runs the full enhancement based on Claude's analysis
func (lch *LoadCommandHandler) ExecuteAutoEnhance() error {
	// This method is called by Claude after analysis
	// Claude will determine which flags to use based on project analysis

	fmt.Println("\n🚀 Auto-enhancing project based on analysis...")

	// Claude should call this with appropriate flags based on analysis
	// For example, after analyzing a Go CLI project:
	// - Enable Serena for code analysis
	// - Enable Context7 for Go documentation
	// - Enable Sequential for complex orchestration logic
	// - Enable code2prompt for comprehensive context
	// - Create CLI-specialist if many CLI commands

	return nil
}
