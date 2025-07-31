package orchestrator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jonwraymond/claude-code-super-crew/pkg/logger"
)

func TestAgentGenerator_GenerateAgents(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "agent-gen-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize logger
	log := logger.NewNamedLogger("test")
	log.SetLevel(logger.DebugLevel)

	// Create test project characteristics
	chars := &ProjectCharacteristics{
		RootPath:     tempDir,
		MainLanguage: "go",
		Frameworks:   []string{"gin", "gorm"},
		DetectedAgents: []string{
			"go-backend-specialist",
			"api-specialist",
			"database-specialist",
		},
		HasBackend:  true,
		HasDatabase: true,
	}

	// Create agent generator
	generator := NewAgentGenerator(tempDir)

	// Generate agents
	if err := generator.GenerateAgents(chars); err != nil {
		t.Fatalf("GenerateAgents failed: %v", err)
	}

	// Check that agents were created
	agentsDir := filepath.Join(tempDir, ".claude", "agents")
	for _, agentType := range chars.DetectedAgents {
		agentFile := filepath.Join(agentsDir, agentType+".md")

		// Check file exists
		if _, err := os.Stat(agentFile); os.IsNotExist(err) {
			t.Errorf("Agent file %s was not created", agentFile)
			continue
		}

		// Read and verify content
		content, err := os.ReadFile(agentFile)
		if err != nil {
			t.Errorf("Failed to read agent file %s: %v", agentFile, err)
			continue
		}

		contentStr := string(content)

		// Check YAML frontmatter
		if !strings.Contains(contentStr, "name: "+agentType) {
			t.Errorf("Agent %s missing name in frontmatter", agentType)
		}

		// Check language reference
		if !strings.Contains(contentStr, "language: \"go\"") {
			t.Errorf("Agent %s missing correct language", agentType)
		}

		// Check framework references
		if !strings.Contains(contentStr, "gin") || !strings.Contains(contentStr, "gorm") {
			t.Errorf("Agent %s missing framework references", agentType)
		}

		// Check agent-specific content
		switch agentType {
		case "go-backend-specialist":
			if !strings.Contains(contentStr, "Go backend development specialist") {
				t.Errorf("Go backend specialist missing correct description")
			}
		case "api-specialist":
			if !strings.Contains(contentStr, "API design specialist") {
				t.Errorf("API specialist missing correct description")
			}
		case "database-specialist":
			if !strings.Contains(contentStr, "Database specialist") {
				t.Errorf("Database specialist missing correct description")
			}
		}
	}
}

func TestAgentGenerator_SkipExistingAgents(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "agent-skip-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize logger
	log := logger.NewNamedLogger("test")
	log.SetLevel(logger.DebugLevel)

	// Create agents directory
	agentsDir := filepath.Join(tempDir, ".claude", "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("Failed to create agents dir: %v", err)
	}

	// Create existing agent file
	existingAgent := "go-backend-specialist.md"
	existingContent := "# Existing agent - should not be overwritten"
	if err := os.WriteFile(filepath.Join(agentsDir, existingAgent), []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to create existing agent: %v", err)
	}

	// Create test project characteristics
	chars := &ProjectCharacteristics{
		RootPath:     tempDir,
		MainLanguage: "go",
		DetectedAgents: []string{
			"go-backend-specialist", // This one exists
			"api-specialist",        // This one doesn't
		},
	}

	// Create agent generator
	generator := NewAgentGenerator(tempDir)

	// Generate agents
	if err := generator.GenerateAgents(chars); err != nil {
		t.Fatalf("GenerateAgents failed: %v", err)
	}

	// Check that existing agent was not overwritten
	content, err := os.ReadFile(filepath.Join(agentsDir, existingAgent))
	if err != nil {
		t.Fatalf("Failed to read existing agent: %v", err)
	}

	if string(content) != existingContent {
		t.Error("Existing agent was overwritten")
	}

	// Check that new agent was created
	newAgentFile := filepath.Join(agentsDir, "api-specialist.md")
	if _, err := os.Stat(newAgentFile); os.IsNotExist(err) {
		t.Error("New agent was not created")
	}
}

func TestGetAgentDescription(t *testing.T) {
	tests := []struct {
		agentType    string
		expectedText string
	}{
		{
			agentType:    "go-backend-specialist",
			expectedText: "Go backend development specialist",
		},
		{
			agentType:    "react-frontend-specialist",
			expectedText: "React frontend specialist",
		},
		{
			agentType:    "devops-specialist",
			expectedText: "DevOps specialist",
		},
		{
			agentType:    "unknown-specialist",
			expectedText: "Specialized agent for unknown specialist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.agentType, func(t *testing.T) {
			desc := GetAgentDescription(tt.agentType)
			if !strings.Contains(desc, tt.expectedText) {
				t.Errorf("Expected description to contain '%s', got '%s'", tt.expectedText, desc)
			}
		})
	}
}
