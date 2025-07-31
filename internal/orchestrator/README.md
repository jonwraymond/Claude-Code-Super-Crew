# Orchestrator Agent Auto-Trigger System

The orchestrator system automatically generates project-specific subagents when the `/crew:onboard` command is executed in a project directory.

## Overview

When a user runs `/crew:onboard` (or variations like `crew:load`, `crew: init`) in a project:

1. The orchestrator agent analyzes the repository to detect:
   - Programming languages (Go, JavaScript, Python, etc.)
   - Frameworks (React, Vue, Django, etc.)
   - Project structure and patterns
   - Infrastructure components (Docker, Kubernetes)

2. Based on the analysis, it generates specialized agent configurations in the **local** `.claude/agents/` directory (not the global `~/.claude/agents/`)

3. Each generated agent is customized for the specific project context with:
   - Project-specific expertise
   - Relevant tools and capabilities
   - Integration with other project agents
   - Awareness of project conventions

## Implementation Details

### Project Analysis (`project_analyzer.go`)

The `ProjectAnalyzer` detects:
- **Languages**: Go, JavaScript/TypeScript, Python, Rust, Java
- **Frameworks**: React, Vue, Angular, Django, Express, etc.
- **Infrastructure**: Docker, Kubernetes, CI/CD pipelines
- **Project Characteristics**: Backend/Frontend, Database, Testing

### Agent Generation (`agent_generator.go`)

The `AgentGenerator` creates specialized agents:
- `go-backend-specialist` - For Go backend projects
- `react-frontend-specialist` - For React applications
- `node-backend-specialist` - For Node.js servers
- `python-backend-specialist` - For Python backends
- `devops-specialist` - For infrastructure projects
- `database-specialist` - When databases are detected
- `qa-specialist` - When testing frameworks are found
- `api-specialist` - For API-focused projects

### Integration (`slash_integration.go`)

The orchestrator integrates with the `/crew:onboard` command:
- Triggers automatically on project load
- Works with any project directory
- Logs progress and results
- Handles errors gracefully

## Usage

1. Navigate to a project directory:
   ```bash
   cd /path/to/my/project
   ```

2. Run the load command:
   ```
   /crew:onboard
   ```

3. The orchestrator will:
   - Analyze the project
   - Generate appropriate agents in `.claude/agents/`
   - Log what was created

## Generated Agent Structure

Each generated agent includes:

```yaml
---
name: <specialist-name>
description: <role description>
version: "1.0.0"
project: <project-name>
language: <primary-language>
frameworks: [<detected-frameworks>]
tags: [<relevant-tags>]
tools: [<available-tools>]
---

[Agent prompt with project-specific context and expertise]
```

## File Placement

All project-specific agents are written to:
```
<project-directory>/.claude/agents/
```

The global orchestrator agent remains at:
```
~/.claude/agents/orchestrator-agent.md
```

## Testing

The orchestrator includes comprehensive tests:
- `project_analyzer_test.go` - Tests project detection logic
- `agent_generator_test.go` - Tests agent generation
- Various project scenarios covered

## Future Enhancements

- Support for more languages and frameworks
- Custom agent templates
- Interactive agent customization
- Team-specific agent configurations
- Agent version management