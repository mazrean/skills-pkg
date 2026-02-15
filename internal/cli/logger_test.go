package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestLogger_Info(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(false)
	logger.out = &buf

	logger.Info("Test message: %s", "hello")

	got := buf.String()
	want := "Test message: hello\n"
	if got != want {
		t.Errorf("Info() = %q, want %q", got, want)
	}
}

func TestLogger_Error(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(false)
	logger.errOut = &buf

	logger.Error("Error: %s", "test error")

	got := buf.String()
	want := "Error: test error\n"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestLogger_Verbose_Enabled(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(true)
	logger.out = &buf

	logger.Verbose("Debug info: %d", 42)

	got := buf.String()
	if !strings.Contains(got, "[VERBOSE]") {
		t.Errorf("Verbose() should include [VERBOSE] prefix, got %q", got)
	}
	if !strings.Contains(got, "Debug info: 42") {
		t.Errorf("Verbose() should include message, got %q", got)
	}
}

func TestLogger_Verbose_Disabled(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(false)
	logger.out = &buf

	logger.Verbose("Debug info: %d", 42)

	got := buf.String()
	if got != "" {
		t.Errorf("Verbose() should not print when disabled, got %q", got)
	}
}

func TestLogger_SetVerbose(t *testing.T) {
	logger := NewLogger(false)

	if logger.IsVerbose() {
		t.Error("Logger should start with verbose disabled")
	}

	logger.SetVerbose(true)
	if !logger.IsVerbose() {
		t.Error("Logger verbose should be enabled after SetVerbose(true)")
	}

	logger.SetVerbose(false)
	if logger.IsVerbose() {
		t.Error("Logger verbose should be disabled after SetVerbose(false)")
	}
}

func TestLogger_IsVerbose(t *testing.T) {
	tests := []struct {
		name    string
		verbose bool
	}{
		{"verbose enabled", true},
		{"verbose disabled", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger(tt.verbose)
			if got := logger.IsVerbose(); got != tt.verbose {
				t.Errorf("IsVerbose() = %v, want %v", got, tt.verbose)
			}
		})
	}
}
