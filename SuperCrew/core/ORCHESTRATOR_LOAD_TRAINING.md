# Orchestrator Training: `/crew:load` Command Operations

## Training Overview

This document provides comprehensive training for both the **Global Orchestrator Agent** and **Local Orchestrator Specialist** on the `/crew:load` command operations, responsibilities, and collaboration protocols.

## Command Authority and Roles

### Global Orchestrator Agent (from ~/.claude/agents/orchestrator-agent.md)
**Primary Responsibilities:**
- **Lead the overall orchestration lifecycle**
- **Perform comprehensive codebase analysis**
- **Generate project-specific specialists**
- **Update and optimize the local orchestrator-specialist**
- **Coordinate with local orchestrator for collaborative analysis**

### Local Orchestrator Specialist (from .claude/agents/orchestrator-specialist.md)
**Primary Responsibilities:**
- **Provide project-specific pattern insights**
- **Assist in specialist requirement analysis**
- **Support collaboration with global orchestrator**
- **Execute optimized routing after enhancement**
- **Maintain local orchestration after setup**

## Phase-by-Phase Execution Guide

### Phase 1: Dual Orchestrator Codebase Analysis

#### Global Orchestrator Tasks:
```yaml
ANALYSIS_RESPONSIBILITIES:
  1. file_structure_analysis:
     - Scan entire project directory tree
     - Identify primary and secondary programming languages
     - Detect framework and library patterns
     - Assess architectural complexity and patterns
  
  2. technology_detection:
     - Parse configuration files (package.json, go.mod, requirements.txt, etc.)
     - Identify build systems and toolchains
     - Detect testing frameworks and methodologies
     - Analyze deployment and infrastructure patterns
  
  3. specialist_requirement_planning:
     - Determine which specialists are essential vs. beneficial
     - Plan specialist generation priority and sequence
     - Consider resource allocation and workflow efficiency
     - Prepare specialist customization requirements
```

#### Local Orchestrator Specialist Tasks:
```yaml
COLLABORATION_RESPONSIBILITIES:
  1. project_context_analysis:
     - Analyze existing project patterns and conventions
     - Identify project-specific workflow requirements
     - Assess current orchestration gaps and needs
     - Understand local development environment constraints
  
  2. specialist_integration_planning:
     - Determine how new specialists will integrate with existing workflows
     - Identify optimal chain patterns for this specific project
     - Plan routing optimization requirements
     - Assess collaboration efficiency opportunities
  
  3. local_optimization_requirements:
     - Define project-specific routing patterns needed
     - Identify workflow templates that should be enhanced
     - Plan integration with existing local development processes
     - Prepare for orchestrator-specialist enhancement phase
```

#### Collaborative Analysis Protocol:
```yaml
COLLABORATION_WORKFLOW:
  1. independent_analysis:
     - Both orchestrators perform initial analysis separately
     - Document findings and recommendations independently
     - Prepare collaboration input and requirements
  
  2. collaborative_synthesis:
     - Share analysis results and findings
     - Reconcile different perspectives and priorities
     - Agree on final specialist requirements and priorities
     - Plan coordinated execution strategy
  
  3. execution_planning:
     - Define specialist generation sequence and responsibilities
     - Plan orchestrator-specialist enhancement requirements
     - Establish validation and testing protocols
     - Set success criteria and completion standards
```

### Phase 2: Specialist Creation & Installation

#### Global Orchestrator Execution:
```yaml
SPECIALIST_GENERATION:
  1. template_utilization:
     - Use generic-specialist-template.md as foundation
     - Customize template with project-specific patterns
     - Generate specialists based on analysis findings
     - Ensure specialists are optimized for THIS codebase
  
  2. specialist_creation_process:
     - Create go-specialist.md (if Go detected)
     - Generate react-specialist.md (if React patterns found)
     - Build cli-specialist.md (if CLI patterns detected)
     - Develop testing-specialist.md (if testing frameworks present)
     - Create additional specialists based on analysis
  
  3. installation_management:
     - Install specialists to .claude/agents/ directory
     - Validate specialist files are properly formatted
     - Test specialist accessibility and functionality
     - Ensure specialists integrate with existing agents
  
  4. specialist_optimization:
     - Customize specialists for detected project patterns
     - Include project-specific examples and contexts
     - Optimize specialist tools and capabilities
     - Validate specialist expertise alignment
```

#### Local Orchestrator Specialist Support:
```yaml
INSTALLATION_SUPPORT:
  1. local_integration_validation:
     - Verify specialists are accessible in local environment
     - Test routing and chain patterns with new specialists
     - Validate specialist precedence and conflict resolution
     - Ensure specialists work with existing local agents
  
  2. workflow_integration_testing:
     - Test common workflow patterns with new specialists
     - Validate chain coordination and handoff protocols
     - Ensure routing efficiency and optimization
     - Test specialist collaboration patterns
```

### Phase 3: Orchestrator Optimization

#### Global Orchestrator Enhancement Tasks:
```yaml
ORCHESTRATOR_SPECIALIST_UPDATES:
  1. routing_pattern_optimization:
     - Update routing algorithms for new specialists
     - Add project-specific routing rules and patterns
     - Optimize decision trees for detected technologies
     - Enhance chain coordination for specialist utilization
  
  2. specialist_registry_updates:
     - Add generated specialists to available agents list
     - Update precedence rules and conflict resolution
     - Define optimal chain patterns for each specialist
     - Create workflow templates utilizing new specialists
  
  3. collaboration_pattern_enhancement:
     - Define optimal multi-agent workflow patterns
     - Create specialist interaction protocols
     - Establish quality gates and validation patterns
     - Optimize resource allocation and efficiency
  
  4. project_customization_integration:
     - Add project-specific context and patterns
     - Include detected architectural considerations
     - Integrate workflow optimizations for this codebase
     - Embed technology-specific best practices
```

#### Local Orchestrator Specialist Receiving Updates:
```yaml
UPDATE_INTEGRATION:
  1. enhanced_routing_capabilities:
     - Receive updated routing algorithms and patterns
     - Integrate new specialist discovery and utilization
     - Update chain coordination for optimal workflows
     - Enhance decision-making for specialist selection
  
  2. specialist_coordination_optimization:
     - Learn optimal collaboration patterns with new specialists
     - Understand precedence rules and conflict resolution
     - Master efficient workflow orchestration
     - Execute enhanced quality assurance patterns
  
  3. project_specific_enhancement:
     - Integrate project-specific routing optimizations
     - Understand technology-specific workflow patterns
     - Master codebase-specific orchestration requirements
     - Execute optimized local development workflows
```

### Phase 4: Validation & Documentation

#### Shared Validation Responsibilities:
```yaml
SYSTEM_VALIDATION:
  1. specialist_functionality_testing:
     - Test each specialist individually for basic functionality
     - Validate specialist expertise and capability alignment
     - Ensure specialists respond appropriately to routing
     - Test specialist integration with existing agents
  
  2. orchestration_workflow_testing:
     - Test common workflow patterns end-to-end
     - Validate routing decisions and specialist selection
     - Test chain coordination and handoff protocols
     - Ensure optimal workflow efficiency and quality
  
  3. integration_validation:
     - Test global → local orchestrator collaboration
     - Validate specialist → orchestrator communication
     - Test multi-agent chain workflows
     - Ensure system scalability and performance
  
  4. documentation_and_completion:
     - Document all generated specialists and capabilities
     - Create usage guidance and best practices
     - Record workflow patterns and optimization results
     - Provide troubleshooting and maintenance guidance
```

## Communication Protocols

### Global ↔ Local Orchestrator Communication:
```yaml
COMMUNICATION_STANDARDS:
  1. analysis_sharing:
     format: "Structured analysis reports with findings and recommendations"
     timing: "After independent analysis, before collaborative synthesis"
     content: "Technology detection, specialist requirements, optimization needs"
  
  2. decision_coordination:
     format: "Collaborative decision-making with rationale documentation"
     timing: "During specialist planning and orchestrator enhancement"
     content: "Specialist priorities, routing patterns, workflow optimizations"
  
  3. progress_reporting:
     format: "Status updates with completion metrics and validation results"
     timing: "Throughout execution phases with milestone confirmations"
     content: "Specialist generation progress, installation status, optimization results"
```

### Error Handling and Escalation:
```yaml
ERROR_PROTOCOLS:
  1. specialist_generation_failures:
     - Global orchestrator retries with adjusted parameters
     - Local orchestrator provides additional context if needed
     - Escalate to user only if fundamental issues detected
  
  2. integration_conflicts:
     - Local orchestrator identifies and reports conflicts
     - Global orchestrator adjusts specialist configurations
     - Collaborative resolution with precedence rule updates
  
  3. performance_optimization_issues:
     - Both orchestrators collaborate on diagnosis
     - Iterative optimization with validation testing
     - Document lessons learned for future improvements
```

## Success Criteria and Quality Standards

### Phase Completion Standards:
```yaml
QUALITY_GATES:
  phase_1_completion:
    - Comprehensive codebase analysis completed by both orchestrators
    - Collaborative specialist requirements agreed upon
    - Execution plan documented and validated
  
  phase_2_completion:
    - All required specialists generated and installed
    - Specialist functionality validated and tested
    - Integration with existing agents confirmed
  
  phase_3_completion:
    - Orchestrator-specialist enhanced with new capabilities
    - Routing patterns optimized for new specialists
    - Workflow efficiency improvements validated
  
  phase_4_completion:
    - All workflows tested and validated
    - Documentation complete and accurate
    - System ready for production use
```

### Overall Success Metrics:
```yaml
SUCCESS_INDICATORS:
  1. specialist_coverage: "All detected technologies have appropriate specialists"
  2. routing_efficiency: "Optimal agent selection for common workflows"
  3. workflow_optimization: "Improved workflow execution with specialist chains"
  4. integration_quality: "Seamless collaboration between all agents"
  5. documentation_completeness: "Clear guidance for ongoing usage and maintenance"
```

## Training Validation

### Global Orchestrator Competency Checklist:
- [ ] Can perform comprehensive codebase analysis
- [ ] Can generate appropriate specialists using templates
- [ ] Can optimize local orchestrator-specialist effectively
- [ ] Can coordinate with local orchestrator collaboratively
- [ ] Can validate and document results comprehensively

### Local Orchestrator Specialist Competency Checklist:
- [ ] Can provide project-specific pattern analysis
- [ ] Can support specialist integration and validation
- [ ] Can execute enhanced routing and workflow patterns
- [ ] Can collaborate effectively with global orchestrator
- [ ] Can maintain optimized orchestration post-setup

## Ongoing Development and Maintenance

### Continuous Improvement Protocol:
```yaml
IMPROVEMENT_CYCLE:
  1. usage_monitoring:
     - Track workflow efficiency and specialist utilization
     - Identify optimization opportunities and gaps
     - Monitor specialist effectiveness and relevance
  
  2. pattern_evolution:
     - Detect new project patterns and requirements
     - Update specialist capabilities and routing rules
     - Enhance workflow patterns based on usage data
  
  3. collaboration_optimization:
     - Refine global ↔ local orchestrator collaboration
     - Improve specialist coordination and chain patterns
     - Enhance error handling and recovery protocols
```

This training ensures both orchestrators understand their roles, responsibilities, and collaboration protocols for successful `/crew:load` execution and ongoing optimization.