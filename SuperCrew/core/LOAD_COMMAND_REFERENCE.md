# `/crew:load` Command Reference Guide

## Quick Reference for Orchestrators

This guide provides both the **Global Orchestrator Agent** and **Local Orchestrator Specialist** with immediate access to `/crew:load` execution procedures, responsibilities, and decision-making frameworks.

## Command Overview
```yaml
command: "/crew:load"
purpose: "Comprehensive orchestration lifecycle: analyze codebase, create specialists, optimize orchestrator"
complexity: "High - Multi-phase, multi-agent coordination"
execution_time: "5-15 minutes depending on project size"
wave_enabled: true
```

## Orchestrator Role Matrix

| Phase | Global Orchestrator | Local Orchestrator Specialist |
|-------|--------------------|---------------------------------|
| **Analysis** | Lead comprehensive analysis | Provide project-specific insights |
| **Planning** | Design specialist requirements | Assist with integration planning |
| **Generation** | Create and install specialists | Validate local integration |
| **Optimization** | Update orchestrator-specialist | Receive and integrate enhancements |
| **Validation** | Test system-wide functionality | Validate local workflow patterns |

## Phase Execution Checklists

### Phase 1: Dual Orchestrator Analysis ✓

#### Global Orchestrator Checklist:
- [ ] Scan project directory structure completely
- [ ] Identify all programming languages and their prevalence
- [ ] Detect framework and library patterns
- [ ] Analyze build systems and configuration files
- [ ] Assess architectural patterns and complexity
- [ ] Document technology stack and requirements
- [ ] Plan specialist generation priority

#### Local Orchestrator Specialist Checklist:
- [ ] Analyze existing project workflow patterns
- [ ] Identify project-specific orchestration needs
- [ ] Assess current development environment
- [ ] Document local constraints and requirements
- [ ] Prepare collaboration input for global orchestrator
- [ ] Plan integration strategy for new specialists

#### Collaboration Checkpoint:
- [ ] Both orchestrators share analysis results
- [ ] Reconcile findings and recommendations
- [ ] Agree on specialist requirements and priorities
- [ ] Document collaborative execution plan

### Phase 2: Specialist Creation & Installation ✓

#### Global Orchestrator Actions:
- [ ] Load `generic-specialist-template.md` from templates
- [ ] Generate required specialists based on analysis:
  - [ ] Language specialists (go, python, javascript, etc.)
  - [ ] Framework specialists (react, django, express, etc.)
  - [ ] Pattern specialists (cli, api, testing, database, etc.)
- [ ] Customize specialists with project-specific context
- [ ] Install specialists to `.claude/agents/` directory
- [ ] Validate specialist file format and accessibility
- [ ] Test specialist basic functionality

#### Local Orchestrator Specialist Actions:
- [ ] Monitor specialist installation progress
- [ ] Validate specialists are accessible locally
- [ ] Test routing and discovery of new specialists
- [ ] Verify specialist precedence and conflict resolution
- [ ] Test basic workflow patterns with new specialists
- [ ] Report any integration issues to global orchestrator

### Phase 3: Orchestrator-Specialist Optimization ✓

#### Global Orchestrator Updates to Local Orchestrator:
- [ ] Update specialist registry with new agents
- [ ] Add project-specific routing patterns
- [ ] Optimize chain coordination algorithms
- [ ] Enhance workflow templates for detected technologies
- [ ] Add project-specific context and patterns
- [ ] Update precedence rules and conflict resolution
- [ ] Integrate quality gates and validation patterns

#### Local Orchestrator Specialist Integration:
- [ ] Receive and integrate routing enhancements
- [ ] Update specialist discovery and utilization
- [ ] Learn new chain coordination patterns
- [ ] Integrate workflow optimizations
- [ ] Test enhanced routing capabilities
- [ ] Validate improved specialist coordination

### Phase 4: Validation & Documentation ✓

#### Shared Validation Tasks:
- [ ] Test each specialist individually
- [ ] Validate common workflow patterns end-to-end
- [ ] Test routing decisions and specialist selection
- [ ] Verify chain coordination and handoffs
- [ ] Test error handling and recovery
- [ ] Validate performance and efficiency
- [ ] Document all generated capabilities
- [ ] Create usage guidance and best practices

## Decision-Making Framework

### Specialist Selection Criteria (Global Orchestrator):
```yaml
ESSENTIAL_SPECIALISTS:
  criteria:
    - Primary language used in >30% of codebase
    - Framework central to application architecture
    - Pattern critical to core functionality
  
BENEFICIAL_SPECIALISTS:
  criteria:
    - Secondary language used in >10% of codebase
    - Framework used for specific components
    - Pattern supporting important workflows
  
AVOID_SPECIALISTS:
  criteria:
    - Technology not present in codebase
    - Patterns that duplicate existing capabilities
    - Specialists that would create resource conflicts
```

### Routing Optimization Criteria (Local Orchestrator Specialist):
```yaml
ROUTING_PRIORITIES:
  1. project_specialists: "Always prefer project-specific agents"
  2. language_alignment: "Route to primary language specialists first"
  3. workflow_efficiency: "Minimize chain complexity when possible"
  4. quality_gates: "Include validation agents in critical workflows"
  5. resource_optimization: "Balance thoroughness with execution speed"
```

## Communication Templates

### Analysis Sharing Template:
```yaml
ANALYSIS_REPORT:
  orchestrator_type: "global|local"
  project_summary:
    primary_languages: []
    frameworks_detected: []
    architectural_patterns: []
    complexity_assessment: "low|medium|high"
  
  specialist_recommendations:
    essential: []
    beneficial: []
    rejected: []
  
  collaboration_input:
    optimization_requirements: []
    integration_considerations: []
    workflow_priorities: []
```

### Progress Update Template:
```yaml
PROGRESS_UPDATE:
  phase: "analysis|generation|optimization|validation"
  status: "in_progress|completed|blocked"
  completed_tasks: []
  current_task: ""
  next_tasks: []
  issues_encountered: []
  assistance_needed: ""
```

## Error Handling Protocols

### Common Issues and Resolutions:

#### Specialist Generation Failures:
```yaml
ISSUE: "Cannot generate specialist from template"
DIAGNOSIS:
  - Check template accessibility and format
  - Validate project analysis data completeness
  - Verify target directory permissions
RESOLUTION:
  - Retry with corrected parameters
  - Use fallback template if available
  - Escalate to user if fundamental issue
```

#### Integration Conflicts:
```yaml
ISSUE: "New specialist conflicts with existing agent"
DIAGNOSIS:
  - Check agent name conflicts
  - Validate precedence rules
  - Assess capability overlaps
RESOLUTION:
  - Rename conflicting specialist
  - Update precedence configuration
  - Merge capabilities if appropriate
```

#### Routing Optimization Failures:
```yaml
ISSUE: "Enhanced routing patterns not working"
DIAGNOSIS:
  - Validate specialist accessibility
  - Check routing pattern syntax
  - Test chain coordination logic
RESOLUTION:
  - Revert to previous working patterns
  - Incrementally apply optimizations
  - Validate each change independently
```

## Quality Assurance Checklist

### Pre-Execution Validation:
- [ ] Both orchestrators accessible and functional
- [ ] Generic specialist template available
- [ ] Target directory writable
- [ ] No conflicting background processes

### Mid-Execution Monitoring:
- [ ] Specialist generation proceeding successfully
- [ ] No file system errors or conflicts
- [ ] Integration testing passing
- [ ] Performance within acceptable limits

### Post-Execution Verification:
- [ ] All required specialists generated and accessible
- [ ] Orchestrator-specialist enhanced successfully
- [ ] Workflow patterns optimized and functional
- [ ] Documentation complete and accurate
- [ ] System ready for production use

## Success Metrics

### Quantitative Measures:
```yaml
METRICS:
  specialist_coverage: "% of detected technologies with specialists"
  routing_efficiency: "Average routing decision time"
  workflow_performance: "Chain execution time improvement"
  error_rate: "% of operations completing successfully"
  user_satisfaction: "Workflow effectiveness rating"
```

### Qualitative Indicators:
```yaml
INDICATORS:
  - Orchestrators collaborate effectively without conflicts
  - Specialists provide relevant, project-specific expertise
  - Workflow patterns are optimized for detected technologies
  - System scales appropriately with project complexity
  - Documentation enables effective ongoing usage
```

## Quick Command Reference

### For Global Orchestrator:
```bash
# Primary responsibilities during /crew:load execution
1. Analyze codebase comprehensively
2. Generate project-appropriate specialists
3. Optimize local orchestrator-specialist
4. Validate system functionality
5. Document results and capabilities
```

### For Local Orchestrator Specialist:
```bash
# Primary responsibilities during /crew:load execution
1. Provide project-specific insights
2. Support specialist integration
3. Receive orchestrator enhancements
4. Validate local functionality
5. Execute optimized workflows
```

## Troubleshooting Quick Reference

| Symptom | Likely Cause | Quick Fix |
|---------|--------------|-----------|
| No specialists generated | Analysis incomplete | Re-run analysis phase |
| Routing not working | Specialist not accessible | Check file permissions |
| Chain patterns failing | Coordination logic error | Revert to basic patterns |
| Performance degraded | Too many specialists | Optimize specialist selection |
| Integration conflicts | Name/precedence issues | Update conflict resolution |

---

**Remember**: Both orchestrators are responsible for the success of `/crew:load` execution. Collaborate effectively, communicate clearly, and validate thoroughly to ensure optimal orchestration environment setup.