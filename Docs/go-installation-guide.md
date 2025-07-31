# Claude Code SuperCrew Go Implementation - Installation Guide 🚀

## Overview

The Go implementation of Claude Code SuperCrew provides a fast, single-binary solution that integrates seamlessly with Claude Code. The installation is a two-step process: global framework installation and project-specific integration.

## Why Two Steps?

The two-step installation provides important benefits:

1. **Global Framework** (`crew install`): Installed once per system, contains all the core SuperCrew commands and personas that are shared across all your projects.

2. **Project Integration** (`crew claude --install`): Enabled per project, allowing each project to have:
   - Its own orchestrator-specialist tuned to the project
   - Project-specific agents created on demand
   - Isolated configuration and agent management
   - Different projects can use different agent configurations

This separation means you can work on multiple projects with different tech stacks, and each gets its own tailored agent setup while sharing the common framework.

## Prerequisites

- **Go 1.21+** (for building from source)
- **Claude Code** (the framework integrates with Claude Code)
- **macOS, Linux, or Windows**

## Installation Process

### Step 1: Global Framework Installation (Once Per System)

First, install the SuperCrew framework globally:

```bash
# Download and run the crew binary
# (Replace with actual download URL when available)
curl -sSL https://github.com/jonwraymond/claude-code-super-crew/releases/latest/download/crew -o crew
chmod +x crew

# Install framework globally
./crew install
```

This will install:
- ✅ SuperCrew framework files to `~/.claude/`
- ✅ Core command definitions
- ✅ Global personas and agents
- ✅ Framework infrastructure

### Step 2: Project-Level Integration (Per Project)

Navigate to each project where you want to use SuperCrew:

```bash
# Navigate to your project
cd /path/to/your/project

# Enable Claude integration for this project
./crew claude --install
```

This will:
- ✅ Create `.claude/agents/` in your project
- ✅ Install orchestrator-specialist for the project
- ✅ Enable `/crew:` commands for this project
- ✅ Set up project-specific agent configuration

### Build from Source

If you prefer to build from source:

```bash
# Clone the repository
git clone https://github.com/jonwraymond/claude-code-super-crew.git
cd claude-code-super-crew

# Build the binary
make build

# Install framework globally
./crew install

# Then enable for your project
cd /path/to/your/project
/path/to/crew claude --install
```

## What Gets Installed

### Global Installation (`crew install`)

Creates the following in `~/.claude/`:

```
~/.claude/
├── SuperCrew/
│   ├── Commands/        # Slash command definitions
│   ├── Core/           # Framework core files
│   └── Hooks/          # Event hooks (optional)
└── claude-integration/  # Claude Code integration files
```

### Project Installation (`crew claude --install`)

Creates the following in your project:

```
your-project/
└── .claude/
    └── agents/
        └── orchestrator-specialist.md  # Project-specific orchestrator
```

## Post-Installation

After installation:

1. **Restart Claude Code** to load the new commands
2. **Test the installation** by typing `/crew:` and pressing Tab
3. **Run `/crew:onboard`** to analyze your project and set up orchestration

## Available Commands

Once installed, you'll have access to these `/crew:` commands:

- `/crew:help` - Show available agents and commands
- `/crew:onboard` - Analyze project and set up orchestration
- `/crew:analyze` - Deep project analysis
- `/crew:build` - Build workflows
- `/crew:implement` - Implementation assistance
- `/crew:orchestrate` - Complex multi-agent coordination
- And many more...

## Managing Your Installation

### Check Status
```bash
# Shows both global and project status
./crew claude --status
```

### Update Installation
```bash
./crew claude --update
```

### List Available Commands
```bash
./crew claude --list
```

### Uninstall
```bash
# Remove project integration
./crew claude --uninstall

# Remove global framework
./crew uninstall
```

## Troubleshooting

### Installation Failed
- Ensure you have write permissions to `~/.claude/`
- Check that Claude Code is installed
- Use `--verbose` flag for detailed output: `./crew claude --install --verbose`

### Commands Not Working
- Restart Claude Code after installation
- Verify installation with `./crew claude --status`
- Check that files exist in `~/.claude/SuperCrew/`

### Custom Installation Directory
To install to a different directory:
```bash
./crew install --install-dir /path/to/custom/dir
```

## Key Differences from Python Version

The Go implementation offers:
- **Single Binary**: No Python dependencies or virtual environments
- **Faster Execution**: Native performance for all operations
- **Two-Step Installation**: Clear separation between global framework and project integration
- **Multi-Project Support**: Enable SuperCrew for multiple projects independently
- **Project Isolation**: Each project has its own agents and configuration

## Next Steps

After installation:
1. Navigate to your project directory
2. Run `/crew:onboard` in Claude Code to analyze your project
3. Use `/crew:help` to explore available agents
4. Start using specialized agents for your tasks!

## Support

For issues or questions:
- Check the [troubleshooting guide](./troubleshooting.md)
- Report issues at: https://github.com/jonwraymond/claude-code-super-crew/issues
- Join the community discussions

---

The Go implementation simplifies the Claude Code SuperCrew experience while maintaining all the powerful features of the framework. Happy coding! 🎯