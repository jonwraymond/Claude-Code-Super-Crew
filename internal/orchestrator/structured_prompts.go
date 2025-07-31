package orchestrator

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// StructuredPrompt represents a comprehensive prompt with full context
type StructuredPrompt struct {
	Type           string                 `json:"type"`           // agent-creation, analysis, implementation
	Target         string                 `json:"target"`         // global or local agent
	Context        ProjectContext         `json:"context"`        
	Task           TaskSpecification      `json:"task"`
	Resources      AvailableResources     `json:"resources"`
	Constraints    []string               `json:"constraints"`
	SuccessCriteria []string              `json:"success_criteria"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// ProjectContext contains comprehensive project information
type ProjectContext struct {
	Summary        string            `json:"summary"`
	Languages      []string          `json:"languages"`
	Frameworks     []string          `json:"frameworks"`
	Patterns       []string          `json:"patterns"`
	PainPoints     []string          `json:"pain_points"`
	CodeStructure  string            `json:"code_structure"`    // from code2prompt
	SymbolOverview string            `json:"symbol_overview"`   // from Serena
	Dependencies   map[string]string `json:"dependencies"`
}

// TaskSpecification defines what needs to be done
type TaskSpecification struct {
	Goal           string   `json:"goal"`
	SubTasks       []string `json:"subtasks"`
	Priority       string   `json:"priority"`
	Complexity     string   `json:"complexity"`
	EstimatedTime  string   `json:"estimated_time"`
}

// AvailableResources lists what the agent can use
type AvailableResources struct {
	MCPServers    []MCPServer    `json:"mcp_servers"`
	CLITools      []CLITool      `json:"cli_tools"`
	ExistingAgents []string      `json:"existing_agents"`
	ProjectAssets  []string      `json:"project_assets"`
}

// MCPServer represents an available MCP server
type MCPServer struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
	Purpose string `json:"purpose"`
}

// PromptBuilder creates structured prompts for agents
type PromptBuilder struct {
	projectRoot string
	analyzer    *ProjectAnalyzer
}

// NewPromptBuilder creates a new prompt builder
func NewPromptBuilder(projectRoot string) *PromptBuilder {
	return &PromptBuilder{
		projectRoot: projectRoot,
		analyzer:    NewProjectAnalyzer(projectRoot),
	}
}

// BuildAgentCreationPrompt creates a prompt for generating a new agent
func (pb *PromptBuilder) BuildAgentCreationPrompt(agentType, purpose string, isGlobal bool) (*StructuredPrompt, error) {
	// Gather comprehensive context
	context, err := pb.gatherProjectContext()
	if err != nil {
		return nil, fmt.Errorf("failed to gather context: %w", err)
	}

	targetType := "local"
	if isGlobal {
		targetType = "global"
	}

	prompt := &StructuredPrompt{
		Type:   "agent-creation",
		Target: targetType,
		Context: *context,
		Task: TaskSpecification{
			Goal: fmt.Sprintf("Create a %s specialist agent for %s", agentType, purpose),
			SubTasks: []string{
				"Analyze the specific patterns in the codebase",
				"Design agent capabilities that solve real problems",
				"Create clear, actionable agent instructions",
				"Include specific tool recommendations",
				"Add example workflows and use cases",
			},
			Priority:   "high",
			Complexity: "medium",
			EstimatedTime: "15 minutes",
		},
		Resources: pb.getAvailableResources(),
		Constraints: []string{
			"Agent must solve specific problems in this codebase",
			"Focus on practical, day-to-day development tasks",
			"Include concrete examples from the actual code",
			"Make the agent genuinely useful, not generic",
		},
		SuccessCriteria: []string{
			"Agent has clear, specific capabilities",
			"Examples reference actual code patterns",
			"Workflows would save real development time",
			"Integration with other agents is defined",
		},
		Metadata: map[string]interface{}{
			"created_at": time.Now().Format(time.RFC3339),
			"framework_version": "2.0",
		},
	}

	return prompt, nil
}

// gatherProjectContext uses tools to create comprehensive context
func (pb *PromptBuilder) gatherProjectContext() (*ProjectContext, error) {
	context := &ProjectContext{
		PainPoints: []string{},
		Patterns:   []string{},
	}

	// Try to use code2prompt if available
	if pb.isToolAvailable("code2prompt") {
		codeContext, err := pb.runCode2Prompt()
		if err == nil {
			context.CodeStructure = codeContext
		}
	}

	// Use basic analysis as fallback
	characteristics, err := pb.analyzer.Analyze()
	if err == nil {
		context.Summary = pb.generateProjectSummary(characteristics)
		context.Languages = []string{characteristics.MainLanguage}
		context.Frameworks = characteristics.Frameworks
		
		// Extract patterns from detected agents
		context.Patterns = characteristics.DetectedAgents
	}

	// Add known pain points based on patterns
	context.PainPoints = pb.identifyPainPoints(context.Patterns)

	return context, nil
}

// runCode2Prompt executes code2prompt to get comprehensive context
func (pb *PromptBuilder) runCode2Prompt() (string, error) {
	// Build appropriate command based on project type
	ext := pb.getMainExtension()
	cmd := exec.Command("code2prompt",
		"--include", fmt.Sprintf("**/*.%s", ext),
		"--exclude", "**/vendor/**",
		"--exclude", "**/node_modules/**",
		"--max-tokens", "20000",
	)
	cmd.Dir = pb.projectRoot

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("code2prompt failed: %w", err)
	}

	return string(output), nil
}

// getMainExtension determines the primary file extension
func (pb *PromptBuilder) getMainExtension() string {
	// Check for common project files
	if _, err := os.Stat(filepath.Join(pb.projectRoot, "go.mod")); err == nil {
		return "go"
	}
	if _, err := os.Stat(filepath.Join(pb.projectRoot, "package.json")); err == nil {
		return "js,ts,jsx,tsx"
	}
	if _, err := os.Stat(filepath.Join(pb.projectRoot, "requirements.txt")); err == nil {
		return "py"
	}
	return "*" // fallback to all files
}

// identifyPainPoints maps patterns to common pain points
func (pb *PromptBuilder) identifyPainPoints(patterns []string) []string {
	painPoints := []string{}
	
	for _, pattern := range patterns {
		switch pattern {
		case "cli":
			painPoints = append(painPoints, 
				"Maintaining consistent command structure",
				"Complex flag validation and handling",
				"Help text and documentation sync",
			)
		case "api":
			painPoints = append(painPoints,
				"Repetitive endpoint boilerplate",
				"Consistent error handling",
				"Request/response validation",
			)
		case "testing":
			painPoints = append(painPoints,
				"Test coverage gaps",
				"Mock generation complexity",
				"Test data management",
			)
		case "database":
			painPoints = append(painPoints,
				"Migration management",
				"Query optimization",
				"Connection pooling issues",
			)
		}
	}
	
	return painPoints
}

// getAvailableResources lists all available resources
func (pb *PromptBuilder) getAvailableResources() AvailableResources {
	return AvailableResources{
		MCPServers: []MCPServer{
			{Name: "context7", Enabled: true, Purpose: "Documentation lookup"},
			{Name: "sequential", Enabled: true, Purpose: "Complex analysis"},
			{Name: "serena", Enabled: true, Purpose: "Semantic code understanding"},
			{Name: "magic", Enabled: false, Purpose: "UI components"},
			{Name: "playwright", Enabled: false, Purpose: "Browser testing"},
		},
		CLITools: []CLITool{
			{Name: "code2prompt", Command: "code2prompt", Description: "Generate code context"},
			{Name: "ast-grep", Command: "ast-grep", Description: "Semantic search"},
		},
		ExistingAgents: pb.listExistingAgents(),
		ProjectAssets: []string{
			"project-analysis.json",
			"CLAUDE.md", 
			"orchestrator-specialist.md",
		},
	}
}

// listExistingAgents finds all agents in the project
func (pb *PromptBuilder) listExistingAgents() []string {
	agents := []string{}
	
	// Check local agents
	localAgentsDir := filepath.Join(pb.projectRoot, ".claude", "agents")
	if entries, err := os.ReadDir(localAgentsDir); err == nil {
		for _, entry := range entries {
			if strings.HasSuffix(entry.Name(), ".md") && entry.Name() != "project-analysis.json" {
				agents = append(agents, entry.Name())
			}
		}
	}
	
	// Add known global agents
	agents = append(agents,
		"architect-persona",
		"frontend-persona", 
		"backend-persona",
		"analyzer-persona",
		"security-persona",
		"qa-persona",
		"performance-persona",
		"refactorer-persona",
		"devops-persona",
		"mentor-persona",
		"scribe-persona",
	)
	
	return agents
}

// isToolAvailable checks if a CLI tool is installed
func (pb *PromptBuilder) isToolAvailable(tool string) bool {
	_, err := exec.LookPath(tool)
	return err == nil
}

// generateProjectSummary creates a concise project summary
func (pb *PromptBuilder) generateProjectSummary(characteristics *ProjectCharacteristics) string {
	frameworks := "no specific frameworks"
	if len(characteristics.Frameworks) > 0 {
		frameworks = strings.Join(characteristics.Frameworks, ", ")
	}
	
	patterns := "no specific patterns"
	if len(characteristics.DetectedAgents) > 0 {
		patterns = strings.Join(characteristics.DetectedAgents, ", ")
	}
	
	return fmt.Sprintf(
		"A %s project using %s. Detected patterns: %s",
		characteristics.MainLanguage,
		frameworks,
		patterns,
	)
}

// CreateAgentPrompt generates a prompt for agent creation with full context
func (pb *PromptBuilder) CreateAgentPrompt(agentName, purpose string) (string, error) {
	structured, err := pb.BuildAgentCreationPrompt(agentName, purpose, false)
	if err != nil {
		return "", err
	}
	
	// Convert to readable prompt
	contextJSON, _ := json.MarshalIndent(structured, "", "  ")
	
	prompt := fmt.Sprintf(`# Agent Creation Request

## Structured Context
%s

## Your Task
Based on the comprehensive context above, create a specialist agent named "%s" that:

1. Solves specific problems identified in the pain points
2. Uses the available tools and MCP servers effectively  
3. Integrates well with existing agents
4. Provides concrete value for day-to-day development

Remember to:
- Reference actual code patterns from the context
- Create practical workflows that save time
- Make the agent specific to this project, not generic
- Include example commands and use cases

The agent should be genuinely useful and address real development needs!`,
		string(contextJSON),
		agentName,
	)
	
	return prompt, nil
}