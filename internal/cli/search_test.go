package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSearchCmd_runWithLoggerAndBaseURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		apiResponse any
		checkFunc   func(t *testing.T, output string)
		name        string
		query       string
		limit       int
		apiStatus   int
		wantErr     bool
	}{
		{
			name:  "success: results found",
			query: "typescript",
			limit: 10,
			apiResponse: searchResponse{Skills: []searchSkill{
				{Name: "typescript-helper", SkillID: "typescript-helper", Source: "github.com/example/typescript-helper", Installs: 42},
				{Name: "ts-tools", SkillID: "ts-tools", Source: "github.com/example/ts-tools", Installs: 7},
			}},
			apiStatus: http.StatusOK,
			wantErr:   false,
			checkFunc: func(t *testing.T, output string) {
				t.Helper()

				if !strings.Contains(output, "typescript-helper") {
					t.Errorf("output should contain 'typescript-helper', got: %s", output)
				}
				if !strings.Contains(output, "ts-tools") {
					t.Errorf("output should contain 'ts-tools', got: %s", output)
				}
				if !strings.Contains(output, "42") {
					t.Errorf("output should contain install count '42', got: %s", output)
				}
				if !strings.Contains(output, "2 result") {
					t.Errorf("output should show 2 results, got: %s", output)
				}
			},
		},
		{
			name:        "success: empty results shows message",
			query:       "nonexistent-skill-xyz",
			limit:       10,
			apiResponse: searchResponse{Skills: []searchSkill{}},
			apiStatus:   http.StatusOK,
			wantErr:     false,
			checkFunc: func(t *testing.T, output string) {
				t.Helper()

				output = strings.ToLower(output)
				if !strings.Contains(output, "no skill") {
					t.Errorf("output should indicate no skills found, got: %s", output)
				}
			},
		},
		{
			name:        "error: API returns non-200",
			query:       "test",
			limit:       10,
			apiResponse: nil,
			apiStatus:   http.StatusInternalServerError,
			wantErr:     true,
			checkFunc: func(t *testing.T, output string) {
				t.Helper()

				output = strings.ToLower(output)
				if !strings.Contains(output, "failed") {
					t.Errorf("output should indicate failure, got: %s", output)
				}
			},
		},
		{
			name:  "success: empty query returns results",
			query: "",
			limit: 10,
			apiResponse: searchResponse{Skills: []searchSkill{
				{Name: "go-tools", SkillID: "go-tools", Source: "github.com/example/go-tools", Installs: 100},
			}},
			apiStatus: http.StatusOK,
			wantErr:   false,
			checkFunc: func(t *testing.T, output string) {
				t.Helper()

				if !strings.Contains(output, "go-tools") {
					t.Errorf("output should contain 'go-tools', got: %s", output)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/search" {
					http.NotFound(w, r)
					return
				}

				w.WriteHeader(tt.apiStatus)
				if tt.apiResponse != nil {
					if err := json.NewEncoder(w).Encode(tt.apiResponse); err != nil {
						t.Errorf("failed to encode response: %v", err)
					}
				}
			}))
			defer server.Close()

			cmd := &SearchCmd{
				Query: tt.query,
				Limit: tt.limit,
			}

			var outBuf, errBuf bytes.Buffer
			logger := &Logger{
				out:    &outBuf,
				errOut: &errBuf,
			}

			err := cmd.runWithLoggerAndBaseURL(context.Background(), logger, server.URL)

			if (err != nil) != tt.wantErr {
				t.Errorf("runWithLoggerAndBaseURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.checkFunc != nil {
				output := outBuf.String() + errBuf.String()
				tt.checkFunc(t, output)
			}
		})
	}
}
