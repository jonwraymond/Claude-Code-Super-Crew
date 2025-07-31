---
name: ast-grep
description: "Advanced code search and transformation using AST-based pattern matching"
version: 1.0.0
tools: [ast-grep]
dependencies: []
enabled-by-default: false
---

# AST-Grep Add-on

The ast-grep add-on provides powerful Abstract Syntax Tree (AST) based code search and transformation capabilities for SuperCrew commands.

## Capabilities

- **AST-based Search**: Find code patterns using AST structure instead of text
- **Semantic Code Analysis**: Understand code semantics beyond syntax
- **Automated Refactoring**: Apply safe code transformations
- **Pattern Matching**: Define complex code patterns with precision
- **Multi-language Support**: Support for JavaScript, TypeScript, Python, Go, and more

## Usage

When enabled, this add-on provides the `ast-grep` tool that can be used in commands to:

1. Search for specific code patterns using AST queries
2. Perform semantic code analysis across the codebase
3. Apply automated refactoring based on patterns
4. Detect code smells and anti-patterns
5. Generate code transformation reports

## Integration

This add-on is particularly useful for:
- Large-scale code refactoring projects
- Code quality analysis and improvement
- Legacy code modernization
- Consistency enforcement across codebases
- Security vulnerability detection

## Configuration

The add-on can be enabled in command frontmatter:
```yaml
add-ons: [ast-grep]
```

## Pattern Examples

```yaml
# Find all console.log statements
pattern: console.log($$$ARGS)

# Find unused variables
pattern: |
  const $VAR = $INIT
  // $VAR is never used

# Find React components without prop types
pattern: |
  function $COMPONENT($PROPS) { $$$BODY }
  // Missing prop types definition