// Package hooks provides hook management for Claude Code Super Crew
package hooks

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jonwraymond/claude-code-super-crew/pkg/logger"
)

// HookType represents the type of Claude Code hook
type HookType string

const (
	PreToolUse        HookType = "PreToolUse"
	PostToolUse       HookType = "PostToolUse"
	UserPromptSubmit  HookType = "UserPromptSubmit"
	Stop              HookType = "Stop"
	SubagentStop      HookType = "SubagentStop"
	PreCompact        HookType = "PreCompact"
	SessionStart      HookType = "SessionStart"
	Notification      HookType = "Notification"
)

// Hook represents a configured hook
type Hook struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Type        HookType          `json:"type"`
	Matcher     string            `json:"matcher,omitempty"`
	Command     string            `json:"command"`
	Enabled     bool              `json:"enabled"`
	Config      map[string]string `json:"config,omitempty"`
}

// HookManager manages SuperCrew hooks
type HookManager struct {
	hooksDir     string
	globalHooks  map[string]*Hook
	enabledHooks map[string]bool
	logger       logger.Logger
}

// NewHookManager creates a new hook manager
func NewHookManager(projectRoot string) *HookManager {
	return &HookManager{
		hooksDir:     filepath.Join(projectRoot, "SuperCrew", "Hooks"),
		globalHooks:  make(map[string]*Hook),
		enabledHooks: make(map[string]bool),
		logger:       logger.GetLogger(),
	}
}

// DiscoverHooks finds all available hooks
func (hm *HookManager) DiscoverHooks() error {
	// Define our global hooks
	hm.globalHooks = map[string]*Hook{
		"git-auto-commit": {
			Name:        "git-auto-commit",
			Description: "Automatically commit changes made by Claude Code",
			Type:        PostToolUse,
			Matcher:     "Write|Edit|MultiEdit",
			Command:     filepath.Join(hm.hooksDir, "git-auto-commit.sh"),
			Enabled:     false,
			Config: map[string]string{
				"SUPERCREW_GIT_AUTO_COMMIT": "true",
			},
		},
		"lint-on-save": {
			Name:        "lint-on-save",
			Description: "Run linters after file modifications",
			Type:        PostToolUse,
			Matcher:     "Write|Edit|MultiEdit",
			Command:     filepath.Join(hm.hooksDir, "lint-on-save.sh"),
			Enabled:     false,
			Config: map[string]string{
				"SUPERCREW_LINT_AUTOFIX": "false",
				"SUPERCREW_LINT_QUIET":   "false",
			},
		},
		"test-on-change": {
			Name:        "test-on-change",
			Description: "Run relevant tests when code files are modified",
			Type:        PostToolUse,
			Matcher:     "Write|Edit|MultiEdit",
			Command:     filepath.Join(hm.hooksDir, "test-on-change.sh"),
			Enabled:     false,
			Config: map[string]string{
				"SUPERCREW_TEST_PATTERN":  "auto",
				"SUPERCREW_TEST_COVERAGE": "false",
			},
		},
		"security-scan": {
			Name:        "security-scan",
			Description: "Scan code changes for security vulnerabilities",
			Type:        PostToolUse,
			Matcher:     "Write|Edit|MultiEdit",
			Command:     filepath.Join(hm.hooksDir, "security-scan.sh"),
			Enabled:     false,
			Config: map[string]string{
				"SUPERCREW_SECURITY_BLOCK": "true",
				"SUPERCREW_SECURITY_LEVEL": "medium",
			},
		},
		"backup-before-change": {
			Name:        "backup-before-change",
			Description: "Create backups before modifying files",
			Type:        PreToolUse,
			Matcher:     "Write|Edit|MultiEdit",
			Command:     filepath.Join(hm.hooksDir, "backup-before-change.sh"),
			Enabled:     false,
			Config: map[string]string{
				"SUPERCREW_BACKUP_DIR":  ".claude/backups",
				"SUPERCREW_BACKUP_DAYS": "7",
			},
		},
	}

	// Load enabled status from config
	hm.loadEnabledStatus()

	return nil
}

// ListHooks returns all discovered hooks
func (hm *HookManager) ListHooks() []*Hook {
	var hooks []*Hook
	for _, hook := range hm.globalHooks {
		hooks = append(hooks, hook)
	}
	return hooks
}

// EnableHook enables a specific hook
func (hm *HookManager) EnableHook(name string) error {
	hook, exists := hm.globalHooks[name]
	if !exists {
		return fmt.Errorf("hook not found: %s", name)
	}

	// Check if hook script exists
	if _, err := os.Stat(hook.Command); os.IsNotExist(err) {
		return fmt.Errorf("hook script not found: %s", hook.Command)
	}

	// Make script executable
	if err := os.Chmod(hook.Command, 0755); err != nil {
		return fmt.Errorf("failed to make hook executable: %w", err)
	}

	hook.Enabled = true
	hm.enabledHooks[name] = true

	// Save to Claude settings
	if err := hm.updateClaudeSettings(); err != nil {
		return fmt.Errorf("failed to update Claude settings: %w", err)
	}

	hm.logger.Successf("Enabled hook: %s", name)
	return nil
}

// DisableHook disables a specific hook
func (hm *HookManager) DisableHook(name string) error {
	hook, exists := hm.globalHooks[name]
	if !exists {
		return fmt.Errorf("hook not found: %s", name)
	}

	hook.Enabled = false
	delete(hm.enabledHooks, name)

	// Update Claude settings
	if err := hm.updateClaudeSettings(); err != nil {
		return fmt.Errorf("failed to update Claude settings: %w", err)
	}

	hm.logger.Successf("Disabled hook: %s", name)
	return nil
}

// InstallHooks copies hook scripts to the appropriate location
func (hm *HookManager) InstallHooks(targetDir string) error {
	// Create hooks directory
	hooksDir := filepath.Join(targetDir, "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return fmt.Errorf("failed to create hooks directory: %w", err)
	}

	// Copy hook scripts
	entries, err := os.ReadDir(hm.hooksDir)
	if err != nil {
		return fmt.Errorf("failed to read hooks directory: %w", err)
	}

	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".sh") {
			src := filepath.Join(hm.hooksDir, entry.Name())
			dst := filepath.Join(hooksDir, entry.Name())

			// Read and copy file
			content, err := os.ReadFile(src)
			if err != nil {
				hm.logger.Warnf("Failed to read hook %s: %v", entry.Name(), err)
				continue
			}

			if err := os.WriteFile(dst, content, 0755); err != nil {
				hm.logger.Warnf("Failed to install hook %s: %v", entry.Name(), err)
				continue
			}

			hm.logger.Infof("Installed hook: %s", entry.Name())
		}
	}

	return nil
}

// updateClaudeSettings updates the Claude Code settings.json with hook configuration
func (hm *HookManager) updateClaudeSettings() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	settingsPath := filepath.Join(homeDir, ".claude", "settings.json")

	// Read existing settings
	var settings map[string]interface{}
	if data, err := os.ReadFile(settingsPath); err == nil {
		json.Unmarshal(data, &settings)
	} else {
		settings = make(map[string]interface{})
	}

	// Build hooks configuration
	hooks := make(map[string]interface{})
	
	// Group hooks by type
	for _, hook := range hm.globalHooks {
		if !hook.Enabled {
			continue
		}

		hookConfig := map[string]interface{}{
			"type":    "command",
			"command": hook.Command,
		}

		// Add environment variables from config
		if len(hook.Config) > 0 {
			env := make(map[string]string)
			for k, v := range hook.Config {
				env[k] = v
			}
			hookConfig["env"] = env
		}

		// Create hook entry
		hookEntry := map[string]interface{}{
			"matcher": hook.Matcher,
			"hooks":   []interface{}{hookConfig},
		}

		// Add to appropriate hook type
		typeName := string(hook.Type)
		if existing, ok := hooks[typeName]; ok {
			hooks[typeName] = append(existing.([]interface{}), hookEntry)
		} else {
			hooks[typeName] = []interface{}{hookEntry}
		}
	}

	// Update settings
	settings["hooks"] = hooks

	// Write settings
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	// Ensure directory exists
	os.MkdirAll(filepath.Dir(settingsPath), 0755)

	return os.WriteFile(settingsPath, data, 0644)
}

// loadEnabledStatus loads which hooks are enabled from Claude settings
func (hm *HookManager) loadEnabledStatus() {
	homeDir, _ := os.UserHomeDir()
	settingsPath := filepath.Join(homeDir, ".claude", "settings.json")

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return
	}

	// Check which hooks are configured
	if hooks, ok := settings["hooks"].(map[string]interface{}); ok {
		for _, hook := range hm.globalHooks {
			if hookType, ok := hooks[string(hook.Type)].([]interface{}); ok {
				for _, entry := range hookType {
					if m, ok := entry.(map[string]interface{}); ok {
						if matcher, ok := m["matcher"].(string); ok && matcher == hook.Matcher {
							if hooksList, ok := m["hooks"].([]interface{}); ok {
								for _, h := range hooksList {
									if hookMap, ok := h.(map[string]interface{}); ok {
										if cmd, ok := hookMap["command"].(string); ok && strings.Contains(cmd, hook.Name) {
											hook.Enabled = true
											hm.enabledHooks[hook.Name] = true
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
}

// GetHookInfo returns detailed information about a specific hook
func (hm *HookManager) GetHookInfo(name string) (*Hook, error) {
	hook, exists := hm.globalHooks[name]
	if !exists {
		return nil, fmt.Errorf("hook not found: %s", name)
	}
	return hook, nil
}

// ConfigureHook updates hook configuration
func (hm *HookManager) ConfigureHook(name string, config map[string]string) error {
	hook, exists := hm.globalHooks[name]
	if !exists {
		return fmt.Errorf("hook not found: %s", name)
	}

	// Update configuration
	for k, v := range config {
		hook.Config[k] = v
	}

	// Update Claude settings if hook is enabled
	if hook.Enabled {
		if err := hm.updateClaudeSettings(); err != nil {
			return fmt.Errorf("failed to update Claude settings: %w", err)
		}
	}

	return nil
}