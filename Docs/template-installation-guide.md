# Claude Code SuperCrew Template Installation Guide

## Overview

Claude Code SuperCrew uses **two distinct types of templates** that are installed in **different locations** using **different commands**. Understanding this separation is crucial for proper usage.

## Template Types and Installation Locations

### 1. Generic Persona Template (Global Installation)

**Location**: `~/.claude/agents/generic-persona-template.md`
**Installation Command**: `crew install`
**Scope**: System-wide, available across all projects

#### Purpose
- **Cross-project utility agent** applicable to all projects
- **General-purpose capabilities** that work across different technologies
- **Consistent global personas** available in any Claude Code session
- **User-level agents** that maintain consistency across projects

#### Use Cases
- Project managers, documentation specialists, mentors
- Quality assurance roles that apply universally
- Analysis and review personas that work across domains
- Cross-technology collaboration agents

#### Installation Process
```bash
# Install global framework (includes generic persona template)
./crew install

# Verify installation
ls ~/.claude/agents/generic-persona-template.md
```

### 2. Generic Specialist Template (Project-Level Installation)

**Location**: `{project}/.claude/agents/templates/generic-specialist-template.md`
**Installation Command**: `crew claude --install`
**Scope**: Project-specific, customized per codebase

#### Purpose
- **Domain-specific specialists** tailored to project technology stack
- **Project-aware agents** that understand local patterns and architecture
- **Technology-focused expertise** (Go, React, Python, etc.)
- **Isolated configurations** per project for different tech stacks

#### Use Cases
- Backend specialists (Go, Python, Node.js)
- Frontend specialists (React, Vue, Angular)
- Database specialists (PostgreSQL, MongoDB)
- DevOps specialists (Docker, Kubernetes, AWS)
- Language-specific experts for each project

#### Installation Process
```bash
# First, ensure global framework is installed
./crew install

# Navigate to your project
cd /path/to/your/project

# Install project-level integration (includes specialist template)
./crew claude --install

# Verify installation
ls .claude/agents/templates/generic-specialist-template.md
```

## Command Distinction

### `crew install` (Global Framework Installation)

**What it does:**
- Installs SuperCrew framework to `~/.claude/`
- Sets up global commands, personas, and infrastructure
- Installs **GenericPersonaTemplate.md** for cross-project agents
- Creates system-wide hooks and automation
- **Run once per system**

**Directory Structure Created:**
```
~/.claude/
├── agents/
│   ├── analyzer-persona.md
│   ├── architect-persona.md
│   ├── generic-persona-template.md          ← Global persona template
│   ├── [other global personas]
│   └── templates/
│       └── [hooks, not agent templates]
├── commands/
│   └── [global slash commands]
├── hooks/
│   └── [system hooks]
└── [framework infrastructure]
```

### `crew claude --install` (Project Integration)

**What it does:**
- Creates project-specific `.claude/` directory
- Installs **generic-specialist-template.md** for project agents
- Sets up project-specific orchestrator
- Enables `/crew:` commands for this project only
- **Run once per project**

**Directory Structure Created:**
```
your-project/
└── .claude/
    ├── agents/
    │   └── templates/
    │       └── generic-specialist-template.md  ← Project specialist template
    ├── commands/
    ├── hooks/
    ├── project-config.json
    └── supercrew-commands.json
```

## Template Usage Examples

### Creating a Global Persona (Using generic-persona-template.md)

```bash
# Copy the global template
cp ~/.claude/agents/generic-persona-template.md ~/.claude/agents/project-manager-persona.md

# Edit to create a cross-project project manager
# This persona will be available in ALL projects
```

### Creating a Project Specialist (Using generic-specialist-template.md)

```bash
# Navigate to your Go project
cd /path/to/go-project

# Copy the project template
cp .claude/agents/templates/generic-specialist-template.md .claude/agents/go-backend-specialist.md

# Edit to create a Go-specific specialist
# This specialist is only available in THIS project
```

## Installation Workflow

### Initial Setup (Once Per System)
```bash
# 1. Install global framework
./crew install

# 2. Verify global installation
./crew status
```

### Per-Project Setup (Once Per Project)
```bash
# 1. Navigate to project
cd /path/to/your/project

# 2. Install project integration
./crew claude --install

# 3. Verify project installation
./crew claude --status
```

## Template Customization Guidelines

### Global Persona Template Guidelines
- **Technology-agnostic**: Should work across all tech stacks
- **Process-focused**: Emphasize workflows and methodologies
- **Cross-functional**: Designed for collaboration between projects
- **Stable**: Changes infrequently, affects all projects

### Project Specialist Template Guidelines  
- **Technology-specific**: Tailored to project's tech stack
- **Context-aware**: Understands project architecture and patterns
- **Isolated**: Changes only affect the current project
- **Evolving**: Can be updated as project needs change

## Best Practices

### Do's
✅ Use global personas for roles that span multiple projects
✅ Use project specialists for technology-specific expertise  
✅ Keep global templates technology-agnostic
✅ Customize project templates for specific tech stacks
✅ Run `crew install` before `crew claude --install`

### Don'ts
❌ Don't put technology-specific logic in global personas
❌ Don't put cross-project utilities in project specialists
❌ Don't modify templates directly - copy them first
❌ Don't skip the global installation step

## Troubleshooting

### "Framework not installed" error
```bash
# Solution: Install global framework first
./crew install
```

### Missing template files
```bash
# Check global installation
ls ~/.claude/agents/generic-persona-template.md

# Check project installation  
ls .claude/agents/templates/generic-specialist-template.md

# Reinstall if missing
./crew install --force               # Global
./crew claude --install --force      # Project
```

### Template in wrong location
- generic-persona-template.md should ONLY be in `~/.claude/agents/`
- generic-specialist-template.md should ONLY be in `{project}/.claude/agents/templates/`

## Summary

| Template Type | Location | Command | Scope | Purpose |
|---------------|----------|---------|-------|---------|
| **Generic Persona** | `~/.claude/agents/` | `crew install` | Global | Cross-project utilities |
| **Generic Specialist** | `{project}/.claude/agents/templates/` | `crew claude --install` | Project | Domain-specific experts |

This separation ensures that global utilities remain consistent across projects while allowing each project to have its own specialized agents tailored to its specific technology stack and requirements.