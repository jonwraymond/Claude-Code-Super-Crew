package core

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// Validator handles system requirement validation
type Validator struct {
	checks map[string]func() (bool, string)
}

// NewValidator creates a new validator
func NewValidator() *Validator {
	v := &Validator{
		checks: make(map[string]func() (bool, string)),
	}

	// Register default checks
	v.registerDefaultChecks()

	return v
}

// registerDefaultChecks registers default system checks
func (v *Validator) registerDefaultChecks() {
	// Go toolchain check
	v.checks["go"] = func() (bool, string) {
		if _, err := exec.LookPath("go"); err == nil {
			if out, err := exec.Command("go", "version").Output(); err == nil {
				version := strings.TrimSpace(string(out))
				return true, version
			}
			return true, "Go found"
		}
		return false, "Go not found in PATH"
	}

	// Claude CLI check
	v.checks["claude"] = func() (bool, string) {
		if _, err := exec.LookPath("claude"); err == nil {
			if out, err := exec.Command("claude", "--version").Output(); err == nil {
				version := strings.TrimSpace(string(out))
				return true, fmt.Sprintf("Claude CLI: %s", version)
			}
			return true, "Claude CLI found"
		}
		return false, "Claude CLI not found in PATH"
	}

	// Git check
	v.checks["git"] = func() (bool, string) {
		if _, err := exec.LookPath("git"); err == nil {
			// Get version
			if out, err := exec.Command("git", "--version").Output(); err == nil {
				version := strings.TrimSpace(string(out))
				return true, version
			}
			return true, "Git found"
		}
		return false, "Git not found in PATH"
	}

	// Directory permissions check
	v.checks["permissions"] = func() (bool, string) {
		homeDir := os.Getenv("HOME")
		if homeDir == "" {
			homeDir = os.Getenv("USERPROFILE") // Windows
		}
		claudeDir := fmt.Sprintf("%s/.claude", homeDir)
		
		// Check if directory exists or can be created
		if _, err := os.Stat(claudeDir); err == nil {
			// Directory exists, check if writable
			testFile := fmt.Sprintf("%s/.test_write", claudeDir)
			if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
				return false, fmt.Sprintf("%s exists but is not writable", claudeDir)
			}
			os.Remove(testFile)
			return true, fmt.Sprintf("%s exists and is writable", claudeDir)
		} else {
			// Try to create directory
			if err := os.MkdirAll(claudeDir, 0755); err != nil {
				return false, fmt.Sprintf("Cannot create %s: %v", claudeDir, err)
			}
			// Clean up test directory
			os.RemoveAll(claudeDir)
			return true, fmt.Sprintf("%s can be created", claudeDir)
		}
	}

	// Node.js check (for MCP servers)
	v.checks["node"] = func() (bool, string) {
		if _, err := exec.LookPath("node"); err == nil {
			if out, err := exec.Command("node", "--version").Output(); err == nil {
				version := strings.TrimSpace(string(out))
				return true, fmt.Sprintf("Node.js %s", version)
			}
			return true, "Node.js found"
		}
		return false, "Node.js not found in PATH"
	}
}

// ValidateComponentRequirements validates requirements for specific components
func (v *Validator) ValidateComponentRequirements(components []string, requirements map[string]map[string]string) (bool, []string) {
	errors := []string{}
	allPassed := true

	// Check global requirements
	if reqs, ok := requirements["global"]; ok {
		for check, minVersion := range reqs {
			if checkFunc, exists := v.checks[check]; exists {
				passed, msg := checkFunc()
				if !passed {
					errors = append(errors, fmt.Sprintf("%s: %s (required: %s)", check, msg, minVersion))
					allPassed = false
				}
			}
		}
	}

	// Check component-specific requirements
	for _, comp := range components {
		if reqs, ok := requirements[comp]; ok {
			for check, minVersion := range reqs {
				if checkFunc, exists := v.checks[check]; exists {
					passed, msg := checkFunc()
					if !passed {
						errors = append(errors, fmt.Sprintf("%s for %s: %s (required: %s)", check, comp, msg, minVersion))
						allPassed = false
					}
				}
			}
		}
	}

	return allPassed, errors
}

// DiagnoseSystem runs comprehensive system diagnostics
func (v *Validator) DiagnoseSystem() map[string]interface{} {
	diagnostics := map[string]interface{}{
		"platform":        fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		"checks":          make(map[string]map[string]string),
		"issues":          []string{},
		"recommendations": []string{},
	}

	checks := diagnostics["checks"].(map[string]map[string]string)
	issues := diagnostics["issues"].([]string)
	recommendations := diagnostics["recommendations"].([]string)

	// Run all checks
	for name, checkFunc := range v.checks {
		passed, msg := checkFunc()
		status := "pass"
		if !passed {
			status = "fail"
		}

		checks[name] = map[string]string{
			"status":  status,
			"message": msg,
		}

		// Add issues and recommendations
		if !passed {
			switch name {
			case "go":
				issues = append(issues, "Go toolchain is not installed")
				recommendations = append(recommendations, v.getGoInstallCommand())
			case "git":
				issues = append(issues, "Git is not installed")
				recommendations = append(recommendations, v.getGitInstallCommand())
			case "node":
				issues = append(issues, "Node.js is not installed (required for MCP components)")
				recommendations = append(recommendations, v.getNodeInstallCommand())
			case "claude":
				issues = append(issues, "Claude CLI is not installed")
				recommendations = append(recommendations, "Install Claude CLI from https://claude.ai/cli")
			case "permissions":
				issues = append(issues, "Cannot access ~/.claude directory")
				recommendations = append(recommendations, "Ensure you have write permissions to your home directory")
			}
		}
	}

	diagnostics["issues"] = issues
	diagnostics["recommendations"] = recommendations

	return diagnostics
}

// Platform-specific installation commands
func (v *Validator) getGitInstallCommand() string {
	switch runtime.GOOS {
	case "darwin":
		return "Install Git:\n  brew install git\n  # or download from https://git-scm.com/"
	case "linux":
		return "Install Git:\n  sudo apt-get install git\n  # or: sudo yum install git\n  # or: sudo pacman -S git"
	case "windows":
		return "Install Git:\n  winget install Git.Git\n  # or download from https://git-scm.com/"
	default:
		return "Install Git from https://git-scm.com/"
	}
}

func (v *Validator) getGoInstallCommand() string {
	switch runtime.GOOS {
	case "darwin":
		return "Install Go:\n  brew install go\n  # or download from https://go.dev/dl/"
	case "linux":
		return "Install Go:\n  # Download from https://go.dev/dl/\n  # Extract: tar -C /usr/local -xzf go*.tar.gz\n  # Add to PATH: export PATH=$PATH:/usr/local/go/bin"
	case "windows":
		return "Install Go:\n  winget install GoLang.Go\n  # or download from https://go.dev/dl/"
	default:
		return "Install Go from https://go.dev/dl/"
	}
}

func (v *Validator) getNodeInstallCommand() string {
	switch runtime.GOOS {
	case "darwin":
		return "Install Node.js:\n  brew install node\n  # or download from https://nodejs.org/"
	case "linux":
		return "Install Node.js:\n  curl -fsSL https://deb.nodesource.com/setup_lts.x | sudo -E bash -\n  sudo apt-get install nodejs\n  # or download from https://nodejs.org/"
	case "windows":
		return "Install Node.js:\n  winget install OpenJS.NodeJS\n  # or download from https://nodejs.org/"
	default:
		return "Install Node.js from https://nodejs.org/"
	}
}
