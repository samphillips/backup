package logging

import (
	"fmt"
	golog "log"
	"os"
	"strings"
)

const (
	// DEBUG is the const string for debug
	DEBUG = "debug"
	// INFO is the const string for info
	INFO = "info"
	// WARN is the const string for warn
	WARN = "warn"
	// ERROR is the const string for error
	ERROR = "error"
	// FATAL is the const string for fatal
	FATAL = "fatal"
)

// Logger holds the pointers used for logging at different levels
type Logger struct {
	Debug *golog.Logger
	Info  *golog.Logger
	Warn  *golog.Logger
	Error *golog.Logger
	Fatal *golog.Logger
	Level int
}

var (
	colourOff    = []byte("\033[0m")
	colourRed    = []byte("\033[0;31m")
	colourGreen  = []byte("\033[0;32m")
	colourOrange = []byte("\033[0;33m")
	colourCyan   = []byte("\033[0;36m")
	logLevels    = map[string]int{
		DEBUG: 0,
		INFO:  1,
		WARN:  2,
		ERROR: 3,
		FATAL: 4,
	}
	log Logger
)

func init() {
	flags := golog.LUTC
	log.Debug = golog.New(os.Stdout, ColourCyan("[DEBUG] "), flags)
	log.Info = golog.New(os.Stdout, ColourGreen("[INFO] "), flags)
	log.Warn = golog.New(os.Stdout, ColourOrange("[WARN] "), flags)
	log.Error = golog.New(os.Stdout, ColourRed("[ERROR] "), flags)
	log.Fatal = golog.New(os.Stdout, ColourRed("[FATAL] "), flags)

	SetLogLevel(INFO)
}

// SetLogLevel sets the logger to log up to a certain level
func SetLogLevel(logLevel string) {
	levelLower := strings.ToLower(logLevel)

	if level, ok := logLevels[levelLower]; ok {
		log.Level = level
		_ = log.Info.Output(2, fmt.Sprintf("Log level set to %s", logLevel))
	} else {
		_ = log.Warn.Output(2, fmt.Sprintf("%s: invalid log level, log level remains at %v", levelLower, log.Level))
	}
}

// Debug logs a debug message
func Debug(logString string, args ...interface{}) {
	if log.Level <= logLevels[DEBUG] {
		_ = log.Debug.Output(2, fmt.Sprintf(logString, args...))
	}
}

// Info logs an info message
func Info(logString string, args ...interface{}) {
	if log.Level <= logLevels[INFO] {
		_ = log.Info.Output(2, fmt.Sprintf(logString, args...))
	}
}

// Warn logs a warning message
func Warn(logString string, args ...interface{}) {
	if log.Level <= logLevels[WARN] {
		_ = log.Warn.Output(2, fmt.Sprintf(logString, args...))
	}
}

// Error logs an error message
func Error(logString string, args ...interface{}) {
	if log.Level <= logLevels[ERROR] {
		_ = log.Error.Output(2, fmt.Sprintf(logString, args...))
	}
}

// Fatal logs a fatal message
func Fatal(logString string, args ...interface{}) {
	if log.Level <= logLevels[FATAL] {
		_ = log.Fatal.Output(2, fmt.Sprintf(logString, args...))
	}
}

// ColourRed colours the given string red
func ColourRed(text string) string {
	return fmt.Sprintf("%s%s%s", colourRed, text, colourOff)
}

// ColourGreen colours the given string green
func ColourGreen(text string) string {
	return fmt.Sprintf("%s%s%s", colourGreen, text, colourOff)
}

// ColourOrange colours the given string orange
func ColourOrange(text string) string {
	return fmt.Sprintf("%s%s%s", colourOrange, text, colourOff)
}

// ColourCyan colours the given string cyan
func ColourCyan(text string) string {
	return fmt.Sprintf("%s%s%s", colourCyan, text, colourOff)
}
