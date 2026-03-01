package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"

	"github.com/alecthomas/kong"
)

const (
	searchAPIBase = "https://skills.sh"
	searchLimit   = 10
)

// SearchCmd searches for available skills on skills.sh.
type SearchCmd struct {
	Query string `arg:"" optional:"" help:"Search query for skills"`
	Limit int    `default:"10" help:"Maximum number of results to show"`
}

// searchSkill represents a skill returned by the skills.sh search API.
type searchSkill struct {
	Name     string `json:"name"`
	SkillID  string `json:"skillId"`
	Source   string `json:"source"`
	Installs int    `json:"installs"`
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
	return c.runWithLoggerAndBaseURL(ctx, logger, searchAPIBase)
}

func (c *SearchCmd) runWithLoggerAndBaseURL(ctx context.Context, logger *Logger, apiBase string) error {
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

	logger.Info("")
	logger.Info("%-30s %-40s %-10s", "NAME", "SOURCE", "INSTALLS")
	logger.Info("%s", "--------------------------------------------------------------------------------")

	for _, s := range skills {
		logger.Info("%-30s %-40s %-10d", s.Name, s.Source, s.Installs)
	}

	logger.Info("")
	logger.Info("Total: %d result(s)", len(skills))

	return nil
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
