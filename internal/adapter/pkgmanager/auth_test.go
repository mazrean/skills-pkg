package pkgmanager

import (
	"testing"

	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
)

func TestIsSSHURL(t *testing.T) {
	tests := []struct {
		url  string
		want bool
	}{
		{"git@github.com:user/repo.git", true},
		{"ssh://git@github.com/user/repo.git", true},
		{"https://github.com/user/repo.git", false},
		{"http://github.com/user/repo.git", false},
		{"git://github.com/user/repo.git", false},
	}

	for _, tt := range tests {
		got := isSSHURL(tt.url)
		if got != tt.want {
			t.Errorf("isSSHURL(%q) = %v, want %v", tt.url, got, tt.want)
		}
	}
}

func TestBuildHTTPSAuth(t *testing.T) {
	tests := []struct {
		name     string
		env      map[string]string
		wantNil  bool
		wantUser string
		wantPass string
	}{
		{
			name:    "no credentials",
			env:     map[string]string{},
			wantNil: true,
		},
		{
			name:     "GITHUB_TOKEN set",
			env:      map[string]string{"GITHUB_TOKEN": "ghp_test123"},
			wantNil:  false,
			wantUser: "token",
			wantPass: "ghp_test123",
		},
		{
			name:     "GIT_TOKEN takes priority over GITHUB_TOKEN",
			env:      map[string]string{"GIT_TOKEN": "git_token", "GITHUB_TOKEN": "github_token"},
			wantNil:  false,
			wantUser: "token",
			wantPass: "git_token",
		},
		{
			name:     "GITLAB_TOKEN set",
			env:      map[string]string{"GITLAB_TOKEN": "glpat_test"},
			wantNil:  false,
			wantUser: "token",
			wantPass: "glpat_test",
		},
		{
			name:     "GITEA_TOKEN set",
			env:      map[string]string{"GITEA_TOKEN": "gitea_test"},
			wantNil:  false,
			wantUser: "token",
			wantPass: "gitea_test",
		},
		{
			name:     "GIT_USERNAME and GIT_PASSWORD set",
			env:      map[string]string{"GIT_USERNAME": "user", "GIT_PASSWORD": "pass"},
			wantNil:  false,
			wantUser: "user",
			wantPass: "pass",
		},
		{
			name:    "GIT_USERNAME without GIT_PASSWORD",
			env:     map[string]string{"GIT_USERNAME": "user"},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set env vars
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			auth := buildHTTPSAuth()

			if tt.wantNil {
				if auth != nil {
					t.Errorf("buildHTTPSAuth() = %v, want nil", auth)
				}
				return
			}

			if auth == nil {
				t.Fatal("buildHTTPSAuth() = nil, want non-nil")
			}

			basic, ok := auth.(*githttp.BasicAuth)
			if !ok {
				t.Fatalf("buildHTTPSAuth() type = %T, want *githttp.BasicAuth", auth)
			}
			if basic.Username != tt.wantUser {
				t.Errorf("Username = %q, want %q", basic.Username, tt.wantUser)
			}
			if basic.Password != tt.wantPass {
				t.Errorf("Password = %q, want %q", basic.Password, tt.wantPass)
			}
		})
	}
}

func TestBuildAuthMethod_HTTPS(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "test_token")

	auth, err := buildAuthMethod("https://github.com/user/repo.git")
	if err != nil {
		t.Fatalf("buildAuthMethod() error = %v", err)
	}
	if auth == nil {
		t.Fatal("buildAuthMethod() = nil, want non-nil for HTTPS with token")
	}

	basic, ok := auth.(*githttp.BasicAuth)
	if !ok {
		t.Fatalf("auth type = %T, want *githttp.BasicAuth", auth)
	}
	if basic.Password != "test_token" {
		t.Errorf("Password = %q, want %q", basic.Password, "test_token")
	}
}

func TestBuildAuthMethod_HTTPS_NoCredentials(t *testing.T) {
	auth, err := buildAuthMethod("https://github.com/user/repo.git")
	if err != nil {
		t.Fatalf("buildAuthMethod() error = %v (want nil for anonymous HTTPS)", err)
	}
	if auth != nil {
		t.Errorf("buildAuthMethod() = %v, want nil when no credentials set", auth)
	}
}
