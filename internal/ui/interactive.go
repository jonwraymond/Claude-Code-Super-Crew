package ui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Menu represents an interactive menu system with keyboard navigation
type Menu struct {
	Title       string
	Options     []string
	MultiSelect bool
	selected    map[int]bool
}

// NewMenu creates a new interactive menu
func NewMenu(title string, options []string, multiSelect bool) *Menu {
	return &Menu{
		Title:       title,
		Options:     options,
		MultiSelect: multiSelect,
		selected:    make(map[int]bool),
	}
}

// Display shows the menu and returns user selection
func (m *Menu) Display() (interface{}, error) {
	scanner := bufio.NewScanner(os.Stdin)
	
	for {
		// Clear screen and display menu
		m.displayMenu()
		
		fmt.Print("> ")
		if !scanner.Scan() {
			return nil, fmt.Errorf("failed to read input")
		}
		
		input := strings.TrimSpace(scanner.Text())
		
		if m.MultiSelect {
			result, err := m.handleMultiSelectInput(input)
			if err != nil {
				fmt.Printf("%s%s%s\n", Colors.Red, err.Error(), Colors.Reset)
				waitForKey("Press Enter to continue...")
				continue
			}
			return result, nil
		} else {
			result, err := m.handleSingleSelectInput(input)
			if err != nil {
				fmt.Printf("%s%s%s\n", Colors.Red, err.Error(), Colors.Reset)
				waitForKey("Press Enter to continue...")
				continue
			}
			return result, nil
		}
	}
}

// displayMenu shows the menu options
func (m *Menu) displayMenu() {
	clearScreen()
	
	// Display header
	fmt.Printf("\n%s%s%s%s\n", Colors.Cyan, Colors.Bright, m.Title, Colors.Reset)
	fmt.Println(strings.Repeat("=", len(m.Title)))
	fmt.Println()
	
	// Display options
	for i, option := range m.Options {
		number := i + 1
		if m.MultiSelect {
			marker := "[ ]"
			if m.selected[i] {
				marker = "[x]"
			}
			fmt.Printf("%s%2d.%s %s%s %s\n", Colors.Yellow, number, Colors.Reset, Colors.Green, marker, option)
		} else {
			fmt.Printf("%s%2d.%s %s\n", Colors.Yellow, number, Colors.Reset, option)
		}
	}
	
	fmt.Println()
	
	// Display instructions
	if m.MultiSelect {
		fmt.Printf("%sEnter numbers separated by commas (e.g., 1,3,5) or 'all' for all options:%s\n", Colors.Blue, Colors.Reset)
		fmt.Printf("%sPress Enter without input to continue with current selection%s\n", Colors.Blue, Colors.Reset)
	} else {
		fmt.Printf("%sEnter your choice (1-%d):%s\n", Colors.Blue, len(m.Options), Colors.Reset)
	}
	fmt.Println()
}

// handleSingleSelectInput processes single selection input
func (m *Menu) handleSingleSelectInput(input string) (int, error) {
	if input == "" {
		return -1, fmt.Errorf("please enter a valid number")
	}
	
	choice, err := strconv.Atoi(input)
	if err != nil {
		return -1, fmt.Errorf("please enter a valid number")
	}
	
	if choice < 1 || choice > len(m.Options) {
		return -1, fmt.Errorf("invalid choice. Please enter a number between 1 and %d", len(m.Options))
	}
	
	return choice - 1, nil
}

// handleMultiSelectInput processes multi-selection input
func (m *Menu) handleMultiSelectInput(input string) ([]int, error) {
	if input == "" {
		// Return current selection
		var selected []int
		for i := range m.Options {
			if m.selected[i] {
				selected = append(selected, i)
			}
		}
		return selected, nil
	}
	
	if strings.ToLower(input) == "all" {
		var all []int
		for i := range m.Options {
			all = append(all, i)
		}
		return all, nil
	}
	
	// Parse comma-separated numbers
	var selections []int
	parts := strings.Split(input, ",")
	
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		
		choice, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid input: %s", part)
		}
		
		if choice < 1 || choice > len(m.Options) {
			return nil, fmt.Errorf("invalid option: %d", choice)
		}
		
		index := choice - 1
		selections = append(selections, index)
		m.selected[index] = true
	}
	
	return selections, nil
}

// Confirm displays a confirmation dialog
func Confirm(message string, defaultResponse bool) bool {
	scanner := bufio.NewScanner(os.Stdin)
	
	suffix := "[Y/n]"
	if !defaultResponse {
		suffix = "[y/N]"
	}
	
	for {
		fmt.Printf("%s%s %s%s ", Colors.Blue, message, suffix, Colors.Reset)
		
		if !scanner.Scan() {
			return false
		}
		
		response := strings.ToLower(strings.TrimSpace(scanner.Text()))
		
		if response == "" {
			return defaultResponse
		}
		
		switch response {
		case "y", "yes", "true", "1":
			return true
		case "n", "no", "false", "0":
			return false
		default:
			fmt.Printf("%sPlease enter 'y' or 'n' (or press Enter for default).%s\n", Colors.Red, Colors.Reset)
		}
	}
}

// PromptString prompts for string input with validation
func PromptString(message string, defaultValue string, validator func(string) error) (string, error) {
	scanner := bufio.NewScanner(os.Stdin)
	
	for {
		prompt := message
		if defaultValue != "" {
			prompt = fmt.Sprintf("%s [%s]", message, defaultValue)
		}
		
		fmt.Printf("%s%s:%s ", Colors.Blue, prompt, Colors.Reset)
		
		if !scanner.Scan() {
			return "", fmt.Errorf("failed to read input")
		}
		
		input := strings.TrimSpace(scanner.Text())
		if input == "" && defaultValue != "" {
			input = defaultValue
		}
		
		if validator != nil {
			if err := validator(input); err != nil {
				fmt.Printf("%s%s%s\n", Colors.Red, err.Error(), Colors.Reset)
				continue
			}
		}
		
		return input, nil
	}
}

// PromptInt prompts for integer input with validation
func PromptInt(message string, defaultValue int, min, max int) (int, error) {
	validator := func(input string) error {
		value, err := strconv.Atoi(input)
		if err != nil {
			return fmt.Errorf("please enter a valid number")
		}
		
		if value < min || value > max {
			return fmt.Errorf("please enter a number between %d and %d", min, max)
		}
		
		return nil
	}
	
	defaultStr := ""
	if defaultValue >= min && defaultValue <= max {
		defaultStr = strconv.Itoa(defaultValue)
	}
	
	result, err := PromptString(message, defaultStr, validator)
	if err != nil {
		return 0, err
	}
	
	return strconv.Atoi(result)
}

// PromptChoice prompts for a choice from a list of options
func PromptChoice(message string, options []string, defaultIndex int) (int, error) {
	menu := NewMenu(message, options, false)
	
	result, err := menu.Display()
	if err != nil {
		return -1, err
	}
	
	if choice, ok := result.(int); ok {
		return choice, nil
	}
	
	return defaultIndex, nil
}

// PromptMultiChoice prompts for multiple choices from a list of options
func PromptMultiChoice(message string, options []string) ([]int, error) {
	menu := NewMenu(message, options, true)
	
	result, err := menu.Display()
	if err != nil {
		return nil, err
	}
	
	if choices, ok := result.([]int); ok {
		return choices, nil
	}
	
	return []int{}, nil
}

// DisplayHeader shows a formatted header
func DisplayHeader(title string, subtitle string) {
	fmt.Printf("\n%s%s%s%s\n", Colors.Cyan, Colors.Bright, strings.Repeat("=", 60), Colors.Reset)
	fmt.Printf("%s%s%s%s\n", Colors.Cyan, Colors.Bright, centerString(title, 60), Colors.Reset)
	if subtitle != "" {
		fmt.Printf("%s%s%s\n", Colors.White, centerString(subtitle, 60), Colors.Reset)
	}
	fmt.Printf("%s%s%s%s\n\n", Colors.Cyan, Colors.Bright, strings.Repeat("=", 60), Colors.Reset)
}

// DisplayInfo shows an info message
func DisplayInfo(message string) {
	fmt.Printf("%s[INFO] %s%s\n", Colors.Blue, message, Colors.Reset)
}

// DisplaySuccess shows a success message
func DisplaySuccess(message string) {
	fmt.Printf("%s[✓] %s%s\n", Colors.Green, message, Colors.Reset)
}

// DisplayWarning shows a warning message
func DisplayWarning(message string) {
	fmt.Printf("%s[!] %s%s\n", Colors.Yellow, message, Colors.Reset)
}

// DisplayError shows an error message
func DisplayError(message string) {
	fmt.Printf("%s[✗] %s%s\n", Colors.Red, message, Colors.Reset)
}

// DisplayStep shows step progress
func DisplayStep(step, total int, message string) {
	fmt.Printf("%s[%d/%d] %s%s\n", Colors.Cyan, step, total, message, Colors.Reset)
}

// DisplayTable shows data in table format
func DisplayTable(headers []string, rows [][]string, title string) {
	if len(rows) == 0 {
		return
	}
	
	// Calculate column widths
	colWidths := make([]int, len(headers))
	for i, header := range headers {
		colWidths[i] = len(header)
	}
	
	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}
	
	// Display title
	if title != "" {
		fmt.Printf("\n%s%s%s%s\n\n", Colors.Cyan, Colors.Bright, title, Colors.Reset)
	}
	
	// Display headers
	headerLine := ""
	for i, header := range headers {
		if i > 0 {
			headerLine += " | "
		}
		headerLine += fmt.Sprintf("%-*s", colWidths[i], header)
	}
	fmt.Printf("%s%s%s\n", Colors.Yellow, headerLine, Colors.Reset)
	fmt.Println(strings.Repeat("-", len(headerLine)))
	
	// Display rows
	for _, row := range rows {
		rowLine := ""
		for i, cell := range row {
			if i > 0 {
				rowLine += " | "
			}
			if i < len(colWidths) {
				rowLine += fmt.Sprintf("%-*s", colWidths[i], cell)
			} else {
				rowLine += cell
			}
		}
		fmt.Println(rowLine)
	}
	
	fmt.Println()
}

// waitForKey waits for user to press Enter
func waitForKey(message string) {
	fmt.Printf("%s%s%s", Colors.Blue, message, Colors.Reset)
	bufio.NewReader(os.Stdin).ReadString('\n')
}

// clearScreen clears the terminal screen
func clearScreen() {
	fmt.Print("\033[2J\033[H")
}

// centerString centers text within a given width
func centerString(text string, width int) string {
	if len(text) >= width {
		return text
	}
	
	padding := width - len(text)
	leftPad := padding / 2
	rightPad := padding - leftPad
	
	return strings.Repeat(" ", leftPad) + text + strings.Repeat(" ", rightPad)
}

// FormatSize formats file size in human-readable format
func FormatSize(sizeBytes int64) string {
	const unit = 1024
	if sizeBytes < unit {
		return fmt.Sprintf("%d B", sizeBytes)
	}
	div, exp := int64(unit), 0
	for n := sizeBytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(sizeBytes)/float64(div), "KMGTPE"[exp])
}

// ProgressBar represents a progress bar for long operations
type ProgressBar struct {
	Total   int
	Current int
	Width   int
	Prefix  string
	Suffix  string
}

// NewProgressBar creates a new progress bar
func NewProgressBar(total int, width int, prefix, suffix string) *ProgressBar {
	return &ProgressBar{
		Total:  total,
		Width:  width,
		Prefix: prefix,
		Suffix: suffix,
	}
}

// Update updates and displays the progress bar
func (pb *ProgressBar) Update(current int, message string) {
	pb.Current = current
	
	percent := float64(current) / float64(pb.Total) * 100
	if pb.Total == 0 {
		percent = 100
	}
	
	// Calculate filled and empty portions
	filledWidth := int(float64(pb.Width) * float64(current) / float64(pb.Total))
	if pb.Total == 0 {
		filledWidth = pb.Width
	}
	
	filled := strings.Repeat("█", filledWidth)
	empty := strings.Repeat("░", pb.Width-filledWidth)
	
	// Format progress line
	status := ""
	if message != "" {
		status = fmt.Sprintf(" %s", message)
	}
	
	progressLine := fmt.Sprintf("\r%s[%s%s%s%s%s] %5.1f%%%s%s",
		pb.Prefix,
		Colors.Green, filled,
		Colors.White, empty,
		Colors.Reset,
		percent,
		status,
		pb.Suffix)
	
	fmt.Print(progressLine)
}

// Increment increments progress by 1
func (pb *ProgressBar) Increment(message string) {
	pb.Update(pb.Current+1, message)
}

// Finish completes the progress bar
func (pb *ProgressBar) Finish(message string) {
	pb.Update(pb.Total, message)
	fmt.Println() // New line after completion
}

// StatusSpinner represents a simple spinner for long operations
type StatusSpinner struct {
	Message  string
	spinning bool
	chars    []string
	current  int
}

// NewStatusSpinner creates a new status spinner
func NewStatusSpinner(message string) *StatusSpinner {
	return &StatusSpinner{
		Message: message,
		chars:   []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
	}
}

// Start starts the spinner (would need goroutine in real implementation)
func (ss *StatusSpinner) Start() {
	ss.spinning = true
	fmt.Printf("%s%s %s%s", Colors.Blue, ss.chars[0], ss.Message, Colors.Reset)
}

// Stop stops the spinner
func (ss *StatusSpinner) Stop(finalMessage string) {
	ss.spinning = false
	
	// Clear spinner line
	fmt.Printf("\r%s\r", strings.Repeat(" ", len(ss.Message)+5))
	
	if finalMessage != "" {
		fmt.Println(finalMessage)
	}
}

// InstallationProfile represents an installation profile for interactive setup
type InstallationProfile struct {
	Name        string
	Description string
	Components  []string
	Config      map[string]interface{}
}

// ProfileSelector provides interactive profile selection
type ProfileSelector struct {
	Profiles []InstallationProfile
}

// NewProfileSelector creates a new profile selector
func NewProfileSelector(profiles []InstallationProfile) *ProfileSelector {
	return &ProfileSelector{
		Profiles: profiles,
	}
}

// SelectProfile displays profile options and returns selected profile
func (ps *ProfileSelector) SelectProfile() (*InstallationProfile, error) {
	options := make([]string, len(ps.Profiles))
	for i, profile := range ps.Profiles {
		options[i] = fmt.Sprintf("%s - %s", profile.Name, profile.Description)
	}
	
	menu := NewMenu("Select Installation Profile", options, false)
	result, err := menu.Display()
	if err != nil {
		return nil, err
	}
	
	if index, ok := result.(int); ok && index >= 0 && index < len(ps.Profiles) {
		return &ps.Profiles[index], nil
	}
	
	return nil, fmt.Errorf("invalid profile selection")
}

// ComponentSelector provides interactive component selection
type ComponentSelector struct {
	Available []string
	Selected  []string
}

// NewComponentSelector creates a new component selector
func NewComponentSelector(available []string, preSelected []string) *ComponentSelector {
	return &ComponentSelector{
		Available: available,
		Selected:  preSelected,
	}
}

// SelectComponents displays component options and returns selected components
func (cs *ComponentSelector) SelectComponents() ([]string, error) {
	menu := NewMenu("Select Components to Install", cs.Available, true)
	
	// Pre-select components
	for i, component := range cs.Available {
		for _, selected := range cs.Selected {
			if component == selected {
				menu.selected[i] = true
				break
			}
		}
	}
	
	result, err := menu.Display()
	if err != nil {
		return nil, err
	}
	
	if indices, ok := result.([]int); ok {
		var selected []string
		for _, index := range indices {
			if index >= 0 && index < len(cs.Available) {
				selected = append(selected, cs.Available[index])
			}
		}
		return selected, nil
	}
	
	return cs.Selected, nil
}