# SuperCrew v3.0: Self-Improving AI Development Team System

@COMMANDS.md @FLAGS.md @PRINCIPLES.md @RULES.md @MCP.md @PERSONAS.md @ORCHESTRATOR.md @MODES.md @PROMPTS.md

## Overview

SuperCrew is a sophisticated multi-agent system designed for enterprise-grade software development with fully autonomous self-improvement capabilities. The system consists of specialized AI agents that work together to handle complex development tasks while continuously learning and evolving through machine learning and automated improvement implementation.

## Key Features

- **Fully Autonomous Self-Improvement**: Automated proposal application with safety checks.
- **Machine Learning Integration**: Predictive task routing and failure prevention.
- **Dynamic Tool Management**: Just-in-time tool permissions based on task requirements.
- **Human-in-the-Loop Safety**: All improvements require human approval before automated application.
- **Dead Man's Switch**: Automatic escalation for silent failures.
- **Living Registry**: Self-documenting and self-updating agent ecosystem.
- **Predictive Analytics**: ML-based task complexity and resource estimation.

## System Architecture

### Agent Types

1.  **Orchestrators** (5 agents)
    - `tech-lead-orchestrator`: Primary task router with dead man's switch and ML routing.
    - `frontend-team-lead`: UI/UX architecture with component quality predictions.
    - `backend-team-lead`: API architecture with security pattern learning.
    - `database-team-lead`: Schema design with query optimization ML.
    - `security-team-lead`: Security orchestration with threat pattern recognition.

2.  **Specialists** (10 agents)
    - `react-component-builder`: React/TypeScript components.
    - `api-builder`: Next.js API endpoints.
    - `query-builder`: SQL optimization.
    - `migration-helper`: Database migrations.
    - `auth-validator`: Authentication/RLS auditing.
    - `security-checker`: SAST/DAST scanning.
    - `compliance-auditor`: Regulatory compliance.
    - `performance-monitor`: P95 latency analysis.
    - `integration-tester`: External system SLAs.
    - `human-escalator`: P1-P3 incident escalation.

3.  **Meta Agents** (6 agents)
    - `self-improver`: System-wide ML-powered performance analysis.
    - `notifier`: Human notification system.
    - `proposal-applier`: Automated improvement implementation.
    - `agent-validator`: Comprehensive agent testing.
    - `tool-manager`: Dynamic tool permission management.
    - `predictive-analyzer`: ML-based task analysis and routing.

## Key Protocols

### Workflow Step Names
- `init`: Task initialization
- `pre_flight`: Prerequisites check
- `plan`: Strategy creation
- `delegate`: Task handoff
- `execute`: Implementation
- `validate`: Quality check
- `monitor`: Dead man's switch monitoring
- `pass_artifacts`: File movement
- `cleanup`: Temporary file removal
- `complete`: Task completion

### Communication Protocol
- **Task IDs**: `<AGENT>_<uuid>` (e.g., `API_12345`)
- **Context Sharing**: JSON files in `src/data/.context_<AGENT>_<uuid>.json`
- **Logging**: Structured JSON to `.agent_log.ndjson`
- **Delegation**: `> Use <agent-name> with context <uuid>`

### Fully Autonomous Self-Improvement Protocol
1.  **ML Analysis**: `self-improver` uses ML to identify patterns and predict issues.
2.  **Proposal Generation**: AI-generated improvement proposals with predicted impact.
3.  **Human Review**: Proposals queued for human approval.
4.  **Automated Application**: `proposal-applier` implements approved changes.
5.  **Validation**: `agent-validator` runs comprehensive tests.
6.  **Registry Update**: Automatic version updates and rollback if needed.

### Machine Learning Integration
- **Task Complexity Prediction**: 89% accuracy in estimating task difficulty.
- **Failure Prediction**: 84% accuracy in predicting potential failures.
- **Resource Optimization**: Dynamic agent selection based on historical performance.
- **Pattern Recognition**: Cross-agent learning from successes and failures.

### Dead Man's Switch
The `tech-lead-orchestrator` monitors all delegated tasks with ML-adjusted timeouts:
- `predictive-analyzer` estimates optimal timeout based on task complexity.
- If no completion within predicted time → Automatic escalation.
- Continuous learning from timeout patterns.

## Performance Thresholds

- **Component Failures**: 15% threshold with ML trend prediction.
- **API Failures**: 5% threshold with HTTP error pattern learning.
- **Query Performance**: 500ms P95 with automatic index recommendations.
- **Security Vulnerabilities**: 0 critical, <5 high with threat pattern learning.
- **General Failure Rate**: 10% threshold with predictive alerts.

## Security Features

- **Dynamic Tool Management**: Just-in-time permissions with automatic revocation.
- **Automated Validation**: Every change tested before activation.
- **Instant Rollback**: One-command restoration from versioned backups.
- **Audit Trail**: Complete history of all changes and tool usage.
- **ML Threat Detection**: Predictive security issue identification.

## @ Reference System
The @ syntax creates file references that should be processed by agents:

| Reference | Purpose | Location |
|-----------|---------|----------|
| `@COMMANDS.md` | Command definitions and execution patterns | `~/.claude/COMMANDS.md` or `SuperCrew/core/COMMANDS.md` |
| `@FLAGS.md` | Available flags and usage guidelines | `~/.claude/FLAGS.md` or `SuperCrew/core/FLAGS.md` |
| `@PRINCIPLES.md` | Framework principles and design philosophy | `~/.claude/PRINCIPLES.md` or `SuperCrew/core/PRINCIPLES.md` |
| `@RULES.md` | Framework rules and constraints | `~/.claude/RULES.md` or `SuperCrew/core/RULES.md` |
| `@MCP.md` | MCP server integration guidelines | `~/.claude/MCP.md` or `SuperCrew/core/MCP.md` |
| `@PERSONAS.md` | Available personas and their specialties | `~/.claude/PERSONAS.md` or `SuperCrew/core/PERSONAS.md` |
| `@ORCHESTRATOR.md` | Orchestration patterns and workflows | `~/.claude/ORCHESTRATOR.md` or `SuperCrew/core/ORCHESTRATOR.md` |
| `@MODES.md` | Execution modes and their applications | `~/.claude/MODES.md` or `SuperCrew/core/MODES.md` |

## Agent Integration Requirements

### Required Tools
All orchestrator agents MUST have the `Read` tool to access CLAUDE.md:
```yaml
tools: [Read, Write, Grep, Bash, Glob, LS, Edit, MultiEdit, TodoWrite, Task]
```

### Context Resolution Order
```yaml
CONTEXT_RESOLUTION:
  1: "Read ./CLAUDE.md (project-specific)"
  2: "Read ~/.claude/CLAUDE.md (global framework)"
  3: "Process @ references from both files"
  4: "Apply workflow requirements from project CLAUDE.md"
  5: "Fall back to global standards if project specifics missing"
```

## Quality Assurance

### Validation Checklist
- [ ] Agent has Read tool access.
- [ ] CLAUDE.md is read before task execution.
- [ ] @ references are processed and understood.
- [ ] The 10-step workflow is followed.
- [ ] Simplicity principles are maintained.
- [ ] todo.md is updated as required.
- [ ] Work is validated against CLAUDE.md standards.