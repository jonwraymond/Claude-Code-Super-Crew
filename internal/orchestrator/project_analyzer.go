package orchestrator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ProjectType represents the detected type of project
type ProjectType int

const (
	ProjectTypeUnknown ProjectType = iota
	ProjectTypeGo
	ProjectTypeJavaScript
	ProjectTypeTypeScript
	ProjectTypePython
	ProjectTypeRust
	ProjectTypeJava
	ProjectTypeReact
	ProjectTypeVue
	ProjectTypeAngular
	ProjectTypeDocker
	ProjectTypeKubernetes
)

// ProjectCharacteristics holds the analyzed characteristics of a project
type ProjectCharacteristics struct {
	RootPath       string
	MainLanguage   string
	Frameworks     []string
	ProjectTypes   []ProjectType
	HasBackend     bool
	HasFrontend    bool
	HasDatabase    bool
	HasDocker      bool
	HasKubernetes  bool
	HasTesting     bool
	HasCI          bool
	DetectedAgents []string // List of agent types to generate
}

// ProjectAnalyzer analyzes project structure and characteristics
type ProjectAnalyzer struct {
	rootPath string
}

// NewProjectAnalyzer creates a new project analyzer
func NewProjectAnalyzer(rootPath string) *ProjectAnalyzer {
	return &ProjectAnalyzer{
		rootPath: rootPath,
	}
}

// Analyze performs comprehensive project analysis
func (pa *ProjectAnalyzer) Analyze() (*ProjectCharacteristics, error) {
	chars := &ProjectCharacteristics{
		RootPath:       pa.rootPath,
		Frameworks:     []string{},
		ProjectTypes:   []ProjectType{},
		DetectedAgents: []string{},
	}

	// Detect main language and frameworks
	if err := pa.detectLanguagesAndFrameworks(chars); err != nil {
		return nil, fmt.Errorf("failed to detect languages: %w", err)
	}

	// Detect infrastructure components
	if err := pa.detectInfrastructure(chars); err != nil {
		return nil, fmt.Errorf("failed to detect infrastructure: %w", err)
	}

	// Determine which agents to generate
	pa.determineAgents(chars)

	return chars, nil
}

// detectLanguagesAndFrameworks detects programming languages and frameworks
func (pa *ProjectAnalyzer) detectLanguagesAndFrameworks(chars *ProjectCharacteristics) error {
	// Check for Go
	if exists(filepath.Join(pa.rootPath, "go.mod")) {
		chars.MainLanguage = "go"
		chars.ProjectTypes = append(chars.ProjectTypes, ProjectTypeGo)
		chars.HasBackend = true
	}

	// Check for JavaScript/TypeScript
	if exists(filepath.Join(pa.rootPath, "package.json")) {
		// Read package.json to detect frameworks
		if exists(filepath.Join(pa.rootPath, "tsconfig.json")) {
			if chars.MainLanguage == "" {
				chars.MainLanguage = "typescript"
			}
			chars.ProjectTypes = append(chars.ProjectTypes, ProjectTypeTypeScript)
		} else if chars.MainLanguage == "" {
			chars.MainLanguage = "javascript"
			chars.ProjectTypes = append(chars.ProjectTypes, ProjectTypeJavaScript)
		}

		// Check for React
		if existsPattern(pa.rootPath, "src/**/*.jsx") || existsPattern(pa.rootPath, "src/**/*.tsx") ||
			exists(filepath.Join(pa.rootPath, "src", "App.js")) || exists(filepath.Join(pa.rootPath, "src", "App.tsx")) {
			chars.Frameworks = append(chars.Frameworks, "react")
			chars.ProjectTypes = append(chars.ProjectTypes, ProjectTypeReact)
			chars.HasFrontend = true
		}

		// Check for Vue
		if exists(filepath.Join(pa.rootPath, "vue.config.js")) || existsPattern(pa.rootPath, "src/**/*.vue") {
			chars.Frameworks = append(chars.Frameworks, "vue")
			chars.ProjectTypes = append(chars.ProjectTypes, ProjectTypeVue)
			chars.HasFrontend = true
		}

		// Check for Angular
		if exists(filepath.Join(pa.rootPath, "angular.json")) {
			chars.Frameworks = append(chars.Frameworks, "angular")
			chars.ProjectTypes = append(chars.ProjectTypes, ProjectTypeAngular)
			chars.HasFrontend = true
		}

		// Check for Node.js backend patterns
		if exists(filepath.Join(pa.rootPath, "server.js")) || exists(filepath.Join(pa.rootPath, "app.js")) ||
			existsPattern(pa.rootPath, "src/controllers/**/*.js") || existsPattern(pa.rootPath, "api/**/*.js") {
			chars.HasBackend = true
		}
	}

	// Check for Python
	if exists(filepath.Join(pa.rootPath, "requirements.txt")) || exists(filepath.Join(pa.rootPath, "setup.py")) ||
		exists(filepath.Join(pa.rootPath, "pyproject.toml")) || exists(filepath.Join(pa.rootPath, "Pipfile")) {
		if chars.MainLanguage == "" {
			chars.MainLanguage = "python"
		}
		chars.ProjectTypes = append(chars.ProjectTypes, ProjectTypePython)
		
		// Check for Django/Flask
		if exists(filepath.Join(pa.rootPath, "manage.py")) {
			chars.Frameworks = append(chars.Frameworks, "django")
			chars.HasBackend = true
		}
		if existsPattern(pa.rootPath, "**/flask_app.py") || existsPattern(pa.rootPath, "**/app.py") {
			chars.HasBackend = true
		}
	}

	// Check for Rust
	if exists(filepath.Join(pa.rootPath, "Cargo.toml")) {
		if chars.MainLanguage == "" {
			chars.MainLanguage = "rust"
		}
		chars.ProjectTypes = append(chars.ProjectTypes, ProjectTypeRust)
		chars.HasBackend = true
	}

	// Check for Java
	if exists(filepath.Join(pa.rootPath, "pom.xml")) || exists(filepath.Join(pa.rootPath, "build.gradle")) {
		if chars.MainLanguage == "" {
			chars.MainLanguage = "java"
		}
		chars.ProjectTypes = append(chars.ProjectTypes, ProjectTypeJava)
		chars.HasBackend = true
	}

	// Check for database
	if existsPattern(pa.rootPath, "**/migrations/**") || existsPattern(pa.rootPath, "**/models/**") ||
		exists(filepath.Join(pa.rootPath, "schema.sql")) || exists(filepath.Join(pa.rootPath, "database")) {
		chars.HasDatabase = true
	}

	// Check for testing
	if existsPattern(pa.rootPath, "**/*_test.go") || existsPattern(pa.rootPath, "**/*.test.js") ||
		existsPattern(pa.rootPath, "**/*.spec.js") || existsPattern(pa.rootPath, "**/test_*.py") ||
		exists(filepath.Join(pa.rootPath, "tests")) || exists(filepath.Join(pa.rootPath, "__tests__")) {
		chars.HasTesting = true
	}

	return nil
}

// detectInfrastructure detects infrastructure and DevOps components
func (pa *ProjectAnalyzer) detectInfrastructure(chars *ProjectCharacteristics) error {
	// Check for Docker
	if exists(filepath.Join(pa.rootPath, "Dockerfile")) || exists(filepath.Join(pa.rootPath, "docker-compose.yml")) ||
		exists(filepath.Join(pa.rootPath, "docker-compose.yaml")) {
		chars.HasDocker = true
		chars.ProjectTypes = append(chars.ProjectTypes, ProjectTypeDocker)
	}

	// Check for Kubernetes
	if existsPattern(pa.rootPath, "k8s/**/*.yaml") || existsPattern(pa.rootPath, "k8s/**/*.yml") ||
		existsPattern(pa.rootPath, "kubernetes/**/*.yaml") || existsPattern(pa.rootPath, ".k8s/**/*.yaml") ||
		exists(filepath.Join(pa.rootPath, "skaffold.yaml")) {
		chars.HasKubernetes = true
		chars.ProjectTypes = append(chars.ProjectTypes, ProjectTypeKubernetes)
	}

	// Check for CI/CD
	if exists(filepath.Join(pa.rootPath, ".github", "workflows")) ||
		exists(filepath.Join(pa.rootPath, ".gitlab-ci.yml")) ||
		exists(filepath.Join(pa.rootPath, "Jenkinsfile")) ||
		exists(filepath.Join(pa.rootPath, ".circleci")) {
		chars.HasCI = true
	}

	return nil
}

// determineAgents determines which agents should be generated based on project characteristics
func (pa *ProjectAnalyzer) determineAgents(chars *ProjectCharacteristics) {
	agents := []string{}

	// Backend agents
	if chars.HasBackend {
		switch chars.MainLanguage {
		case "go":
			agents = append(agents, "go-backend-specialist")
		case "javascript", "typescript":
			agents = append(agents, "node-backend-specialist")
		case "python":
			agents = append(agents, "python-backend-specialist")
		case "rust":
			agents = append(agents, "rust-backend-specialist")
		case "java":
			agents = append(agents, "java-backend-specialist")
		default:
			if chars.HasBackend {
				agents = append(agents, "backend-specialist")
			}
		}
	}

	// Frontend agents
	if chars.HasFrontend {
		for _, framework := range chars.Frameworks {
			switch framework {
			case "react":
				agents = append(agents, "react-frontend-specialist")
			case "vue":
				agents = append(agents, "vue-frontend-specialist")
			case "angular":
				agents = append(agents, "angular-frontend-specialist")
			}
		}
		if len(agents) == 0 && chars.HasFrontend {
			agents = append(agents, "frontend-specialist")
		}
	}

	// Infrastructure agents
	if chars.HasDocker || chars.HasKubernetes || chars.HasCI {
		agents = append(agents, "devops-specialist")
	}

	// Database agent
	if chars.HasDatabase {
		agents = append(agents, "database-specialist")
	}

	// Testing agent
	if chars.HasTesting {
		agents = append(agents, "qa-specialist")
	}

	// API agent for backend projects
	if chars.HasBackend && !contains(agents, "api-specialist") {
		agents = append(agents, "api-specialist")
	}

	chars.DetectedAgents = agents
}

// Helper functions

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func existsPattern(root, pattern string) bool {
	matches, err := filepath.Glob(filepath.Join(root, pattern))
	if err != nil {
		return false
	}
	return len(matches) > 0
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// GetProjectTypeName returns the string name of a project type
func GetProjectTypeName(pt ProjectType) string {
	switch pt {
	case ProjectTypeGo:
		return "Go"
	case ProjectTypeJavaScript:
		return "JavaScript"
	case ProjectTypeTypeScript:
		return "TypeScript"
	case ProjectTypePython:
		return "Python"
	case ProjectTypeRust:
		return "Rust"
	case ProjectTypeJava:
		return "Java"
	case ProjectTypeReact:
		return "React"
	case ProjectTypeVue:
		return "Vue"
	case ProjectTypeAngular:
		return "Angular"
	case ProjectTypeDocker:
		return "Docker"
	case ProjectTypeKubernetes:
		return "Kubernetes"
	default:
		return "Unknown"
	}
}

// GetAgentDescription returns a description for an agent type
func GetAgentDescription(agentType string) string {
	descriptions := map[string]string{
		"go-backend-specialist":       "Go backend development specialist with expertise in Go patterns, concurrency, and performance",
		"node-backend-specialist":     "Node.js backend specialist with expertise in Express, APIs, and async patterns",
		"python-backend-specialist":   "Python backend specialist with expertise in Django/Flask, data processing, and APIs",
		"rust-backend-specialist":     "Rust backend specialist with expertise in performance, safety, and systems programming",
		"java-backend-specialist":     "Java backend specialist with expertise in Spring, enterprise patterns, and JVM optimization",
		"backend-specialist":          "General backend development specialist with API design and server architecture expertise",
		"react-frontend-specialist":   "React frontend specialist with expertise in hooks, state management, and component patterns",
		"vue-frontend-specialist":     "Vue.js frontend specialist with expertise in composition API, reactivity, and components",
		"angular-frontend-specialist": "Angular frontend specialist with expertise in TypeScript, RxJS, and enterprise patterns",
		"frontend-specialist":         "Frontend development specialist with expertise in modern web technologies and UX",
		"devops-specialist":           "DevOps specialist with expertise in Docker, Kubernetes, CI/CD, and infrastructure",
		"database-specialist":         "Database specialist with expertise in SQL, NoSQL, migrations, and query optimization",
		"qa-specialist":               "Quality assurance specialist with expertise in testing strategies, automation, and coverage",
		"api-specialist":              "API design specialist with expertise in REST, GraphQL, authentication, and documentation",
	}

	if desc, ok := descriptions[agentType]; ok {
		return desc
	}
	return "Specialized agent for " + strings.ReplaceAll(agentType, "-", " ")
}