# Claude-Driven Project Analysis

## Overview

The Claude Code SuperCrew framework now uses Claude's intelligence to analyze projects rather than rigid Go code patterns. This provides maximum flexibility and adaptability.

## How It Works

### 1. Template Creation (Go Code)
When `/crew:onboard` is executed, the Go code simply:
- Installs/validates the orchestrator-specialist
- Creates an empty `project-analysis.json` template
- Provides instructions for Claude

### 2. Claude Analysis (AI Intelligence)
Claude then analyzes the project using available tools:
- **Glob**: Find all source files and count by extension
- **Read**: Check for framework indicators (go.mod, package.json, etc.)
- **Grep**: Search for architectural patterns
- **LS**: Assess project structure and complexity

### 3. Analysis Template Structure

```json
{
  "analysis_version": "2.0",
  "analyzed_by": "claude",
  "analysis_date": "2025-01-29",
  "project_path": "/path/to/project",
  
  "languages": [
    {
      "name": "Go",
      "file_count": 42,
      "primary": true,
      "patterns_observed": ["error handling", "goroutines", "interfaces"]
    }
  ],
  
  "frameworks": [
    {
      "name": "Go Modules",
      "type": "dependency-manager",
      "version": "1.21",
      "usage_patterns": ["internal packages", "vendor directory"]
    }
  ],
  
  "architectural_patterns": [
    {
      "pattern": "CLI Application",
      "confidence": "high",
      "evidence": ["cmd directory", "cobra usage", "flag parsing"],
      "implications": ["command structure important", "user interaction patterns"]
    }
  ],
  
  "complexity_assessment": {
    "overall_complexity": "complex",
    "factors": {
      "size": "large",
      "domain_count": 3,
      "integration_points": 5,
      "architectural_layers": 4
    },
    "orchestration_benefit": "high"
  },
  
  "specialist_recommendations": [
    {
      "specialist_name": "cli-specialist",
      "reason": "Heavy CLI command patterns detected (15+ command files)",
      "priority": "medium",
      "trigger_conditions": ["repeated command implementation", "complex flag handling"]
    }
  ],
  
  "usage_patterns": {
    "detected_workflows": ["build-test-deploy", "command-generation"],
    "common_tasks": ["agent creation", "command routing", "orchestration"],
    "pain_points": ["complex command parsing", "multi-agent coordination"]
  }
}
```

## Key Principles

### 1. Intelligence Over Automation
- Claude's judgment determines what's important
- No hardcoded assumptions about project types
- Flexible interpretation of patterns

### 2. Conservative Recommendations
- Only suggest specialists for **repeated** patterns
- Require evidence of complexity before recommending
- Personas often sufficient for most tasks

### 3. Contextual Understanding
- Claude can identify subtle patterns automated code would miss
- Considers project evolution and user needs
- Adapts to unique project characteristics

## Analysis Guidelines for Claude

### Language Detection
```
✓ Count actual files, not just presence
✓ Identify primary language (most files)
✓ Note language-specific patterns (not just extension)
✗ Don't assume importance from file count alone
```

### Framework Detection
```
✓ Check dependency files (go.mod, package.json, etc.)
✓ Read configuration to understand usage
✓ Note version information when available
✗ Don't assume framework defines architecture
```

### Pattern Recognition
```
✓ Look for consistent usage across multiple files
✓ Consider confidence levels (low/medium/high)
✓ Document evidence for patterns found
✗ Don't overinterpret isolated examples
```

### Specialist Recommendations
```
✓ Require 5+ files showing same pattern
✓ Ensure pattern is complex enough to warrant specialist
✓ Include clear trigger conditions
✗ Don't recommend for every detected pattern
✗ Don't create specialists personas can handle
```

## Benefits of Claude-Driven Analysis

### 1. Adaptability
- Handles any project type or structure
- Evolves with new languages and frameworks
- No need to update Go code for new patterns

### 2. Intelligence
- Understands context and nuance
- Makes judgment calls about importance
- Considers user's actual needs

### 3. Simplicity
- Go code just creates template
- Claude does the complex analysis
- Easy to understand and modify

## Usage Flow

```
User: /crew:onboard
  ↓
Go Code: Creates orchestrator + empty template
  ↓
Claude: Analyzes project with tools
  ↓
Claude: Fills out project-analysis.json
  ↓
Orchestrator: Uses analysis for routing
  ↓
User: Benefits from intelligent orchestration
```

## Example Analysis Process

```bash
# Claude's analysis workflow:

1. Language Analysis:
   Glob "**/*.go" → 42 files found
   Glob "**/*.py" → 1 file found
   → Primary: Go, Secondary: Python

2. Framework Check:
   Read "go.mod" → Go Modules detected
   Read "Makefile" → Build automation present

3. Pattern Search:
   Grep "cobra.Command" → CLI framework usage
   Grep "handler|endpoint" → API patterns
   → Patterns: CLI Application, API Development

4. Complexity Assessment:
   LS -R → 96 files, 4 directory levels
   → Size: large, Complexity: high

5. Recommendations:
   CLI pattern in 15+ files → Maybe cli-specialist
   API pattern in 8 files → Watch for growth
   → Conservative: Personas sufficient for now
```

## Best Practices

### For Claude
1. Always analyze thoroughly before recommending
2. Be conservative with specialist suggestions
3. Document evidence for findings
4. Consider evolution - what patterns are emerging?

### For Users
1. Run `/crew:onboard` when starting a project
2. Let patterns emerge naturally
3. Request specialists explicitly if needed
4. Trust orchestrator routing recommendations

### For Developers
1. Keep Go code simple - just templates
2. Let Claude handle intelligence
3. Don't hardcode assumptions
4. Focus on clear instructions

---

The Claude-driven analysis approach ensures each project gets exactly the analysis it needs, with intelligent interpretation that adapts to any codebase or technology stack.