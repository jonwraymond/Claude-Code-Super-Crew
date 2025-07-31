// Package orchestrator provides intelligent command routing and agent orchestration
package orchestrator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jonwraymond/claude-code-super-crew/pkg/logger"
	"gopkg.in/yaml.v3"
)

// LocalCommandRouter handles routing for local project commands
type LocalCommandRouter struct {
	projectPath      string
	globalCommands   map[string]*Command
	localCommands    map[string]*LocalCommand
	shadowCommands   map[string]*ShadowCommand
	specialists      map[string]*Specialist
	orchestratorPath string
	logger           logger.Logger
}

// LocalCommand represents a project-specific command
type LocalCommand struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Shadows     bool              `yaml:"shadows"`
	Routing     RoutingConfig     `yaml:"routing"`
	Tools       []string          `yaml:"tools"`
	Flags       []CommandFlag     `yaml:"flags"`
	Content     string            `yaml:"-"` // Command documentation content
}

// ShadowCommand enhances a global command
type ShadowCommand struct {
	Name        string           `yaml:"name"`
	Description string           `yaml:"description"`
	Shadows     bool             `yaml:"shadows"`
	BaseCommand string           `yaml:"base_command"`
	Routing     ShadowRouting    `yaml:"routing"`
	PreHooks    []string         `yaml:"pre_hooks"`
	PostHooks   []string         `yaml:"post_hooks"`
	Content     string           `yaml:"-"`
}

// RoutingConfig defines how commands route to agents
type RoutingConfig struct {
	Primary     string                 `yaml:"primary"`
	Fallback    string                 `yaml:"fallback"`
	Orchestrate bool                   `yaml:"orchestrate"`
	Workflow    []WorkflowStep         `yaml:"workflow,omitempty"`
	Context     map[string]interface{} `yaml:"context,omitempty"`
}

// ShadowRouting extends base command routing
type ShadowRouting struct {
	Inherit    bool     `yaml:"inherit"`
	Additional []string `yaml:"additional"`
	PreHooks   []string `yaml:"pre_hooks"`
	PostHooks  []string `yaml:"post_hooks"`
}

// WorkflowStep defines a step in multi-agent workflow
type WorkflowStep struct {
	Name   string   `yaml:"name"`
	Agents []string `yaml:"agents"`
}

// CommandFlag represents a command flag
type CommandFlag struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Type        string `yaml:"type,omitempty"`
	Default     string `yaml:"default,omitempty"`
}

// RoutingDecision represents the routing decision
type RoutingDecision struct {
	Command     string
	Type        CommandType
	Agents      []string
	Workflow    []WorkflowStep
	Context     map[string]interface{}
	Explanation string
}

// Command represents a base command structure
type Command struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Category    string   `yaml:"category,omitempty"`
	Tools       []string `yaml:"tools,omitempty"`
}

// Specialist represents a project-specific specialist agent
type Specialist struct {
	Name         string   `yaml:"name"`
	Description  string   `yaml:"description"`
	Domain       string   `yaml:"domain"`
	Expertise    []string `yaml:"expertise"`
	Tools        []string `yaml:"tools"`
	Personas     []string `yaml:"personas,omitempty"`
	FilePath     string   `yaml:"-"`
}

// CommandType represents the type of command
type CommandType string

const (
	LocalCommandType   CommandType = "local"
	ShadowCommandType  CommandType = "shadow"
	GlobalCommandType  CommandType = "global"
	DynamicCommandType CommandType = "dynamic"
)

// NewLocalCommandRouter creates a new router
func NewLocalCommandRouter(projectPath string) *LocalCommandRouter {
	return &LocalCommandRouter{
		projectPath:      projectPath,
		globalCommands:   make(map[string]*Command),
		localCommands:    make(map[string]*LocalCommand),
		shadowCommands:   make(map[string]*ShadowCommand),
		specialists:      make(map[string]*Specialist),
		orchestratorPath: filepath.Join(projectPath, ".claude", "agents", "orchestrator-specialist.md"),
		logger:           logger.GetLogger(),
	}
}

// Initialize loads commands and specialists
func (r *LocalCommandRouter) Initialize() error {
	// First check if local orchestrator exists
	if err := r.ensureLocalOrchestrator(); err != nil {
		return fmt.Errorf("ensuring local orchestrator: %w", err)
	}

	// Load global commands
	if err := r.loadGlobalCommands(); err != nil {
		return fmt.Errorf("loading global commands: %w", err)
	}

	// Load local commands
	if err := r.loadLocalCommands(); err != nil {
		return fmt.Errorf("loading local commands: %w", err)
	}

	// Discover specialists
	if err := r.discoverSpecialists(); err != nil {
		return fmt.Errorf("discovering specialists: %w", err)
	}

	return nil
}

// ensureLocalOrchestrator creates local orchestrator from global template if needed
func (r *LocalCommandRouter) ensureLocalOrchestrator() error {
	// Check if local orchestrator exists
	if _, err := os.Stat(r.orchestratorPath); err == nil {
		r.logger.Debug("Local orchestrator already exists")
		return nil
	}

	r.logger.Info("Creating local orchestrator from global template...")

	// This is where Claude Code would be prompted to create the local orchestrator
	// based on the global template and project analysis
	fmt.Println("\n🎯 Orchestrator Setup Required")
	fmt.Println("The global orchestrator template will guide creation of your project-specific orchestrator.")
	fmt.Println("\nClaude will:")
	fmt.Println("1. Use the global orchestrator as a template")
	fmt.Println("2. Analyze your project structure and patterns")
	fmt.Println("3. Create a customized local orchestrator at .claude/agents/orchestrator-specialist.md")
	fmt.Println("4. Configure it with project-specific routing rules and patterns")
	fmt.Println("\nThis enables intelligent command routing tailored to your project!")

	return nil
}

// loadGlobalCommands loads commands from ~/.claude/commands
func (r *LocalCommandRouter) loadGlobalCommands() error {
	homeDir, _ := os.UserHomeDir()
	globalPath := filepath.Join(homeDir, ".claude", "commands")
	
	if _, err := os.Stat(globalPath); os.IsNotExist(err) {
		return nil // No global commands is OK
	}

	entries, err := os.ReadDir(globalPath)
	if err != nil {
		return fmt.Errorf("reading global commands: %w", err)
	}

	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		cmdName := strings.TrimSuffix(entry.Name(), ".md")
		// For now, just track that these exist
		r.globalCommands[cmdName] = &Command{Name: cmdName}
	}

	return nil
}

// loadLocalCommands loads project-specific commands
func (r *LocalCommandRouter) loadLocalCommands() error {
	localPath := filepath.Join(r.projectPath, ".claude", "commands")
	
	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		return nil // No local commands is OK
	}

	// Load regular local commands
	entries, err := os.ReadDir(localPath)
	if err != nil {
		return fmt.Errorf("reading local commands: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		cmdPath := filepath.Join(localPath, entry.Name())
		cmd, err := r.loadLocalCommand(cmdPath)
		if err != nil {
			r.logger.Warn(fmt.Sprintf("Failed to load command %s: %v", entry.Name(), err))
			continue
		}

		r.localCommands[cmd.Name] = cmd
	}

	// Load shadow commands
	shadowPath := filepath.Join(localPath, "shadows")
	if _, err := os.Stat(shadowPath); err == nil {
		shadowEntries, err := os.ReadDir(shadowPath)
		if err == nil {
			for _, entry := range shadowEntries {
				if !strings.HasSuffix(entry.Name(), ".md") {
					continue
				}

				cmdPath := filepath.Join(shadowPath, entry.Name())
				cmd, err := r.loadShadowCommand(cmdPath)
				if err != nil {
					r.logger.Warn(fmt.Sprintf("Failed to load shadow command %s: %v", entry.Name(), err))
					continue
				}

				r.shadowCommands[cmd.Name] = cmd
			}
		}
	}

	return nil
}

// loadLocalCommand loads a single local command
func (r *LocalCommandRouter) loadLocalCommand(path string) (*LocalCommand, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Parse YAML front matter
	parts := strings.SplitN(string(content), "---", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid command format")
	}

	var cmd LocalCommand
	if err := yaml.Unmarshal([]byte(parts[1]), &cmd); err != nil {
		return nil, fmt.Errorf("parsing command metadata: %w", err)
	}

	cmd.Content = parts[2]
	return &cmd, nil
}

// loadShadowCommand loads a shadow command
func (r *LocalCommandRouter) loadShadowCommand(path string) (*ShadowCommand, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Parse YAML front matter
	parts := strings.SplitN(string(content), "---", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid shadow command format")
	}

	var cmd ShadowCommand
	if err := yaml.Unmarshal([]byte(parts[1]), &cmd); err != nil {
		return nil, fmt.Errorf("parsing shadow command metadata: %w", err)
	}

	cmd.Content = parts[2]
	return &cmd, nil
}

// discoverSpecialists finds all project specialists
func (r *LocalCommandRouter) discoverSpecialists() error {
	agentsPath := filepath.Join(r.projectPath, ".claude", "agents")
	
	if _, err := os.Stat(agentsPath); os.IsNotExist(err) {
		return nil // No agents is OK
	}

	entries, err := os.ReadDir(agentsPath)
	if err != nil {
		return fmt.Errorf("reading agents directory: %w", err)
	}

	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), "-specialist.md") {
			continue
		}

		// For now, just track that these exist
		specName := strings.TrimSuffix(entry.Name(), ".md")
		r.specialists[specName] = &Specialist{Name: specName}
	}

	return nil
}

// Route determines how to handle a command
func (r *LocalCommandRouter) Route(cmd string, args []string) (*RoutingDecision, error) {
	// 1. Check for exact local command
	if local, exists := r.localCommands[cmd]; exists {
		return r.routeLocal(local, args)
	}

	// 2. Check for shadow command
	if shadow, exists := r.shadowCommands[cmd]; exists {
		return r.routeShadow(shadow, args)
	}

	// 3. Check global command
	if _, exists := r.globalCommands[cmd]; exists {
		return r.routeGlobal(cmd, args)
	}

	// 4. Use orchestrator for intelligent routing
	return r.routeDynamic(cmd, args)
}

// routeLocal handles local command routing
func (r *LocalCommandRouter) routeLocal(cmd *LocalCommand, args []string) (*RoutingDecision, error) {
	decision := &RoutingDecision{
		Command:     cmd.Name,
		Type:        LocalCommandType,
		Context:     cmd.Routing.Context,
		Explanation: fmt.Sprintf("Using local command '%s'", cmd.Name),
	}

	// Handle orchestrated workflows
	if cmd.Routing.Orchestrate || len(cmd.Routing.Workflow) > 0 {
		decision.Workflow = cmd.Routing.Workflow
		decision.Explanation = fmt.Sprintf("Orchestrating workflow for '%s'", cmd.Name)
		return decision, nil
	}

	// Simple routing
	if cmd.Routing.Primary != "" {
		decision.Agents = []string{cmd.Routing.Primary}
		if cmd.Routing.Fallback != "" {
			decision.Agents = append(decision.Agents, cmd.Routing.Fallback)
		}
	}

	return decision, nil
}

// routeShadow handles shadow command routing
func (r *LocalCommandRouter) routeShadow(cmd *ShadowCommand, args []string) (*RoutingDecision, error) {
	decision := &RoutingDecision{
		Command:     cmd.Name,
		Type:        ShadowCommandType,
		Explanation: fmt.Sprintf("Using shadow command '%s' (enhances global)", cmd.Name),
	}

	// Start with global command agents if inheriting
	if cmd.Routing.Inherit {
		// Would get agents from global command
		decision.Agents = []string{} // Placeholder
	}

	// Add additional agents
	decision.Agents = append(decision.Agents, cmd.Routing.Additional...)

	return decision, nil
}

// routeGlobal handles global command routing
func (r *LocalCommandRouter) routeGlobal(cmd string, args []string) (*RoutingDecision, error) {
	return &RoutingDecision{
		Command:     cmd,
		Type:        GlobalCommandType,
		Explanation: fmt.Sprintf("Using global command '%s'", cmd),
		// Agents would be determined by global command metadata
	}, nil
}

// routeDynamic uses orchestrator for intelligent routing
func (r *LocalCommandRouter) routeDynamic(cmd string, args []string) (*RoutingDecision, error) {
	// Check if orchestrator exists
	if _, exists := r.specialists["orchestrator-specialist"]; !exists {
		return nil, fmt.Errorf("command '%s' not found and no orchestrator available", cmd)
	}

	return &RoutingDecision{
		Command:     cmd,
		Type:        DynamicCommandType,
		Agents:      []string{"orchestrator-specialist"},
		Explanation: fmt.Sprintf("No command '%s' found, using orchestrator for intelligent routing", cmd),
	}, nil
}

// GetAvailableCommands returns all available commands
func (r *LocalCommandRouter) GetAvailableCommands() map[string]string {
	commands := make(map[string]string)

	// Add local commands
	for name, cmd := range r.localCommands {
		commands[name] = fmt.Sprintf("[Local] %s", cmd.Description)
	}

	// Add shadow commands
	for name, cmd := range r.shadowCommands {
		commands[name] = fmt.Sprintf("[Shadow] %s", cmd.Description)
	}

	// Add global commands not shadowed
	for name := range r.globalCommands {
		if _, shadowed := r.shadowCommands[name]; !shadowed {
			if _, local := r.localCommands[name]; !local {
				commands[name] = "[Global] Available"
			}
		}
	}

	return commands
}