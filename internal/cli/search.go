package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"sync"

	"github.com/alecthomas/kong"
)

const (
	searchAPIBase = "https://skills.sh"
	searchLimit   = 10
	rawGitHubBase = "https://raw.githubusercontent.com"
)

// SearchCmd searches for available skills on skills.sh.
type SearchCmd struct {
	Query string `arg:"" optional:"" help:"Search query for skills"`
	Limit int    `default:"10" help:"Maximum number of results to show"`
}

// searchSkill represents a skill returned by the skills.sh search API.
type searchSkill struct {
	Name        string `json:"name"`
	SkillID     string `json:"skillId"`
	Source      string `json:"source"`
	Description string `json:"-"`
	Installs    int    `json:"installs"`
}

// searchResponse is the top-level envelope returned by the skills.sh search API.
type searchResponse struct {
	Skills []searchSkill `json:"skills"`
}

func (c *SearchCmd) Run(ctx *kong.Context) error {
	verbose := false
	if model := ctx.Model; model != nil && model.Target.IsValid() {
		if verboseField := model.Target.FieldByName("Verbose"); verboseField.IsValid() && verboseField.Kind() == reflect.Bool {
			verbose = verboseField.Bool()
		}
	}

	return c.runWithLogger(context.Background(), NewLogger(verbose))
}

func (c *SearchCmd) runWithLogger(ctx context.Context, logger *Logger) error {
	return c.runWithLoggerAndBaseURLs(ctx, logger, searchAPIBase, rawGitHubBase)
}

func (c *SearchCmd) runWithLoggerAndBaseURLs(ctx context.Context, logger *Logger, apiBase, rawBase string) error {
	limit := c.Limit
	if limit <= 0 {
		limit = searchLimit
	}

	logger.Verbose("Searching skills on skills.sh (query=%q, limit=%d)", c.Query, limit)

	skills, err := c.fetchSkills(ctx, c.Query, limit, apiBase)
	if err != nil {
		logger.Error("Failed to search skills: %v", err)
		return err
	}

	if len(skills) == 0 {
		logger.Info("No skills found")
		logger.Info("Try a different query or browse https://skills.sh")
		return nil
	}

	// Fetch descriptions from SKILL.md concurrently
	var wg sync.WaitGroup
	for i := range skills {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			skills[i].Description = c.fetchDescription(ctx, rawBase, skills[i].Source, skills[i].SkillID)
		}(i)
	}
	wg.Wait()

	logger.Info("")
	logger.Info("%-30s %-40s %-10s", "NAME", "SOURCE", "INSTALLS")
	logger.Info("%s", "--------------------------------------------------------------------------------")

	for _, s := range skills {
		logger.Info("%-30s %-40s %-10d", s.Name, s.Source, s.Installs)
		if s.Description != "" {
			logger.Info("  %s", s.Description)
		}
	}

	logger.Info("")
	logger.Info("Total: %d result(s)", len(skills))

	return nil
}

// fetchDescription retrieves the description field from a skill's SKILL.md.
// It first tries skills/{skillID}/SKILL.md (multi-skill repos), then falls back
// to SKILL.md at the repository root (single-skill repos).
func (c *SearchCmd) fetchDescription(ctx context.Context, rawBase, source, skillID string) string {
	primaryURL := fmt.Sprintf("%s/%s/main/skills/%s/SKILL.md", rawBase, source, skillID)
	if desc := c.tryFetchDescription(ctx, primaryURL); desc != "" {
		return desc
	}

	fallbackURL := fmt.Sprintf("%s/%s/main/SKILL.md", rawBase, source)
	return c.tryFetchDescription(ctx, fallbackURL)
}

// tryFetchDescription fetches a SKILL.md from rawURL and extracts the description field value.
func (c *SearchCmd) tryFetchDescription(ctx context.Context, rawURL string) string {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return ""
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ""
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	return parseSkillMDDescription(resp.Body)
}

// parseSkillMDDescription parses an MDX/Markdown document and extracts the description
// field from the YAML frontmatter block (delimited by ---). If no frontmatter is
// present, it falls back to scanning all lines for a bare "description:" key.
func parseSkillMDDescription(r io.Reader) string {
	scanner := bufio.NewScanner(r)
	if !scanner.Scan() {
		return ""
	}
	firstLine := strings.TrimRight(scanner.Text(), "\r")

	if firstLine == "---" {
		// Frontmatter block: scan until closing --- or ...
		for scanner.Scan() {
			line := strings.TrimRight(scanner.Text(), "\r")
			if line == "---" || line == "..." {
				break
			}
			if after, ok := strings.CutPrefix(line, "description:"); ok {
				return strings.TrimSpace(after)
			}
		}
		return ""
	}

	// No frontmatter delimiter — treat the whole file as bare YAML metadata.
	if after, ok := strings.CutPrefix(firstLine, "description:"); ok {
		return strings.TrimSpace(after)
	}
	for scanner.Scan() {
		line := scanner.Text()
		if after, ok := strings.CutPrefix(line, "description:"); ok {
			return strings.TrimSpace(after)
		}
	}
	return ""
}

func (c *SearchCmd) fetchSkills(ctx context.Context, query string, limit int, apiBase string) ([]searchSkill, error) {
	apiURL, err := url.Parse(apiBase + "/api/search")
	if err != nil {
		return nil, fmt.Errorf("parse API URL: %w", err)
	}

	params := url.Values{}
	params.Set("q", query)
	params.Set("limit", fmt.Sprintf("%d", limit))
	apiURL.RawQuery = params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search API request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search API returned status %d", resp.StatusCode)
	}

	var result searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode search response: %w", err)
	}

	return result.Skills, nil
}
