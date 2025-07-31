# ORCHESTRATOR.md - Claude Code Super Crew Intelligent Routing System

## 🎯 Overview
Intelligent routing system for Claude Code Super Crew framework that provides comprehensive orchestration capabilities for analyzing codebases, creating specialists, and optimizing workflows.

## 🧠 Detection Engine

### Pre-Operation Validation Checks
**Resource Validation**:
- Token usage prediction based on operation complexity and scope
- Memory and processing requirements estimation
- File system permissions and available space verification
- MCP server availability and response time checks

**Compatibility Validation**:
- Flag combination conflict detection
- Persona + command compatibility verification
- Tool availability for requested operations
- Project structure requirements validation

**Risk Assessment**:
- Operation complexity scoring (0.0-1.0 scale)
- Failure probability based on historical patterns
- Resource exhaustion likelihood prediction
- Cascading failure potential analysis

### Pattern Recognition Rules

#### Complexity Detection
```yaml
simple:
  indicators:
    - single file operations
    - basic CRUD tasks
    - straightforward queries
    - < 3 step workflows
  token_budget: 5K
  time_estimate: < 5 min

moderate:
  indicators:
    - multi-file operations
    - analysis tasks
    - refactoring requests
    - 3-10 step workflows
  token_budget: 15K
  time_estimate: 5-30 min

complex:
  indicators:
    - system-wide changes
    - architectural decisions
    - performance optimization
    - > 10 step workflows
  token_budget: 30K+
  time_estimate: > 30 min
```

#### Domain Identification
```yaml
frontend:
  keywords: [UI, component, React, Vue, CSS, responsive, accessibility]
  file_patterns: ["*.jsx", "*.tsx", "*.vue", "*.css", "*.scss"]
  typical_operations: [create, implement, style, optimize, test]

backend:
  keywords: [API, database, server, endpoint, authentication, performance]
  file_patterns: ["*.js", "*.ts", "*.py", "*.go", "controllers/*", "models/*"]
  typical_operations: [implement, optimize, secure, scale]

infrastructure:
  keywords: [deploy, Docker, CI/CD, monitoring, scaling, configuration]
  file_patterns: ["Dockerfile", "*.yml", "*.yaml", ".github/*", "terraform/*"]
  typical_operations: [setup, configure, automate, monitor]

security:
  keywords: [vulnerability, authentication, encryption, audit, compliance]
  file_patterns: ["*auth*", "*security*", "*.pem", "*.key"]
  typical_operations: [scan, harden, audit, fix]

documentation:
  keywords: [document, README, wiki, guide, manual, instructions]
  file_patterns: ["*.md", "*.rst", "*.txt", "docs/*", "README*", "CHANGELOG*"]
  typical_operations: [write, document, explain, translate, localize]
```

## 🚦 Routing Intelligence

### Wave Orchestration Engine
Multi-stage command execution with compound intelligence. Automatic complexity assessment or explicit flag control.

**Wave Control Matrix**:
```yaml
wave-activation:
  automatic: "complexity >= 0.7"
  explicit: "--wave-mode, --force-waves"
  override: "--single-wave, --wave-dry-run"

wave-strategies:
  progressive: "Incremental enhancement"
  systematic: "Methodical analysis"
  adaptive: "Dynamic configuration"
```

### Master Routing Table

| Pattern | Complexity | Domain | Auto-Activates | Confidence |
|---------|------------|---------|----------------|------------|
| "analyze architecture" | complex | infrastructure | architect persona, --ultrathink, Sequential | 95% |
| "create component" | simple | frontend | frontend persona, Magic, --uc | 90% |
| "implement feature" | moderate | any | domain-specific persona, Context7, Sequential | 88% |
| "implement API" | moderate | backend | backend persona, --seq, Context7 | 92% |
| "implement UI component" | simple | frontend | frontend persona, Magic, --c7 | 94% |
| "implement authentication" | complex | security | security persona, backend persona, --validate | 90% |
| "fix bug" | moderate | any | analyzer persona, --think, Sequential | 85% |
| "optimize performance" | complex | backend | performance persona, --think-hard, Playwright | 90% |
| "security audit" | complex | security | security persona, --ultrathink, Sequential | 95% |
| "write documentation" | moderate | documentation | scribe persona, --persona-scribe=en, Context7 | 95% |

## 🎯 Training Materials

### Core Training Documents

#### 1. ORCHESTRATOR_onboard_TRAINING.md - Comprehensive Training
**Purpose**: Complete training on `/crew:onboard` command operations
**Audience**: Both Global and Local Orchestrators
**Content**:
- Phase-by-phase execution guide
- Role responsibilities and collaboration protocols
- Communication standards and error handling
- Quality assurance and success criteria

#### 2. ONBOARD_COMMAND_REFERENCE.md - Quick Reference
**Purpose**: Immediate access guide for execution
**Audience**: Both Global and Local Orchestrators
**Content**:
- Execution checklists and decision frameworks
- Communication templates and error protocols
- Quality metrics and troubleshooting guides
- Command reference and success indicators

### Training Path by Role

#### For Global Orchestrator Agent

**Required Reading (In Order)**:
1. [Agent Selection Guidelines](./PERSONAS.md#agent-selection-guidelines) - Understand authority and responsibility for agent selection
2. [Comprehensive Training](#training-materials) - Learn comprehensive execution procedures
3. [Quick Reference](#quick-reference-summary) - Access quick reference for execution

**Key Competencies to Master**:
- [ ] Comprehensive codebase analysis and technology detection
- [ ] Intelligent specialist generation using templates
- [ ] Local orchestrator-specialist optimization and enhancement
- [ ] Collaborative coordination with local orchestrator
- [ ] System validation and quality assurance

#### For Local Orchestrator Specialist

**Required Reading (In Order)**:
1. [Comprehensive Training](#training-materials) - Understand collaboration and support role
2. [Quick Reference](#quick-reference-summary) - Access execution checklists and protocols
3. [Agent Selection Guidelines](./PERSONAS.md#agent-selection-guidelines) - Understand global orchestrator's decision authority

**Key Competencies to Master**:
- [ ] Project-specific pattern analysis and insight provision
- [ ] Specialist integration validation and testing
- [ ] Effective collaboration with global orchestrator
- [ ] Enhancement reception and workflow optimization
- [ ] Local validation and ongoing orchestration

## 🔧 Configuration

### Orchestrator Settings
```yaml
orchestrator_config:
  # Performance
  enable_caching: true
  cache_ttl: 3600
  parallel_operations: true
  max_parallel: 3
  
  # Intelligence
  learning_enabled: true
  confidence_threshold: 0.7
  pattern_detection: aggressive
  
  # Resource Management
  token_reserve: 10%
  emergency_threshold: 90%
  compression_threshold: 75%
  
  # Wave Mode Settings
  wave_mode:
    enable_auto_detection: true
    wave_score_threshold: 0.7
    max_waves_per_operation: 5
    adaptive_wave_sizing: true
    wave_validation_required: true
```

## 🚨 Emergency Protocols

### Resource Management
Threshold-based resource management follows unified Resource Management Thresholds:
- **Green Zone** (0-60%): Full operations
- **Yellow Zone** (60-75%): Resource optimization
- **Orange Zone** (75-85%): Warning alerts
- **Red Zone** (85-95%): Force efficiency modes
- **Critical Zone** (95%+): Emergency protocols

### Graceful Degradation
- **Level 1**: Reduce verbosity, skip optional enhancements
- **Level 2**: Disable advanced features, simplify operations
- **Level 3**: Essential operations only, maximum compression

## 📚 Quick Reference Summary

### `/crew:onboard` Execution Phases
```yaml
PHASE_1_ANALYSIS:
  global: "Lead comprehensive codebase analysis"
  local: "Provide project-specific insights and patterns"
  
PHASE_2_GENERATION:
  global: "Generate and install project-appropriate specialists"
  local: "Validate specialist integration and accessibility"
  
PHASE_3_OPTIMIZATION:
  global: "Update local orchestrator-specialist with enhancements"
  local: "Receive and integrate routing and workflow optimizations"
  
PHASE_4_VALIDATION:
  both: "Test workflows, validate functionality, document results"
```

### Communication Protocol
```yaml
COLLABORATION_FLOW:
  1. independent_analysis: "Both orchestrators analyze separately"
  2. collaborative_synthesis: "Share findings and plan together"
  3. coordinated_execution: "Execute with clear role delineation"
  4. joint_validation: "Test and validate results together"
  5. shared_documentation: "Document outcomes and guidance"
```

### Success Criteria
```yaml
SUCCESS_INDICATORS:
  - All detected technologies have appropriate specialists
  - Orchestrator-specialist optimized for project workflows
  - Routing patterns efficient and effective
  - System validated and documented thoroughly
  - Both orchestrators can operate enhanced system effectively
```

---

**Note**: The training materials from ORCHESTRATOR_TRAINING_INDEX.md, ONBOARD_COMMAND_REFERENCE.md, and ORCHESTRATOR_ONBOARD_TRAINING.md have been consolidated into this single comprehensive ORCHESTRATOR.md file. The AGENT_SELECTION_PROMPT.md content has been integrated into the PERSONAS.md file as indicated in the training materials.
