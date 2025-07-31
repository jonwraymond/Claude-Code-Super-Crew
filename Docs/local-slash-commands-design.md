# Local Custom Slash Commands Layer Design

## Overview

A new intelligent routing layer that enables project-specific slash commands to leverage both global personas and local specialists, with the ability to shadow global commands and customize them for local project patterns.

## Architecture

### 1. Command Resolution Hierarchy

```yaml
command_resolution:
  1_local_explicit:
    - path: .claude/commands/{command}.md
    - priority: highest
    - behavior: Always executes if found
    
  2_local_shadow:
    - path: .claude/commands/shadows/{command}.md
    - priority: high
    - behavior: Enhances/modifies global command
    
  3_global_command:
    - path: ~/.claude/commands/{command}.md
    - priority: normal
    - behavior: Default execution
    
  4_dynamic_routing:
    - orchestrator: Analyzes and routes
    - priority: fallback
    - behavior: Intelligent agent selection
```

### 2. Command Structure with Sub-Agent Chaining

#### Local Command Format with Chaining (.claude/commands/api.md)
```yaml
---
name: api
description: Generate API endpoints following project conventions using multi-agent workflow
shadows: false  # Not shadowing a global command
routing:
  orchestrate: true           # Enable multi-agent orchestration
  # Option 1: Simple agent specification
  primary: backend-specialist  # Single agent (fallback)
  fallback: backend-persona    # If specialist unavailable
  
  # Option 2: POWERFUL Sub-Agent Chain (Recommended!)
  workflow:
    - name: "analyze"
      agent: "analyzer-persona"
      task: "Analyze existing API patterns and conventions"
    - name: "design"
      agent: "architect-persona"
      task: "Design API structure and contracts"
    - name: "implement"
      agent: "backend-specialist"
      task: "Generate API endpoint code"
    - name: "test"
      agent: "test-generator-specialist"
      task: "Create comprehensive tests"
    - name: "document"
      agent: "scribe-persona"
      task: "Generate API documentation"
tools:
  - Read
  - Write
  - MultiEdit
  - Serena
  - Task  # Important for sub-agent chaining!
flags:
  - name: --rest
    description: Generate RESTful endpoint
  - name: --graphql
    description: Generate GraphQL resolver
  - name: --grpc
    description: Generate gRPC service
---

# API Command - Project-Specific

This command understands our project's API patterns:
- Authentication middleware
- Error response format
- Validation patterns
- Database transaction handling

## Usage Examples

### Generate REST endpoint
```
/api users --rest
```

### Generate GraphQL resolver
```
/api products --graphql
```

## Project Patterns

### Authentication
All endpoints use our custom JWT middleware:
```go
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
    // Automatically includes auth check
    user := middleware.GetUser(r.Context())
    // ... implementation
}
```

### Error Handling
Consistent error format:
```go
return responses.Error(w, http.StatusBadRequest, "INVALID_INPUT", "Email is required")
```
```

#### Shadow Command Format (.claude/commands/shadows/build.md)
```yaml
---
name: build
description: Enhanced build command with project-specific steps
shadows: true
base_command: build  # Which global command to enhance
routing:
  inherit: true      # Use global routing
  additional:
    - docker-specialist  # Add local specialist
  pre_hooks:
    - validate-env
    - check-dependencies
  post_hooks:
    - run-tests
    - build-docker
---

# Build Command Shadow - Project Enhancements

This shadows the global /build command and adds:
1. Environment validation
2. Dependency checking
3. Automatic test execution
4. Docker image building

## Project-Specific Build Steps

### Pre-Build Validation
```bash
# Automatically run before build
./scripts/validate-env.sh
go mod verify
```

### Post-Build Actions
```bash
# Automatically run after successful build
go test ./...
docker build -t myapp:latest .
```

## Flags

Inherits all global build flags plus:
- `--skip-docker`: Skip Docker image creation
- `--production`: Use production build settings
```

### 3. Intelligent Routing System

#### Command Router Enhancement
```go
type LocalCommandRouter struct {
    globalCommands   map[string]*Command
    localCommands    map[string]*LocalCommand
    shadowCommands   map[string]*ShadowCommand
    orchestrator     *Orchestrator
    projectContext   *ProjectContext
}

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
    if global, exists := r.globalCommands[cmd]; exists {
        return r.routeGlobal(global, args)
    }
    
    // 4. Use orchestrator for intelligent routing
    return r.orchestrator.IntelligentRoute(cmd, args, r.projectContext)
}
```

#### Routing Decision Structure
```go
type RoutingDecision struct {
    Command      string
    Type         CommandType  // Local, Shadow, Global, Dynamic
    Agents       []Agent      // Ordered list of agents to use
    Workflow     *Workflow    // Multi-agent workflow if needed
    Context      map[string]interface{}
    Explanation  string       // Why this routing was chosen
}
```

### 4. Agent Selection Intelligence

#### Smart Agent Resolution
```yaml
agent_resolution:
  1_explicit_routing:
    - source: Command metadata
    - example: "primary: api-specialist"
    
  2_pattern_matching:
    - source: Project patterns
    - example: "API files → api-specialist"
    
  3_contextual_inference:
    - source: Task analysis
    - example: "Database work → db-specialist or backend-persona"
    
  4_orchestrator_override:
    - source: Complexity analysis
    - example: "Multi-domain → orchestrated workflow"
```

#### Dynamic Specialist Discovery
```go
func (r *LocalCommandRouter) discoverSpecialists() []Specialist {
    specialists := []Specialist{}
    
    // Scan .claude/agents/ for specialists
    files, _ := filepath.Glob(".claude/agents/*-specialist.md")
    for _, file := range files {
        spec := parseSpecialist(file)
        specialists = append(specialists, spec)
    }
    
    return specialists
}
```

### 5. Command Enhancement Features

#### Pre/Post Hooks
```yaml
hooks:
  pre_execution:
    - validate_environment
    - check_dependencies
    - setup_context
    
  post_execution:
    - run_tests
    - update_documentation
    - notify_team
```

#### Context Injection
```yaml
context_injection:
  project_patterns:
    - error_format: "project-specific"
    - auth_method: "JWT"
    - db_connection: "connection-pool"
    
  local_conventions:
    - import_style: "grouped"
    - test_pattern: "table-driven"
    - doc_format: "godoc"
```

### 6. Implementation Examples

#### Example 1: Local Migration Command
```yaml
# .claude/commands/migrate.md
---
name: migrate
description: Database migration helper for our Postgres setup
routing:
  primary: db-migration-specialist
  fallback: backend-persona
  orchestrate: false
tools: [Read, Write, Bash, Grep]
---

# Migrate Command

Handles our specific migration workflow:
1. Validates schema changes
2. Generates migration files
3. Runs migrations with rollback support

## Usage
/migrate create add_user_roles
/migrate up
/migrate down --steps 2
```

#### Example 2: Shadow Test Command
```yaml
# .claude/commands/shadows/test.md
---
name: test
shadows: true
base_command: test
routing:
  additional:
    - integration-test-specialist
    - performance-test-specialist
---

# Enhanced Test Command

Adds project-specific testing:
- Runs integration tests against local containers
- Includes performance benchmarks
- Generates coverage reports in our format
```

#### Example 3: Complex Deployment Command
```yaml
# .claude/commands/deploy.md
---
name: deploy
description: Multi-stage deployment with validation
routing:
  orchestrate: true  # Always use orchestrator
  workflow:
    - validate: [security-persona, qa-persona]
    - build: [build-specialist, docker-specialist]
    - deploy: [devops-persona, k8s-specialist]
    - verify: [monitoring-specialist, qa-persona]
---
```

### 7. Claude Code Integration

#### Command Discovery Prompt
```
When a slash command is used:
1. Check local commands first (.claude/commands/)
2. Check shadow commands (.claude/commands/shadows/)
3. Fall back to global commands
4. If no exact match, use orchestrator for intelligent routing

Local commands can:
- Use any combination of global personas and local specialists
- Override or enhance global behavior
- Define complex multi-agent workflows
- Include project-specific context and patterns
```

#### Dynamic Learning
```yaml
learning_system:
  pattern_detection:
    - Track: Common command sequences
    - Identify: Repeated agent combinations
    - Suggest: New local commands
    
  effectiveness_tracking:
    - Monitor: Command success rates
    - Analyze: Agent performance
    - Optimize: Routing decisions
```

### 8. Benefits of Sub-Agent Chaining

1. **Superior Results**: Multiple specialists contribute their expertise
2. **Project Customization**: Commands tailored to project needs
3. **Flexible Shadowing**: Enhance global commands without replacing
4. **Intelligent Routing**: Best agent(s) for each task
5. **Progressive Enhancement**: Start simple, add complexity as needed
6. **Context Awareness**: Commands understand project patterns
7. **Multi-Agent Coordination**: Complex workflows made simple
8. **Quality Through Specialization**: Each agent focuses on what they do best

### Why Sub-Agent Chaining is Game-Changing

#### Single Agent Approach (Limited)
```
/api create-user
→ backend-persona does everything
→ Result: Functional but may miss patterns, tests, docs
```

#### Multi-Agent Chain (Powerful!)
```
/api create-user
→ analyzer: Studies existing patterns
→ architect: Designs optimal structure
→ backend: Implements with patterns
→ tester: Creates comprehensive tests
→ error-handler: Adds robust error handling
→ documenter: Generates complete docs
→ Result: Production-ready, well-tested, documented API
```

### Real-World Chaining Examples

#### Debugging Chain
```yaml
/debug performance-issue:
  workflow:
    - analyzer: "Profile and identify bottlenecks"
    - performance: "Suggest optimizations"
    - implementer: "Apply fixes"
    - tester: "Verify improvements"
    - documenter: "Record changes and learnings"
```

#### Feature Implementation Chain
```yaml
/feature user-notifications:
  workflow:
    - analyzer: "Study requirements and existing patterns"
    - architect: "Design notification system"
    - backend: "Implement server logic"
    - frontend: "Create UI components"
    - tester: "Test all scenarios"
    - security: "Audit for vulnerabilities"
```

#### Refactoring Chain
```yaml
/refactor legacy-module:
  workflow:
    - analyzer: "Map dependencies and usage"
    - architect: "Design new structure"
    - refactorer: "Extract and reorganize"
    - tester: "Ensure no regressions"
    - optimizer: "Improve performance"
```

### 9. Example Project Structure

```
.claude/
├── agents/
│   ├── orchestrator-specialist.md
│   ├── api-specialist.md
│   ├── db-migration-specialist.md
│   └── integration-test-specialist.md
├── commands/
│   ├── api.md           # Local command
│   ├── migrate.md       # Local command
│   ├── deploy.md        # Local command
│   └── shadows/
│       ├── build.md     # Shadows global build
│       └── test.md      # Shadows global test
└── CLAUDE.md            # Project configuration
```

### 10. Future Enhancements

1. **Command Composition**: Combine multiple commands
2. **Conditional Routing**: Route based on file types or patterns
3. **Interactive Mode**: Commands can ask clarifying questions
4. **Command Templates**: Generate new commands from templates
5. **Cross-Project Learning**: Share effective patterns

This design provides maximum flexibility while maintaining simplicity. Projects can start with zero local commands and progressively add them as patterns emerge, with the orchestrator providing intelligent fallback routing throughout.