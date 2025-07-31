# SuperCrew Add-ons

This directory contains modular add-on components that extend the core SuperCrew functionality. Add-ons provide optional tools and capabilities that can be selectively enabled based on project needs.

## Structure

```
SuperCrew/addons/
├── README.md                    # This file
├── code2prompt/                 # Code2Prompt add-on
│   ├── addon.md                 # Add-on definition
│   └── templates/               # Optional templates
├── ast-grep/                    # AST-Grep add-on
│   ├── addon.md                 # Add-on definition
│   └── patterns/                # Optional patterns
└── registry/                    # Add-on registry management
    └── registry.go              # Registry implementation
```

## Add-on Format

Each add-on is defined by an `addon.md` file with the following structure:

```yaml
---
name: add-on-name
description: Brief description of the add-on's purpose
version: 1.0.0
tools: [tool1, tool2]           # List of tools provided
dependencies: [dep1, dep2]      # Optional dependencies
enabled-by-default: false       # Whether enabled by default
---
```

## Usage

Add-ons are referenced in command frontmatter using the `add-ons:` field:

```yaml
---
add-ons: [code2prompt, ast-grep]
---
```

## Creating New Add-ons

1. Create a new directory under `SuperCrew/addons/`
2. Add an `addon.md` file with the required metadata
3. Implement any necessary templates or patterns
4. Update the registry if needed