---
allowed-tools: [Read, Grep, Glob, Bash, Write, Edit, MultiEdit, TodoWrite, Task]
description: "Comprehensive project onboarding: analyze codebase, create specialists, install hooks, and optimize the local orchestrator."
wave-enabled: true
complexity-threshold: 0.8
performance-profile: complex
personas: [orchestrator, analyzer, architect, scribe]
mcp-servers: [sequential, context7, serena]
add-ons: [code2prompt, ast-grep]
argument-hint: "[target] [--force-refresh] [--dry-run]"
thinking-budget: 16000
---

# /crew:onboard - Project Onboarding & Specialist Generation

## Purpose
Execute a comprehensive orchestration lifecycle that analyzes the codebase, creates relevant specialists, installs best-practice hooks, and optimizes the local orchestrator-specialist for effective project collaboration. This command uses extended thinking blocks for transparency throughout the entire process.

## Usage
```
/crew:onboard [target] [--force-refresh] [--dry-run]
```

## Arguments
- `target` - Project directory to analyze (default: current directory)
- `--force-refresh` - Force complete re-analysis and regeneration
- `--dry-run` - Preview what specialists and hooks would be created without making changes

## Onboarding Lifecycle Execution

The entire process is executed as an interleaved sequence of analysis, generation, and installation steps, with thinking blocks providing transparency at each stage.

### Phase 1: CLAUDE.md Initialization & Context Setup

<thinking>
First, I need to check if the project has a CLAUDE.md file. This is the foundational step that ensures all subsequent analysis follows project-specific workflows.
</thinking>

1. **CLAUDE.md Detection & Creation**:
   - Check for existing `CLAUDE.md` in the project root
   - If missing, execute `/init` command to create comprehensive project CLAUDE.md
   - Load and parse CLAUDE.md to understand project-specific workflow requirements
   - Process @ references (@COMMANDS.md, @FLAGS.md, etc.) for comprehensive context

### Phase 2: Extended Analysis with Dual Orchestrators

<thinking>
Now I'll perform collaborative analysis using both global and local orchestrators, with thinking blocks to show the reasoning process at each step.
</thinking>

2. **Global Orchestrator Deep Analysis**: 
   - Activate global `orchestrator` agent from `~/.claude/agents/`
   - Load global `~/.claude/CLAUDE.md` for framework context
   - Perform comprehensive codebase analysis using extended thinking blocks
   - Analyze programming languages, frameworks, patterns, and architecture
   - Identify security considerations and code quality requirements
   - Map project-specific specialist needs with detailed reasoning

3. **Local Orchestrator Context Analysis**:
   - Activate local `orchestrator-specialist` from `.claude/agents/`
   - Conduct project-specific pattern analysis following CLAUDE.md workflow
   - Identify codebase-specific orchestration needs and constraints
   - Analyze existing hooks and configuration requirements

4. **Collaborative Intelligence**:
   - Both orchestrators collaborate using thinking blocks for transparency
   - Synthesize global patterns with local project needs
   - Create comprehensive specialist requirements with detailed justification
   - Plan hook installation strategy based on detected technologies

### Phase 3: Specialist Generation & Installation

<thinking>
Based on the analysis, I'll now generate specialists that are specifically tailored to this project's technology stack and patterns.
</thinking>

5. **Specialist Creation**:
   - Generate specialists using `generic-specialist-template.md` with project-specific customizations
   - Create specialists for each detected technology stack
   - Install specialists to local `.claude/agents/` directory
   - Ensure specialists include project-specific context and patterns

6. **Specialist Validation**:
   - Verify all generated specialists are properly installed
   - Test specialist accessibility and routing capabilities
   - Document specialist capabilities and usage patterns

### Phase 4: Automated Hook Installation

<thinking>
Now I'll detect project needs and install appropriate hooks to enhance development workflow, security, and code quality.
</thinking>

7. **Hook Detection & Installation**:
   - Analyze project structure to detect language and framework needs
   - Install security hooks (e.g., pre-commit security scanning)
   - Install code formatting hooks (e.g., lint-on-save, format-on-save)
   - Install testing hooks (e.g., test-on-change, pre-push testing)
   - Configure hooks in `.claude/settings.json` with project-specific settings
   - Ensure hooks follow security best practices and minimal performance impact

8. **Hook Configuration**:
   - Create comprehensive hook configuration in `.claude/settings.json`
   - Map hooks to appropriate triggers and conditions
   - Ensure compatibility with existing development workflows
   - Document hook purposes and configuration options

### Phase 5: Orchestrator Optimization

<thinking>
With specialists and hooks installed, I'll now optimize the local orchestrator-specialist to work effectively with the new components.
</thinking>

9. **Orchestrator-Specialist Enhancement**:
   - Update local `orchestrator-specialist.md` with new specialist routing
   - Integrate hook awareness into orchestrator workflows
   - Optimize chain coordination for project-specific patterns
   - Update specialist registry with hook integration capabilities

10. **Workflow Optimization**:
    - Fine-tune orchestrator-specialist for effective collaboration
    - Create project-specific workflow patterns incorporating hooks
    - Establish optimal chain patterns for common operations
    - Ensure seamless integration between specialists and hooks

### Phase 6: Comprehensive Validation

<thinking>
Finally, I'll validate the entire setup to ensure all components work together correctly and provide a smooth development experience.
</thinking>

11. **System Integration Testing**:
    - Test orchestrator → specialist → chain workflows with hooks active
    - Validate all specialists are discoverable and routable
    - Test hook triggers and ensure they execute correctly
    - Verify hook integration with specialist workflows

12. **End-to-End Validation**:
    - Execute sample workflows to test complete integration
    - Validate security and formatting hooks trigger appropriately
    - Test specialist routing with hook-aware orchestrator
    - Ensure all components maintain project-specific optimization

## Execution Strategy

### Transparent Multi-Agent Coordination Pattern
```yaml
ORCHESTRATION_CHAIN:
  step1: orchestrator - "Initialize CLAUDE.md and perform global analysis with thinking blocks"
  step2: orchestrator-specialist - "Local pattern analysis with hook requirements"
  step3: orchestrator + orchestrator-specialist - "Collaborative specialist and hook planning"
  step4: orchestrator - "Generate specialists with project-specific customizations"
  step5: orchestrator - "Install and configure hooks based on detected needs"
  step6: orchestrator - "Optimize orchestrator-specialist with hook integration"
  step7: scribe - "Document complete setup and validation results"
```

### Specialist & Hook Detection Logic
```yaml
DETECTION_MAPPING:
  language_detection:
    go_files: ["go-specialist", "go-security-hooks", "go-format-hooks"]
    js_ts_files: ["javascript-specialist", "eslint-hooks", "prettier-hooks"]
    py_files: ["python-specialist", "black-hooks", "pytest-hooks"]
    java_files: ["java-specialist", "checkstyle-hooks", "maven-hooks"]
    rs_files: ["rust-specialist", "clippy-hooks", "cargo-fmt-hooks"]
  
  framework_detection:
    react_patterns: ["react-specialist", "jsx-lint-hooks", "testing-hooks"]
    vue_patterns: ["vue-specialist", "vue-lint-hooks", "testing-hooks"]
    django_patterns: ["django-specialist", "python-security-hooks", "migration-hooks"]
    express_patterns: ["nodejs-specialist", "security-scan-hooks", "testing-hooks"]
  
  security_patterns:
    auth_patterns: ["security-specialist", "security-scan-hooks"]
    api_patterns: ["api-specialist", "security-scan-hooks"]
    database_patterns: ["database-specialist", "migration-hooks"]
```

### Hook Installation Examples
- **Security Hooks**: Pre-commit security scanning, dependency vulnerability checks
- **Code Quality Hooks**: Lint-on-save, format-on-save, import organization
- **Testing Hooks**: Test-on-change, pre-push test execution, coverage reporting
- **Git Hooks**: Auto-commit messages, branch protection, backup creation

## Expected Outcomes

### Generated Specialists
- Technology-specific specialists optimized for your project stack
- Specialists include hook awareness and integration capabilities
- Project-specific context embedded in each specialist

### Installed Hooks
- Security-focused hooks for vulnerability detection
- Code quality hooks for consistent formatting
- Testing hooks for continuous validation
- Git hooks for workflow automation

### Enhanced Orchestrator-Specialist
- Updated with specialist and hook routing
- Optimized for project-specific workflows
- Integrated hook trigger awareness
- Enhanced collaboration patterns

### Validation Results
- Complete system integration test results
- Specialist routing verification
- Hook trigger validation
- Performance impact assessment

## Quality Assurance

- **Deterministic Generation**: All specialists and hooks are generated deterministically
- **Security Validation**: All hooks are security-reviewed and sandboxed
- **Performance Testing**: Hook impact on development workflow is measured
- **Integration Testing**: Complete end-to-end workflow validation
- **Documentation**: Comprehensive setup documentation and usage guides

This enhanced command transforms your project into a fully orchestrated environment with specialized agents, automated hooks, and optimized workflows tailored to your specific codebase and development practices.