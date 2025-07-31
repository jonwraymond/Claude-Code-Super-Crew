package managers

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// SecurityValidator provides security checks for the installation system
type SecurityValidator struct{}

// NewSecurityValidator creates a new security validator
func NewSecurityValidator() *SecurityValidator {
	return &SecurityValidator{}
}

// ValidateInstallationTarget checks if an installation target is safe
func (sv *SecurityValidator) ValidateInstallationTarget(targetPath string) (bool, []string) {
	var errors []string

	// Ensure target path is absolute
	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		errors = append(errors, fmt.Sprintf("Cannot resolve absolute path: %v", err))
		return false, errors
	}

	// Check for dangerous paths
	dangerousPaths := []string{
		"/",
		"/bin",
		"/boot",
		"/dev",
		"/etc",
		"/lib",
		"/lib64",
		"/proc",
		"/root",
		"/sbin",
		"/sys",
		"/usr",
		"/usr/bin",
		"/usr/sbin",
		"/var",
		"C:\\Windows",
		"C:\\Program Files",
		"C:\\Program Files (x86)",
	}

	cleanPath := filepath.Clean(absPath)
	for _, dangerous := range dangerousPaths {
		if cleanPath == dangerous || strings.HasPrefix(cleanPath, dangerous+string(filepath.Separator)) {
			errors = append(errors, fmt.Sprintf("Installation to system directory '%s' is not allowed", dangerous))
			return false, errors
		}
	}

	// Check if path contains suspicious patterns
	suspiciousPatterns := []string{
		"..",
		"~",
		"$",
		"`",
		"|",
		">",
		"<",
		"&",
		";",
		"\\",
	}

	for _, pattern := range suspiciousPatterns {
		if strings.Contains(cleanPath, pattern) {
			errors = append(errors, fmt.Sprintf("Path contains suspicious pattern '%s'", pattern))
			return false, errors
		}
	}

	return len(errors) == 0, errors
}

// CheckPermissions verifies the required permissions on a path
func (sv *SecurityValidator) CheckPermissions(path string, requiredPerms []string) (bool, []string) {
	var errors []string

	// Check if path exists
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Path doesn't exist, check parent directory permissions
			parent := filepath.Dir(path)
			return sv.CheckPermissions(parent, requiredPerms)
		}
		errors = append(errors, fmt.Sprintf("Cannot stat path: %v", err))
		return false, errors
	}

	// Check permissions
	for _, perm := range requiredPerms {
		switch perm {
		case "read":
			// Try to open for reading
			if info.IsDir() {
				if _, err := os.ReadDir(path); err != nil {
					errors = append(errors, "No read permission")
				}
			} else {
				if file, err := os.Open(path); err != nil {
					errors = append(errors, "No read permission")
				} else {
					file.Close()
				}
			}

		case "write":
			// Try to create a test file
			testPath := filepath.Join(path, ".claude_write_test")
			if !info.IsDir() {
				testPath = path + ".claude_write_test"
			}

			if file, err := os.Create(testPath); err != nil {
				errors = append(errors, "No write permission")
			} else {
				file.Close()
				os.Remove(testPath)
			}

		case "execute":
			// Check execute bit
			if info.Mode()&0111 == 0 {
				errors = append(errors, "No execute permission")
			}
		}
	}

	return len(errors) == 0, errors
}

// ValidateComponentFiles checks if component files are safe to install
func (sv *SecurityValidator) ValidateComponentFiles(files []string, sourceDir, targetDir string) (bool, []string) {
	var errors []string

	for _, file := range files {
		// Validate file name
		if err := sv.validateFileName(file); err != nil {
			errors = append(errors, err.Error())
			continue
		}

		// Check source file exists and is readable
		sourcePath := filepath.Join(sourceDir, file)
		if info, err := os.Stat(sourcePath); err != nil {
			errors = append(errors, fmt.Sprintf("Cannot access source file %s: %v", file, err))
		} else if info.IsDir() {
			errors = append(errors, fmt.Sprintf("Source path %s is a directory, not a file", file))
		}

		// Validate target path
		targetPath := filepath.Join(targetDir, file)
		if isValid, validationErrors := sv.ValidateInstallationTarget(targetPath); !isValid {
			errors = append(errors, validationErrors...)
		}
	}

	return len(errors) == 0, errors
}

// validateFileName checks if a filename is safe
func (sv *SecurityValidator) validateFileName(filename string) error {
	// Check for empty filename
	if filename == "" {
		return fmt.Errorf("empty filename")
	}

	// Check for path traversal
	if strings.Contains(filename, "..") {
		return fmt.Errorf("filename contains path traversal '..'")
	}

	// Check for absolute paths
	if filepath.IsAbs(filename) {
		return fmt.Errorf("filename must not be an absolute path")
	}

	// Check for suspicious characters
	suspiciousChars := []string{
		"|", ">", "<", "&", ";", "`", "$", "~",
	}

	for _, char := range suspiciousChars {
		if strings.Contains(filename, char) {
			return fmt.Errorf("filename contains suspicious character '%s'", char)
		}
	}

	// Check for hidden files (optional - may want to allow these)
	if strings.HasPrefix(filepath.Base(filename), ".") && filename != ".claude" {
		// Allow .claude directory but warn about other hidden files
		// This is informational, not an error
	}

	return nil
}

// IsPathSafe performs a comprehensive safety check on a path
func (sv *SecurityValidator) IsPathSafe(path string) bool {
	isValid, _ := sv.ValidateInstallationTarget(path)
	return isValid
}

// CheckFileIntegrity verifies file integrity using SHA-256 hash
func (sv *SecurityValidator) CheckFileIntegrity(path string, expectedHash string) (bool, error) {
	// Check if file exists
	if _, err := os.Stat(path); err != nil {
		return false, fmt.Errorf("file not found: %w", err)
	}
	
	// If no expected hash provided, just verify file exists
	if expectedHash == "" {
		return true, nil
	}
	
	// Open file for reading
	file, err := os.Open(path)
	if err != nil {
		return false, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	
	// Calculate SHA-256 hash
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return false, fmt.Errorf("failed to read file: %w", err)
	}
	
	// Compare hashes
	actualHash := fmt.Sprintf("%x", hasher.Sum(nil))
	if actualHash != expectedHash {
		return false, fmt.Errorf("hash mismatch: expected %s, got %s", expectedHash, actualHash)
	}
	
	return true, nil
}