package log

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

var (
	globalLogger zerolog.Logger
)

func init() {
	// Set up console writer with color support
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
		NoColor:    false,
	}

	// Initialize logger with console writer
	globalLogger = zerolog.New(output).With().Timestamp().Logger()

	// Set default level to info
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

// SetLevel sets the logging level
func SetLevel(level string) {
	switch level {
	case "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case "panic":
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

// SetQuiet sets the logger to only show errors
func SetQuiet(quiet bool) {
	if quiet {
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

// SetDebug sets the logger to debug level
func SetDebug(debug bool) {
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

// SetOutput sets the logger output
func SetOutput(w io.Writer) {
	globalLogger = zerolog.New(w).With().Timestamp().Logger()
}

// Trace logs a trace message
func Trace(msg string, args ...interface{}) {
	globalLogger.Trace().Msgf(msg, args...)
}

// Debug logs a debug message
func Debug(msg string, args ...interface{}) {
	globalLogger.Debug().Msgf(msg, args...)
}

// Info logs an info message
func Info(msg string, args ...interface{}) {
	globalLogger.Info().Msgf(msg, args...)
}

// Warn logs a warning message
func Warn(msg string, args ...interface{}) {
	globalLogger.Warn().Msgf(msg, args...)
}

// Error logs an error message
func Error(msg string, args ...interface{}) {
	globalLogger.Error().Msgf(msg, args...)
}

// Fatal logs a fatal message and exits
func Fatal(msg string, args ...interface{}) {
	globalLogger.Fatal().Msgf(msg, args...)
}

// Panic logs a panic message and panics
func Panic(msg string, args ...interface{}) {
	globalLogger.Panic().Msgf(msg, args...)
}

// WithField adds a field to the logger and returns a new event
func WithField(key string, value interface{}) {
	globalLogger.Info().Interface(key, value).Send()
}

// WithFields adds multiple fields to the logger and returns a new event
func WithFields(fields map[string]interface{}) {
	event := globalLogger.Info()
	for k, v := range fields {
		event.Interface(k, v)
	}
	event.Send()
}

// WithError adds an error field to the logger and returns a new event
func WithError(err error) {
	globalLogger.Error().Err(err).Send()
}
