package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/jonwraymond/claude-code-super-crew/internal/metadata"
)

type StatusCommand struct {
	installDir string
	format     string
	verbose    bool
	components []string
}

func NewStatusCommand() *cobra.Command {
	sc := &StatusCommand{}

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show detailed status of all components and features",
		Long: `Display comprehensive status information including:
- Component versions and installation status
- Document versions and integrity
- Feature flags and configuration
- Installation metadata and totals`,
		RunE: sc.Execute,
	}

	cmd.Flags().StringVar(&sc.format, "format", "table", "Output format: table, json, yaml")
	cmd.Flags().BoolVarP(&sc.verbose, "verbose", "v", false, "Show detailed information")
	cmd.Flags().StringSliceVar(&sc.components, "components", []string{}, "Show status for specific components only")
	cmd.Flags().StringVar(&sc.installDir, "install-dir", "", "Installation directory (default: ~/.claude)")

	return cmd
}

func (sc *StatusCommand) Execute(cmd *cobra.Command, args []string) error {
	// Set default install directory
	if sc.installDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		sc.installDir = filepath.Join(homeDir, ".claude")
	}

	// Check if installation exists
	if _, err := os.Stat(sc.installDir); os.IsNotExist(err) {
		fmt.Printf("❌ Claude Code Super Crew is not installed in %s\n", sc.installDir)
		fmt.Println("Run 'crew install' to install the framework")
		return nil
	}

	// Load metadata
	metaMgr := metadata.NewMetadataManager(sc.installDir)
	metadata, err := metaMgr.RefreshMetadata()
	if err != nil {
		return fmt.Errorf("failed to load metadata: %w", err)
	}

	// Display status based on format
	switch sc.format {
	case "json":
		return sc.displayJSON(metadata)
	case "yaml":
		return sc.displayYAML(metadata)
	default:
		return sc.displayTable(metadata)
	}
}

func (sc *StatusCommand) displayTable(meta *metadata.UnifiedMetadata) error {
	fmt.Printf("\n%s\n", colorize("🔍 Claude Code Super Crew Status", "blue", true))
	fmt.Printf("%s\n\n", strings.Repeat("=", 60))

	// Framework Information
	fmt.Printf("%s\n", colorize("📋 Framework Information", "cyan", true))
	fmt.Printf("├─ Version: %s\n", colorize(meta.Framework.Version, "green", false))
	fmt.Printf("├─ Release Date: %s\n", meta.Framework.ReleaseDate)
	fmt.Printf("├─ Last Updated: %s\n", meta.Framework.UpdatedAt.Format("2006-01-02 15:04:05"))
	if meta.Framework.PreviousVersion != "" {
		fmt.Printf("└─ Previous Version: %s\n", meta.Framework.PreviousVersion)
	}
	fmt.Println()

	// Installation Information
	fmt.Printf("%s\n", colorize("🏗️  Installation Information", "cyan", true))
	fmt.Printf("├─ Install Directory: %s\n", meta.Installation.InstallDir)
	fmt.Printf("├─ Installed At: %s\n", meta.Installation.InstalledAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("├─ Installer Version: %s\n", meta.Installation.InstallerVersion)
	fmt.Printf("├─ Total Size: %s\n", formatBytes(meta.Installation.TotalSize))
	fmt.Printf("└─ Total Files: %d\n", meta.Installation.TotalFiles)
	fmt.Println()

	// Components Status
	fmt.Printf("%s\n", colorize("🧩 Components Status", "cyan", true))
	sc.displayComponentsTable(meta.Components)
	fmt.Println()

	// Documents Status
	if sc.verbose {
		fmt.Printf("%s\n", colorize("📄 Documents Status", "cyan", true))
		sc.displayDocumentsTable(meta.Documents)
		fmt.Println()
	}

	// Features Status
	if len(meta.Features) > 0 {
		fmt.Printf("%s\n", colorize("🎛️  Feature Flags", "cyan", true))
		sc.displayFeaturesTable(meta.Features)
		fmt.Println()
	}

	return nil
}

func (sc *StatusCommand) displayComponentsTable(components map[string]metadata.ComponentMeta) {
	if len(components) == 0 {
		fmt.Println("└─ No components found")
		return
	}

	// Sort components by name
	var names []string
	for name := range components {
		if len(sc.components) == 0 || contains(sc.components, name) {
			names = append(names, name)
		}
	}
	sort.Strings(names)

	for i, name := range names {
		comp := components[name]
		isLast := i == len(names)-1
		prefix := "├─"
		if isLast {
			prefix = "└─"
		}

		status := sc.getStatusIcon(comp.Status)
		fmt.Printf("%s %s %s (v%s) - %s\n", prefix, status, name, comp.Version, comp.Status)

		if sc.verbose {
			subPrefix := "│  "
			if isLast {
				subPrefix = "   "
			}
			
			if len(comp.Dependencies) > 0 {
				fmt.Printf("%s   Dependencies: %s\n", subPrefix, strings.Join(comp.Dependencies, ", "))
			}
			fmt.Printf("%s   Size: %s (%d files)\n", subPrefix, formatBytes(comp.Size), comp.FileCount)
			fmt.Printf("%s   Updated: %s\n", subPrefix, comp.UpdatedAt.Format("2006-01-02 15:04:05"))
			if comp.PreviousVersion != "" {
				fmt.Printf("%s   Previous: v%s\n", subPrefix, comp.PreviousVersion)
			}
		}
	}
}

func (sc *StatusCommand) displayDocumentsTable(documents map[string]metadata.DocumentMeta) {
	if len(documents) == 0 {
		fmt.Println("└─ No documents found")
		return
	}

	// Group documents by component
	byComponent := make(map[string][]string)
	for docName, doc := range documents {
		byComponent[doc.Component] = append(byComponent[doc.Component], docName)
	}

	// Sort components
	var components []string
	for comp := range byComponent {
		components = append(components, comp)
	}
	sort.Strings(components)

	for i, comp := range components {
		isLastComp := i == len(components)-1
		prefix := "├─"
		if isLastComp {
			prefix = "└─"
		}

		fmt.Printf("%s %s:\n", prefix, colorize(comp, "yellow", false))

		docs := byComponent[comp]
		sort.Strings(docs)

		for j, docName := range docs {
			doc := documents[docName]
			isLastDoc := j == len(docs)-1
			
			var subPrefix string
			if isLastComp {
				subPrefix = "   "
			} else {
				subPrefix = "│  "
			}
			
			if isLastDoc {
				subPrefix += "└─"
			} else {
				subPrefix += "├─"
			}

			status := sc.getStatusIcon(doc.Status)
			fmt.Printf("%s %s %s (v%s) - %s\n", subPrefix, status, docName, doc.Version, formatBytes(doc.Size))
		}
	}
}

func (sc *StatusCommand) displayFeaturesTable(features map[string]metadata.FeatureMeta) {
	// Sort features by name
	var names []string
	for name := range features {
		names = append(names, name)
	}
	sort.Strings(names)

	for i, name := range names {
		feature := features[name]
		isLast := i == len(names)-1
		prefix := "├─"
		if isLast {
			prefix = "└─"
		}

		enabledIcon := "❌"
		enabledText := "disabled"
		if feature.Enabled {
			enabledIcon = "✅"
			enabledText = "enabled"
		}

		fmt.Printf("%s %s %s (%s)\n", prefix, enabledIcon, name, enabledText)

		if sc.verbose && feature.Description != "" {
			subPrefix := "│  "
			if isLast {
				subPrefix = "   "
			}
			fmt.Printf("%s   Description: %s\n", subPrefix, feature.Description)
			if len(feature.Flags) > 0 {
				fmt.Printf("%s   Flags: %s\n", subPrefix, strings.Join(feature.Flags, ", "))
			}
		}
	}
}

func (sc *StatusCommand) displayJSON(meta *metadata.UnifiedMetadata) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(meta)
}

func (sc *StatusCommand) displayYAML(meta *metadata.UnifiedMetadata) error {
	// For now, convert to JSON and display as indented JSON
	// In a real implementation, you'd use a YAML library
	return sc.displayJSON(meta)
}

func (sc *StatusCommand) getStatusIcon(status string) string {
	switch status {
	case "installed", "present":
		return "✅"
	case "missing":
		return "❌"
	case "corrupted":
		return "⚠️ "
	case "outdated":
		return "🔄"
	default:
		return "❓"
	}
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func colorize(text, color string, bold bool) string {
	// ANSI color codes
	colors := map[string]string{
		"red":    "31",
		"green":  "32",
		"yellow": "33",
		"blue":   "34",
		"purple": "35",
		"cyan":   "36",
		"white":  "37",
	}

	code, exists := colors[color]
	if !exists {
		return text
	}

	if bold {
		return fmt.Sprintf("\033[1;%sm%s\033[0m", code, text)
	}
	return fmt.Sprintf("\033[%sm%s\033[0m", code, text)
}

// contains function already exists in utils.go, removing duplicate