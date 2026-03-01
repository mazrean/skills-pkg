package domain_test

import (
	"errors"
	"testing"

	"github.com/mazrean/skills-pkg/internal/domain"
)

func TestSkill_Validate(t *testing.T) {
	tests := []struct {
		wantErrCheck func(error) bool
		skill        *domain.Skill
		name         string
	}{
		{
			name: "valid git source skill",
			skill: &domain.Skill{
				Name:      "test-skill",
				Source:    "git",
				URL:       "https://github.com/example/skill.git",
				Version:   "v1.0.0",

				HashValue: "abc123",
			},
			wantErrCheck: nil,
		},
		{
			name: "valid go-mod source skill",
			skill: &domain.Skill{
				Name:           "test-skill",
				Source:         "go-mod",
				URL:            "github.com/example/skill",
				Version:        "v1.0.0",

				HashValue:      "def456",
			},
			wantErrCheck: nil,
		},
		{
			name: "valid go-mod source skill",
			skill: &domain.Skill{
				Name:           "test-skill",
				Source:         "go-mod",
				URL:            "github.com/example/skill",
				Version:        "v1.0.0",

				HashValue:      "ghi789",
			},
			wantErrCheck: nil,
		},
		{
			name: "invalid source type",
			skill: &domain.Skill{
				Name:      "test-skill",
				Source:    "invalid",
				URL:       "https://example.com",
				Version:   "1.0.0",

				HashValue: "abc123",
			},
			wantErrCheck: func(err error) bool {
				_, ok := errors.AsType[*domain.ErrorInvalidSource](err)
				return ok
			},
		},
		{
			name: "empty name",
			skill: &domain.Skill{
				Name:      "",
				Source:    "git",
				URL:       "https://github.com/example/skill.git",
				Version:   "v1.0.0",

				HashValue: "abc123",
			},
			wantErrCheck: func(err error) bool {
				_, ok := errors.AsType[*domain.ErrorInvalidSkill](err)
				return ok
			},
		},
		{
			name: "empty URL",
			skill: &domain.Skill{
				Name:      "test-skill",
				Source:    "git",
				URL:       "",
				Version:   "v1.0.0",

				HashValue: "abc123",
			},
			wantErrCheck: func(err error) bool {
				_, ok := errors.AsType[*domain.ErrorInvalidSkill](err)
				return ok
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.skill.Validate()
			if tt.wantErrCheck != nil {
				if err == nil {
					t.Error("Skill.Validate() expected error, got nil")
				} else if !tt.wantErrCheck(err) {
					t.Errorf("Skill.Validate() error = %v, did not match expected error type", err)
				}
			} else if err != nil {
				t.Errorf("Skill.Validate() unexpected error = %v", err)
			}
		})
	}
}

func TestConfig_FindSkillByName(t *testing.T) {
	config := &domain.Config{
		Skills: []*domain.Skill{
			{Name: "skill1"},
			{Name: "skill2"},
			{Name: "skill3"},
		},
		InstallTargets: []string{"/path/to/dir"},
	}

	tests := []struct {
		name      string
		skillName string
		wantFound bool
	}{
		{
			name:      "existing skill",
			skillName: "skill2",
			wantFound: true,
		},
		{
			name:      "non-existing skill",
			skillName: "skill4",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			skill := config.FindSkillByName(tt.skillName)
			found := skill != nil
			if found != tt.wantFound {
				t.Errorf("Config.FindSkillByName() found = %v, wantFound %v", found, tt.wantFound)
			}
			if found && skill.Name != tt.skillName {
				t.Errorf("Config.FindSkillByName() returned skill with name = %v, want %v", skill.Name, tt.skillName)
			}
		})
	}
}

func TestConfig_HasSkill(t *testing.T) {
	config := &domain.Config{
		Skills: []*domain.Skill{
			{Name: "skill1"},
			{Name: "skill2"},
		},
		InstallTargets: []string{"/path/to/dir"},
	}

	tests := []struct {
		name      string
		skillName string
		want      bool
	}{
		{
			name:      "existing skill",
			skillName: "skill1",
			want:      true,
		},
		{
			name:      "non-existing skill",
			skillName: "skill3",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := config.HasSkill(tt.skillName); got != tt.want {
				t.Errorf("Config.HasSkill() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		wantErrCheck func(error) bool
		config       *domain.Config
		name         string
	}{
		{
			name: "valid config",
			config: &domain.Config{
				Skills: []*domain.Skill{
					{
						Name:      "skill1",
						Source:    "git",
						URL:       "https://github.com/example/skill.git",
						Version:   "v1.0.0",

						HashValue: "abc123",
					},
				},
				InstallTargets: []string{"/path/to/dir"},
			},
			wantErrCheck: nil,
		},
		{
			name: "duplicate skill names",
			config: &domain.Config{
				Skills: []*domain.Skill{
					{Name: "skill1", Source: "git", URL: "url1", Version: "v1.0.0", HashValue: "abc"},
					{Name: "skill1", Source: "go-mod", URL: "url2", Version: "v1.0.0", HashValue: "def"},
				},
				InstallTargets: []string{"/path/to/dir"},
			},
			wantErrCheck: func(err error) bool {
				_, ok := errors.AsType[*domain.ErrorSkillExists](err)
				return ok
			},
		},
		{
			name: "invalid skill in config",
			config: &domain.Config{
				Skills: []*domain.Skill{
					{
						Name:      "skill1",
						Source:    "invalid-source",
						URL:       "url",
						Version:   "v1.0.0",

						HashValue: "abc",
					},
				},
				InstallTargets: []string{"/path/to/dir"},
			},
			wantErrCheck: func(err error) bool {
				_, ok := errors.AsType[*domain.ErrorInvalidSource](err)
				return ok
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErrCheck != nil {
				if err == nil {
					t.Error("Config.Validate() expected error, got nil")
				} else if !tt.wantErrCheck(err) {
					t.Errorf("Config.Validate() error = %v, did not match expected error type", err)
				}
			} else if err != nil {
				t.Errorf("Config.Validate() unexpected error = %v", err)
			}
		})
	}
}
