package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestLogger_Info(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		format string
		args   []interface{}
		want   string
	}{
		{
			name:   "simple message",
			format: "Test message: %s",
			args:   []interface{}{"hello"},
			want:   "Test message: hello\n",
		},
		{
			name:   "message without arguments",
			format: "Simple message",
			args:   []interface{}{},
			want:   "Simple message\n",
		},
		{
			name:   "message with multiple arguments",
			format: "User %s has %d items",
			args:   []interface{}{"Alice", 5},
			want:   "User Alice has 5 items\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			logger := NewLogger(false)
			logger.out = &buf

			logger.Info(tt.format, tt.args...)

			got := buf.String()
			if got != tt.want {
				t.Errorf("Info() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLogger_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		format string
		args   []interface{}
		want   string
	}{
		{
			name:   "simple error message",
			format: "Error: %s",
			args:   []interface{}{"test error"},
			want:   "Error: test error\n",
		},
		{
			name:   "error without arguments",
			format: "An error occurred",
			args:   []interface{}{},
			want:   "An error occurred\n",
		},
		{
			name:   "error with multiple arguments",
			format: "Failed to process %s: %v",
			args:   []interface{}{"file.txt", "permission denied"},
			want:   "Failed to process file.txt: permission denied\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			logger := NewLogger(false)
			logger.errOut = &buf

			logger.Error(tt.format, tt.args...)

			got := buf.String()
			if got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLogger_Verbose(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		verboseEnabled bool
		format         string
		args           []interface{}
		wantEmpty      bool
		wantContains   []string
	}{
		{
			name:           "verbose enabled with message",
			verboseEnabled: true,
			format:         "Debug info: %d",
			args:           []interface{}{42},
			wantEmpty:      false,
			wantContains:   []string{"[VERBOSE]", "Debug info: 42"},
		},
		{
			name:           "verbose disabled",
			verboseEnabled: false,
			format:         "Debug info: %d",
			args:           []interface{}{42},
			wantEmpty:      true,
			wantContains:   []string{},
		},
		{
			name:           "verbose enabled without arguments",
			verboseEnabled: true,
			format:         "Starting process",
			args:           []interface{}{},
			wantEmpty:      false,
			wantContains:   []string{"[VERBOSE]", "Starting process"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			logger := NewLogger(tt.verboseEnabled)
			logger.out = &buf

			logger.Verbose(tt.format, tt.args...)

			got := buf.String()
			if tt.wantEmpty {
				if got != "" {
					t.Errorf("Verbose() should not print when disabled, got %q", got)
				}
			} else {
				for _, want := range tt.wantContains {
					if !strings.Contains(got, want) {
						t.Errorf("Verbose() output should contain %q, got %q", want, got)
					}
				}
			}
		})
	}
}

func TestLogger_SetVerbose(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		initialValue bool
		setValue     bool
		wantValue    bool
	}{
		{
			name:         "enable verbose from disabled",
			initialValue: false,
			setValue:     true,
			wantValue:    true,
		},
		{
			name:         "disable verbose from enabled",
			initialValue: true,
			setValue:     false,
			wantValue:    false,
		},
		{
			name:         "keep verbose enabled",
			initialValue: true,
			setValue:     true,
			wantValue:    true,
		},
		{
			name:         "keep verbose disabled",
			initialValue: false,
			setValue:     false,
			wantValue:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger := NewLogger(tt.initialValue)
			logger.SetVerbose(tt.setValue)

			if got := logger.IsVerbose(); got != tt.wantValue {
				t.Errorf("IsVerbose() = %v, want %v", got, tt.wantValue)
			}
		})
	}
}

func TestLogger_IsVerbose(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		verbose bool
	}{
		{
			name:    "verbose enabled",
			verbose: true,
		},
		{
			name:    "verbose disabled",
			verbose: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger := NewLogger(tt.verbose)
			if got := logger.IsVerbose(); got != tt.verbose {
				t.Errorf("IsVerbose() = %v, want %v", got, tt.verbose)
			}
		})
	}
}
