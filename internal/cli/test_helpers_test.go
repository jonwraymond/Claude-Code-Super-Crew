package cli

import (
	"os"
	"testing"
)

// TestMain runs before all tests in the cli package
func TestMain(m *testing.M) {
	// Enable test mode to bypass certain validations
	SetTestMode(true)
	
	// Run tests
	code := m.Run()
	
	// Cleanup
	SetTestMode(false)
	
	os.Exit(code)
}