---
name: code2prompt
description: "Generate context-rich prompts from codebases using code2prompt tool"
version: 1.0.0
tools: [code2prompt]
dependencies: []
enabled-by-default: false
---

# Code2Prompt Add-on

The code2prompt add-on provides intelligent code analysis and prompt generation capabilities for SuperCrew commands.

## Capabilities

- **Code Context Generation**: Extract relevant code snippets with full context
- **Repository Analysis**: Analyze entire repositories for patterns and structure
- **Prompt Optimization**: Generate optimized prompts for AI interactions
- **Multi-language Support**: Support for various programming languages

## Usage

When enabled, this add-on provides the `code2prompt` tool that can be used in commands to:

1. Generate comprehensive code context for AI interactions
2. Create structured prompts from existing codebases
3. Analyze code patterns and dependencies
4. Extract relevant documentation from source code

## Integration

This add-on is particularly useful for:
- Project onboarding and understanding
- Code review and analysis workflows
- Documentation generation
- Legacy codebase exploration

## Configuration

The add-on can be enabled in command frontmatter:
```yaml
add-ons: [code2prompt]