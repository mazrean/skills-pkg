// Package domain provides core domain models and business logic for skills-pkg.
// It defines the configuration structures, validation rules, and domain-level errors.
package domain

// Config represents the entire .skillspkg.toml configuration.
// It manages the list of skills and their installation targets.
// Requirements: 2.1, 2.2, 10.1
type Config struct {
	Skills         []*Skill `toml:"skills"`
	InstallTargets []string `toml:"install_targets"`
}

// Skill represents a single skill entry in the configuration.
// It contains all metadata required for skill installation and verification.
// Requirements: 2.2, 2.3, 2.4, 5.2, 11.4
type Skill struct {
	Name      string `toml:"name"`
	Source    string `toml:"source"`                 // "git", "go-mod"
	URL       string `toml:"url"`                    // Git URL, Go module path
	Version   string `toml:"version,omitempty"`      // Tag, commit hash, or semantic version
	HashValue string `toml:"hash_value,omitempty"`   // Hash value with algorithm prefix (e.g., "h1:<base64>")
	SubDir    string `toml:"subdir,omitempty"`       // Subdirectory within the downloaded source (e.g., "skills/my-agent")
}

// Validate validates the skill configuration.
// It checks that all required fields are present and that the source type is valid.
// Requirements: 2.2, 11.4, 12.2, 12.3
func (s *Skill) Validate() error {
	// Check required fields
	if s.Name == "" {
		return ErrInvalidSkill
	}
	if s.URL == "" {
		return ErrInvalidSkill
	}

	// Validate source type (requirement 11.4)
	validSources := map[string]bool{
		"git":    true,
		"go-mod": true,
	}
	if !validSources[s.Source] {
		return ErrInvalidSource
	}

	return nil
}

// FindSkillByName finds a skill by its name.
// Returns nil if the skill is not found.
// Requirements: 8.1, 9.3
func (c *Config) FindSkillByName(name string) *Skill {
	for _, skill := range c.Skills {
		if skill.Name == name {
			return skill
		}
	}
	return nil
}

// FindSkillsBySource は指定ソースタイプに一致するスキルのスライスを返す。
// 一致するスキルが存在しない場合は空スライスを返す。
func (c *Config) FindSkillsBySource(sourceType string) []*Skill {
	result := make([]*Skill, 0)
	for _, skill := range c.Skills {
		if skill.Source == sourceType {
			result = append(result, skill)
		}
	}
	return result
}

// HasSkill checks if a skill with the given name exists.
// Requirements: 2.3
func (c *Config) HasSkill(name string) bool {
	return c.FindSkillByName(name) != nil
}

// Validate validates the entire configuration.
// It checks for duplicate skill names and validates each skill.
// Requirements: 2.1, 2.2, 12.2, 12.3
func (c *Config) Validate() error {
	// Check for duplicate skill names (requirement 2.2)
	nameMap := make(map[string]bool)
	for _, skill := range c.Skills {
		if nameMap[skill.Name] {
			return ErrSkillExists
		}
		nameMap[skill.Name] = true

		// Validate each skill
		if err := skill.Validate(); err != nil {
			return err
		}
	}

	return nil
}
