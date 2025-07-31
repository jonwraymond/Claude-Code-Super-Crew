# How to Use Local Shadow Commands

## Overview

Local shadow commands enhance global commands with project-specific workflows. They're activated through Claude Code using the same command name as the global, but Claude automatically uses the local enhanced version when it exists.

## Activation Methods

### Method 1: Direct Command Usage (Recommended)

Simply use the command as normal in Claude Code:

```bash
# If you have .claude/commands/shadows/build.md
# Claude will automatically use the enhanced local version
/build

# With flags
/build --release --cross-compile
```

### Method 2: Explicit Routing Through Orchestrator

For complex workflows, route through the orchestrator:

```bash
/crew:orchestrate "build the project with full validation"
# Orchestrator will use the enhanced local build command
```

### Method 3: Direct Agent Chain Invocation

You can also invoke the chain directly:

```bash
# This triggers the multi-agent workflow defined in the shadow command
/crew:chain "analyze deps → build → test → package"
```

## How It Works

### 1. Command Resolution

When you type `/build` in Claude Code:

```
1. Claude checks .claude/commands/build.md (local command)
2. Claude checks .claude/commands/shadows/build.md (shadow command) ✓
3. Claude checks ~/.claude/commands/build.md (global command)
4. Shadow command takes precedence and enhances global
```

### 2. Multi-Agent Execution

The shadow command's workflow executes:

```yaml
workflow:
  - analyzer: "Check environment"
  - security: "Scan dependencies"
  - builder: "Execute build" (inherits global)
  - tester: "Run tests"
  - packager: "Create artifacts"
```

### 3. Local Integration

Project scripts and tools are automatically integrated:

```bash
# Pre-build scripts run automatically
./scripts/check-environment.sh

# Build happens with enhancements

# Post-build scripts run automatically
./scripts/package-release.sh
```

## Example Walkthrough

### Step 1: Create Shadow Command

First, ensure the shadow command exists:

```bash
# Check if shadow exists
ls .claude/commands/shadows/build.md
```

If not, ask Claude to create it:

```
"Create a shadow command for /build that adds testing and security scanning"
```

### Step 2: Use the Command

Simply use it like any slash command:

```
/build
```

Claude will:
1. Detect the shadow command
2. Execute the multi-agent workflow
3. Run integrated scripts
4. Provide enhanced results

### Step 3: Monitor Execution

You'll see the multi-agent chain in action:

```
🎯 Executing enhanced /build command...

1️⃣ analyzer-persona: Checking environment...
   ✓ Go 1.21.5 detected
   ✓ All dependencies available

2️⃣ security-persona: Scanning dependencies...
   ✓ No vulnerabilities found

3️⃣ backend-persona: Building project...
   ✓ Build successful

4️⃣ test-generator: Running tests...
   ✓ All tests passed (coverage: 67%)

5️⃣ devops-persona: Creating artifacts...
   ✓ Release package created

✅ Enhanced build completed successfully!
```

## Command Discovery

### List Available Shadow Commands

Ask Claude:

```
"Show me all shadow commands in this project"
```

Or check manually:

```bash
ls .claude/commands/shadows/
```

### Understand a Shadow Command

Ask Claude:

```
"Explain what the local /build command does"
```

Claude will show:
- What it shadows
- What enhancements it adds
- The agent workflow
- Integrated tools/scripts

## Creating New Shadow Commands

### Quick Creation

Ask Claude to create shadows for common commands:

```
"Create a shadow for /test that adds coverage reporting and benchmarks"

"Shadow the /deploy command to add pre-deployment validation"

"Enhance /analyze with project-specific patterns"
```

### Custom Creation

Provide specific requirements:

```
"Create a shadow for /release that:
1. Runs all tests
2. Builds for multiple platforms
3. Creates changelog
4. Tags git repository
5. Uploads artifacts"
```

## Best Practices

### 1. Start Simple

Begin with basic enhancements:

```yaml
# Simple shadow adding one step
workflow:
  - inherit: true  # Do everything global does
  - linter: "Run project linter"
```

### 2. Build Up Gradually

Add more agents as needed:

```yaml
workflow:
  - analyzer: "Pre-checks"
  - inherit: true
  - tester: "Additional tests"
  - documenter: "Update docs"
```

### 3. Use Project Resources

Integrate existing scripts:

```yaml
scripts:
  pre: ["make pre-build"]
  post: ["make post-build"]
```

### 4. Document Value

Always explain enhancements:

```markdown
## Enhancements
- Adds security scanning
- Runs comprehensive tests
- Generates documentation
```

## Troubleshooting

### Shadow Command Not Working?

1. **Check it exists**: `ls .claude/commands/shadows/`
2. **Validate syntax**: Ask Claude to check the YAML
3. **Test components**: Try individual agents
4. **Check scripts**: Ensure local scripts are executable

### Want to Use Global Instead?

Explicitly request global version:

```
"Use the global /build command without local enhancements"
```

Or temporarily rename shadow:

```bash
mv .claude/commands/shadows/build.md .claude/commands/shadows/build.md.disabled
```

## Summary

- **Automatic**: Shadow commands activate automatically when present
- **Enhanced**: They add project-specific workflows and tools
- **Powerful**: Multi-agent chains provide superior results
- **Integrated**: Local scripts and tools work seamlessly
- **Flexible**: Can always fall back to global if needed

The beauty is that it's transparent - just use commands normally and get enhanced behavior automatically!