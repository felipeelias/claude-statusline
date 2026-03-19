package usage

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// CredentialSource retrieves an OAuth access token for the Anthropic API.
type CredentialSource interface {
	AccessToken() (string, error)
}

// keychainCredentials is the JSON structure stored in macOS Keychain.
type keychainCredentials struct {
	ClaudeAiOAuth struct {
		AccessToken string `json:"accessToken"`
	} `json:"claudeAiOauth"`
}

// keychainSource retrieves credentials from macOS Keychain.
type keychainSource struct{}

func (keychainSource) AccessToken() (string, error) {
	cmd := exec.Command("security", "find-generic-password", "-s", "Claude Code-credentials", "-w")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("keychain access failed: %w", err)
	}

	var creds keychainCredentials
	if err := json.Unmarshal(output, &creds); err != nil {
		return "", fmt.Errorf("failed to parse keychain credentials: %w", err)
	}

	if creds.ClaudeAiOAuth.AccessToken == "" {
		return "", fmt.Errorf("no access token in keychain credentials")
	}

	return creds.ClaudeAiOAuth.AccessToken, nil
}

// fileSource retrieves credentials from a JSON file on disk.
type fileSource struct {
	path string
}

func (f fileSource) AccessToken() (string, error) {
	data, err := os.ReadFile(f.path)
	if err != nil {
		return "", fmt.Errorf("failed to read credentials file: %w", err)
	}

	var creds keychainCredentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return "", fmt.Errorf("failed to parse credentials file: %w", err)
	}

	if creds.ClaudeAiOAuth.AccessToken == "" {
		return "", fmt.Errorf("no access token in credentials file")
	}

	return creds.ClaudeAiOAuth.AccessToken, nil
}

// NewCredentialSource returns the appropriate CredentialSource for the current platform.
// On macOS, it reads from the Keychain. On Linux, it reads from ~/.claude/.credentials.json.
func NewCredentialSource() (CredentialSource, error) {
	switch runtime.GOOS {
	case "darwin":
		return keychainSource{}, nil
	case "linux":
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("cannot determine home directory: %w", err)
		}

		return fileSource{path: filepath.Join(home, ".claude", ".credentials.json")}, nil
	default:
		return nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}
