package modules

import (
	"os"
	"strings"

	"github.com/felipeelias/claude-statusline/internal/config"
	"github.com/felipeelias/claude-statusline/internal/input"
)

const defaultEnvVarFormat = "{{.Value}}"

// EnvVarModule renders the env_var instance named by Instance, or every
// instance in definition order, space-separated, when Instance is empty.
type EnvVarModule struct {
	Instance string
}

func (EnvVarModule) Name() string { return "env_var" }

func (m EnvVarModule) Render(_ input.Data, cfg config.Config) (string, error) {
	if m.Instance != "" {
		return renderEnvVar(m.Instance, cfg)
	}

	var parts []string

	for _, name := range cfg.EnvVarNames() {
		rendered, err := renderEnvVar(name, cfg)
		if err != nil {
			return "", err
		}

		if rendered != "" {
			parts = append(parts, rendered)
		}
	}

	return strings.Join(parts, " "), nil
}

func renderEnvVar(instance string, cfg config.Config) (string, error) {
	envCfg, configured := cfg.EnvVar[instance]
	if !configured || envCfg.Disabled {
		return "", nil
	}

	variable := envCfg.Variable
	if variable == "" {
		variable = instance
	}

	value, set := os.LookupEnv(variable)
	if !set {
		if envCfg.Default == nil {
			return "", nil
		}

		value = *envCfg.Default
	}

	format := envCfg.Format
	if format == "" {
		format = defaultEnvVarFormat
	}

	templateData := struct{ Name, Value string }{Name: variable, Value: value}

	result, err := renderTemplate("env_var."+instance, format, templateData)
	if err != nil || result == "" {
		return "", err
	}

	return wrapStyle(result, envCfg.Style), nil
}
