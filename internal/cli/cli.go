package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/felipeelias/claude-statusline/internal/config"
	"github.com/felipeelias/claude-statusline/internal/input"
	"github.com/felipeelias/claude-statusline/internal/render"
	ucli "github.com/urfave/cli/v2"
)

const (
	configDirPerms  = 0750
	configFilePerms = 0600
)

// New creates the CLI application.
func New(version string) *ucli.App {
	return &ucli.App{
		Name:    "claude-statusline",
		Usage:   "Configurable status line for Claude Code",
		Version: version,
		Flags: []ucli.Flag{
			&ucli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Path to config file",
				Value:   config.DefaultPath(),
				EnvVars: []string{"CLAUDE_STATUSLINE_CONFIG"},
			},
		},
		Action:   promptAction,
		Commands: []*ucli.Command{
			promptCommand(),
			initCommand(),
			testCommand(),
			themesCommand(),
		},
	}
}

func promptAction(cmd *ucli.Context) error {
	configPath := cmd.String("config")

	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintln(cmd.App.ErrWriter, "config error:", err)

		return nil
	}

	reader := cmd.App.Reader
	if reader == nil {
		reader = os.Stdin
	}

	data, err := input.Parse(reader)
	if err != nil {
		fmt.Fprintln(cmd.App.ErrWriter, "input error:", err)

		return nil
	}

	output, err := render.Render(cfg, data)
	if err != nil {
		fmt.Fprintln(cmd.App.ErrWriter, "render error:", err)

		return nil
	}

	_, _ = fmt.Fprint(cmd.App.Writer, output)

	return nil
}

func promptCommand() *ucli.Command {
	return &ucli.Command{
		Name:   "prompt",
		Usage:  "Render the status line (default action)",
		Action: promptAction,
	}
}

func initCommand() *ucli.Command {
	return &ucli.Command{
		Name:  "init",
		Usage: "Create default config file",
		Action: func(cmd *ucli.Context) error {
			configPath := cmd.String("config")

			_, err := os.Stat(configPath)
			if err == nil {
				return fmt.Errorf("config already exists at %s", configPath)
			}

			err = os.MkdirAll(filepath.Dir(configPath), configDirPerms)
			if err != nil {
				return fmt.Errorf("creating config directory: %w", err)
			}

			sample := config.SampleConfig()
			err = os.WriteFile(configPath, []byte(sample), configFilePerms)
			if err != nil {
				return fmt.Errorf("writing config: %w", err)
			}

			_, _ = fmt.Fprintf(cmd.App.Writer, "Config created at %s\n", configPath)

			return nil
		},
	}
}

func testCommand() *ucli.Command {
	return &ucli.Command{
		Name:  "test",
		Usage: "Render with your config and mock data",
		Action: func(cmd *ucli.Context) error {
			configPath := cmd.String("config")

			cfg, err := config.Load(configPath)
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			output, err := render.Render(cfg, mockInput())
			if err != nil {
				return fmt.Errorf("rendering: %w", err)
			}

			_, _ = fmt.Fprintln(cmd.App.Writer, output)

			return nil
		},
	}
}

func themesCommand() *ucli.Command {
	return &ucli.Command{
		Name:  "themes",
		Usage: "Preview all built-in presets with mock data",
		Action: func(cmd *ucli.Context) error {
			writer := cmd.App.Writer
			data := mockInput()

			// Show user's current config first.
			configPath := cmd.String("config")

			userCfg, err := config.Load(configPath)
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			output, err := render.Render(userCfg, data)
			if err != nil {
				return fmt.Errorf("rendering current: %w", err)
			}

			_, _ = fmt.Fprintf(writer, "current:\n  %s\n\n", output)

			for _, name := range config.PresetNames() {
				cfg, _ := config.ApplyPreset(name)
				output, err := render.Render(cfg, data)
				if err != nil {
					return fmt.Errorf("rendering %s: %w", name, err)
				}

				_, _ = fmt.Fprintf(writer, "%s:\n  %s\n\n", name, output)
			}

			return nil
		},
	}
}

//nolint:mnd // mock data uses literal values by design
func mockInput() input.Data {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "/tmp/project"
	}

	return input.Data{
		SessionID: "test-session",
		Version:   "1.0.0",
		Model: input.Model{
			ID:          "claude-opus-4-20250514",
			DisplayName: "Claude Opus 4",
		},
		Cwd: cwd,
		Cost: input.Cost{
			TotalCostUSD:      0.42,
			TotalDurationMs:   180000,
			TotalLinesAdded:   42,
			TotalLinesRemoved: 7,
		},
		ContextWindow: input.ContextWindow{
			UsedPercentage:      42.5,
			RemainingPercentage: 57.5,
			ContextWindowSize:   200000,
		},
		RateLimits: &input.RateLimits{
			FiveHour: input.RateLimitWindow{
				UsedPercentage: 42,
				ResetsAt:       time.Now().Add(2*time.Hour + 13*time.Minute).Unix(),
			},
			SevenDay: input.RateLimitWindow{
				UsedPercentage: 15,
				ResetsAt:       time.Now().Add(3*24*time.Hour + 2*time.Hour).Unix(),
			},
		},
	}
}
