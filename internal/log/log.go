package log

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

var (
	quiet     bool
	debug     bool
	logFile   *os.File
	logMu     sync.Mutex
	startTime = time.Now()
)

const (
	levelInfo    = "INFO"
	levelError   = "ERROR"
	levelDebug   = "DEBUG"
	levelSuccess = "SUCCESS"
)

// Init initializes the logger with an optional log file
func Init(logPath string) error {
	if logPath != "" {
		// Create log directory if it doesn't exist
		logDir := filepath.Dir(logPath)
		if err := os.MkdirAll(logDir, 0700); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}

		// Open log file with restricted permissions
		f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}

		logFile = f
	}
	return nil
}

// Close closes the log file if it's open
func Close() error {
	if logFile != nil {
		return logFile.Close()
	}
	return nil
}

// SetQuiet sets the quiet mode
func SetQuiet(q bool) {
	quiet = q
}

// SetDebug sets the debug mode
func SetDebug(d bool) {
	debug = d
}

// log writes a log message with the given level
func log(level, format string, args ...interface{}) {
	logMu.Lock()
	defer logMu.Unlock()

	// Get caller information
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "unknown"
		line = 0
	}
	file = filepath.Base(file)

	// Format message
	msg := fmt.Sprintf(format, args...)

	// Sanitize message for logging
	msg = sanitizeLogMessage(msg)

	// Create log entry
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	elapsed := time.Since(startTime).Round(time.Millisecond)
	entry := fmt.Sprintf("%s [%s] (%s:%d) [%s] %s\n",
		timestamp, level, file, line, elapsed, msg)

	// Write to log file if configured
	if logFile != nil {
		if _, err := logFile.WriteString(entry); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write to log file: %v\n", err)
		}
	}

	// Write to stdout/stderr unless in quiet mode
	if !quiet || level == levelError {
		if level == levelError {
			fmt.Fprint(os.Stderr, entry)
		} else {
			fmt.Print(entry)
		}
	}
}

// sanitizeLogMessage removes sensitive information from log messages
func sanitizeLogMessage(msg string) string {
	// Remove potential sensitive information
	sensitivePatterns := []string{
		`password=[\S]+`,
		`token=[\S]+`,
		`key=[\S]+`,
		`secret=[\S]+`,
		`auth=[\S]+`,
	}

	sanitized := msg
	for _, pattern := range sensitivePatterns {
		sanitized = strings.ReplaceAll(sanitized, pattern, "[REDACTED]")
	}
	return sanitized
}

// Info logs an info message
func Info(format string, args ...interface{}) {
	log(levelInfo, format, args...)
}

// Error logs an error message
func Error(format string, args ...interface{}) {
	log(levelError, format, args...)
}

// Debug logs a debug message if debug mode is enabled
func Debug(format string, args ...interface{}) {
	if debug {
		log(levelDebug, format, args...)
	}
}

// Success logs a success message
func Success(format string, args ...interface{}) {
	log(levelSuccess, format, args...)
}
