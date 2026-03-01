package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestParseSkillMDDescription(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    string
	}{
		{
			name:  "frontmatter: description extracted",
			input: "---\nname: my-skill\ndescription: A helpful skill.\n---\n# Content\n",
			want:  "A helpful skill.",
		},
		{
			name:  "frontmatter: closing ... supported",
			input: "---\ndescription: Dot-terminated.\n...\n# Body\n",
			want:  "Dot-terminated.",
		},
		{
			name:  "frontmatter: description not present",
			input: "---\nname: my-skill\n---\n",
			want:  "",
		},
		{
			name:  "frontmatter: body description ignored",
			input: "---\nname: my-skill\n---\ndescription: should be ignored\n",
			want:  "",
		},
		{
			name:  "no frontmatter: bare yaml description extracted",
			input: "name: my-skill\ndescription: Bare YAML description.\n",
			want:  "Bare YAML description.",
		},
		{
			name:  "no frontmatter: description on first line",
			input: "description: First line description.\n",
			want:  "First line description.",
		},
		{
			name:  "empty file",
			input: "",
			want:  "",
		},
		{
			name:  "frontmatter: trailing CR stripped",
			input: "---\r\ndescription: Windows line endings.\r\n---\r\n",
			want:  "Windows line endings.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := parseSkillMDDescription(strings.NewReader(tt.input))
			if got != tt.want {
				t.Errorf("parseSkillMDDescription() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSearchCmd_runWithLoggerAndBaseURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		apiResponse    any
		skillMdContent map[string]string // path -> content
		checkFunc      func(t *testing.T, output string)
		name           string
		query          string
		limit          int
		apiStatus      int
		wantErr        bool
	}{
		{
			name:  "success: results found",
			query: "typescript",
			limit: 10,
			apiResponse: searchResponse{Skills: []searchSkill{
				{Name: "typescript-helper", SkillID: "typescript-helper", Source: "example/typescript-helper", Installs: 42},
				{Name: "ts-tools", SkillID: "ts-tools", Source: "example/ts-tools", Installs: 7},
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
			name:  "success: description shown from skills subdirectory",
			query: "golang",
			limit: 10,
			apiResponse: searchResponse{Skills: []searchSkill{
				{Name: "golang-pro", SkillID: "golang-pro", Source: "example/claude-skills", Installs: 100},
			}},
			skillMdContent: map[string]string{
				"/example/claude-skills/main/skills/golang-pro/SKILL.md": "name: golang-pro\ndescription: Use when building Go applications.\n",
			},
			apiStatus: http.StatusOK,
			wantErr:   false,
			checkFunc: func(t *testing.T, output string) {
				t.Helper()

				if !strings.Contains(output, "Use when building Go applications.") {
					t.Errorf("output should contain description, got: %s", output)
				}
			},
		},
		{
			name:  "success: description shown from repo root fallback",
			query: "golang",
			limit: 10,
			apiResponse: searchResponse{Skills: []searchSkill{
				{Name: "golang", SkillID: "golang", Source: "example/golang-skill", Installs: 50},
			}},
			skillMdContent: map[string]string{
				"/example/golang-skill/main/SKILL.md": "---\nname: golang\ndescription: Best practices for writing production Go code.\n---\n",
			},
			apiStatus: http.StatusOK,
			wantErr:   false,
			checkFunc: func(t *testing.T, output string) {
				t.Helper()

				if !strings.Contains(output, "Best practices for writing production Go code.") {
					t.Errorf("output should contain description from root SKILL.md, got: %s", output)
				}
			},
		},
		{
			name:  "success: no description when SKILL.md missing",
			query: "golang",
			limit: 10,
			apiResponse: searchResponse{Skills: []searchSkill{
				{Name: "golang-tools", SkillID: "golang-tools", Source: "example/golang-tools", Installs: 30},
			}},
			skillMdContent: map[string]string{},
			apiStatus:      http.StatusOK,
			wantErr:        false,
			checkFunc: func(t *testing.T, output string) {
				t.Helper()

				if !strings.Contains(output, "golang-tools") {
					t.Errorf("output should contain skill name, got: %s", output)
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
				{Name: "go-tools", SkillID: "go-tools", Source: "example/go-tools", Installs: 100},
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

			skillMdContent := tt.skillMdContent

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/search" {
					w.WriteHeader(tt.apiStatus)
					if tt.apiResponse != nil {
						if err := json.NewEncoder(w).Encode(tt.apiResponse); err != nil {
							t.Errorf("failed to encode response: %v", err)
						}
					}
					return
				}

				if content, ok := skillMdContent[r.URL.Path]; ok {
					_, _ = fmt.Fprint(w, content)
					return
				}

				http.NotFound(w, r)
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

			err := cmd.runWithLoggerAndBaseURLs(context.Background(), logger, server.URL, server.URL)

			if (err != nil) != tt.wantErr {
				t.Errorf("runWithLoggerAndBaseURLs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.checkFunc != nil {
				output := outBuf.String() + errBuf.String()
				tt.checkFunc(t, output)
			}
		})
	}
}
