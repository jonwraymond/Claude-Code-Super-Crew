# Unified Agent System Documentation

## Overview

The Claude Code SuperCrew framework now features a unified agent system that combines behavioral personas with task-specific agents, all enhanced with visual identities for clear attribution and improved user experience.

## Agent Types

### 1. Persona Agents (Behavioral Modes)

Located in `~/.claude/agents/`, these are specialized behavioral modes that provide domain expertise:

| Emoji | Agent | Description | Activation Keywords |
|-------|-------|-------------|-------------------|
| 🏗️ | architect-persona | Systems design & architecture | architecture, design, scalability |
| 🎨 | frontend-persona | UX/UI & accessibility | component, responsive, ui |
| ⚙️ | backend-persona | Reliability & APIs | api, database, service |
| 🛡️ | security-persona | Threat modeling & compliance | vulnerability, threat, audit |
| 🔍 | analyzer-persona | Root cause analysis | analyze, investigate, debug |
| ⚡ | performance-persona | Optimization & speed | optimize, performance, speed |
| 🎯 | qa-persona | Quality & testing | test, quality, validation |
| 🔧 | refactorer-persona | Code quality & cleanup | refactor, cleanup, simplify |
| 🚀 | devops-persona | Infrastructure & deployment | deploy, infrastructure, ci/cd |
| 📚 | mentor-persona | Teaching & knowledge transfer | explain, learn, understand |
| ✍️ | scribe-persona | Documentation & writing | document, write, guide |

### 2. Project-Specific Agents (Task Tools)

Generated automatically based on project analysis and stored in `.claude/agents/`:

- **backend-specialist**: Language-specific backend development
- **frontend-specialist**: UI framework expertise
- **cli-specialist**: Command-line interface development
- **api-specialist**: API design and integration
- **installer-specialist**: Installation and deployment
- **testing-specialist**: Test suite development

## Visual Identity System

Each agent has a unique visual identity consisting of:

```yaml
visual_identity:
  emoji: "🏗️"              # Visual marker for quick identification
  background_color: "#1e3a8a"  # Agent-specific color (when supported)
  text_color: "#fbbf24"     # Contrasting text color
```

### Visual Attribution in Outputs

When agents contribute to outputs, they are clearly identified:

```
🏗️ [Architect]: I recommend a microservices architecture for this system...
🎨 [Frontend]: For the UI, we should implement a component-based design...
⚙️ [Backend]: The API should follow RESTful principles with proper versioning...
```

## Agent Activation

### 1. Automatic Activation

Agents activate based on:
- **Keyword Matching**: Primary, secondary, and contextual keywords
- **Project Context**: Current files, language, frameworks
- **Task Complexity**: Complexity scoring triggers appropriate agents
- **Command Context**: Specific commands activate relevant agents

### 2. Manual Activation

Use the Task tool with explicit agent specification:

```
Task(description="Design the system architecture", 
     prompt="Create a scalable microservices design",
     subagent_type="architect-persona")
```

### 3. Agent Chaining

Agents can hand off tasks to specialized agents:
- Architect → Backend: Design to implementation
- Analyzer → Refactorer: Problem identification to solution
- QA → Security: Testing to vulnerability assessment
- Mentor → Scribe: Knowledge to documentation

## Project Setup

### 1. Installing Personas

Persona agents are pre-installed in `~/.claude/agents/` and available globally.

### 2. Generating Project Agents

When you run `/crew:onboard` or initialize a project:

1. **Project Analysis**: Detects languages, frameworks, and project type
2. **Agent Generation**: Creates relevant project-specific agents
3. **Visual Enhancement**: Adds emojis and visual identities
4. **Activation Setup**: Configures keyword triggers

### 3. Agent Configuration

Each agent file follows this structure:

```markdown
---
name: agent-name
description: Agent purpose and expertise
version: "1.0.0"
created: "2025-01-28"
visual_identity:
  emoji: "🎨"
  background_color: "#7c3aed"
  text_color: "#fde68a"
activation_keywords:
  primary: ["main", "keywords"]
  secondary: ["related", "terms"]
  contextual: ["action phrases"]
tools: [Read, Write, Edit, ...]
---

# 🎨 Agent Name

[Agent prompt and instructions...]
```

## Orchestrator Agent

The 🎭 orchestrator-agent manages the entire agent ecosystem:

1. **Agent Discovery**: Finds and registers all available agents
2. **Project Analysis**: Determines which agents to activate
3. **Visual Coordination**: Manages agent visual identities
4. **Workflow Management**: Coordinates multi-agent operations

## Best Practices

### 1. Agent Selection
- Let automatic activation handle most cases
- Use manual activation for specific expertise needs
- Chain agents for complex multi-domain tasks

### 2. Visual Clarity
- Agents always identify themselves with their emoji
- Use visual markers consistently in outputs
- Maintain clear attribution for multi-agent responses

### 3. Project Integration
- Run `/crew:onboard` to set up project agents
- Keep project agents in `.claude/agents/`
- Personas are shared across all projects

### 4. Custom Agents
- Follow the visual identity pattern
- Choose unique emojis and colors
- Define clear activation keywords
- Document the agent's expertise

## Example Workflow

1. **Initialize Project**:
   ```
   /crew:onboard
   ```
   Output shows available personas and generated project agents.

2. **Complex Task with Multiple Agents**:
   ```
   User: "Design and implement a secure API with documentation"
   
   🏗️ [Architect]: I'll design the API structure...
   ⚙️ [Backend]: Implementing the endpoints...
   🛡️ [Security]: Adding authentication and security measures...
   ✍️ [Scribe]: Documenting the API...
   ```

3. **Specific Agent Request**:
   ```
   Task(subagent_type="performance-persona", 
        prompt="Optimize this database query")
   ```

## Troubleshooting

### Agent Not Activating
- Check activation keywords in agent file
- Verify agent file is in correct location
- Ensure proper YAML frontmatter format

### Visual Identity Not Showing
- Confirm visual_identity section in YAML
- Check for emoji support in terminal
- Verify agent file was generated with latest system

### Multiple Agents Responding
- This is intentional for multi-domain tasks
- Each agent provides specialized perspective
- Visual markers distinguish contributions

## Future Enhancements

- **Dynamic Visual Themes**: User-customizable color schemes
- **Agent Learning**: Agents adapt to project patterns
- **Cross-Agent Memory**: Shared context between agents
- **Visual Agent Gallery**: UI for browsing available agents
- **Agent Marketplace**: Community-contributed agents

---

The unified agent system transforms Claude Code into a multi-personality assistant, with each agent bringing specialized expertise and clear visual identity to your development workflow.