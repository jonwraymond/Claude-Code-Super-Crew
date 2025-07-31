package ui

// Color constants for terminal output
const (
	ColorReset   = "\033[0m"
	ColorBright  = "\033[1m"
	ColorDim     = "\033[2m"
	
	ColorRed     = "\033[31m"
	ColorGreen   = "\033[32m"
	ColorYellow  = "\033[33m"
	ColorBlue    = "\033[34m"
	ColorMagenta = "\033[35m"
	ColorCyan    = "\033[36m"
	ColorWhite   = "\033[37m"
	ColorGray    = "\033[90m"
)

// Colors provides a struct-based interface to color constants
var Colors = struct {
	Reset   string
	Bright  string
	Red     string
	Green   string
	Yellow  string
	Blue    string
	Magenta string
	Cyan    string
	White   string
	Gray    string
}{
	Reset:   ColorReset,
	Bright:  ColorBright,
	Red:     ColorRed,
	Green:   ColorGreen,
	Yellow:  ColorYellow,
	Blue:    ColorBlue,
	Magenta: ColorMagenta,
	Cyan:    ColorCyan,
	White:   ColorWhite,
	Gray:    ColorGray,
}