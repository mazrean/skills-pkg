package domain_test

import (
	"testing"

	"github.com/mazrean/skills-pkg/internal/domain"
)

func TestSkill_Validate(t *testing.T) {
	tests := []struct {
		wantErr error
		skill   *domain.Skill
		name    string
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
			wantErr: nil,
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
			wantErr: nil,
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
			wantErr: nil,
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
			wantErr: domain.ErrInvalidSource,
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
			wantErr: domain.ErrInvalidSkill,
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

func TestConfig_FindSkillsBySource(t *testing.T) {
	skills := []*domain.Skill{
		{Name: "git-skill1", Source: "git", URL: "https://github.com/example/skill1.git"},
		{Name: "git-skill2", Source: "git", URL: "https://github.com/example/skill2.git"},
		{Name: "gomod-skill1", Source: "go-mod", URL: "github.com/example/skill3"},
	}
	config := &domain.Config{
		Skills:         skills,
		InstallTargets: []string{"/path/to/dir"},
	}

	tests := []struct {
		name       string
		sourceType string
		wantNames  []string
	}{
		{
			name:       "filter git skills",
			sourceType: "git",
			wantNames:  []string{"git-skill1", "git-skill2"},
		},
		{
			name:       "filter go-mod skills",
			sourceType: "go-mod",
			wantNames:  []string{"gomod-skill1"},
		},
		{
			name:       "no matching skills returns empty slice",
			sourceType: "nonexistent",
			wantNames:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := config.FindSkillsBySource(tt.sourceType)
			if len(got) != len(tt.wantNames) {
				t.Errorf("Config.FindSkillsBySource() returned %d skills, want %d", len(got), len(tt.wantNames))
				return
			}
			for i, skill := range got {
				if skill.Name != tt.wantNames[i] {
					t.Errorf("Config.FindSkillsBySource()[%d].Name = %v, want %v", i, skill.Name, tt.wantNames[i])
				}
			}
		})
	}
}

func TestConfig_FindSkillsBySource_EmptyConfig(t *testing.T) {
	config := &domain.Config{
		Skills:         []*domain.Skill{},
		InstallTargets: []string{"/path/to/dir"},
	}
	got := config.FindSkillsBySource("git")
	if got == nil {
		t.Error("Config.FindSkillsBySource() returned nil, want empty slice")
	}
	if len(got) != 0 {
		t.Errorf("Config.FindSkillsBySource() returned %d skills, want 0", len(got))
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		wantErr error
		config  *domain.Config
		name    string
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
			wantErr: nil,
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
			wantErr: domain.ErrSkillExists,
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
