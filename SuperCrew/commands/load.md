---
allowed-tools: [Read, Grep, Glob, Bash, Write, Edit, MultiEdit, TodoWrite, Task]
description: "Comprehensive orchestration lifecycle: analyze codebase, create specialists, and optimize local orchestrator"
wave-enabled: true
complexity-threshold: 0.8
performance-profile: complex
personas: [orchestrator, analyzer, architect, scribe]
mcp-servers: [sequential, context7]
---

# /crew:load - Orchestration Lifecycle & Specialist Generation

## Purpose
Execute a comprehensive orchestration lifecycle that analyzes the codebase, creates relevant specialists, and optimizes the local orchestrator-specialist for effective project collaboration.

## Usage
```
/crew:load [target] [--generate-specialists] [--optimize-orchestrator] [--force-refresh]
```

## Arguments
- `target` - Project directory to analyze (default: current directory)
- `--generate-specialists` - Force generation of new specialists (default: true)
- `--optimize-orchestrator` - Update orchestrator-specialist for new specialists (default: true)
- `--force-refresh` - Force complete re-analysis and regeneration
- `--dry-run` - Preview what specialists would be created without generating them

## Orchestration Lifecycle Execution

### Phase 1: CLAUDE.md Integration & Dual Orchestrator Analysis
1. **CLAUDE.md Context Loading**:
   - Read project `CLAUDE.md` file to understand project-specific workflow requirements
   - Process @ references (@COMMANDS.md, @FLAGS.md, etc.) for comprehensive context
   - Validate project follows the 9-step workflow defined in CLAUDE.md
   - Create todo.md if required by CLAUDE.md workflow

2. **Global Orchestrator Analysis**: 
   - Activate global `orchestrator` agent from `~/.claude/agents/`
   - Load global `~/.claude/CLAUDE.md` for framework context
   - Perform comprehensive codebase analysis for specialist identification
   - Analyze programming languages, frameworks, patterns, and architecture
   - Apply CLAUDE.md workflow principles throughout analysis

3. **Local Orchestrator Analysis**:
   - Activate local `orchestrator-specialist` from `.claude/agents/`
   - Ensure CLAUDE.md integration protocol is active (requires Read tool)
   - Conduct project-specific pattern analysis following CLAUDE.md workflow
   - Identify codebase-specific orchestration needs within simplicity constraints

4. **Collaborative Analysis**:
   - Both orchestrators collaborate with CLAUDE.md workflow compliance
   - Create comprehensive specialist requirements following quality standards
   - Determine priority and urgency of each specialist type
   - Plan specialist generation strategy with minimal code impact principle

### Phase 2: Specialist Creation & Installation
4. **Specialist Generation**:
   - Global orchestrator creates specialists using `generic-specialist-template.md`
   - Generate specialists for detected technologies (e.g., `go-specialist`, `react-specialist`)
   - Install specialists to local `.claude/agents/` directory
   - Ensure specialists are optimized for THIS specific codebase

5. **Specialist Activation**:
   - Validate all generated specialists are properly installed
   - Test specialist accessibility and functionality
   - Create specialist interaction patterns

### Phase 3: Orchestrator Optimization
6. **Orchestrator-Specialist Enhancement**:
   - Global orchestrator updates the local `orchestrator-specialist.md`
   - Integrate new specialist routing patterns
   - Optimize chain coordination for project-specific workflows
   - Update specialist registry and precedence rules

7. **Collaboration Optimization**:
   - Fine-tune orchestrator-specialist for effective collaboration with new specialists
   - Create project-specific workflow patterns
   - Establish optimal chain patterns for common operations

### Phase 4: Validation & Documentation
8. **System Validation**:
   - Test orchestrator → specialist → chain workflows
   - Validate all specialists are discoverable and functional
   - Ensure orchestrator-specialist can route effectively

9. **Documentation & Completion**:
   - Document generated specialists and their capabilities
   - Update project documentation with new orchestration capabilities
   - Provide usage guidance for optimal specialist utilization

## Execution Strategy

### Multi-Agent Coordination Pattern
```yaml
ORCHESTRATION_CHAIN:
  step1: orchestrator - "Global analysis and specialist planning"
  step2: orchestrator-specialist - "Local pattern analysis and requirements"
  step3: orchestrator + orchestrator-specialist - "Collaborative specialist generation"
  step4: orchestrator - "Orchestrator-specialist optimization"
  step5: scribe - "Documentation of new orchestration capabilities"
```

### Specialist Detection Logic
```yaml
SPECIALIST_MAPPING:
  language_detection:
    go_files: "go-specialist"
    js_ts_files: "javascript-specialist" 
    py_files: "python-specialist"
    java_files: "java-specialist"
    rs_files: "rust-specialist"
  
  framework_detection:
    react_patterns: "react-specialist"
    vue_patterns: "vue-specialist"
    django_patterns: "django-specialist"
    express_patterns: "nodejs-specialist"
  
  pattern_detection:
    api_patterns: "api-specialist"
    cli_patterns: "cli-specialist"
    database_patterns: "database-specialist"
    testing_patterns: "testing-specialist"
```

### Orchestrator-Specialist Updates
The global orchestrator will update the local orchestrator-specialist with:
- **Dynamic Specialist Registry**: List of generated specialists and their capabilities
- **Project-Specific Routing**: Optimized routing patterns for this codebase
- **Chain Optimization**: Enhanced chain patterns utilizing new specialists
- **Workflow Templates**: Common workflows optimized for this project

## Expected Outcomes

### Generated Specialists (Examples)
- `go-specialist.md` - Go language and ecosystem expertise
- `cli-specialist.md` - Command-line interface development
- `testing-specialist.md` - Testing frameworks and methodologies
- `devops-specialist.md` - Infrastructure and deployment

### Enhanced Orchestrator-Specialist
- Updated with project-specific specialist routing
- Optimized chain patterns for detected technologies
- Enhanced workflow coordination for this codebase
- Improved collaboration patterns with new specialists

### Project Capabilities
- Intelligent routing to appropriate specialists
- Technology-specific expertise for all detected patterns
- Optimized multi-agent workflows for common operations
- Enhanced code quality through specialized review chains

## Claude Code Integration
- Leverages global and local orchestrators for comprehensive analysis
- Uses Task tool for complex specialist generation workflows
- Applies Write/Edit for specialist creation and orchestrator updates
- Maintains project-specific optimization while preserving global patterns

## Quality Assurance
- Validates all generated specialists against project requirements
- Ensures orchestrator-specialist optimization is deterministic
- Tests all routing patterns and chain workflows
- Documents all changes and enhancements for future reference

This command transforms your project into a fully orchestrated environment with specialized agents tailored to your specific codebase, coordinated by an optimized local orchestrator.