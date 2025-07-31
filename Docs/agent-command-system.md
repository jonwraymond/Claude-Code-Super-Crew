# Agent-Command System Documentation

## Overview

The Claude Code Super Crew framework features an enhanced agent-command system that automatically generates slash commands for each agent and intelligently routes commands to the most appropriate specialists.

## Key Features

### 1. Automatic Slash Command Generation

When agents are created, they automatically receive:
- **Primary commands** based on their specialization
- **Cross-promotion** to related agents
- **Visual routing** with emoji identifiers
- **Clear delegation syntax** for Claude Code

Example from go-backend-specialist:
```yaml
slash_commands:
  - name: "/crew:go"
    description: "Go development tasks - builds, tests, and optimizations"
    promotes: ["self", "qa-persona"]
  - name: "/crew:backend"
    description: "Backend development - APIs, services, and data processing"
    promotes: ["self", "api-specialist"]
  - name: "/crew:goroutine"
    description: "Concurrent programming with goroutines and channels"
    promotes: ["self", "performance-persona"]
```

### 2. Orchestrator-Specialist

The project-level orchestrator-specialist acts as the command router:

```markdown
🎯 orchestrator-specialist
- Routes /crew: commands to appropriate agents
- Suggests Task() calls with correct subagent_type
- Promotes agent discovery
- Coordinates multi-agent workflows
```

### 3. Command Routing Logic

When a slash command is executed:

1. **Command Analysis**: Parse intent and parameters
2. **Agent Matching**: Find best specialist for the task
3. **Visual Routing**: Show delegation with emojis
4. **Syntax Provision**: Provide exact Task() call

Example flow:
```
User: /crew:go
System: 🎯 [Orchestrator]: Delegating to ⚙️ go-backend-specialist. 
        Use: Task(subagent_type='go-backend-specialist', prompt='Handle Go task')
```

## Agent Command Mappings

### Project Specialists
| Agent | Emoji | Commands | Promotes |
|-------|-------|----------|----------|
| go-backend-specialist | ⚙️ | `/crew:go`, `/crew:backend`, `/crew:goroutine` | qa-persona, api-specialist, performance-persona |
| api-specialist | 🔌 | `/crew:endpoint`, `/crew:rest` | backend-persona, architect-persona |
| cli-specialist | 💻 | `/crew:command`, `/crew:flag` | frontend-persona |
| installer-specialist | 📦 | `/crew:install`, `/crew:setup` | devops-persona, scribe-persona |
| orchestrator-specialist | 🎯 | `/crew:orchestrate`, `/crew:agent-help` | all-agents |

### Global Personas
| Persona | Emoji | Suggested Commands |
|---------|-------|-------------------|
| architect-persona | 🏗️ | `/crew:design`, `/crew:architecture` |
| frontend-persona | 🎨 | `/crew:ui`, `/crew:component` |
| backend-persona | ⚙️ | `/crew:api`, `/crew:service` |
| security-persona | 🛡️ | `/crew:audit`, `/crew:secure` |
| analyzer-persona | 🔍 | `/crew:analyze`, `/crew:debug` |
| performance-persona | ⚡ | `/crew:optimize`, `/crew:profile` |
| qa-persona | 🎯 | `/crew:test`, `/crew:validate` |
| refactorer-persona | 🔧 | `/crew:cleanup`, `/crew:refactor` |
| devops-persona | 🚀 | `/crew:deploy`, `/crew:build` |
| mentor-persona | 📚 | `/crew:explain`, `/crew:teach` |
| scribe-persona | ✍️ | `/crew:document`, `/crew:readme` |

## Implementation Details

### Agent Generator Enhancement

The `agent_generator.go` now includes:

```go
// getSlashCommands returns slash commands that promote this agent
func (ag *AgentGenerator) getSlashCommands(agentType string, chars *ProjectCharacteristics) string {
    // Generate commands based on agent type
    // Include self-promotion and cross-agent promotion
    // Return YAML-formatted command list
}
```

### Slash Command Handler

The command handler in `slash_commands.go` now:

```go
// suggestAgentForCommand analyzes command and suggests appropriate agent
func (r *SlashCommandRegistry) suggestAgentForCommand(name string, args []string) string {
    // Route based on command patterns
    // Return orchestrator delegation message
    // Include Task() syntax
}
```

## Usage Examples

### 1. Direct Agent Commands

```bash
/crew:go
# Output: 🎯 [Orchestrator]: Delegating to ⚙️ go-backend-specialist...

/crew:optimize
# Output: 🎯 [Orchestrator]: Delegating to ⚡ performance-persona...
```

### 2. Agent Discovery

```bash
/crew:agent-help
# Output: Lists all available agents with emojis and specialties
```

### 3. Multi-Agent Coordination

```bash
/crew:orchestrate "build and test API"
# Output: 🎯 [Orchestrator]: Coordinating ⚙️ backend + 🔌 api + 🎯 qa agents...
```

## Best Practices

### For Agent Creation

1. **Define Meaningful Commands**: Choose commands that reflect agent expertise
2. **Cross-Promote Wisely**: Link to complementary agents
3. **Use Clear Descriptions**: Help users understand command purpose
4. **Include Visual Identity**: Always use the agent's emoji

### For Command Usage

1. **Start with Orchestrator**: Let it route to the right agent
2. **Use Suggested Syntax**: Copy the Task() call provided
3. **Chain Agents**: Some tasks benefit from multiple specialists
4. **Discover via Help**: Use `/crew:agent-help` to explore

## Future Enhancements

### Planned Features

1. **Dynamic Command Registration**: Agents register commands at runtime
2. **Command Aliases**: Multiple ways to invoke same agent
3. **Context-Aware Routing**: Consider current files/task
4. **Command History**: Learn from usage patterns
5. **Interactive Menus**: Choose from agent suggestions

### Integration Points

- **IDE Integration**: VSCode/Cursor command palette
- **CLI Completion**: Tab completion for commands
- **Chat Integration**: Natural language to command mapping
- **Workflow Automation**: Command chains and scripts

## Troubleshooting

### Common Issues

**Command Not Recognized**
- Check if agent has the command in slash_commands
- Verify agent is properly generated
- Use `/crew:agent-help` to see available commands

**Wrong Agent Selected**
- Be more specific with command
- Use agent name directly
- Check command mappings in orchestrator

**No Visual Identity**
- Regenerate agent with latest generator
- Ensure visual_identity in YAML frontmatter
- Check emoji support in terminal

---

The agent-command system transforms slash commands into intelligent agent delegation, making it easy to leverage the right specialist for every task. Each command promotes agent discovery and provides clear visual feedback about which expert is handling your request.