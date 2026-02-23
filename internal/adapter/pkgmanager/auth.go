package pkgmanager

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/transport"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	gitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

// isSSHURL reports whether repoURL uses the SSH protocol.
func isSSHURL(repoURL string) bool {
	return strings.HasPrefix(repoURL, "git@") || strings.HasPrefix(repoURL, "ssh://")
}

// buildAuthMethod returns an auth method appropriate for the given URL.
//
// For SSH URLs (git@... or ssh://...) it tries SSH agent then key files in ~/.ssh/.
// An error is returned if no SSH credentials are available.
//
// For HTTPS/HTTP URLs it reads credentials from environment variables and
// returns nil when none are set (allowing anonymous access for public repos).
func buildAuthMethod(repoURL string) (transport.AuthMethod, error) {
	if isSSHURL(repoURL) {
		return buildSSHAuth()
	}
	return buildHTTPSAuth(), nil
}

// buildSSHAuth creates an SSH auth method, trying SSH agent first then key files.
func buildSSHAuth() (transport.AuthMethod, error) {
	auth, err := gitssh.NewSSHAgentAuth(gitssh.DefaultUsername)
	if err == nil {
		return auth, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("SSH authentication unavailable: SSH agent not running and cannot determine home directory: %w", err)
	}

	for _, keyFile := range []string{
		filepath.Join(home, ".ssh", "id_ed25519"),
		filepath.Join(home, ".ssh", "id_rsa"),
		filepath.Join(home, ".ssh", "id_ecdsa"),
		filepath.Join(home, ".ssh", "id_dsa"),
	} {
		if _, statErr := os.Stat(keyFile); statErr != nil {
			continue
		}
		auth, keyErr := gitssh.NewPublicKeysFromFile(gitssh.DefaultUsername, keyFile, "")
		if keyErr == nil {
			return auth, nil
		}
	}

	return nil, fmt.Errorf("SSH authentication unavailable: SSH agent not running and no usable key files found in %s/.ssh/", home)
}

// buildHTTPSAuth returns an HTTP BasicAuth built from environment variables,
// or nil when no credentials are configured.
// Checked variables (in order): GIT_TOKEN, GITHUB_TOKEN, GITLAB_TOKEN, GITEA_TOKEN,
// then GIT_USERNAME + GIT_PASSWORD.
func buildHTTPSAuth() transport.AuthMethod {
	for _, envVar := range []string{"GIT_TOKEN", "GITHUB_TOKEN", "GITLAB_TOKEN", "GITEA_TOKEN"} {
		if token := os.Getenv(envVar); token != "" {
			return &githttp.BasicAuth{
				Username: "token",
				Password: token,
			}
		}
	}

	if username := os.Getenv("GIT_USERNAME"); username != "" {
		if password := os.Getenv("GIT_PASSWORD"); password != "" {
			return &githttp.BasicAuth{
				Username: username,
				Password: password,
			}
		}
	}

	return nil
}
