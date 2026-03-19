package usage

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCredentialSource(t *testing.T) {
	src, err := NewCredentialSource()

	switch runtime.GOOS {
	case "darwin":
		require.NoError(t, err)
		assert.IsType(t, keychainSource{}, src)
	case "linux":
		require.NoError(t, err)
		assert.IsType(t, fileSource{}, src)
	default:
		assert.Error(t, err)
		assert.Nil(t, src)
	}
}

func TestFileSource_AccessToken(t *testing.T) {
	t.Run("valid credentials", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "credentials.json")

		content := `{"claudeAiOauth":{"accessToken":"test-token-123"}}`
		require.NoError(t, os.WriteFile(path, []byte(content), 0600))

		src := fileSource{path: path}
		token, err := src.AccessToken()

		require.NoError(t, err)
		assert.Equal(t, "test-token-123", token)
	})

	t.Run("missing file", func(t *testing.T) {
		src := fileSource{path: "/nonexistent/credentials.json"}
		_, err := src.AccessToken()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read credentials file")
	})

	t.Run("malformed JSON", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "credentials.json")

		require.NoError(t, os.WriteFile(path, []byte("not json"), 0600))

		src := fileSource{path: path}
		_, err := src.AccessToken()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse credentials file")
	})

	t.Run("empty access token", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "credentials.json")

		content := `{"claudeAiOauth":{"accessToken":""}}`
		require.NoError(t, os.WriteFile(path, []byte(content), 0600))

		src := fileSource{path: path}
		_, err := src.AccessToken()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no access token")
	})

	t.Run("missing oauth field", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "credentials.json")

		content := `{"someOtherField":"value"}`
		require.NoError(t, os.WriteFile(path, []byte(content), 0600))

		src := fileSource{path: path}
		_, err := src.AccessToken()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no access token")
	})
}
