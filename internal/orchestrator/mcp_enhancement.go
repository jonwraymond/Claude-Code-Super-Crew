package orchestrator

import (
	"fmt"
	"strings"
)

// MCPIntegration represents an MCP server integration
type MCPIntegration struct {
	Name        string
	Description string
	Tools       []string
	UsageGuide  string
	Examples    []MCPExample
	Enabled     bool
}

// MCPExample represents a usage example for an MCP integration
type MCPExample struct {
	Description string
	Code        string
}

// MCPEnhancer handles the enhancement of CLAUDE.md with MCP integrations
type MCPEnhancer struct {
	integrations map[string]MCPIntegration
}

// NewMCPEnhancer creates a new MCP enhancer
func NewMCPEnhancer() *MCPEnhancer {
	return &MCPEnhancer{
		integrations: initializeMCPIntegrations(),
	}
}

// initializeMCPIntegrations sets up all available MCP integrations
func initializeMCPIntegrations() map[string]MCPIntegration {
	return map[string]MCPIntegration{
		"context7": {
			Name:        "Context7",
			Description: "Official library documentation, code examples, best practices",
			Tools: []string{
				"mcp__context7__resolve-library-id",
				"mcp__context7__get-library-docs",
			},
			UsageGuide: `## Context7 Integration

Context7 provides access to up-to-date documentation for any library or framework. Use it for:
- Finding official documentation and examples
- Understanding best practices and patterns
- Resolving version-specific implementations
- Getting framework-specific conventions`,
			Examples: []MCPExample{
				{
					Description: "Get React hooks documentation",
					Code: `# First resolve the library ID
mcp__context7__resolve-library-id:
  libraryName: "react"

# Then get specific documentation
mcp__context7__get-library-docs:
  context7CompatibleLibraryID: "/facebook/react"
  topic: "hooks"
  tokens: 10000`,
				},
				{
					Description: "Find Next.js routing patterns",
					Code: `mcp__context7__get-library-docs:
  context7CompatibleLibraryID: "/vercel/next.js"
  topic: "routing"`,
				},
			},
			Enabled: true,
		},
		"sequential": {
			Name:        "Sequential Thinking",
			Description: "Multi-step problem solving, architectural analysis, systematic debugging",
			Tools: []string{
				"mcp__sequential-thinking__sequentialthinking",
			},
			UsageGuide: `## Sequential Thinking Integration

Sequential provides structured, multi-step analysis for complex problems. Use it for:
- Breaking down complex problems into manageable steps
- Architectural decision-making with revision capability
- Root cause analysis with hypothesis testing
- Planning implementations with adaptive thinking`,
			Examples: []MCPExample{
				{
					Description: "Analyze a complex bug",
					Code: `mcp__sequential-thinking__sequentialthinking:
  thought: "First, let me understand the error pattern..."
  nextThoughtNeeded: true
  thoughtNumber: 1
  totalThoughts: 5
  isRevision: false`,
				},
				{
					Description: "Plan a feature implementation",
					Code: `mcp__sequential-thinking__sequentialthinking:
  thought: "Breaking down the feature requirements..."
  nextThoughtNeeded: true
  thoughtNumber: 1
  totalThoughts: 8
  needsMoreThoughts: true`,
				},
			},
			Enabled: true,
		},
		"magic": {
			Name:        "Magic UI Components",
			Description: "Modern UI component generation, design system integration",
			Tools: []string{
				"mcp__magic__21st_magic_component_builder",
				"mcp__magic__21st_magic_component_inspiration",
				"mcp__magic__21st_magic_component_refiner",
				"mcp__magic__logo_search",
			},
			UsageGuide: `## Magic UI Integration

Magic provides AI-powered UI component generation and refinement. Use it for:
- Creating modern, accessible UI components
- Finding design inspiration from 21st.dev
- Refining existing components for better UX
- Adding company logos to projects`,
			Examples: []MCPExample{
				{
					Description: "Create a dashboard component",
					Code: `mcp__magic__21st_magic_component_builder:
  message: "Create a modern analytics dashboard"
  searchQuery: "dashboard analytics"
  absolutePathToCurrentFile: "/src/components/Dashboard.tsx"
  absolutePathToProjectDirectory: "/path/to/project"
  standaloneRequestQuery: "Analytics dashboard with charts and metrics"`,
				},
				{
					Description: "Add company logos",
					Code: `mcp__magic__logo_search:
  queries: ["github", "slack", "discord"]
  format: "TSX"`,
				},
			},
			Enabled: true,
		},
		"playwright": {
			Name:        "Playwright Testing",
			Description: "Browser automation, E2E testing, performance monitoring",
			Tools: []string{
				"playwright_navigate",
				"playwright_screenshot",
				"playwright_click",
				"playwright_fill",
			},
			UsageGuide: `## Playwright Integration

Playwright enables browser automation and testing. Use it for:
- End-to-end testing of web applications
- Visual regression testing with screenshots
- Performance monitoring and metrics
- Cross-browser compatibility testing`,
			Examples: []MCPExample{
				{
					Description: "Test a login flow",
					Code: `# Navigate to login page
playwright_navigate:
  url: "https://app.example.com/login"

# Fill in credentials
playwright_fill:
  selector: "#email"
  value: "test@example.com"

# Take screenshot for visual validation
playwright_screenshot:
  name: "login-page"`,
				},
			},
			Enabled: false, // Disabled by default
		},
		"serena": {
			Name:        "Serena Code Intelligence",
			Description: "Powerful coding agent toolkit providing semantic code retrieval and editing",
			Tools: []string{
				"mcp__serena__get_symbols_overview",
				"mcp__serena__find_symbol",
				"mcp__serena__find_referencing_symbols",
				"mcp__serena__find_referencing_code_snippets",
				"mcp__serena__search_for_pattern",
				"mcp__serena__read_file",
				"mcp__serena__create_text_file",
				"mcp__serena__replace_symbol_body",
				"mcp__serena__insert_before_symbol",
				"mcp__serena__insert_after_symbol",
				"mcp__serena__replace_lines",
				"mcp__serena__delete_lines",
			},
			UsageGuide: `## Serena Code Intelligence Integration

Serena provides powerful semantic code understanding and editing capabilities. Use it for:
- AST-aware code navigation and understanding
- Finding all usages and references of symbols
- Semantic code modifications (replace symbol bodies, insert around symbols)
- Project-wide pattern searching
- Intelligent code refactoring with context preservation

Key advantages over basic tools:
- Understands code structure, not just text
- Can modify code at the symbol level
- Maintains semantic correctness during edits
- Works across multiple programming languages`,
			Examples: []MCPExample{
				{
					Description: "Find and understand code structure",
					Code: `# Get overview of symbols in a module
mcp__serena__get_symbols_overview:
  relative_path: "pkg/your_package"

# Find a specific symbol globally
mcp__serena__find_symbol:
  name_path: "YourFunction"
  search_type: "global"

# Find all references to a symbol
mcp__serena__find_referencing_symbols:
  name_path: "YourFunction"
  relative_path: "pkg/your_package"

# Find code snippets that reference a symbol
mcp__serena__find_referencing_code_snippets:
  name_path: "YourFunction"
  relative_path: "pkg/your_package"`,
				},
				{
					Description: "Semantic code modifications",
					Code: `# Replace entire function body
mcp__serena__replace_symbol_body:
  name_path: "YourFunction"
  relative_path: "pkg/your_package"
  new_body: |
    func YourFunction(param string) error {
        // New implementation
        return nil
    }

# Insert code before a symbol
mcp__serena__insert_before_symbol:
  name_path: "YourStruct"
  relative_path: "pkg/your_package"
  content: |
    // NewFunction performs important operation
    func NewFunction() error {
        return nil
    }

# Insert code after a symbol
mcp__serena__insert_after_symbol:
  name_path: "YourStruct"
  relative_path: "pkg/your_package"
  content: |
    // Helper method for YourStruct
    func (s *YourStruct) Helper() string {
        return s.Name
    }`,
				},
				{
					Description: "Search for patterns across codebase",
					Code: `# Search for error handling patterns
mcp__serena__search_for_pattern:
  substring_pattern: "if err != nil"
  restrict_search_to_code_files: true
  relative_path: "."

# Search for TODO comments
mcp__serena__search_for_pattern:
  substring_pattern: "TODO"
  restrict_search_to_code_files: true`,
				},
			},
			Enabled: true, // Official tool in our framework
		},
		"astgrep": {
			Name:        "AST-GREP",
			Description: "Pattern-based semantic code search",
			Tools: []string{
				"ast-grep",
			},
			UsageGuide: `## AST-GREP Integration

AST-GREP provides pattern-based semantic searching. Use it for:
- Finding structural code patterns
- Identifying similar implementations
- Refactoring assistance
- Code quality checks`,
			Examples: []MCPExample{
				{
					Description: "Find error handling patterns",
					Code: `ast-grep --pattern 'if err != nil { return err }'`,
				},
				{
					Description: "Find all struct definitions",
					Code: `ast-grep --pattern 'type $NAME struct { $$$ }'`,
				},
			},
			Enabled: false, // Optional tool
		},
	}
}

// GenerateEnhancedSection generates the MCP section for CLAUDE.md
func (e *MCPEnhancer) GenerateEnhancedSection(enabledServers []string, projectType string) string {
	var sections []string
	
	// Header
	sections = append(sections, "## 🔧 MCP Server Integrations\n")
	sections = append(sections, "This project has the following MCP (Model Context Protocol) servers available:")
	sections = append(sections, "")
	
	// Enable specified servers
	for _, server := range enabledServers {
		if integration, exists := e.integrations[strings.ToLower(server)]; exists {
			integration.Enabled = true
			e.integrations[strings.ToLower(server)] = integration
		}
	}
	
	// Add project-specific recommendations
	recommendations := e.getProjectRecommendations(projectType)
	if len(recommendations) > 0 {
		sections = append(sections, "### Recommended for this project")
		for _, rec := range recommendations {
			sections = append(sections, fmt.Sprintf("- **%s**: %s", rec.Name, rec.Description))
		}
		sections = append(sections, "")
	}
	
	// Add enabled integrations
	for _, integration := range e.integrations {
		if integration.Enabled {
			sections = append(sections, e.formatIntegration(integration))
		}
	}
	
	// Add workflow examples
	sections = append(sections, e.generateWorkflowExamples(projectType))
	
	return strings.Join(sections, "\n")
}

// formatIntegration formats a single MCP integration
func (e *MCPEnhancer) formatIntegration(integration MCPIntegration) string {
	var parts []string
	
	parts = append(parts, fmt.Sprintf("### %s", integration.Name))
	parts = append(parts, "")
	parts = append(parts, integration.UsageGuide)
	parts = append(parts, "")
	parts = append(parts, "#### Available Tools:")
	for _, tool := range integration.Tools {
		parts = append(parts, fmt.Sprintf("- `%s`", tool))
	}
	parts = append(parts, "")
	parts = append(parts, "#### Examples:")
	for _, example := range integration.Examples {
		parts = append(parts, fmt.Sprintf("\n**%s:**", example.Description))
		parts = append(parts, "```yaml")
		parts = append(parts, example.Code)
		parts = append(parts, "```")
	}
	parts = append(parts, "")
	
	return strings.Join(parts, "\n")
}

// getProjectRecommendations returns recommended MCP servers based on project type
func (e *MCPEnhancer) getProjectRecommendations(projectType string) []MCPIntegration {
	var recommendations []MCPIntegration
	
	switch projectType {
	case "react", "vue", "angular", "frontend":
		if integration, exists := e.integrations["magic"]; exists {
			recommendations = append(recommendations, integration)
		}
		if integration, exists := e.integrations["context7"]; exists {
			recommendations = append(recommendations, integration)
		}
	case "go", "rust", "backend":
		if integration, exists := e.integrations["serena"]; exists {
			recommendations = append(recommendations, integration)
		}
		if integration, exists := e.integrations["sequential"]; exists {
			recommendations = append(recommendations, integration)
		}
		if integration, exists := e.integrations["context7"]; exists {
			recommendations = append(recommendations, integration)
		}
	case "fullstack":
		if integration, exists := e.integrations["context7"]; exists {
			recommendations = append(recommendations, integration)
		}
		if integration, exists := e.integrations["magic"]; exists {
			recommendations = append(recommendations, integration)
		}
		if integration, exists := e.integrations["sequential"]; exists {
			recommendations = append(recommendations, integration)
		}
	}
	
	return recommendations
}

// generateWorkflowExamples generates project-specific workflow examples
func (e *MCPEnhancer) generateWorkflowExamples(projectType string) string {
	var examples []string
	
	examples = append(examples, "## 📋 Common MCP Workflows\n")
	
	switch projectType {
	case "react", "frontend":
		examples = append(examples, "### Component Development Workflow")
		examples = append(examples, "```yaml")
		examples = append(examples, "1. Get framework documentation:")
		examples = append(examples, "   mcp__context7__get-library-docs:")
		examples = append(examples, "     context7CompatibleLibraryID: \"/facebook/react\"")
		examples = append(examples, "     topic: \"components\"")
		examples = append(examples, "")
		examples = append(examples, "2. Create component with Magic:")
		examples = append(examples, "   mcp__magic__21st_magic_component_builder:")
		examples = append(examples, "     searchQuery: \"form input\"")
		examples = append(examples, "")
		examples = append(examples, "3. Test with Playwright:")
		examples = append(examples, "   playwright_navigate:")
		examples = append(examples, "     url: \"http://localhost:3000\"")
		examples = append(examples, "```")
		
	case "go", "backend":
		examples = append(examples, "### Backend Development Workflow")
		examples = append(examples, "```yaml")
		examples = append(examples, "1. Understand existing code structure:")
		examples = append(examples, "   mcp__serena__get_symbols_overview:")
		examples = append(examples, "     relative_path: \"internal/api\"")
		examples = append(examples, "")
		examples = append(examples, "2. Find similar patterns:")
		examples = append(examples, "   mcp__serena__search_for_pattern:")
		examples = append(examples, "     substring_pattern: \"handler\"")
		examples = append(examples, "     restrict_search_to_code_files: true")
		examples = append(examples, "")
		examples = append(examples, "3. Plan implementation:")
		examples = append(examples, "   mcp__sequential-thinking__sequentialthinking:")
		examples = append(examples, "     thought: \"Design API endpoint structure...\"")
		examples = append(examples, "")
		examples = append(examples, "4. Get framework best practices:")
		examples = append(examples, "   mcp__context7__get-library-docs:")
		examples = append(examples, "     context7CompatibleLibraryID: \"/golang/go\"")
		examples = append(examples, "     topic: \"http handlers\"")
		examples = append(examples, "")
		examples = append(examples, "5. Implement with semantic precision:")
		examples = append(examples, "   mcp__serena__replace_symbol_body:")
		examples = append(examples, "     name_path: \"HandleUserCreate\"")
		examples = append(examples, "     relative_path: \"internal/api/users.go\"")
		examples = append(examples, "```")
	}
	
	examples = append(examples, "")
	return strings.Join(examples, "\n")
}

// EnhanceProjectCLAUDE enhances a project's CLAUDE.md with MCP integrations
func (e *MCPEnhancer) EnhanceProjectCLAUDE(existingContent string, enabledServers []string, projectType string) string {
	// Generate MCP section
	mcpSection := e.GenerateEnhancedSection(enabledServers, projectType)
	
	// If existing content has MCP section, replace it
	if strings.Contains(existingContent, "## 🔧 MCP Server Integrations") {
		// Find and replace existing MCP section
		startIdx := strings.Index(existingContent, "## 🔧 MCP Server Integrations")
		endIdx := strings.Index(existingContent[startIdx:], "\n## ")
		if endIdx == -1 {
			// MCP section is at the end
			return existingContent[:startIdx] + mcpSection
		}
		endIdx += startIdx
		return existingContent[:startIdx] + mcpSection + existingContent[endIdx:]
	}
	
	// Otherwise, append MCP section
	return existingContent + "\n\n" + mcpSection
}