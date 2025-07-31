package orchestrator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jonwraymond/claude-code-super-crew/pkg/logger"
)

// AgentGenerator generates project-specific agent configurations
type AgentGenerator struct {
	projectPath string
	logger      logger.Logger
}

// NewAgentGenerator creates a new agent generator
func NewAgentGenerator(projectPath string) *AgentGenerator {
	return &AgentGenerator{
		projectPath: projectPath,
		logger:      logger.GetLogger(),
	}
}

// GenerateAgents generates agent configuration files based on project characteristics
func (ag *AgentGenerator) GenerateAgents(chars *ProjectCharacteristics) error {
	// Create local .claude/agents directory
	agentsDir := filepath.Join(ag.projectPath, ".claude", "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		return fmt.Errorf("failed to create agents directory: %w", err)
	}

	// Generate each detected agent
	generatedAgents := []string{}
	for _, agentType := range chars.DetectedAgents {
		agentFile := filepath.Join(agentsDir, agentType+".md")

		// Skip if agent already exists
		if _, err := os.Stat(agentFile); err == nil {
			ag.logger.Info(fmt.Sprintf("Agent %s already exists, skipping", agentType))
			continue
		}

		// Generate agent content
		content := ag.generateAgentContent(agentType, chars)

		// Write agent file
		if err := os.WriteFile(agentFile, []byte(content), 0644); err != nil {
			ag.logger.Error(fmt.Sprintf("Failed to write agent %s: %v", agentType, err))
			continue
		}

		generatedAgents = append(generatedAgents, agentType)
		ag.logger.Info(fmt.Sprintf("Generated agent: %s", agentType))
	}

	// Log summary
	if len(generatedAgents) > 0 {
		ag.logger.Success(fmt.Sprintf("Project-specific agents created in .claude/agents/: %s",
			strings.Join(generatedAgents, ", ")))
	} else {
		ag.logger.Info("No new agents generated (may already exist)")
	}

	return nil
}

// generateAgentContent generates the content for a specific agent type
func (ag *AgentGenerator) generateAgentContent(agentType string, chars *ProjectCharacteristics) string {
	// Get visual identity for this agent type
	emoji, bgColor, textColor := ag.getVisualIdentity(agentType)

	// Base template with visual identity
	template := `---
name: %s
description: %s
version: "1.0.0"
created: "%s"
project: "%s"
language: "%s"
frameworks: %s
visual_identity:
  emoji: "%s"
  background_color: "%s"
  text_color: "%s"
tags: ["%s", "project-specific", "auto-generated"]
activation_keywords:
  primary: %s
  secondary: %s
  contextual: %s
slash_commands: %s
tools:
  - Read
  - Write
  - Edit
  - MultiEdit
  - Grep
  - Glob
  - Bash
  - TodoWrite
  - Task
---

# %s %s

You are a specialized %s agent for the %s project. You have deep expertise in %s and understand the specific patterns, conventions, and requirements of this codebase.

When contributing to outputs, your responses are marked with the %s emoji for clear visual identification.

## Project Context

- **Primary Language**: %s
- **Frameworks**: %s
- **Project Path**: %s

## Core Responsibilities

%s

## Technical Expertise

%s

## Best Practices

%s

## Project-Specific Guidelines

1. Follow the existing code style and conventions in this repository
2. Respect the project's architecture and design patterns
3. Ensure compatibility with the project's dependency versions
4. Maintain consistency with existing implementations
5. Prioritize solutions that align with the project's technical stack

## Collaboration

Work seamlessly with other project agents:
%s

## Quality Standards

- Write clean, maintainable, and well-documented code
- Ensure comprehensive test coverage for new features
- Follow security best practices specific to %s
- Optimize for performance while maintaining readability
- Document complex logic and architectural decisions

Remember: You are specifically tuned for this project's needs. Always consider the project's context, existing patterns, and technical constraints when providing solutions.
`

	// Generate content based on agent type
	name := agentType
	description := GetAgentDescription(agentType)
	created := time.Now().Format("2006-01-02")
	projectName := filepath.Base(ag.projectPath)
	language := chars.MainLanguage
	frameworks := formatFrameworks(chars.Frameworks)
	frameworksList := strings.Join(chars.Frameworks, ", ")
	if frameworksList == "" {
		frameworksList = "none"
	}

	tag := strings.Split(agentType, "-")[0] // Extract main tag (e.g., "go" from "go-backend-specialist")

	responsibilities := ag.getResponsibilities(agentType, chars)
	expertise := ag.getTechnicalExpertise(agentType, chars)
	bestPractices := ag.getBestPractices(agentType, chars)
	collaborators := ag.getCollaborators(agentType, chars)

	// Get activation keywords
	primaryKeywords := ag.getPrimaryKeywords(agentType, chars)
	secondaryKeywords := ag.getSecondaryKeywords(agentType, chars)
	contextualKeywords := ag.getContextualKeywords(agentType, chars)

	// Get slash commands for this agent
	slashCommands := ag.getSlashCommands(agentType, chars)

	return fmt.Sprintf(template,
		name,
		description,
		created,
		projectName,
		language,
		frameworks,
		emoji,
		bgColor,
		textColor,
		tag,
		primaryKeywords,
		secondaryKeywords,
		contextualKeywords,
		slashCommands,
		emoji,
		name,
		description,
		projectName,
		ag.getExpertiseSummary(agentType, chars),
		emoji,
		language,
		frameworksList,
		ag.projectPath,
		responsibilities,
		expertise,
		bestPractices,
		collaborators,
		language,
	)
}

// Helper methods for generating specific sections

func (ag *AgentGenerator) getResponsibilities(agentType string, chars *ProjectCharacteristics) string {
	switch {
	case strings.Contains(agentType, "backend"):
		return `### Backend Development
- Design and implement robust server-side logic
- Create and maintain RESTful APIs or GraphQL endpoints
- Manage database connections and queries
- Implement authentication and authorization
- Ensure data validation and security
- Optimize performance and scalability`

	case strings.Contains(agentType, "frontend"):
		return `### Frontend Development
- Build responsive and accessible user interfaces
- Implement state management solutions
- Optimize client-side performance
- Ensure cross-browser compatibility
- Create reusable component libraries
- Implement user experience best practices`

	case strings.Contains(agentType, "devops"):
		return `### DevOps & Infrastructure
- Manage containerization and orchestration
- Implement CI/CD pipelines
- Monitor system performance and reliability
- Manage infrastructure as code
- Ensure security and compliance
- Optimize deployment processes`

	case strings.Contains(agentType, "database"):
		return `### Database Management
- Design efficient database schemas
- Optimize query performance
- Implement data migrations
- Ensure data integrity and consistency
- Manage database backups and recovery
- Implement caching strategies`

	case strings.Contains(agentType, "qa"):
		return `### Quality Assurance
- Design comprehensive test strategies
- Implement unit, integration, and E2E tests
- Ensure code coverage targets are met
- Identify and document bugs
- Implement automated testing workflows
- Maintain testing documentation`

	case strings.Contains(agentType, "api"):
		return `### API Design & Development
- Design RESTful or GraphQL APIs
- Implement API versioning strategies
- Ensure API security and authentication
- Create comprehensive API documentation
- Implement rate limiting and caching
- Monitor API performance and usage`

	default:
		return `### General Development
- Implement features according to specifications
- Maintain code quality and standards
- Collaborate with team members
- Document implementations
- Ensure testing coverage
- Optimize performance`
	}
}

func (ag *AgentGenerator) getTechnicalExpertise(agentType string, chars *ProjectCharacteristics) string {
	base := "### Core Technologies\n"

	// Language-specific expertise
	switch chars.MainLanguage {
	case "go":
		base += "- Go idioms, concurrency patterns, and error handling\n"
		base += "- Go modules and dependency management\n"
		base += "- Performance optimization and profiling\n"
	case "javascript", "typescript":
		base += "- Modern JavaScript/TypeScript features and patterns\n"
		base += "- Async/await and Promise handling\n"
		base += "- NPM/Yarn package management\n"
	case "python":
		base += "- Python best practices and PEP standards\n"
		base += "- Virtual environments and dependency management\n"
		base += "- Type hints and modern Python features\n"
	case "rust":
		base += "- Rust ownership and borrowing concepts\n"
		base += "- Memory safety and performance optimization\n"
		base += "- Cargo and crate management\n"
	case "java":
		base += "- Java design patterns and best practices\n"
		base += "- JVM optimization and memory management\n"
		base += "- Maven/Gradle build systems\n"
	}

	// Framework-specific expertise
	if len(chars.Frameworks) > 0 {
		base += "\n### Framework Expertise\n"
		for _, framework := range chars.Frameworks {
			switch framework {
			case "react":
				base += "- React hooks, context, and component patterns\n"
				base += "- State management (Redux, MobX, Zustand)\n"
				base += "- React performance optimization\n"
			case "vue":
				base += "- Vue 3 Composition API and reactivity\n"
				base += "- Vuex/Pinia state management\n"
				base += "- Vue component patterns\n"
			case "angular":
				base += "- Angular modules, services, and dependency injection\n"
				base += "- RxJS and reactive programming\n"
				base += "- Angular performance optimization\n"
			case "django":
				base += "- Django ORM and database migrations\n"
				base += "- Django REST framework\n"
				base += "- Django security best practices\n"
			}
		}
	}

	return base
}

func (ag *AgentGenerator) getBestPractices(agentType string, chars *ProjectCharacteristics) string {
	practices := []string{}

	// General best practices
	practices = append(practices,
		"- Follow SOLID principles and clean code practices",
		"- Write self-documenting code with clear naming",
		"- Implement comprehensive error handling",
		"- Maintain consistent code formatting",
	)

	// Type-specific practices
	if strings.Contains(agentType, "backend") {
		practices = append(practices,
			"- Implement proper input validation and sanitization",
			"- Use environment variables for configuration",
			"- Implement proper logging and monitoring",
		)
	}

	if strings.Contains(agentType, "frontend") {
		practices = append(practices,
			"- Ensure responsive design across devices",
			"- Implement proper accessibility (WCAG compliance)",
			"- Optimize bundle size and loading performance",
		)
	}

	if strings.Contains(agentType, "devops") {
		practices = append(practices,
			"- Use infrastructure as code principles",
			"- Implement proper secret management",
			"- Ensure high availability and disaster recovery",
		)
	}

	return strings.Join(practices, "\n")
}

func (ag *AgentGenerator) getCollaborators(agentType string, chars *ProjectCharacteristics) string {
	collaborators := []string{}

	for _, agent := range chars.DetectedAgents {
		if agent != agentType {
			collaborators = append(collaborators, fmt.Sprintf("- **%s**: %s", agent, GetAgentDescription(agent)))
		}
	}

	if len(collaborators) == 0 {
		return "- Work independently on this project"
	}

	return strings.Join(collaborators, "\n")
}

func (ag *AgentGenerator) getExpertiseSummary(agentType string, chars *ProjectCharacteristics) string {
	parts := []string{}

	if chars.MainLanguage != "" {
		parts = append(parts, chars.MainLanguage)
	}

	if len(chars.Frameworks) > 0 {
		parts = append(parts, strings.Join(chars.Frameworks, ", "))
	}

	if strings.Contains(agentType, "backend") {
		parts = append(parts, "backend development")
	} else if strings.Contains(agentType, "frontend") {
		parts = append(parts, "frontend development")
	} else if strings.Contains(agentType, "devops") {
		parts = append(parts, "DevOps and infrastructure")
	}

	if len(parts) == 0 {
		return "software development"
	}

	return strings.Join(parts, " ")
}

// Helper function to format frameworks for YAML
func formatFrameworks(frameworks []string) string {
	if len(frameworks) == 0 {
		return "[]"
	}

	formatted := []string{}
	for _, f := range frameworks {
		formatted = append(formatted, fmt.Sprintf(`"%s"`, f))
	}

	return "[" + strings.Join(formatted, ", ") + "]"
}

// getVisualIdentity returns emoji, background color, and text color for an agent type
func (ag *AgentGenerator) getVisualIdentity(agentType string) (string, string, string) {
	// Define visual identities for different agent types
	switch {
	case strings.Contains(agentType, "backend"):
		return "⚙️", "#059669", "#f3f4f6"
	case strings.Contains(agentType, "frontend"):
		return "🎨", "#7c3aed", "#fde68a"
	case strings.Contains(agentType, "cli"):
		return "💻", "#0891b2", "#fef3c7"
	case strings.Contains(agentType, "installer"):
		return "📦", "#dc2626", "#fef3c7"
	case strings.Contains(agentType, "orchestrator"):
		return "🎭", "#581c87", "#fde68a"
	case strings.Contains(agentType, "api"):
		return "🔌", "#2563eb", "#fef3c7"
	case strings.Contains(agentType, "database"):
		return "🗄️", "#7c2d12", "#fef3c7"
	case strings.Contains(agentType, "qa") || strings.Contains(agentType, "test"):
		return "🎯", "#10b981", "#f9fafb"
	case strings.Contains(agentType, "devops") || strings.Contains(agentType, "infrastructure"):
		return "🚀", "#0891b2", "#fef3c7"
	case strings.Contains(agentType, "security"):
		return "🛡️", "#dc2626", "#fef3c7"
	default:
		// Default identity for unknown types
		return "🤖", "#6b7280", "#f9fafb"
	}
}

// getPrimaryKeywords returns primary activation keywords for an agent type
func (ag *AgentGenerator) getPrimaryKeywords(agentType string, chars *ProjectCharacteristics) string {
	keywords := []string{}

	// Extract base type from agent name
	parts := strings.Split(agentType, "-")
	if len(parts) > 0 {
		keywords = append(keywords, fmt.Sprintf(`"%s"`, parts[0]))
	}

	// Add type-specific keywords
	switch {
	case strings.Contains(agentType, "backend"):
		keywords = append(keywords, `"api"`, `"server"`, `"service"`)
	case strings.Contains(agentType, "frontend"):
		keywords = append(keywords, `"ui"`, `"component"`, `"interface"`)
	case strings.Contains(agentType, "cli"):
		keywords = append(keywords, `"command"`, `"terminal"`, `"console"`)
	case strings.Contains(agentType, "installer"):
		keywords = append(keywords, `"install"`, `"setup"`, `"deploy"`)
	case strings.Contains(agentType, "orchestrator"):
		keywords = append(keywords, `"coordinate"`, `"workflow"`, `"automate"`)
	}

	// Add language if relevant
	if chars.MainLanguage != "" && strings.Contains(agentType, strings.ToLower(chars.MainLanguage)) {
		keywords = append(keywords, fmt.Sprintf(`"%s"`, strings.ToLower(chars.MainLanguage)))
	}

	return "[" + strings.Join(keywords, ", ") + "]"
}

// getSecondaryKeywords returns secondary activation keywords
func (ag *AgentGenerator) getSecondaryKeywords(agentType string, chars *ProjectCharacteristics) string {
	keywords := []string{}

	// Add framework-specific keywords
	for _, framework := range chars.Frameworks {
		keywords = append(keywords, fmt.Sprintf(`"%s"`, strings.ToLower(framework)))
	}

	// Add technology-specific keywords
	switch {
	case strings.Contains(agentType, "backend"):
		keywords = append(keywords, `"rest"`, `"graphql"`, `"middleware"`)
	case strings.Contains(agentType, "frontend"):
		keywords = append(keywords, `"react"`, `"vue"`, `"angular"`, `"css"`)
	case strings.Contains(agentType, "cli"):
		keywords = append(keywords, `"cobra"`, `"flag"`, `"args"`)
	case strings.Contains(agentType, "installer"):
		keywords = append(keywords, `"package"`, `"dependency"`, `"configuration"`)
	}

	if len(keywords) == 0 {
		return "[]"
	}

	return "[" + strings.Join(keywords, ", ") + "]"
}

// getContextualKeywords returns contextual activation keywords
func (ag *AgentGenerator) getContextualKeywords(agentType string, chars *ProjectCharacteristics) string {
	keywords := []string{}

	// Add action-based keywords
	switch {
	case strings.Contains(agentType, "backend"):
		keywords = append(keywords, `"build api"`, `"create service"`, `"implement endpoint"`)
	case strings.Contains(agentType, "frontend"):
		keywords = append(keywords, `"build ui"`, `"create component"`, `"style interface"`)
	case strings.Contains(agentType, "cli"):
		keywords = append(keywords, `"add command"`, `"parse arguments"`, `"handle flags"`)
	case strings.Contains(agentType, "installer"):
		keywords = append(keywords, `"setup project"`, `"install dependencies"`, `"configure system"`)
	case strings.Contains(agentType, "orchestrator"):
		keywords = append(keywords, `"coordinate agents"`, `"manage workflow"`, `"automate tasks"`)
	}

	if len(keywords) == 0 {
		return "[]"
	}

	return "[" + strings.Join(keywords, ", ") + "]"
}

// getSlashCommands returns slash commands that promote this agent
func (ag *AgentGenerator) getSlashCommands(agentType string, chars *ProjectCharacteristics) string {
	commands := []string{}

	// Generate slash commands based on agent type
	switch {
	case strings.Contains(agentType, "go-backend"):
		commands = append(commands,
			`  - name: "/crew:go"`,
			`    description: "Go development tasks - builds, tests, and optimizations"`,
			`    promotes: ["self", "qa-persona"]`,
			`  - name: "/crew:backend"`,
			`    description: "Backend development - APIs, services, and data processing"`,
			`    promotes: ["self", "api-specialist"]`,
			`  - name: "/crew:goroutine"`,
			`    description: "Concurrent programming with goroutines and channels"`,
			`    promotes: ["self", "performance-persona"]`,
		)
	case strings.Contains(agentType, "api"):
		commands = append(commands,
			`  - name: "/crew:endpoint"`,
			`    description: "Create or modify API endpoints"`,
			`    promotes: ["self", "backend-persona"]`,
			`  - name: "/crew:rest"`,
			`    description: "RESTful API design and implementation"`,
			`    promotes: ["self", "architect-persona"]`,
		)
	case strings.Contains(agentType, "cli"):
		commands = append(commands,
			`  - name: "/crew:command"`,
			`    description: "Add or modify CLI commands"`,
			`    promotes: ["self", "frontend-persona"]`,
			`  - name: "/crew:flag"`,
			`    description: "Implement command flags and parsing"`,
			`    promotes: ["self"]`,
		)
	case strings.Contains(agentType, "installer"):
		commands = append(commands,
			`  - name: "/crew:install"`,
			`    description: "Installation and setup procedures"`,
			`    promotes: ["self", "devops-persona"]`,
			`  - name: "/crew:setup"`,
			`    description: "Project setup and configuration"`,
			`    promotes: ["self", "scribe-persona"]`,
		)
	case strings.Contains(agentType, "orchestrator"):
		commands = append(commands,
			`  - name: "/crew:orchestrate"`,
			`    description: "Analyze task and delegate to appropriate agents"`,
			`    promotes: ["all-agents"]`,
			`  - name: "/crew:agent-help"`,
			`    description: "Show available agents and their specialties"`,
			`    promotes: ["self"]`,
		)
	default:
		// Default commands for unknown types
		baseName := strings.Split(agentType, "-")[0]
		commands = append(commands,
			fmt.Sprintf(`  - name: "/crew:%s"`, baseName),
			fmt.Sprintf(`    description: "%s-specific tasks and operations"`, baseName),
			`    promotes: ["self"]`,
		)
	}

	if len(commands) == 0 {
		return "[]"
	}

	return "[\n" + strings.Join(commands, "\n") + "\n]"
}
