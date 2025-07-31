package orchestrator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProjectAnalyzer_Analyze(t *testing.T) {
	tests := []struct {
		name           string
		setupFunc      func(dir string) error
		expectedLang   string
		expectedAgents []string
		hasBackend     bool
		hasFrontend    bool
	}{
		{
			name: "Go Backend Project",
			setupFunc: func(dir string) error {
				// Create go.mod
				return os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n\ngo 1.21"), 0644)
			},
			expectedLang:   "go",
			expectedAgents: []string{"go-backend-specialist", "api-specialist"},
			hasBackend:     true,
			hasFrontend:    false,
		},
		{
			name: "React Frontend Project",
			setupFunc: func(dir string) error {
				// Create package.json with React
				packageJSON := `{
					"name": "test-app",
					"dependencies": {
						"react": "^18.0.0",
						"react-dom": "^18.0.0"
					}
				}`
				if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(packageJSON), 0644); err != nil {
					return err
				}
				
				// Create src directory with App.tsx
				srcDir := filepath.Join(dir, "src")
				if err := os.MkdirAll(srcDir, 0755); err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(srcDir, "App.tsx"), []byte("export default function App() {}"), 0644)
			},
			expectedLang:   "javascript",
			expectedAgents: []string{"react-frontend-specialist"},
			hasBackend:     false,
			hasFrontend:    true,
		},
		{
			name: "Python Django Project",
			setupFunc: func(dir string) error {
				// Create requirements.txt
				if err := os.WriteFile(filepath.Join(dir, "requirements.txt"), []byte("django==4.2\n"), 0644); err != nil {
					return err
				}
				// Create manage.py (Django indicator)
				return os.WriteFile(filepath.Join(dir, "manage.py"), []byte("#!/usr/bin/env python\n"), 0644)
			},
			expectedLang:   "python",
			expectedAgents: []string{"python-backend-specialist", "api-specialist"},
			hasBackend:     true,
			hasFrontend:    false,
		},
		{
			name: "Full Stack TypeScript Project",
			setupFunc: func(dir string) error {
				// Create package.json
				packageJSON := `{
					"name": "fullstack-app",
					"dependencies": {
						"react": "^18.0.0",
						"express": "^4.18.0"
					}
				}`
				if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(packageJSON), 0644); err != nil {
					return err
				}
				
				// Create tsconfig.json
				if err := os.WriteFile(filepath.Join(dir, "tsconfig.json"), []byte("{}"), 0644); err != nil {
					return err
				}
				
				// Create server.js
				if err := os.WriteFile(filepath.Join(dir, "server.js"), []byte("const express = require('express')"), 0644); err != nil {
					return err
				}
				
				// Create React component
				srcDir := filepath.Join(dir, "src")
				if err := os.MkdirAll(srcDir, 0755); err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(srcDir, "App.tsx"), []byte("export default function App() {}"), 0644)
			},
			expectedLang:   "typescript",
			expectedAgents: []string{"node-backend-specialist", "react-frontend-specialist", "api-specialist"},
			hasBackend:     true,
			hasFrontend:    true,
		},
		{
			name: "DevOps Project with Docker",
			setupFunc: func(dir string) error {
				// Create Dockerfile
				dockerfile := `FROM node:18-alpine
WORKDIR /app
COPY . .
RUN npm install
CMD ["npm", "start"]`
				if err := os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte(dockerfile), 0644); err != nil {
					return err
				}
				
				// Create docker-compose.yml
				dockerCompose := `version: '3.8'
services:
  app:
    build: .
    ports:
      - "3000:3000"`
				return os.WriteFile(filepath.Join(dir, "docker-compose.yml"), []byte(dockerCompose), 0644)
			},
			expectedLang:   "",
			expectedAgents: []string{"devops-specialist"},
			hasBackend:     false,
			hasFrontend:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			tempDir, err := os.MkdirTemp("", "analyzer-test-*")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tempDir)

			// Setup test project
			if err := tt.setupFunc(tempDir); err != nil {
				t.Fatalf("Failed to setup test project: %v", err)
			}

			// Run analyzer
			analyzer := NewProjectAnalyzer(tempDir)
			chars, err := analyzer.Analyze()
			if err != nil {
				t.Fatalf("Analyze failed: %v", err)
			}

			// Check results
			if chars.MainLanguage != tt.expectedLang {
				t.Errorf("Expected language %s, got %s", tt.expectedLang, chars.MainLanguage)
			}

			if chars.HasBackend != tt.hasBackend {
				t.Errorf("Expected HasBackend=%v, got %v", tt.hasBackend, chars.HasBackend)
			}

			if chars.HasFrontend != tt.hasFrontend {
				t.Errorf("Expected HasFrontend=%v, got %v", tt.hasFrontend, chars.HasFrontend)
			}

			// Check detected agents
			if len(chars.DetectedAgents) != len(tt.expectedAgents) {
				t.Errorf("Expected %d agents, got %d: %v", 
					len(tt.expectedAgents), len(chars.DetectedAgents), chars.DetectedAgents)
			}

			// Check each expected agent is present
			for _, expectedAgent := range tt.expectedAgents {
				found := false
				for _, detectedAgent := range chars.DetectedAgents {
					if detectedAgent == expectedAgent {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected agent %s not found in %v", expectedAgent, chars.DetectedAgents)
				}
			}
		})
	}
}

func TestProjectAnalyzer_DetectTesting(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "test-detection-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	testFiles := []string{
		"main_test.go",
		"utils.test.js",
		"component.spec.ts",
		filepath.Join("tests", "test_utils.py"),
	}

	for _, file := range testFiles {
		dir := filepath.Dir(filepath.Join(tempDir, file))
		if dir != tempDir {
			if err := os.MkdirAll(dir, 0755); err != nil {
				t.Fatalf("Failed to create dir %s: %v", dir, err)
			}
		}
		if err := os.WriteFile(filepath.Join(tempDir, file), []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	// Run analyzer
	analyzer := NewProjectAnalyzer(tempDir)
	chars, err := analyzer.Analyze()
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// Check testing was detected
	if !chars.HasTesting {
		t.Error("Expected HasTesting=true")
	}

	// Check qa-specialist agent was added
	found := false
	for _, agent := range chars.DetectedAgents {
		if agent == "qa-specialist" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected qa-specialist agent to be detected")
	}
}