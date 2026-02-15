package domain_test

import (
	"testing"

	"github.com/mazrean/skills-pkg/internal/domain"
)

func TestSkill_Validate(t *testing.T) {
	tests := []struct {
		name    string
		skill   *domain.Skill
		wantErr error
	}{
		{
			name: "valid git source skill",
			skill: &domain.Skill{
				Name:      "test-skill",
				Source:    "git",
				URL:       "https://github.com/example/skill.git",
				Version:   "v1.0.0",
				HashAlgo:  "sha256",
				HashValue: "abc123",
			},
			wantErr: nil,
		},
		{
			name: "valid npm source skill",
			skill: &domain.Skill{
				Name:           "test-skill",
				Source:         "npm",
				URL:            "example-skill",
				Version:        "1.0.0",
				HashAlgo:       "sha256",
				HashValue:      "def456",
				PackageManager: "npm",
			},
			wantErr: nil,
		},
		{
			name: "valid go-module source skill",
			skill: &domain.Skill{
				Name:           "test-skill",
				Source:         "go-module",
				URL:            "github.com/example/skill",
				Version:        "v1.0.0",
				HashAlgo:       "sha256",
				HashValue:      "ghi789",
				PackageManager: "go-module",
			},
			wantErr: nil,
		},
		{
			name: "invalid source type",
			skill: &domain.Skill{
				Name:      "test-skill",
				Source:    "invalid",
				URL:       "https://example.com",
				Version:   "1.0.0",
				HashAlgo:  "sha256",
				HashValue: "abc123",
			},
			wantErr: domain.ErrInvalidSource,
		},
		{
			name: "empty name",
			skill: &domain.Skill{
				Name:      "",
				Source:    "git",
				URL:       "https://github.com/example/skill.git",
				Version:   "v1.0.0",
				HashAlgo:  "sha256",
				HashValue: "abc123",
			},
			wantErr: domain.ErrInvalidSkill,
		},
		{
			name: "empty URL",
			skill: &domain.Skill{
				Name:      "test-skill",
				Source:    "git",
				URL:       "",
				Version:   "v1.0.0",
				HashAlgo:  "sha256",
				HashValue: "abc123",
			},
			wantErr: domain.ErrInvalidSkill,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.skill.Validate()
			if err != tt.wantErr {
				t.Errorf("Skill.Validate() error = %v, wantErr %v", err, tt.wantErr)
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
		name    string
		config  *domain.Config
		wantErr error
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
						HashAlgo:  "sha256",
						HashValue: "abc123",
					},
				},
				InstallTargets: []string{"/path/to/dir"},
			},
			wantErr: nil,
		},
		{
			name: "duplicate skill names",
			config: &domain.Config{
				Skills: []*domain.Skill{
					{Name: "skill1", Source: "git", URL: "url1", Version: "v1.0.0", HashAlgo: "sha256", HashValue: "abc"},
					{Name: "skill1", Source: "npm", URL: "url2", Version: "1.0.0", HashAlgo: "sha256", HashValue: "def"},
				},
				InstallTargets: []string{"/path/to/dir"},
			},
			wantErr: domain.ErrDuplicateSkill,
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
						HashAlgo:  "sha256",
						HashValue: "abc",
					},
				},
				InstallTargets: []string{"/path/to/dir"},
			},
			wantErr: domain.ErrInvalidSource,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if err != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
