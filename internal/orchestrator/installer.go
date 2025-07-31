package orchestrator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// OrchestratorInstaller handles creation prompts for orchestrator-specialist
// Note: The actual creation is done by Claude, not deterministically
type OrchestratorInstaller struct {
	ProjectRoot string
	AgentsDir   string
}

// NewOrchestratorInstaller creates a new installer instance
func NewOrchestratorInstaller(projectRoot string) *OrchestratorInstaller {
	return &OrchestratorInstaller{
		ProjectRoot: projectRoot,
		AgentsDir:   filepath.Join(projectRoot, ".claude", "agents"),
	}
}

// InstallOrValidate prompts Claude to create orchestrator if it doesn't exist
// This is called by 'crew claude --install' for project-specific setup
func (oi *OrchestratorInstaller) InstallOrValidate() error {
	// Ensure .claude/agents directory exists
	if err := os.MkdirAll(oi.AgentsDir, 0755); err != nil {
		return fmt.Errorf("failed to create agents directory: %w", err)
	}

	orchestratorPath := filepath.Join(oi.AgentsDir, "orchestrator-specialist.md")

	// Check if orchestrator exists
	if _, err := os.Stat(orchestratorPath); os.IsNotExist(err) {
		// Prompt Claude to create from global template
		return oi.promptOrchestratorCreation()
	}

	// Orchestrator exists, validate it has required features
	return oi.validateExistingOrchestrator(orchestratorPath)
}

// promptOrchestratorCreation guides Claude to create the orchestrator
func (oi *OrchestratorInstaller) promptOrchestratorCreation() error {
	fmt.Println("\n🎯 Orchestrator Creation Required")
	fmt.Println("═══════════════════════════════════")
	fmt.Println()
	fmt.Println("The orchestrator-specialist needs to be created by Claude.")
	fmt.Println("This ensures it's customized for YOUR project's specific needs.")
	fmt.Println()
	fmt.Println("📋 Instructions for Claude:")
	fmt.Println()
	fmt.Println("1. Look for the global orchestrator template in one of:")
	fmt.Println("   - SuperCrew/Agents/orchestrator.agent.md (global deterministic)")
	fmt.Println("   - templates/agents/orchestrator.template.md (framework template)")
	fmt.Println()
	fmt.Println("2. Analyze this project to understand:")
	fmt.Println("   - Primary programming languages")
	fmt.Println("   - Frameworks and libraries in use")
	fmt.Println("   - Architecture patterns")
	fmt.Println("   - Common workflows")
	fmt.Println()
	fmt.Println("3. Create .claude/agents/orchestrator-specialist.md with:")
	fmt.Println("   - Project-specific routing rules")
	fmt.Println("   - Custom workflow patterns")
	fmt.Println("   - Double/triple check completion verification")
	fmt.Println("   - Support for local slash commands")
	fmt.Println("   - Integration with existing specialists")
	fmt.Println()
	fmt.Println("Remember: Global components are deterministic (same for everyone)")
	fmt.Println("Local components are intelligent (customized by Claude)")
	fmt.Println()
	return nil
}

// validateExistingOrchestrator checks if existing orchestrator has required features
func (oi *OrchestratorInstaller) validateExistingOrchestrator(path string) error {
	// Read current file
	current, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read existing orchestrator: %w", err)
	}

	// Check for required features
	missingFeatures := []string{}
	
	if !strings.Contains(string(current), "Completion Verification Process") &&
	   !strings.Contains(string(current), "Double Check Phase") {
		missingFeatures = append(missingFeatures, "Double/triple check completion verification")
	}
	
	if !strings.Contains(string(current), "Local Command") &&
	   !strings.Contains(string(current), "local command") {
		missingFeatures = append(missingFeatures, "Local slash command support")
	}
	
	if len(missingFeatures) > 0 {
		fmt.Println("\n⚠️  Orchestrator needs updates:")
		for _, feature := range missingFeatures {
			fmt.Printf("   - Missing: %s\n", feature)
		}
		fmt.Println("\nClaude, please update the orchestrator to include these features.")
		return nil
	}

	fmt.Println("✅ Orchestrator-specialist is up-to-date")
	return nil
}

// hasRequiredSections checks if orchestrator has all required capabilities
func (oi *OrchestratorInstaller) hasRequiredSections(content string) bool {
	requiredSections := []string{
		"## Agent Chaining Patterns",
		"## Intelligent Routing Algorithm",
		"## Concrete Workflow Examples",
		"## When to Use the Orchestrator",
		"slash_commands:",
		"/crew:chain",
		"/crew:multimodal",
		"/crew:workflow",
	}

	for _, section := range requiredSections {
		if !strings.Contains(content, section) {
			return false
		}
	}

	return true
}

// This is no longer needed - Claude handles all customization

// getProjectName extracts project name from path
func (oi *OrchestratorInstaller) getProjectName() string {
	return filepath.Base(oi.ProjectRoot)
}

// ProjectAnalysisTemplate creates an empty template for Claude to fill out
func (oi *OrchestratorInstaller) CreateAnalysisTemplate() string {
	template := `{
  "analysis_version": "2.0",
  "analyzed_by": "claude",
  "analysis_date": "{{DATE}}",
  "project_path": "{{PROJECT_PATH}}",
  
  "languages": [
    {
      "name": "TO_BE_ANALYZED",
      "file_count": 0,
      "primary": false,
      "patterns_observed": []
    }
  ],
  
  "frameworks": [
    {
      "name": "TO_BE_DETECTED",
      "type": "TO_BE_CATEGORIZED",
      "version": "unknown",
      "usage_patterns": []
    }
  ],
  
  "architectural_patterns": [
    {
      "pattern": "TO_BE_IDENTIFIED",
      "confidence": "low|medium|high",
      "evidence": [],
      "implications": []
    }
  ],
  
  "complexity_assessment": {
    "overall_complexity": "simple|moderate|complex|very_complex",
    "factors": {
      "size": "small|medium|large|very_large",
      "domain_count": 0,
      "integration_points": 0,
      "architectural_layers": 0
    },
    "orchestration_benefit": "low|medium|high"
  },
  
  "specialist_recommendations": [
    {
      "specialist_name": "example-specialist",
      "reason": "why this specialist would help",
      "priority": "low|medium|high",
      "trigger_conditions": []
    }
  ],
  
  "usage_patterns": {
    "detected_workflows": [],
    "common_tasks": [],
    "pain_points": []
  },
  
  "notes": "Claude will fill this analysis based on actual project inspection"
}`

	// Replace placeholders
	template = strings.ReplaceAll(template, "{{DATE}}", time.Now().Format("2006-01-02"))
	template = strings.ReplaceAll(template, "{{PROJECT_PATH}}", oi.ProjectRoot)

	return template
}

// detectLanguages identifies programming languages in use
func (oi *OrchestratorInstaller) detectLanguages() []LanguageInfo {
	languages := []LanguageInfo{}

	// Map of extensions to language info
	langMap := map[string]LanguageInfo{
		".go":   {Name: "Go", FileCount: 0, Primary: false},
		".py":   {Name: "Python", FileCount: 0, Primary: false},
		".js":   {Name: "JavaScript", FileCount: 0, Primary: false},
		".ts":   {Name: "TypeScript", FileCount: 0, Primary: false},
		".java": {Name: "Java", FileCount: 0, Primary: false},
		".rb":   {Name: "Ruby", FileCount: 0, Primary: false},
		".rs":   {Name: "Rust", FileCount: 0, Primary: false},
		".cpp":  {Name: "C++", FileCount: 0, Primary: false},
	}

	// Count files for each language
	for ext, lang := range langMap {
		count := oi.countFilesWithExtension(ext)
		if count > 0 {
			lang.FileCount = count
			languages = append(languages, lang)
		}
	}

	// Mark primary language (most files)
	if len(languages) > 0 {
		maxCount := 0
		maxIdx := 0
		for i, lang := range languages {
			if lang.FileCount > maxCount {
				maxCount = lang.FileCount
				maxIdx = i
			}
		}
		languages[maxIdx].Primary = true
	}

	return languages
}

// detectFrameworks identifies frameworks and libraries
func (oi *OrchestratorInstaller) detectFrameworks() []FrameworkInfo {
	frameworks := []FrameworkInfo{}

	// Check for framework indicators
	indicators := map[string][]string{
		"package.json":     {"Node.js", "npm/yarn"},
		"go.mod":           {"Go Modules"},
		"requirements.txt": {"Python", "pip"},
		"Gemfile":          {"Ruby", "Bundler"},
		"pom.xml":          {"Java", "Maven"},
		"build.gradle":     {"Java", "Gradle"},
		"Cargo.toml":       {"Rust", "Cargo"},
		"composer.json":    {"PHP", "Composer"},
	}

	for file, fwList := range indicators {
		if oi.fileExists(file) {
			for _, fw := range fwList {
				frameworks = append(frameworks, FrameworkInfo{
					Name:     fw,
					Type:     "dependency-manager",
					Detected: true,
				})
			}
		}
	}

	return frameworks
}

// detectPatterns identifies architectural patterns
func (oi *OrchestratorInstaller) detectPatterns() []PatternInfo {
	patterns := []PatternInfo{}

	// Pattern detection logic - Claude will interpret these
	patternChecks := []struct {
		Name       string
		Type       string
		CheckFunc  func() bool
		Confidence float64
	}{
		{"API/REST", "architecture", oi.hasAPIIndicators, 0.0},
		{"CLI Application", "architecture", oi.hasCLIIndicators, 0.0},
		{"Web Application", "architecture", oi.hasWebIndicators, 0.0},
		{"Microservices", "architecture", oi.hasMicroserviceIndicators, 0.0},
		{"Testing", "practice", oi.hasTestingIndicators, 0.0},
		{"CI/CD", "practice", oi.hasCICDIndicators, 0.0},
	}

	for _, check := range patternChecks {
		if check.CheckFunc() {
			patterns = append(patterns, PatternInfo{
				Name:       check.Name,
				Type:       check.Type,
				Confidence: check.Confidence,
			})
		}
	}

	return patterns
}

// assessComplexity provides complexity metrics for Claude
func (oi *OrchestratorInstaller) assessComplexity() ComplexityAssessment {
	return ComplexityAssessment{
		FileCount:      oi.countTotalFiles(),
		DirectoryDepth: oi.getMaxDirectoryDepth(),
		LanguageCount:  len(oi.detectLanguages()),
		PatternCount:   len(oi.detectPatterns()),
		EstimatedSize:  oi.estimateProjectSize(),
	}
}

// generateRecommendations provides hints for Claude
func (oi *OrchestratorInstaller) generateRecommendations(analysis ProjectAnalysis) []string {
	recs := []string{}

	// Language-based recommendations
	for _, lang := range analysis.Languages {
		if lang.Primary {
			recs = append(recs, fmt.Sprintf("Primary language is %s - consider language-specific patterns", lang.Name))
		}
	}

	// Pattern-based recommendations
	for _, pattern := range analysis.Patterns {
		switch pattern.Name {
		case "API/REST":
			recs = append(recs, "API patterns detected - consider API design and testing needs")
		case "CLI Application":
			recs = append(recs, "CLI patterns detected - consider command structure and user interaction")
		case "Microservices":
			recs = append(recs, "Microservice architecture - consider service coordination and deployment")
		}
	}

	// Complexity-based recommendations
	if analysis.Complexity.EstimatedSize == "large" {
		recs = append(recs, "Large project - consider modular agent approach")
	}

	return recs
}

// Type definitions for project analysis
type LanguageInfo struct {
	Name      string `json:"name"`
	FileCount int    `json:"file_count"`
	Primary   bool   `json:"primary"`
}

type FrameworkInfo struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Detected bool   `json:"detected"`
}

type PatternInfo struct {
	Name       string  `json:"name"`
	Type       string  `json:"type"`
	Confidence float64 `json:"confidence"`
}

type ComplexityAssessment struct {
	FileCount      int    `json:"file_count"`
	DirectoryDepth int    `json:"directory_depth"`
	LanguageCount  int    `json:"language_count"`
	PatternCount   int    `json:"pattern_count"`
	EstimatedSize  string `json:"estimated_size"`
}

type ProjectAnalysis struct {
	Languages  []LanguageInfo       `json:"languages"`
	Frameworks []FrameworkInfo      `json:"frameworks"`
	Patterns   []PatternInfo        `json:"patterns"`
	Complexity ComplexityAssessment `json:"complexity"`
}

// Helper methods for pattern detection
func (oi *OrchestratorInstaller) hasFilesWithExtension(ext string) bool {
	found := false
	filepath.Walk(oi.ProjectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil || found {
			return filepath.SkipDir
		}
		if strings.HasSuffix(path, ext) {
			found = true
		}
		return nil
	})
	return found
}

func (oi *OrchestratorInstaller) hasAnyPattern(patterns []string) bool {
	return oi.hasAnyFileContaining(patterns)
}

// Additional helper methods
func (oi *OrchestratorInstaller) countFilesWithExtension(ext string) int {
	count := 0
	filepath.Walk(oi.ProjectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ext) {
			count++
		}
		return nil
	})
	return count
}

func (oi *OrchestratorInstaller) fileExists(filename string) bool {
	path := filepath.Join(oi.ProjectRoot, filename)
	_, err := os.Stat(path)
	return err == nil
}

func (oi *OrchestratorInstaller) hasAPIIndicators() bool {
	indicators := []string{"handler", "route", "endpoint", "controller"}
	return oi.hasAnyFileContaining(indicators)
}

func (oi *OrchestratorInstaller) hasCLIIndicators() bool {
	indicators := []string{"cmd", "command", "flag", "args"}
	return oi.hasAnyFileContaining(indicators)
}

func (oi *OrchestratorInstaller) hasWebIndicators() bool {
	return oi.fileExists("index.html") || oi.fileExists("package.json")
}

func (oi *OrchestratorInstaller) hasMicroserviceIndicators() bool {
	return oi.fileExists("docker-compose.yml") || oi.hasMultipleServices()
}

func (oi *OrchestratorInstaller) hasTestingIndicators() bool {
	return oi.hasFilesWithExtension("_test.go") ||
		oi.hasFilesWithExtension(".test.js") ||
		oi.hasFilesWithExtension("_spec.rb")
}

func (oi *OrchestratorInstaller) hasCICDIndicators() bool {
	return oi.fileExists(".github/workflows") ||
		oi.fileExists(".gitlab-ci.yml") ||
		oi.fileExists("Jenkinsfile")
}

func (oi *OrchestratorInstaller) hasMultipleServices() bool {
	// Simplified check - look for multiple main/server files
	mainCount := 0
	filepath.Walk(oi.ProjectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if strings.Contains(path, "main") || strings.Contains(path, "server") {
			mainCount++
		}
		return nil
	})
	return mainCount > 1
}

func (oi *OrchestratorInstaller) hasAnyFileContaining(patterns []string) bool {
	// This is a simplified check - in production would do actual content search
	for _, pattern := range patterns {
		// Check file names for now
		found := false
		filepath.Walk(oi.ProjectRoot, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || found {
				return nil
			}
			if strings.Contains(strings.ToLower(path), pattern) {
				found = true
			}
			return nil
		})
		if found {
			return true
		}
	}
	return false
}

func (oi *OrchestratorInstaller) countTotalFiles() int {
	count := 0
	filepath.Walk(oi.ProjectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		// Skip hidden files and vendor directories
		if !strings.Contains(path, "/.") && !strings.Contains(path, "/vendor/") {
			count++
		}
		return nil
	})
	return count
}

func (oi *OrchestratorInstaller) getMaxDirectoryDepth() int {
	maxDepth := 0
	filepath.Walk(oi.ProjectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			depth := strings.Count(strings.TrimPrefix(path, oi.ProjectRoot), string(os.PathSeparator))
			if depth > maxDepth {
				maxDepth = depth
			}
		}
		return nil
	})
	return maxDepth
}

func (oi *OrchestratorInstaller) estimateProjectSize() string {
	fileCount := oi.countTotalFiles()
	switch {
	case fileCount < 10:
		return "small"
	case fileCount < 50:
		return "medium"
	case fileCount < 200:
		return "large"
	default:
		return "very-large"
	}
}
