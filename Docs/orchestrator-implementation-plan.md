# Orchestrator-Specialist Implementation Plan

## Executive Summary

This plan establishes the orchestrator-specialist as the sole deterministic agent in the Claude Code SuperCrew framework. All other project-specific specialists are generated on-demand by Claude based on actual project needs, not pre-determined templates.

## Core Architecture

### 1. Deterministic Components

#### Orchestrator-Specialist (Always Present)
- **Location**: `.claude/agents/orchestrator-specialist.md`
- **Installation**: Automatic on `/crew:onboard`
- **Updates**: Version-controlled with backup
- **Customization**: Claude can modify based on project context

#### Global Personas (Always Available)
- **Location**: `~/.claude/agents/`
- **Availability**: All projects
- **Examples**: architect-persona, frontend-persona, backend-persona, etc.

### 2. On-Demand Components

#### Project Specialists (Generated as Needed)
- **Location**: `.claude/agents/[specialist-name].md`
- **Creation**: Only when orchestrator/Claude determines need
- **Examples**: Determined by project (e.g., go-specialist, api-specialist)
- **Trigger**: Repeated patterns, explicit requests, complexity thresholds

## Implementation Steps

### Step 1: Orchestrator Installation (/crew:onboard)

```go
// When /crew:onboard is executed:
1. Install/update orchestrator-specialist.md
2. Analyze project structure
3. Save analysis to project-analysis.json
4. Display recommendations
5. Guide Claude on when to create specialists
```

### Step 2: Project Analysis Framework

```yaml
project_analysis:
  languages:
    - Detect primary and secondary languages
    - Count files per language
    - Identify language-specific patterns
  
  frameworks:
    - Package managers (npm, go.mod, etc.)
    - Build tools
    - Testing frameworks
  
  patterns:
    - API/REST indicators
    - CLI application markers
    - Microservice architecture
    - Testing practices
    - CI/CD setup
  
  complexity:
    - File count and directory depth
    - Cross-cutting concerns
    - Multi-domain indicators
```

### Step 3: Agent Generation Decision Logic

```python
def should_generate_specialist(task_history, project_analysis):
    """Claude uses this logic to decide on specialist creation"""
    
    # Frequency threshold
    if count_similar_tasks(task_history) >= 3:
        return True
    
    # Complexity threshold
    if task_complexity > 0.8 and no_existing_specialist():
        return True
    
    # Explicit request
    if user_requests_specialist():
        return True
    
    # Pattern match
    if repeated_pattern_detected() and would_benefit():
        return True
    
    return False
```

### Step 4: Orchestrator Routing Intelligence

```yaml
routing_decision_tree:
  analyze_request:
    - Extract domains and complexity
    - Check for multi-step indicators
    - Assess uncertainty level
  
  single_domain_simple:
    - Route directly to persona
    - Example: "fix typo" → scribe-persona
  
  single_domain_complex:
    - Check for existing specialist
    - If none, use persona + guidance
    - Consider specialist generation
  
  multi_domain:
    - Always use orchestrator
    - Design workflow (sequential/parallel)
    - Coordinate agent handoffs
  
  uncertain:
    - Use orchestrator for analysis
    - Provide routing options
    - Let user/Claude decide
```

## Command Enhancements

### /crew:orchestrate
Primary command for complex coordination:
```bash
/crew:orchestrate "implement secure API with tests"

🎯 Orchestrator Response:
- Analyzing task complexity...
- Domains detected: Backend, Security, Testing
- Recommended workflow: Sequential with validation
- No specialists exist yet - using personas
- [May suggest creating api-specialist if pattern repeats]
```

### /crew:analyze
Check project state and specialist recommendations:
```bash
/crew:analyze

🎯 Project Status:
Languages: Go (primary), JavaScript
Patterns: CLI application, API endpoints
Current Specialists: None
Recommendations: 
- Consider api-specialist if API tasks increase
- Current personas sufficient for most tasks
```

### /crew:chain
Explicit multi-agent workflow:
```bash
/crew:chain "design → implement → test → deploy"

🎯 Creating workflow:
1. architect-persona → System design
2. [specialist or persona] → Implementation
3. qa-persona → Testing
4. devops-persona → Deployment
```

## Promotion Strategy

### In Command Responses
```yaml
simple_task_response:
  text: "Delegating to frontend-persona..."
  hint: "Simple task - direct routing"

complex_task_response:
  text: "This requires coordination..."
  action: "Using orchestrator for multi-agent workflow"
  hint: "Use /crew:orchestrate for complex tasks"
```

### In Help Documentation
```markdown
## When to Use What

**Direct Persona/Specialist Calls**
- Clear, single-domain tasks
- Obvious tool choice
- No coordination needed

**Orchestrator (/crew:orchestrate)**
- Multi-step workflows
- Cross-domain tasks
- Uncertain approach
- Need coordination
```

### In Onboarding
```
Welcome to Claude Code SuperCrew!

✅ Orchestrator-specialist installed
✅ Global personas available
✅ Project analyzed

For complex tasks, use:
  /crew:orchestrate "your complex task"

The orchestrator will:
- Analyze complexity
- Route to appropriate agents
- Suggest specialists if patterns emerge
```

## Claude AI Integration

### Agent Generation Instructions

When Claude detects need for a specialist:

```markdown
Claude, based on repeated [pattern] tasks, create a specialist:

1. Create .claude/agents/[name]-specialist.md
2. Include:
   - Unique emoji and visual identity
   - Specific expertise for this project's [pattern]
   - Integration with project's tools/frameworks
   - Slash commands that promote the specialist
3. Update orchestrator's agent registry
4. Announce availability to user
```

### Flexibility Principles

1. **No Hardcoded Specialists**: Only orchestrator is predetermined
2. **Context-Aware Generation**: Specialists match actual project needs
3. **Evolutionary Design**: Specialists created as patterns emerge
4. **Claude Autonomy**: AI decides when specialists add value

## Success Metrics

### Orchestrator Effectiveness
- Correctly routes 95%+ of requests
- Identifies complexity accurately
- Suggests specialists appropriately

### Specialist Generation
- Only created when truly needed
- Match project patterns closely
- Reduce task completion time

### User Experience
- Clear when to use orchestrator
- Smooth multi-agent workflows
- Intuitive specialist discovery

## Maintenance and Evolution

### Orchestrator Updates
- Version tracking in YAML frontmatter
- Backup before updates
- Preserve customizations

### Specialist Lifecycle
- Track usage patterns
- Suggest consolidation if overlap
- Archive if no longer needed

### Continuous Improvement
- Orchestrator learns from routing patterns
- Claude refines generation criteria
- User feedback shapes evolution

---

This implementation ensures the orchestrator-specialist serves as an intelligent, flexible guardrail that adapts to each project's unique needs while maintaining the Claude Code SuperCrew framework's consistency and power.