package modules_test

import (
	"testing"

	"github.com/felipeelias/claude-statusline/internal/config"
	"github.com/felipeelias/claude-statusline/internal/input"
	"github.com/felipeelias/claude-statusline/internal/modules"
	"github.com/felipeelias/claude-statusline/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func envVarConfig(instances map[string]config.EnvVarConfig) config.Config {
	cfg := config.Default()
	cfg.EnvVar = instances

	return cfg
}

func renderEnvVar(t *testing.T, instance string, cfg config.Config) string {
	t.Helper()

	result, err := modules.EnvVarModule{Instance: instance}.Render(input.Data{}, cfg)
	require.NoError(t, err)

	return result
}

func TestEnvVarModule_Name(t *testing.T) {
	assert.Equal(t, "env_var", modules.EnvVarModule{}.Name())
}

func TestEnvVarModule_Render(t *testing.T) {
	t.Run("reads the variable named by the instance", func(t *testing.T) {
		t.Setenv("CSL_TEST_USER", "felipe")
		cfg := envVarConfig(map[string]config.EnvVarConfig{"CSL_TEST_USER": {}})

		assert.Equal(t, "felipe", renderEnvVar(t, "CSL_TEST_USER", cfg))
	})

	t.Run("variable option overrides the instance name", func(t *testing.T) {
		t.Setenv("CSL_TEST_PROFILE", "prod")
		cfg := envVarConfig(map[string]config.EnvVarConfig{
			"aws": {Variable: "CSL_TEST_PROFILE", Format: "{{.Name}}={{.Value}}"},
		})

		assert.Equal(t, "CSL_TEST_PROFILE=prod", renderEnvVar(t, "aws", cfg))
	})

	t.Run("unset variable without default renders empty", func(t *testing.T) {
		testutil.UnsetEnv(t, "CSL_TEST_UNSET")
		cfg := envVarConfig(map[string]config.EnvVarConfig{"CSL_TEST_UNSET": {Format: "x{{.Value}}"}})

		assert.Empty(t, renderEnvVar(t, "CSL_TEST_UNSET", cfg))
	})

	t.Run("unset variable uses default", func(t *testing.T) {
		testutil.UnsetEnv(t, "CSL_TEST_UNSET")
		fallback := "none"
		cfg := envVarConfig(map[string]config.EnvVarConfig{"CSL_TEST_UNSET": {Default: &fallback}})

		assert.Equal(t, "none", renderEnvVar(t, "CSL_TEST_UNSET", cfg))
	})

	t.Run("empty default still counts as configured", func(t *testing.T) {
		testutil.UnsetEnv(t, "CSL_TEST_UNSET")
		fallback := ""
		cfg := envVarConfig(map[string]config.EnvVarConfig{
			"CSL_TEST_UNSET": {Default: &fallback, Format: "[{{.Value}}]"},
		})

		assert.Equal(t, "[]", renderEnvVar(t, "CSL_TEST_UNSET", cfg))
	})

	t.Run("set but empty variable renders", func(t *testing.T) {
		t.Setenv("CSL_TEST_EMPTY", "")
		cfg := envVarConfig(map[string]config.EnvVarConfig{"CSL_TEST_EMPTY": {Format: "set{{.Value}}"}})

		assert.Equal(t, "set", renderEnvVar(t, "CSL_TEST_EMPTY", cfg))
	})

	t.Run("disabled instance renders empty", func(t *testing.T) {
		t.Setenv("CSL_TEST_USER", "felipe")
		cfg := envVarConfig(map[string]config.EnvVarConfig{"CSL_TEST_USER": {Disabled: true}})

		assert.Empty(t, renderEnvVar(t, "CSL_TEST_USER", cfg))
	})

	t.Run("unknown instance renders empty", func(t *testing.T) {
		assert.Empty(t, renderEnvVar(t, "missing", config.Default()))
	})

	t.Run("template matches on value", func(t *testing.T) {
		cfg := envVarConfig(map[string]config.EnvVarConfig{
			"CSL_TEST_SANDBOX": {Format: `{{if eq .Value "agent-safehouse"}}📦{{end}}`, Style: "bold"},
		})

		t.Setenv("CSL_TEST_SANDBOX", "agent-safehouse")
		assert.Equal(t, "\033[1m📦\033[0m", renderEnvVar(t, "CSL_TEST_SANDBOX", cfg))

		t.Setenv("CSL_TEST_SANDBOX", "other")
		assert.Empty(t, renderEnvVar(t, "CSL_TEST_SANDBOX", cfg))
	})

	t.Run("empty instance renders all instances", func(t *testing.T) {
		t.Setenv("CSL_TEST_A", "a")
		t.Setenv("CSL_TEST_B", "b")
		testutil.UnsetEnv(t, "CSL_TEST_UNSET")
		cfg := envVarConfig(map[string]config.EnvVarConfig{
			"CSL_TEST_B":     {Format: "[{{.Value}}]"},
			"CSL_TEST_A":     {Format: "[{{.Value}}]"},
			"CSL_TEST_UNSET": {Format: "[{{.Value}}]"},
		})

		assert.Equal(t, "[a] [b]", renderEnvVar(t, "", cfg))
	})

	t.Run("invalid template returns error", func(t *testing.T) {
		t.Setenv("CSL_TEST_USER", "felipe")
		cfg := envVarConfig(map[string]config.EnvVarConfig{"CSL_TEST_USER": {Format: "{{.Value"}})

		_, err := modules.EnvVarModule{Instance: "CSL_TEST_USER"}.Render(input.Data{}, cfg)
		require.Error(t, err)
	})
}
