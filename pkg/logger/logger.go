// Package logger provides a unified structured logging interface using zerolog.
// It supports multiple log levels, colored console output, file logging with rotation,
// statistics tracking, and special formatting for success messages.
package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

// LogLevel represents the logging level
type LogLevel int

const (
	// DebugLevel shows all messages including debug
	DebugLevel LogLevel = iota
	// InfoLevel shows info, warnings, and errors
	InfoLevel
	// WarnLevel shows warnings and errors
	WarnLevel
	// ErrorLevel shows only errors
	ErrorLevel
	// CriticalLevel shows only critical errors
	CriticalLevel
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARNING"
	case ErrorLevel:
		return "ERROR"
	case CriticalLevel:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// ToZerologLevel converts LogLevel to zerolog.Level
func (l LogLevel) ToZerologLevel() zerolog.Level {
	switch l {
	case DebugLevel:
		return zerolog.DebugLevel
	case InfoLevel:
		return zerolog.InfoLevel
	case WarnLevel:
		return zerolog.WarnLevel
	case ErrorLevel:
		return zerolog.ErrorLevel
	case CriticalLevel:
		return zerolog.ErrorLevel // zerolog doesn't have critical, use error
	default:
		return zerolog.InfoLevel
	}
}

// Logger interface defines the logging contract for the application
type Logger interface {
	Debug(msg string)
	Debugf(format string, args ...interface{})
	Info(msg string)
	Infof(format string, args ...interface{})
	Warn(msg string)
	Warnf(format string, args ...interface{})
	Error(msg string)
	Errorf(format string, args ...interface{})
	Critical(msg string)
	Criticalf(format string, args ...interface{})
	Success(msg string)
	Successf(format string, args ...interface{})
	Exception(msg string, err error)
	Step(step, total int, message string)
	Section(title string)
	LogSystemInfo(info map[string]interface{})
	LogOperationStart(operation string, details map[string]interface{})
	LogOperationEnd(operation string, success bool, duration time.Duration, details map[string]interface{})
	SetLevel(level LogLevel)
	SetVerbose(verbose bool)
	SetQuiet(quiet bool)
	SetConsoleLevel(level LogLevel)
	SetFileLevel(level LogLevel)
	InitializeFileLogging(logDir string) error
	GetStatistics() map[string]interface{}
	Flush()
	Close()
}

// UnifiedLogger wraps zerolog to implement our Logger interface.
// It provides thread-safe logging with support for console and file output,
// file rotation, statistics tracking, and enhanced formatting.
type UnifiedLogger struct {
	name         string
	logDir       string
	consoleLevel LogLevel
	fileLevel    LogLevel
	sessionStart time.Time
	logFile      *os.File
	logger       zerolog.Logger
	logCounts    map[string]int
	statistics   map[string]interface{}
	verbose      bool
	quiet        bool
	console      zerolog.ConsoleWriter
	mu           sync.Mutex
}

var (
	globalLogger Logger
	once         sync.Once
)

// GetLogger returns the global logger instance
func GetLogger() Logger {
	once.Do(func() {
		globalLogger = NewLogger()
	})
	return globalLogger
}

// NewLogger creates a new unified logger
func NewLogger() Logger {
	return NewNamedLogger("supercrew")
}

// NewNamedLogger creates a new logger with a specific name
func NewNamedLogger(name string) Logger {
	logger := &UnifiedLogger{
		name:         name,
		consoleLevel: InfoLevel,
		fileLevel:    DebugLevel,
		sessionStart: time.Now(),
		logCounts: map[string]int{
			"debug":    0,
			"info":     0,
			"warning":  0,
			"error":    0,
			"critical": 0,
		},
		statistics: make(map[string]interface{}),
		verbose:    false,
		quiet:      false,
	}

	// Setup zerolog with stack trace support
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.TimeFieldFormat = time.RFC3339

	// Setup console writer with custom formatting
	logger.console = zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "15:04:05",
		NoColor:    false,
	}

	// Custom level formatting
	logger.console.FormatLevel = func(i interface{}) string {
		if i == nil {
			return ""
		}
		level := strings.ToUpper(fmt.Sprintf("%s", i))
		switch level {
		case "DEBUG":
			return "\033[90mDEBUG  \033[0m" // Gray
		case "INFO":
			return "\033[36mINFO   \033[0m" // Cyan
		case "WARN":
			return "\033[33mWARN   \033[0m" // Yellow
		case "ERROR":
			return "\033[31mERROR  \033[0m" // Red
		case "CRITICAL":
			return "\033[35mCRITICAL\033[0m" // Magenta
		default:
			return ""
		}
	}

	// Custom message formatting
	logger.console.FormatMessage = func(i interface{}) string {
		if i == nil {
			return ""
		}
		return fmt.Sprintf("%s", i)
	}

	// Create initial logger
	logger.logger = zerolog.New(logger.console).With().
		Timestamp().
		Str("logger", name).
		Logger().
		Level(logger.consoleLevel.ToZerologLevel())

	return logger
}

// InitializeFileLogging initializes file logging with rotation
func (l *UnifiedLogger) InitializeFileLogging(logDir string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if logDir == "" {
		homeDir, _ := os.UserHomeDir()
		// Use the new .crew/logs directory structure
		logDir = filepath.Join(homeDir, ".claude", ".crew", "logs")
	}
	l.logDir = logDir

	// Ensure log directory exists
	if err := os.MkdirAll(l.logDir, 0755); err != nil {
		return err
	}

	// Create timestamped log file
	timestamp := l.sessionStart.Format("20060102_150405")
	logFilePath := filepath.Join(l.logDir, fmt.Sprintf("%s_%s.log", l.name, timestamp))

	var err error
	l.logFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	// Clean up old log files (keep last 10)
	l.cleanupOldLogs(10)

	// Create multi-writer for both console and file
	multi := io.MultiWriter(l.console, l.logFile)
	l.logger = l.logger.Output(multi)

	return nil
}

// cleanupOldLogs removes old log files keeping only the specified count
func (l *UnifiedLogger) cleanupOldLogs(keepCount int) {
	pattern := filepath.Join(l.logDir, l.name+"_*.log")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return
	}

	if len(matches) <= keepCount {
		return
	}

	// Sort by modification time
	type fileInfo struct {
		path    string
		modTime time.Time
	}

	var files []fileInfo
	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil {
			continue
		}
		files = append(files, fileInfo{
			path:    match,
			modTime: info.ModTime(),
		})
	}

	// Sort by modification time, newest first
	sort.Slice(files, func(i, j int) bool {
		return files[i].modTime.After(files[j].modTime)
	})

	// Remove old files
	for i := keepCount; i < len(files); i++ {
		os.Remove(files[i].path)
	}
}

// SetLevel sets the logging level
func (l *UnifiedLogger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.consoleLevel = level
	l.updateZerologLevel()
}

// SetConsoleLevel changes console logging level
func (l *UnifiedLogger) SetConsoleLevel(level LogLevel) {
	l.SetLevel(level)
}

// SetFileLevel changes file logging level
func (l *UnifiedLogger) SetFileLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.fileLevel = level
}

// SetVerbose enables verbose logging
func (l *UnifiedLogger) SetVerbose(verbose bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.verbose = verbose
	if verbose {
		l.consoleLevel = DebugLevel
		// Show timestamp and level in verbose mode
		l.console.TimeFormat = "15:04:05"
		l.console.FormatTimestamp = func(i interface{}) string {
			t := i.(string)
			return fmt.Sprintf("[%s]", t)
		}
	} else {
		// Hide timestamp and level in non-verbose mode
		l.console.TimeFormat = ""
		l.console.FormatTimestamp = func(i interface{}) string {
			return ""
		}
		l.console.FormatLevel = func(i interface{}) string {
			return ""
		}
	}
	l.updateZerologLevel()
}

// SetQuiet enables quiet mode
func (l *UnifiedLogger) SetQuiet(quiet bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.quiet = quiet
	if quiet {
		// In quiet mode, disable console output
		l.logger = l.logger.Output(io.Discard)
		if l.logFile != nil {
			l.logger = l.logger.Output(l.logFile)
		}
	} else {
		// Restore console output
		if l.logFile != nil {
			multi := io.MultiWriter(l.console, l.logFile)
			l.logger = l.logger.Output(multi)
		} else {
			l.logger = l.logger.Output(l.console)
		}
	}
}

func (l *UnifiedLogger) updateZerologLevel() {
	switch l.consoleLevel {
	case DebugLevel:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case InfoLevel:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case WarnLevel:
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case ErrorLevel:
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case CriticalLevel:
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	}
}

// Debug logs a debug message
func (l *UnifiedLogger) Debug(msg string) {
	l.logCounts["debug"]++
	l.logger.Debug().Msg(msg)
}

// Debugf logs a formatted debug message
func (l *UnifiedLogger) Debugf(format string, args ...interface{}) {
	l.logCounts["debug"]++
	l.logger.Debug().Msgf(format, args...)
}

// Info logs an info message
func (l *UnifiedLogger) Info(msg string) {
	l.logCounts["info"]++
	if !l.verbose && !l.quiet {
		fmt.Println(msg)
	} else {
		l.logger.Info().Msg(msg)
	}
}

// Infof logs a formatted info message
func (l *UnifiedLogger) Infof(format string, args ...interface{}) {
	l.logCounts["info"]++
	if !l.verbose && !l.quiet {
		fmt.Printf(format+"\n", args...)
	} else {
		l.logger.Info().Msgf(format, args...)
	}
}

// Warn logs a warning message
func (l *UnifiedLogger) Warn(msg string) {
	l.logCounts["warning"]++
	if !l.verbose && !l.quiet {
		fmt.Printf("\033[33m⚠️  %s\033[0m\n", msg)
	} else {
		l.logger.Warn().Msg(msg)
	}
}

// Warnf logs a formatted warning message
func (l *UnifiedLogger) Warnf(format string, args ...interface{}) {
	l.logCounts["warning"]++
	if !l.verbose && !l.quiet {
		fmt.Printf("\033[33m⚠️  %s\033[0m\n", fmt.Sprintf(format, args...))
	} else {
		l.logger.Warn().Msgf(format, args...)
	}
}

// Error logs an error message
func (l *UnifiedLogger) Error(msg string) {
	l.logCounts["error"]++
	if !l.verbose {
		fmt.Printf("\033[31m❌ %s\033[0m\n", msg)
	} else {
		l.logger.Error().Msg(msg)
	}
}

// Errorf logs a formatted error message
func (l *UnifiedLogger) Errorf(format string, args ...interface{}) {
	l.logCounts["error"]++
	if !l.verbose {
		fmt.Printf("\033[31m❌ %s\033[0m\n", fmt.Sprintf(format, args...))
	} else {
		l.logger.Error().Msgf(format, args...)
	}
}

// Critical logs a critical message
func (l *UnifiedLogger) Critical(msg string) {
	l.logCounts["critical"]++
	l.logger.Error().Str("level", "CRITICAL").Msg(msg)
}

// Criticalf logs a formatted critical message
func (l *UnifiedLogger) Criticalf(format string, args ...interface{}) {
	l.logCounts["critical"]++
	l.logger.Error().Str("level", "CRITICAL").Msgf(format, args...)
}

// Success logs a success message
func (l *UnifiedLogger) Success(msg string) {
	l.logCounts["info"]++
	if !l.verbose && !l.quiet {
		fmt.Printf("\033[32m✅ %s\033[0m\n", msg)
	} else if !l.quiet {
		// Use info level with custom prefix for file logging
		l.logger.Info().Str("type", "SUCCESS").Msg(msg)
	}
}

// Successf logs a formatted success message
func (l *UnifiedLogger) Successf(format string, args ...interface{}) {
	l.logCounts["info"]++
	if !l.verbose && !l.quiet {
		fmt.Printf("\033[32m✅ %s\033[0m\n", fmt.Sprintf(format, args...))
	} else if !l.quiet {
		l.logger.Info().Str("type", "SUCCESS").Msgf(format, args...)
	}
}

// Exception logs an exception with error details
func (l *UnifiedLogger) Exception(msg string, err error) {
	l.logCounts["error"]++
	l.logger.Error().Stack().Err(err).Msg(msg)
	if l.verbose && err != nil {
		l.logger.Debug().Str("error_detail", fmt.Sprintf("%+v", err)).Msg("Stack trace")
	}
}

// Step logs step progress
func (l *UnifiedLogger) Step(step, total int, message string) {
	stepMsg := fmt.Sprintf("[%d/%d] %s", step, total, message)
	l.Info(stepMsg)
}

// Section logs section header
func (l *UnifiedLogger) Section(title string) {
	separator := strings.Repeat("=", min(50, len(title)+4))
	l.Info(separator)
	l.Info(fmt.Sprintf("  %s", title))
	l.Info(separator)
}

// LogSystemInfo logs system information
func (l *UnifiedLogger) LogSystemInfo(info map[string]interface{}) {
	l.Section("System Information")
	for key, value := range info {
		l.Info(fmt.Sprintf("%s: %v", key, value))
	}
}

// LogOperationStart logs start of operation
func (l *UnifiedLogger) LogOperationStart(operation string, details map[string]interface{}) {
	l.Section(fmt.Sprintf("Starting: %s", operation))
	if details != nil {
		for key, value := range details {
			l.Info(fmt.Sprintf("%s: %v", key, value))
		}
	}
}

// LogOperationEnd logs end of operation
func (l *UnifiedLogger) LogOperationEnd(operation string, success bool, duration time.Duration, details map[string]interface{}) {
	status := "SUCCESS"
	if !success {
		status = "FAILED"
	}
	l.Info(fmt.Sprintf("Operation %s completed: %s (Duration: %.2fs)", operation, status, duration.Seconds()))

	if details != nil {
		for key, value := range details {
			l.Info(fmt.Sprintf("%s: %v", key, value))
		}
	}
}

// GetStatistics returns logging statistics
func (l *UnifiedLogger) GetStatistics() map[string]interface{} {
	runtime := time.Since(l.sessionStart)

	return map[string]interface{}{
		"session_start":   l.sessionStart.Format(time.RFC3339),
		"runtime_seconds": runtime.Seconds(),
		"log_counts":      l.logCounts,
		"total_messages":  l.getTotalMessages(),
		"log_file":        l.getLogFilePath(),
		"has_errors":      l.logCounts["error"]+l.logCounts["critical"] > 0,
	}
}

// Flush flushes all log outputs
func (l *UnifiedLogger) Flush() {
	if l.logFile != nil {
		l.logFile.Sync()
	}
}

// Close closes logger and handlers
func (l *UnifiedLogger) Close() {
	l.Section("Session Complete")
	stats := l.GetStatistics()

	l.Info(fmt.Sprintf("Total runtime: %.1f seconds", stats["runtime_seconds"].(float64)))
	l.Info(fmt.Sprintf("Messages logged: %d", stats["total_messages"].(int)))

	if stats["has_errors"].(bool) {
		l.Warn(fmt.Sprintf("Errors/warnings: %d", l.logCounts["error"]+l.logCounts["warning"]))
	}

	if logFile := stats["log_file"]; logFile != nil {
		l.Info(fmt.Sprintf("Full log saved to: %s", logFile.(string)))
	}

	if l.logFile != nil {
		l.logFile.Close()
	}
}

// Helper functions

func (l *UnifiedLogger) getTotalMessages() int {
	total := 0
	for _, count := range l.logCounts {
		total += count
	}
	return total
}

func (l *UnifiedLogger) getLogFilePath() interface{} {
	if l.logFile != nil {
		return l.logFile.Name()
	}
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ParseLogLevel parses a string into a LogLevel
func ParseLogLevel(level string) LogLevel {
	switch level {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn", "warning":
		return WarnLevel
	case "error":
		return ErrorLevel
	case "critical":
		return CriticalLevel
	default:
		return InfoLevel
	}
}

// Convenience functions for backward compatibility
func Debug(msg string) {
	GetLogger().Debug(msg)
}

func Debugf(format string, args ...interface{}) {
	GetLogger().Debugf(format, args...)
}

func Info(msg string) {
	GetLogger().Info(msg)
}

func Infof(format string, args ...interface{}) {
	GetLogger().Infof(format, args...)
}

func Warn(msg string) {
	GetLogger().Warn(msg)
}

func Warnf(format string, args ...interface{}) {
	GetLogger().Warnf(format, args...)
}

func Error(msg string) {
	GetLogger().Error(msg)
}

func Errorf(format string, args ...interface{}) {
	GetLogger().Errorf(format, args...)
}

func Critical(msg string) {
	GetLogger().Critical(msg)
}

func Criticalf(format string, args ...interface{}) {
	GetLogger().Criticalf(format, args...)
}

func Success(msg string) {
	GetLogger().Success(msg)
}

func Successf(format string, args ...interface{}) {
	GetLogger().Successf(format, args...)
}
