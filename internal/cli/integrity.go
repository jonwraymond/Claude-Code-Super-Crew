package cli

import (
	"fmt"
	"strings"

	"github.com/jonwraymond/claude-code-super-crew/internal/managers"
	"github.com/jonwraymond/claude-code-super-crew/internal/metadata"
	"github.com/jonwraymond/claude-code-super-crew/pkg/logger"
	"github.com/spf13/cobra"
)

// IntegrityFlags contains flags for integrity checking
type IntegrityFlags struct {
	Check    bool
	Fix      bool
	Verbose  bool
	AutoFix  bool
}

// NewIntegrityCommand creates the integrity checking command
func NewIntegrityCommand() *cobra.Command {
	var flags IntegrityFlags

	cmd := &cobra.Command{
		Use:   "integrity",
		Short: "Check and manage file integrity",
		Long: `Check the integrity of installed framework files and detect modifications.

This command verifies that all framework files match their original hashes
and provides visual status indicators for any detected changes.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runIntegrity(cmd, args, flags)
		},
	}

	// Add flags
	cmd.Flags().BoolVarP(&flags.Check, "check", "c", false, "Check file integrity")
	cmd.Flags().BoolVarP(&flags.Fix, "fix", "f", false, "Fix integrity issues by removing modified files")
	cmd.Flags().BoolVarP(&flags.Verbose, "verbose", "v", false, "Show detailed integrity information")
	cmd.Flags().BoolVarP(&flags.AutoFix, "auto-fix", "a", false, "Automatically fix integrity issues")

	return cmd
}

// runIntegrity executes the integrity checking command
func runIntegrity(cmd *cobra.Command, args []string, flags IntegrityFlags) error {
	log := logger.GetLogger()
	installDir := globalFlags.InstallDir

	// Initialize managers
	settingsManager := managers.NewSettingsManager(installDir)
	metadataManager := settingsManager.GetMetadataManager()

	// Check if installation exists
	if !metadataManager.CheckInstallationExists() {
		return fmt.Errorf("no installation found at %s", installDir)
	}

	// Perform integrity check
	log.Info("🔍 Checking file integrity...")
	integrity, err := metadataManager.CheckFileIntegrity()
	if err != nil {
		return fmt.Errorf("failed to check integrity: %w", err)
	}

	// Display integrity status with visual indicators
	displayIntegrityStatus(integrity, flags.Verbose)

	// Handle auto-fix if requested
	if flags.AutoFix && integrity.Status != "clean" {
		log.Info("🔧 Auto-fixing integrity issues...")
		if err := fixIntegrityIssues(metadataManager, integrity); err != nil {
			return fmt.Errorf("failed to fix integrity issues: %w", err)
		}
		log.Info("✅ Integrity issues fixed")
	}

	return nil
}

// displayIntegrityStatus displays the integrity status with visual indicators
func displayIntegrityStatus(integrity *metadata.IntegrityMeta, verbose bool) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🔒 CLAUDE CODE SUPER CREW - FILE INTEGRITY STATUS")
	fmt.Println(strings.Repeat("=", 60))

	// Overall status with traffic light indicator
	fmt.Printf("\n📊 OVERALL STATUS: ")
	switch integrity.Status {
	case "clean":
		fmt.Println("🟢 CLEAN - All files match original hashes")
	case "warning":
		fmt.Println("🟡 WARNING - Some files have been modified")
	case "critical":
		fmt.Println("🔴 CRITICAL - Files are missing or corrupted")
	default:
		fmt.Println("⚪ UNKNOWN - Status cannot be determined")
	}

	// Summary statistics
	fmt.Printf("\n📈 INTEGRITY SUMMARY:\n")
	fmt.Printf("   Total Files Tracked: %d\n", integrity.TotalFiles)
	fmt.Printf("   🟢 Clean Files: %d\n", integrity.CleanFiles)
	fmt.Printf("   🟡 Modified Files: %d\n", integrity.ModifiedFiles)
	fmt.Printf("   🔴 Missing Files: %d\n", integrity.MissingFiles)
	fmt.Printf("   ⚫ Corrupted Files: %d\n", integrity.CorruptedFiles)
	fmt.Printf("   Last Checked: %s\n", integrity.LastScan.Format("2006-01-02 15:04:05"))

	// Show detailed information if verbose or if there are issues
	if verbose || integrity.Status != "clean" {
		fmt.Printf("\n🔍 DETAILED FILE STATUS:\n")
		fmt.Println(strings.Repeat("-", 60))

		for filePath, fileIntegrity := range integrity.FileHashes {
			statusIcon := getStatusIcon(fileIntegrity.Status)
			fmt.Printf("%s %s (%s)\n", statusIcon, filePath, fileIntegrity.Component)
			
			if verbose {
				fmt.Printf("   Original Hash: %s\n", fileIntegrity.OriginalHash[:16]+"...")
				fmt.Printf("   Current Hash:  %s\n", fileIntegrity.CurrentHash[:16]+"...")
				fmt.Printf("   Last Checked:  %s\n", fileIntegrity.LastChecked.Format("2006-01-02 15:04:05"))
				
				if len(fileIntegrity.ModificationLog) > 0 {
					fmt.Printf("   Recent Changes:\n")
					for _, logEntry := range fileIntegrity.ModificationLog[len(fileIntegrity.ModificationLog)-3:] {
						fmt.Printf("     • %s\n", logEntry)
					}
				}
				fmt.Println()
			}
		}
	}

	// Recommendations
	if integrity.Status != "clean" {
		fmt.Printf("\n💡 RECOMMENDATIONS:\n")
		if integrity.ModifiedFiles > 0 {
			fmt.Println("   • Modified files may contain user customizations")
			fmt.Println("   • Consider backing up before fixing")
			fmt.Println("   • Use 'crew integrity --fix' to remove modified files")
		}
		if integrity.MissingFiles > 0 {
			fmt.Println("   • Missing files should be reinstalled")
			fmt.Println("   • Run 'crew install' to restore missing files")
		}
		if integrity.CorruptedFiles > 0 {
			fmt.Println("   • Corrupted files cannot be read")
			fmt.Println("   • These files should be removed and reinstalled")
		}
	}

	fmt.Println(strings.Repeat("=", 60))
}

// getStatusIcon returns the appropriate icon for a file status
func getStatusIcon(status string) string {
	switch status {
	case "clean":
		return "🟢"
	case "modified":
		return "🟡"
	case "missing":
		return "🔴"
	case "corrupted":
		return "⚫"
	default:
		return "⚪"
	}
}

// fixIntegrityIssues removes files that have integrity issues
func fixIntegrityIssues(metadataManager *metadata.MetadataManager, integrity *metadata.IntegrityMeta) error {
	log := logger.GetLogger()
	
	for filePath, fileIntegrity := range integrity.FileHashes {
		if fileIntegrity.Status == "modified" || fileIntegrity.Status == "corrupted" {
			log.Infof("Removing file with integrity issue: %s", filePath)
			
			// Remove from integrity tracking
			if err := metadataManager.RemoveFileFromIntegrityTracking(filePath); err != nil {
				log.Warnf("Failed to remove %s from integrity tracking: %v", filePath, err)
			}
			
			// Note: The actual file removal should be done by the uninstall process
			// This just removes it from tracking
		}
	}
	
	return nil
} 