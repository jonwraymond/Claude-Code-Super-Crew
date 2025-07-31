package versioning

import (
	"os"
	"testing"
)

func TestVersionManager(t *testing.T) {
	// Create temporary directory for tests
	tempDir, err := os.MkdirTemp("", "version-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	vm := NewVersionManager(tempDir)

	t.Run("StandardizeAllVersions", func(t *testing.T) {
		// Standardize all versions to 1.0.0
		if err := vm.StandardizeAllVersions(); err != nil {
			t.Errorf("Failed to standardize versions: %v", err)
		}
		
		// Check framework version
		version, err := vm.GetCurrentVersion()
		if err != nil {
			t.Errorf("Failed to get version: %v", err)
		}
		if version != "1.0.0" {
			t.Errorf("Expected version 1.0.0, got %s", version)
		}
		
		// Check component versions
		components := []string{"core", "commands", "hooks", "mcp"}
		for _, comp := range components {
			compVersion, err := vm.GetComponentVersion(comp)
			if err != nil {
				t.Errorf("Failed to get %s version: %v", comp, err)
			}
			if compVersion != "1.0.0" {
				t.Errorf("Expected %s version 1.0.0, got %s", comp, compVersion)
			}
		}
	})

	t.Run("VersionComparison", func(t *testing.T) {
		tests := []struct {
			v1       string
			v2       string
			expected int
		}{
			{"1.0.0", "1.0.0", 0},
			{"1.0.0", "1.0.1", -1},
			{"1.0.1", "1.0.0", 1},
			{"1.0.0", "2.0.0", -1},
			{"2.0.0", "1.9.9", 1},
			{"1.2.3", "1.2.3", 0},
		}
		
		for _, test := range tests {
			result := vm.CompareVersions(test.v1, test.v2)
			if result != test.expected {
				t.Errorf("CompareVersions(%s, %s) = %d, expected %d", 
					test.v1, test.v2, result, test.expected)
			}
		}
	})

	t.Run("ValidVersionCheck", func(t *testing.T) {
		tests := []struct {
			version string
			valid   bool
		}{
			{"1.0.0", true},
			{"0.0.1", true},
			{"10.20.30", true},
			{"1.0", false},
			{"1.0.0.0", false},
			{"v1.0.0", false},
			{"1.a.0", false},
			{"", false},
		}
		
		for _, test := range tests {
			result := vm.IsValidVersion(test.version)
			if result != test.valid {
				t.Errorf("IsValidVersion(%s) = %v, expected %v", 
					test.version, result, test.valid)
			}
		}
	})

	t.Run("UpdateCheck", func(t *testing.T) {
		// Set current version
		if err := vm.SetVersion("1.0.0"); err != nil {
			t.Fatalf("Failed to set version: %v", err)
		}
		
		// Check for updates
		needsUpdate, err := vm.CheckForUpdates("1.0.1")
		if err != nil {
			t.Errorf("Failed to check updates: %v", err)
		}
		if !needsUpdate {
			t.Error("Should need update from 1.0.0 to 1.0.1")
		}
		
		// Check when no update needed
		needsUpdate, err = vm.CheckForUpdates("1.0.0")
		if err != nil {
			t.Errorf("Failed to check updates: %v", err)
		}
		if needsUpdate {
			t.Error("Should not need update when versions are equal")
		}
		
		// Check when current version is newer
		needsUpdate, err = vm.CheckForUpdates("0.9.0")
		if err != nil {
			t.Errorf("Failed to check updates: %v", err)
		}
		if needsUpdate {
			t.Error("Should not need update when current version is newer")
		}
	})

	t.Run("MetadataPersistence", func(t *testing.T) {
		// Set version with metadata
		if err := vm.SetVersion("1.0.0"); err != nil {
			t.Fatalf("Failed to set version: %v", err)
		}
		
		// Set component versions
		if err := vm.SetComponentVersion("core", "1.0.0"); err != nil {
			t.Fatalf("Failed to set component version: %v", err)
		}
		
		// Create new version manager to test persistence
		vm2 := NewVersionManager(tempDir)
		
		// Check that data persisted
		version, err := vm2.GetCurrentVersion()
		if err != nil {
			t.Errorf("Failed to get persisted version: %v", err)
		}
		if version != "1.0.0" {
			t.Errorf("Expected persisted version 1.0.0, got %s", version)
		}
		
		// Check component version persisted
		compVersion, err := vm2.GetComponentVersion("core")
		if err != nil {
			t.Errorf("Failed to get persisted component version: %v", err)
		}
		if compVersion != "1.0.0" {
			t.Errorf("Expected persisted component version 1.0.0, got %s", compVersion)
		}
	})

	t.Run("VersionHistory", func(t *testing.T) {
		// Set initial version
		if err := vm.SetVersion("0.9.0"); err != nil {
			t.Fatalf("Failed to set version: %v", err)
		}
		
		// Update to new version
		if err := vm.SetVersion("1.0.0"); err != nil {
			t.Fatalf("Failed to update version: %v", err)
		}
		
		// Check history
		history, err := vm.GetVersionHistory()
		if err != nil {
			t.Errorf("Failed to get version history: %v", err)
		}
		
		if len(history) < 1 {
			t.Error("Version history should have at least one entry")
		}
		if history[0] != "1.0.0" {
			t.Errorf("Expected current version 1.0.0 in history, got %s", history[0])
		}
	})
}