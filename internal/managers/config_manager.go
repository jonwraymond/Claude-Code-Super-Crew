package managers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

// Profile represents an installation profile
type Profile struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Components  []string `json:"components"`
	Settings    map[string]interface{} `json:"settings,omitempty"`
}

// Requirements represents system requirements
type Requirements struct {
	Global map[string]string            `json:"global"`
	Components map[string]map[string]string `json:"components"`
}

// Config represents the main configuration
type Config struct {
	Version     string                 `json:"version"`
	InstallDir  string                 `json:"install_dir"`
	Components  []string               `json:"components"`
	Settings    map[string]interface{} `json:"settings"`
	LastUpdated string                 `json:"last_updated"`
}

// ConfigManager provides advanced configuration management with JSON schema validation
type ConfigManager struct {
	configDir     string
	configFile    string
	schema        *ConfigSchema
	validator     *validator.Validate
	config        map[string]interface{}
	lastModified  time.Time
	watchEnabled  bool
}

// ConfigSchema defines the structure and validation rules for configuration
type ConfigSchema struct {
	Type       string                    `json:"type"`
	Properties map[string]*PropertySchema `json:"properties,omitempty"`
	Required   []string                  `json:"required,omitempty"`
	Default    interface{}               `json:"default,omitempty"`
}

// PropertySchema defines validation rules for individual properties
type PropertySchema struct {
	Type        string                     `json:"type"`
	Format      string                     `json:"format,omitempty"`
	Pattern     string                     `json:"pattern,omitempty"`
	Minimum     *float64                   `json:"minimum,omitempty"`
	Maximum     *float64                   `json:"maximum,omitempty"`
	MinLength   *int                       `json:"minLength,omitempty"`
	MaxLength   *int                       `json:"maxLength,omitempty"`
	Enum        []interface{}              `json:"enum,omitempty"`
	Properties  map[string]*PropertySchema `json:"properties,omitempty"`
	Items       *PropertySchema            `json:"items,omitempty"`
	Required    []string                   `json:"required,omitempty"`
	Default     interface{}                `json:"default,omitempty"`
	Description string                     `json:"description,omitempty"`
}

// NewConfigManager creates a new configuration manager
func NewConfigManager(configDir string, schemaPath string) (*ConfigManager, error) {
	configFile := filepath.Join(configDir, "config.json")
	
	cm := &ConfigManager{
		configDir:  configDir,
		configFile: configFile,
		validator:  validator.New(),
		config:     make(map[string]interface{}),
	}
	
	// Load schema if provided
	if schemaPath != "" {
		if err := cm.LoadSchema(schemaPath); err != nil {
			return nil, fmt.Errorf("failed to load schema: %w", err)
		}
	}
	
	// Create config directory
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}
	
	// Load existing config or create default
	if err := cm.Load(); err != nil {
		// If config doesn't exist, create with defaults
		if os.IsNotExist(err) {
			cm.applyDefaults()
			if err := cm.Save(); err != nil {
				return nil, fmt.Errorf("failed to save default config: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
	}
	
	return cm, nil
}

// LoadSchema loads and validates a JSON schema
func (cm *ConfigManager) LoadSchema(schemaPath string) error {
	data, err := os.ReadFile(schemaPath)
	if err != nil {
		return err
	}
	
	var schema ConfigSchema
	if err := json.Unmarshal(data, &schema); err != nil {
		return fmt.Errorf("invalid JSON schema: %w", err)
	}
	
	cm.schema = &schema
	return nil
}

// Load loads configuration from file
func (cm *ConfigManager) Load() error {
	data, err := os.ReadFile(cm.configFile)
	if err != nil {
		return err
	}
	
	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("invalid JSON config: %w", err)
	}
	
	// Validate against schema if available
	if cm.schema != nil {
		if err := cm.validateConfig(config); err != nil {
			return fmt.Errorf("config validation failed: %w", err)
		}
	}
	
	cm.config = config
	
	// Update last modified time
	if info, err := os.Stat(cm.configFile); err == nil {
		cm.lastModified = info.ModTime()
	}
	
	return nil
}

// Save saves configuration to file
func (cm *ConfigManager) Save() error {
	// Validate before saving
	if cm.schema != nil {
		if err := cm.validateConfig(cm.config); err != nil {
			return fmt.Errorf("config validation failed: %w", err)
		}
	}
	
	data, err := json.MarshalIndent(cm.config, "", "  ")
	if err != nil {
		return err
	}
	
	// Create backup of existing config
	if _, err := os.Stat(cm.configFile); err == nil {
		backupFile := cm.configFile + ".backup"
		if err := copyFile(cm.configFile, backupFile); err != nil {
			// Log warning but don't fail
			fmt.Printf("Warning: failed to create config backup: %v\n", err)
		}
	}
	
	if err := os.WriteFile(cm.configFile, data, 0644); err != nil {
		return err
	}
	
	// Update last modified time
	cm.lastModified = time.Now()
	
	return nil
}

// Get retrieves a configuration value by key path (dot notation supported)
func (cm *ConfigManager) Get(keyPath string) (interface{}, error) {
	return cm.getNestedValue(cm.config, keyPath)
}

// GetString retrieves a string configuration value
func (cm *ConfigManager) GetString(keyPath string) (string, error) {
	value, err := cm.Get(keyPath)
	if err != nil {
		return "", err
	}
	
	if str, ok := value.(string); ok {
		return str, nil
	}
	
	return "", fmt.Errorf("value at %s is not a string", keyPath)
}

// GetInt retrieves an integer configuration value
func (cm *ConfigManager) GetInt(keyPath string) (int, error) {
	value, err := cm.Get(keyPath)
	if err != nil {
		return 0, err
	}
	
	switch v := value.(type) {
	case int:
		return v, nil
	case float64:
		return int(v), nil
	default:
		return 0, fmt.Errorf("value at %s is not a number", keyPath)
	}
}

// GetBool retrieves a boolean configuration value
func (cm *ConfigManager) GetBool(keyPath string) (bool, error) {
	value, err := cm.Get(keyPath)
	if err != nil {
		return false, err
	}
	
	if b, ok := value.(bool); ok {
		return b, nil
	}
	
	return false, fmt.Errorf("value at %s is not a boolean", keyPath)
}

// GetArray retrieves an array configuration value
func (cm *ConfigManager) GetArray(keyPath string) ([]interface{}, error) {
	value, err := cm.Get(keyPath)
	if err != nil {
		return nil, err
	}
	
	if arr, ok := value.([]interface{}); ok {
		return arr, nil
	}
	
	return nil, fmt.Errorf("value at %s is not an array", keyPath)
}

// GetObject retrieves an object configuration value
func (cm *ConfigManager) GetObject(keyPath string) (map[string]interface{}, error) {
	value, err := cm.Get(keyPath)
	if err != nil {
		return nil, err
	}
	
	if obj, ok := value.(map[string]interface{}); ok {
		return obj, nil
	}
	
	return nil, fmt.Errorf("value at %s is not an object", keyPath)
}

// Set sets a configuration value by key path (dot notation supported)
func (cm *ConfigManager) Set(keyPath string, value interface{}) error {
	// Validate the new value if schema is available
	if cm.schema != nil {
		if err := cm.validateValue(keyPath, value); err != nil {
			return fmt.Errorf("validation failed for %s: %w", keyPath, err)
		}
	}
	
	return cm.setNestedValue(cm.config, keyPath, value)
}

// Delete removes a configuration value by key path
func (cm *ConfigManager) Delete(keyPath string) error {
	return cm.deleteNestedValue(cm.config, keyPath)
}

// Has checks if a configuration key exists
func (cm *ConfigManager) Has(keyPath string) bool {
	_, err := cm.Get(keyPath)
	return err == nil
}

// GetAll returns all configuration data
func (cm *ConfigManager) GetAll() map[string]interface{} {
	return cm.deepCopy(cm.config)
}

// SetAll replaces all configuration data
func (cm *ConfigManager) SetAll(config map[string]interface{}) error {
	// Validate entire config
	if cm.schema != nil {
		if err := cm.validateConfig(config); err != nil {
			return fmt.Errorf("config validation failed: %w", err)
		}
	}
	
	cm.config = cm.deepCopy(config)
	return nil
}

// Merge merges configuration data
func (cm *ConfigManager) Merge(config map[string]interface{}) error {
	merged := cm.deepMerge(cm.config, config)
	
	// Validate merged config
	if cm.schema != nil {
		if err := cm.validateConfig(merged); err != nil {
			return fmt.Errorf("merged config validation failed: %w", err)
		}
	}
	
	cm.config = merged
	return nil
}

// Reset resets configuration to defaults
func (cm *ConfigManager) Reset() error {
	cm.config = make(map[string]interface{})
	cm.applyDefaults()
	return cm.Save()
}

// IsModified checks if config file has been modified since last load
func (cm *ConfigManager) IsModified() (bool, error) {
	info, err := os.Stat(cm.configFile)
	if err != nil {
		return false, err
	}
	
	return info.ModTime().After(cm.lastModified), nil
}

// Reload reloads configuration from file if modified
func (cm *ConfigManager) Reload() error {
	modified, err := cm.IsModified()
	if err != nil {
		return err
	}
	
	if modified {
		return cm.Load()
	}
	
	return nil
}

// Backup creates a backup of the current configuration
func (cm *ConfigManager) Backup(backupPath string) error {
	return copyFile(cm.configFile, backupPath)
}

// Restore restores configuration from backup
func (cm *ConfigManager) Restore(backupPath string) error {
	if err := copyFile(backupPath, cm.configFile); err != nil {
		return err
	}
	
	return cm.Load()
}

// GetConfigPath returns the path to the configuration file
func (cm *ConfigManager) GetConfigPath() string {
	return cm.configFile
}

// GetConfigDir returns the configuration directory
func (cm *ConfigManager) GetConfigDir() string {
	return cm.configDir
}

// Private helper methods

func (cm *ConfigManager) getNestedValue(data map[string]interface{}, keyPath string) (interface{}, error) {
	keys := strings.Split(keyPath, ".")
	current := data
	
	for i, key := range keys {
		if i == len(keys)-1 {
			if value, exists := current[key]; exists {
				return value, nil
			}
			return nil, fmt.Errorf("key %s not found", keyPath)
		}
		
		if next, exists := current[key]; exists {
			if nextMap, ok := next.(map[string]interface{}); ok {
				current = nextMap
			} else {
				return nil, fmt.Errorf("key %s is not an object", strings.Join(keys[:i+1], "."))
			}
		} else {
			return nil, fmt.Errorf("key %s not found", keyPath)
		}
	}
	
	return nil, fmt.Errorf("invalid key path: %s", keyPath)
}

func (cm *ConfigManager) setNestedValue(data map[string]interface{}, keyPath string, value interface{}) error {
	keys := strings.Split(keyPath, ".")
	current := data
	
	for i, key := range keys {
		if i == len(keys)-1 {
			current[key] = value
			return nil
		}
		
		if next, exists := current[key]; exists {
			if nextMap, ok := next.(map[string]interface{}); ok {
				current = nextMap
			} else {
				return fmt.Errorf("key %s is not an object", strings.Join(keys[:i+1], "."))
			}
		} else {
			// Create nested object
			newMap := make(map[string]interface{})
			current[key] = newMap
			current = newMap
		}
	}
	
	return nil
}

func (cm *ConfigManager) deleteNestedValue(data map[string]interface{}, keyPath string) error {
	keys := strings.Split(keyPath, ".")
	current := data
	
	for i, key := range keys {
		if i == len(keys)-1 {
			delete(current, key)
			return nil
		}
		
		if next, exists := current[key]; exists {
			if nextMap, ok := next.(map[string]interface{}); ok {
				current = nextMap
			} else {
				return fmt.Errorf("key %s is not an object", strings.Join(keys[:i+1], "."))
			}
		} else {
			return fmt.Errorf("key %s not found", keyPath)
		}
	}
	
	return nil
}

func (cm *ConfigManager) validateConfig(config map[string]interface{}) error {
	if cm.schema == nil {
		return nil
	}
	
	return cm.validateObject(config, cm.schema)
}

func (cm *ConfigManager) validateValue(keyPath string, value interface{}) error {
	if cm.schema == nil {
		return nil
	}
	
	// Find the schema for this key path
	schema := cm.findSchemaForPath(keyPath)
	if schema == nil {
		return nil // No schema validation for this path
	}
	
	return cm.validateValueAgainstSchema(value, schema)
}

func (cm *ConfigManager) validateObject(obj map[string]interface{}, schema *ConfigSchema) error {
	if schema.Type != "object" {
		return fmt.Errorf("expected object type")
	}
	
	// Check required fields
	for _, required := range schema.Required {
		if _, exists := obj[required]; !exists {
			return fmt.Errorf("required field %s is missing", required)
		}
	}
	
	// Validate properties
	if schema.Properties != nil {
		for key, value := range obj {
			if propSchema, ok := schema.Properties[key]; ok {
				if err := cm.validateValueAgainstSchema(value, propSchema); err != nil {
					return fmt.Errorf("validation failed for property %s: %w", key, err)
				}
			}
		}
	}
	
	return nil
}

func (cm *ConfigManager) validateValueAgainstSchema(value interface{}, schema *PropertySchema) error {
	// Type validation
	switch schema.Type {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
		str := value.(string)
		
		// Length validation
		if schema.MinLength != nil && len(str) < *schema.MinLength {
			return fmt.Errorf("string too short, minimum length: %d", *schema.MinLength)
		}
		if schema.MaxLength != nil && len(str) > *schema.MaxLength {
			return fmt.Errorf("string too long, maximum length: %d", *schema.MaxLength)
		}
		
	case "number":
		var num float64
		switch v := value.(type) {
		case int:
			num = float64(v)
		case float64:
			num = v
		default:
			return fmt.Errorf("expected number, got %T", value)
		}
		
		// Range validation
		if schema.Minimum != nil && num < *schema.Minimum {
			return fmt.Errorf("number too small, minimum: %f", *schema.Minimum)
		}
		if schema.Maximum != nil && num > *schema.Maximum {
			return fmt.Errorf("number too large, maximum: %f", *schema.Maximum)
		}
		
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected boolean, got %T", value)
		}
		
	case "array":
		arr, ok := value.([]interface{})
		if !ok {
			return fmt.Errorf("expected array, got %T", value)
		}
		
		// Validate items if schema provided
		if schema.Items != nil {
			for i, item := range arr {
				if err := cm.validateValueAgainstSchema(item, schema.Items); err != nil {
					return fmt.Errorf("validation failed for array item %d: %w", i, err)
				}
			}
		}
		
	case "object":
		obj, ok := value.(map[string]interface{})
		if !ok {
			return fmt.Errorf("expected object, got %T", value)
		}
		
		// Validate object properties
		objectSchema := &ConfigSchema{
			Type:       "object",
			Properties: schema.Properties,
			Required:   schema.Required,
		}
		return cm.validateObject(obj, objectSchema)
	}
	
	// Enum validation
	if len(schema.Enum) > 0 {
		found := false
		for _, enumValue := range schema.Enum {
			if reflect.DeepEqual(value, enumValue) {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("value not in allowed enum values")
		}
	}
	
	return nil
}

func (cm *ConfigManager) findSchemaForPath(keyPath string) *PropertySchema {
	if cm.schema == nil || cm.schema.Properties == nil {
		return nil
	}
	
	keys := strings.Split(keyPath, ".")
	current := cm.schema.Properties
	
	for i, key := range keys {
		if propSchema, ok := current[key]; ok {
			if i == len(keys)-1 {
				return propSchema
			}
			
			if propSchema.Properties != nil {
				current = propSchema.Properties
			} else {
				return nil
			}
		} else {
			return nil
		}
	}
	
	return nil
}

func (cm *ConfigManager) applyDefaults() {
	if cm.schema != nil {
		cm.applyDefaultsFromSchema(cm.config, cm.schema)
	}
}

func (cm *ConfigManager) applyDefaultsFromSchema(config map[string]interface{}, schema *ConfigSchema) {
	if schema.Properties == nil {
		return
	}
	
	for key, propSchema := range schema.Properties {
		if propSchema.Default != nil {
			if _, exists := config[key]; !exists {
				config[key] = propSchema.Default
			}
		}
		
		// Apply defaults to nested objects
		if propSchema.Type == "object" && propSchema.Properties != nil {
			if obj, ok := config[key].(map[string]interface{}); ok {
				objectSchema := &ConfigSchema{
					Type:       "object",
					Properties: propSchema.Properties,
				}
				cm.applyDefaultsFromSchema(obj, objectSchema)
			}
		}
	}
}

func (cm *ConfigManager) deepCopy(data map[string]interface{}) map[string]interface{} {
	copy := make(map[string]interface{})
	
	for key, value := range data {
		switch v := value.(type) {
		case map[string]interface{}:
			copy[key] = cm.deepCopy(v)
		case []interface{}:
			copySlice := make([]interface{}, len(v))
			for i, item := range v {
				if itemMap, ok := item.(map[string]interface{}); ok {
					copySlice[i] = cm.deepCopy(itemMap)
				} else {
					copySlice[i] = item
				}
			}
			copy[key] = copySlice
		default:
			copy[key] = value
		}
	}
	
	return copy
}

func (cm *ConfigManager) deepMerge(dst, src map[string]interface{}) map[string]interface{} {
	result := cm.deepCopy(dst)
	
	for key, value := range src {
		if existing, exists := result[key]; exists {
			if existingMap, ok := existing.(map[string]interface{}); ok {
				if valueMap, ok := value.(map[string]interface{}); ok {
					result[key] = cm.deepMerge(existingMap, valueMap)
					continue
				}
			}
		}
		result[key] = value
	}
	
	return result
}

// LoadProfile loads a profile from file
func (cm *ConfigManager) LoadProfile(profilePath string) (*Profile, error) {
	data, err := os.ReadFile(profilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read profile: %w", err)
	}
	
	var profile Profile
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("failed to parse profile: %w", err)
	}
	
	return &profile, nil
}

// GetRequirementsForComponents gets requirements for specific components
func (cm *ConfigManager) GetRequirementsForComponents(components []string) map[string]map[string]string {
	// Default requirements
	reqs := map[string]map[string]string{
		"global": {
			"go": ">=1.20",
			"git": ">=2.0.0",
			"claude": "latest",
			"permissions": "required",
		},
	}
	
	// Component-specific requirements
	for _, comp := range components {
		switch comp {
		case "mcp":
			if reqs["mcp"] == nil {
				reqs["mcp"] = make(map[string]string)
			}
			reqs["mcp"]["node"] = ">=18.0.0"
		}
	}
	
	return reqs
}

// ValidateConfigFiles validates configuration files
func (cm *ConfigManager) ValidateConfigFiles() []string {
	errors := []string{}
	
	// Check if config directory exists
	if _, err := os.Stat(cm.configDir); os.IsNotExist(err) {
		errors = append(errors, fmt.Sprintf("config directory not found: %s", cm.configDir))
		return errors
	}
	
	// Validate specific config files
	configFiles := []string{
		"requirements.json",
		"defaults.json",
	}
	
	for _, file := range configFiles {
		path := filepath.Join(cm.configDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			// Optional files, not an error
			continue
		}
		
		// Try to parse JSON
		data, err := os.ReadFile(path)
		if err != nil {
			errors = append(errors, fmt.Sprintf("failed to read %s: %v", file, err))
			continue
		}
		
		var temp interface{}
		if err := json.Unmarshal(data, &temp); err != nil {
			errors = append(errors, fmt.Sprintf("invalid JSON in %s: %v", file, err))
		}
	}
	
	return errors
}

// LoadConfig loads the main configuration
func (cm *ConfigManager) LoadConfig(installDir string) (*Config, error) {
	configPath := filepath.Join(installDir, ".claude", "config.json")
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	
	return &config, nil
}

// SaveConfig saves the main configuration
func (cm *ConfigManager) SaveConfig(installDir string, config *Config) error {
	configPath := filepath.Join(installDir, ".claude", "config.json")
	
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return err
	}
	
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(configPath, data, 0644)
}

// GetDefaultSettings returns default settings
func (cm *ConfigManager) GetDefaultSettings() map[string]interface{} {
	return map[string]interface{}{
		"auto_update":     false,
		"telemetry":       false,
		"log_level":       "info",
		"backup_on_update": true,
		"theme":           "dark",
	}
}

// Utility function to copy files
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	
	return os.WriteFile(dst, data, 0644)
}