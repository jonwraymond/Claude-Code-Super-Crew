# Global vs Local Architecture - Claude Code SuperCrew

## Core Architecture Principle

**Global = Deterministic | Local = Intelligent**

## Global Components (Deterministic)

### Installation: `crew install`
All global components are installed deterministically via `crew install`:

```
~/.claude/
├── commands/          # Fixed global commands
├── agents/            # Fixed global personas
├── hooks/             # Fixed global hooks
├── SuperCrew/         # Fixed framework files
└── CLAUDE.md          # Fixed global configuration
```

### Global Personas (Fixed)
These 11 personas are always the same across all installations:
- architect-persona
- frontend-persona
- backend-persona
- security-persona
- analyzer-persona
- performance-persona
- qa-persona
- refactorer-persona
- devops-persona
- mentor-persona
- scribe-persona

### Global Commands (Fixed)
Standard commands that work the same everywhere:
- /analyze
- /build
- /implement
- /improve
- /test
- /document
- etc.

### Global Orchestrator Agent (Fixed)
The **global orchestrator agent** is a deterministic template that never changes:
- Located in framework: `SuperCrew/Agents/orchestrator.agent.md`
- Installed via: `crew install`
- Purpose: Serves as a template/guide for creating local orchestrators
- Never modified by Claude

## Local Components (Intelligent)

### Installation: `crew claude --install`
Local components are created intelligently by Claude:

```
project/.claude/
├── commands/          # Claude-created local commands
│   └── shadows/       # Claude-created shadow commands
├── agents/            # Claude-created specialists
│   └── orchestrator-specialist.md  # Claude-created from template
├── hooks/             # Claude-created local hooks
└── CLAUDE.md          # Claude-enhanced configuration
```

### Local Orchestrator-Specialist (Intelligent)
Created deterministically but customized intelligently:
1. **Triggered by**: `crew claude --install`
2. **Process**:
   - Framework prompts Claude to read global orchestrator template
   - Claude analyzes the specific project
   - Claude creates customized orchestrator-specialist
   - Includes project-specific routing rules
3. **Can be modified**: Claude can update it based on project evolution
4. **Always includes**: Core features like double/triple checking

### Local Specialists (Intelligent)
Created by Claude based on project needs:
- error-handling-specialist
- component-builder-specialist
- test-generator-specialist
- api-specialist
- migration-specialist
- etc.

### Local Commands (Intelligent)
Created by Claude for project-specific workflows:
- Can use any combination of agents
- Can shadow global commands
- Include project context and patterns

## The Key Distinction

### Global Framework (Deterministic)
```go
// crew install - Always the same
func InstallGlobalFramework() {
    // Copy fixed files from embedded resources
    copyGlobalPersonas()      // Always 11 personas
    copyGlobalCommands()      // Always same commands
    copyGlobalOrchestrator()  // Always same template
    copyGlobalHooks()         // Always same hooks
}
```

### Local Project (Intelligent)
```go
// crew claude --install - Claude-driven
func InstallLocalProject() {
    // Step 1: Create orchestrator deterministically
    promptClaudeToCreateOrchestrator() // From template
    
    // Step 2: Claude analyzes and customizes
    // Claude reads template
    // Claude analyzes project
    // Claude creates customized orchestrator
    
    // Step 3: Claude creates specialists as needed
    // Based on project patterns
    // Based on pain points
    // Based on workflows
}
```

## Workflow Example

### 1. Global Install (Deterministic)
```bash
$ crew install
Installing Claude Code Super Crew...
✓ Copying global personas (11 fixed)
✓ Copying global commands
✓ Copying global orchestrator template
✓ Copying global hooks
Done! Same for everyone.
```

### 2. Local Install (Intelligent)
```bash
$ crew claude --install
Creating project integration...

🎯 Step 1: Creating orchestrator-specialist
Claude, please:
1. Read global template at SuperCrew/Agents/orchestrator.agent.md
2. Analyze this project
3. Create customized orchestrator at .claude/agents/orchestrator-specialist.md

[Claude analyzes and creates...]

✓ Created project-specific orchestrator
✓ Ready for further customization
```

### 3. Project Evolution (Intelligent)
As the project grows, Claude can:
- Create new specialists
- Add local commands
- Update orchestrator routing rules
- Create shadow commands
- All while global components remain unchanged

## Summary

- **Global**: Fixed, deterministic, same for everyone
- **Local**: Intelligent, customized, project-specific
- **Orchestrator Template**: Global and fixed
- **Orchestrator-Specialist**: Local and intelligent (created from template)
- **Framework**: Provides structure
- **Claude**: Provides intelligence

This separation ensures consistency across installations while allowing infinite customization per project!