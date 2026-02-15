package cli

import (
	"fmt"
	"io"
	"os"
)

// Logger provides logging functionality with verbose support
type Logger struct {
	out     io.Writer
	errOut  io.Writer
	verbose bool
}

// NewLogger creates a new Logger instance
func NewLogger(verbose bool) *Logger {
	return &Logger{
		out:     os.Stdout,
		errOut:  os.Stderr,
		verbose: verbose,
	}
}

// Info prints an informational message to stdout
func (l *Logger) Info(format string, args ...interface{}) {
	fmt.Fprintf(l.out, format+"\n", args...)
}

// Error prints an error message to stderr
func (l *Logger) Error(format string, args ...interface{}) {
	fmt.Fprintf(l.errOut, format+"\n", args...)
}

// Verbose prints a verbose debug message to stdout if verbose mode is enabled
func (l *Logger) Verbose(format string, args ...interface{}) {
	if l.verbose {
		fmt.Fprintf(l.out, "[VERBOSE] "+format+"\n", args...)
	}
}

// SetVerbose enables or disables verbose logging
func (l *Logger) SetVerbose(verbose bool) {
	l.verbose = verbose
}

// IsVerbose returns whether verbose mode is enabled
func (l *Logger) IsVerbose() bool {
	return l.verbose
}
