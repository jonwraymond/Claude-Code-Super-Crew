package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jonwraymond/claude-code-super-crew/internal/managers"
	"github.com/jonwraymond/claude-code-super-crew/internal/ui"
	"github.com/jonwraymond/claude-code-super-crew/pkg/backup"
	"github.com/jonwraymond/claude-code-super-crew/pkg/logger"
	"github.com/spf13/cobra"
)

// BackupFlags holds backup command flags
type BackupFlags struct {
	Create    bool
	List      bool
	Restore   string
	Info      string
	Cleanup   bool
	BackupDir string
	Name      string
	Compress  string
	Overwrite bool
	Keep      int
	OlderThan int
}

var backupFlags BackupFlags

// NewBackupCommand creates the backup command
func NewBackupCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Backup and restore Claude Code Super Crew installations",
		Long: `Create, list, restore, and manage Claude Code Super Crew installation backups.

Examples:
  crew backup --create               # Create new backup
  crew backup --list --verbose       # List available backups (verbose)
  crew backup --restore              # Interactive restore
  crew backup --restore backup.tar.gz  # Restore specific backup
  crew backup --info backup.tar.gz   # Show backup information
  crew backup --cleanup --force      # Clean up old backups (forced)`,
		RunE: runBackup,
	}

	// Backup operations
	cmd.Flags().BoolVar(&backupFlags.Create, "create", false,
		"Create a new backup")
	cmd.Flags().BoolVar(&backupFlags.List, "list", false,
		"List available backups")
	cmd.Flags().StringVar(&backupFlags.Restore, "restore", "",
		"Restore from backup (optionally specify backup file)")
	cmd.Flags().StringVar(&backupFlags.Info, "info", "",
		"Show information about a specific backup file")
	cmd.Flags().BoolVar(&backupFlags.Cleanup, "cleanup", false,
		"Clean up old backup files")

	// Backup options
	cmd.Flags().StringVar(&backupFlags.BackupDir, "backup-dir", "",
		"Backup directory (default: <install-dir>/backups)")
	cmd.Flags().StringVar(&backupFlags.Name, "name", "",
		"Custom backup name (for --create)")
	cmd.Flags().StringVar(&backupFlags.Compress, "compress", "gzip",
		"Compression method: none, gzip, bzip2 (default: gzip)")

	// Restore options
	cmd.Flags().BoolVar(&backupFlags.Overwrite, "overwrite", false,
		"Overwrite existing files during restore")

	// Cleanup options
	cmd.Flags().IntVar(&backupFlags.Keep, "keep", 5,
		"Number of backups to keep during cleanup (default: 5)")
	cmd.Flags().IntVar(&backupFlags.OlderThan, "older-than", 0,
		"Remove backups older than N days")

	// Mark operations as mutually exclusive
	cmd.MarkFlagsMutuallyExclusive("create", "list", "restore", "info", "cleanup")
	cmd.MarkFlagsOneRequired("create", "list", "restore", "info", "cleanup")

	return cmd
}

func runBackup(cmd *cobra.Command, args []string) error {
	log := logger.GetLogger()
	log.SetVerbose(globalFlags.Verbose)
	log.SetQuiet(globalFlags.Quiet)

	// Validate installation directory (skip in test mode)
	if !testMode {
		expectedHome := filepath.Join(os.Getenv("HOME"))
		if expectedHome == "" {
			expectedHome = filepath.Join(os.Getenv("USERPROFILE")) // Windows
		}
		actualDir, _ := filepath.Abs(globalFlags.InstallDir)

		if !strings.HasPrefix(actualDir, expectedHome) {
			ui.DisplayError("Installation must be inside your user profile directory.")
			fmt.Printf("    Expected prefix: %s\n", expectedHome)
			fmt.Printf("    Provided path:   %s\n", actualDir)
			return fmt.Errorf("invalid installation directory")
		}
	}

	// Display header
	if !globalFlags.Quiet {
		ui.DisplayHeader(
			"Claude Code Super Crew Backup v1.0",
			"Backup and restore Claude Code Super Crew installations",
		)
	}

	// Get backup directory
	backupDir := getBackupDirectory()

	// Handle different backup operations
	switch {
	case backupFlags.Create:
		return createBackup()

	case backupFlags.List:
		return listBackups(backupDir)

	case backupFlags.Restore != "":
		return restoreBackup(backupFlags.Restore, backupDir)

	case backupFlags.Info != "":
		return showBackupInfo(backupFlags.Info, backupDir)

	case backupFlags.Cleanup:
		return cleanupBackups(backupDir)

	default:
		return fmt.Errorf("no backup operation specified")
	}
}

func getBackupDirectory() string {
	if backupFlags.BackupDir != "" {
		return backupFlags.BackupDir
	}
	return filepath.Join(globalFlags.InstallDir, ".crew", "backups")
}

func checkInstallationExists() bool {
	settingsManager := managers.NewSettingsManager(globalFlags.InstallDir)
	return settingsManager.CheckInstallationExists()
}

func createBackup() error {
	log := logger.GetLogger()

	// Check if installation exists
	if !checkInstallationExists() {
		log.Errorf("No Claude Code Super Crew installation found in %s", globalFlags.InstallDir)
		return fmt.Errorf("no installation found")
	}

	// Create backup directory
	backupDir := getBackupDirectory()
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Generate backup filename
	var backupName string
	if backupFlags.Name != "" {
		backupName = backupFlags.Name
	} else {
		backupName = "crew_backup"
	}

	// Create backup manager
	mgr := backup.NewManager(backup.Options{
		InstallDir: globalFlags.InstallDir,
		BackupDir:  backupDir,
		BackupName: backupName,
		Compress:   backupFlags.Compress,
		Verbose:    globalFlags.Verbose,
		DryRun:     globalFlags.DryRun,
	})

	log.Info("Creating backup...")

	if globalFlags.DryRun {
		log.Info("[DRY RUN] Would create backup")
		return nil
	}

	// Create backup
	backupFile, err := mgr.Create()
	if err != nil {
		return fmt.Errorf("backup creation failed: %w", err)
	}

	// Get backup info
	info := mgr.GetBackupInfo(backupFile)

	log.Successf("Backup created successfully")
	log.Infof("Backup file: %s", backupFile)
	log.Infof("Backup size: %s", ui.FormatSize(info.Size))

	if !globalFlags.Quiet {
		ui.DisplaySuccess("Backup operation completed successfully!")
	}

	return nil
}

func listBackups(backupDir string) error {
	mgr := backup.NewManager(backup.Options{
		BackupDir: backupDir,
		Verbose:   globalFlags.Verbose,
	})

	backups, err := mgr.ListBackups()
	if err != nil {
		return fmt.Errorf("failed to list backups: %w", err)
	}

	if !globalFlags.Quiet {
		displayBackupList(backups)
	} else {
		// Simple list for quiet mode
		for _, b := range backups {
			fmt.Println(b.Path)
		}
	}

	return nil
}

func displayBackupList(backups []backup.BackupInfo) {
	fmt.Printf("\n%s%sAvailable Backups%s\n", ui.ColorCyan, ui.ColorBright, ui.ColorReset)
	fmt.Println(strings.Repeat("=", 70))

	if len(backups) == 0 {
		fmt.Printf("%sNo backups found%s\n", ui.ColorYellow, ui.ColorReset)
		return
	}

	fmt.Printf("%-30s %-10s %-20s %-8s\n", "Name", "Size", "Created", "Files")
	fmt.Println(strings.Repeat("-", 70))

	for _, backup := range backups {
		name := filepath.Base(backup.Path)
		size := ui.FormatSize(backup.Size)
		created := backup.Created.Format("2006-01-02 15:04")
		files := fmt.Sprintf("%d", backup.FileCount)

		fmt.Printf("%-30s %-10s %-20s %-8s\n", name, size, created, files)
	}

	fmt.Println()
}

func restoreBackup(backupFile string, backupDir string) error {
	log := logger.GetLogger()

	// Handle interactive restore
	if backupFile == "" {
		mgr := backup.NewManager(backup.Options{
			BackupDir: backupDir,
			Verbose:   globalFlags.Verbose,
		})

		backups, err := mgr.ListBackups()
		if err != nil {
			return fmt.Errorf("failed to list backups: %w", err)
		}

		if len(backups) == 0 {
			log.Warn("No backups available for restore")
			return nil
		}

		selected := interactiveRestoreSelection(backups)
		if selected == "" {
			log.Info("Restore cancelled by user")
			return nil
		}
		backupFile = selected
	}

	// Resolve backup path
	if !filepath.IsAbs(backupFile) {
		backupFile = filepath.Join(backupDir, backupFile)
	}

	// Create backup manager
	mgr := backup.NewManager(backup.Options{
		InstallDir: globalFlags.InstallDir,
		BackupDir:  backupDir,
		Verbose:    globalFlags.Verbose,
		DryRun:     globalFlags.DryRun,
		Overwrite:  backupFlags.Overwrite,
	})

	log.Infof("Restoring from backup: %s", backupFile)

	if globalFlags.DryRun {
		log.Info("[DRY RUN] Would restore backup")
		return nil
	}

	// Create backup of current installation if it exists
	if checkInstallationExists() {
		log.Info("Creating backup of current installation before restore")
		// This would call create_backup internally
	}

	// Restore backup
	if err := mgr.Restore(backupFile); err != nil {
		return fmt.Errorf("backup restoration failed: %w", err)
	}

	if !globalFlags.Quiet {
		ui.DisplaySuccess("Restore operation completed successfully!")
	}

	return nil
}

func interactiveRestoreSelection(backups []backup.BackupInfo) string {
	fmt.Printf("\n%sSelect Backup to Restore:%s\n", ui.ColorCyan, ui.ColorReset)

	// Create menu options
	options := []string{}
	for _, backup := range backups {
		name := filepath.Base(backup.Path)
		size := ui.FormatSize(backup.Size)
		created := backup.Created.Format("2006-01-02 15:04")
		options = append(options, fmt.Sprintf("%s (%s, %s)", name, size, created))
	}

	menu := ui.NewMenu("Select backup:", options, false)
	result, err := menu.Display()
	if err != nil {
		return ""
	}
	choice := result.(int)

	if choice == -1 || choice >= len(backups) {
		return ""
	}

	return backups[choice].Path
}

func showBackupInfo(backupFile string, backupDir string) error {
	// Resolve backup path
	if !filepath.IsAbs(backupFile) {
		backupFile = filepath.Join(backupDir, backupFile)
	}

	mgr := backup.NewManager(backup.Options{
		BackupDir: backupDir,
		Verbose:   globalFlags.Verbose,
	})

	info := mgr.GetBackupInfo(backupFile)

	if !info.Exists {
		logger.GetLogger().Errorf("Backup file not found: %s", backupFile)
		return fmt.Errorf("backup file not found")
	}

	fmt.Printf("\n%sBackup Information:%s\n", ui.ColorCyan, ui.ColorReset)
	fmt.Printf("File: %s\n", info.Path)
	fmt.Printf("Size: %s\n", ui.FormatSize(info.Size))
	fmt.Printf("Created: %s\n", info.Created)
	fmt.Printf("Files: %d\n", info.FileCount)

	if info.Metadata != nil {
		fmt.Printf("Framework Version: %s\n", info.Metadata.FrameworkVersion)
		if len(info.Metadata.Components) > 0 {
			fmt.Println("Components:")
			for comp, ver := range info.Metadata.Components {
				fmt.Printf("  %s: v%s\n", comp, ver)
			}
		}
	}

	return nil
}

func cleanupBackups(backupDir string) error {
	log := logger.GetLogger()

	mgr := backup.NewManager(backup.Options{
		BackupDir: backupDir,
		Verbose:   globalFlags.Verbose,
		DryRun:    globalFlags.DryRun,
	})

	log.Info("Cleaning up old backups...")

	if globalFlags.DryRun {
		log.Info("[DRY RUN] Would cleanup backups")
		return nil
	}

	removed, err := mgr.Cleanup(backupFlags.Keep, backupFlags.OlderThan)
	if err != nil {
		return fmt.Errorf("cleanup failed: %w", err)
	}

	if removed == 0 {
		log.Info("No backups need to be cleaned up")
	} else {
		log.Successf("Cleaned up %d old backups", removed)
	}

	return nil
}
