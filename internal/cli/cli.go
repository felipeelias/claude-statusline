package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/frank-bee/claude-statusline/internal/anthropic"
	"github.com/frank-bee/claude-statusline/internal/config"
	"github.com/frank-bee/claude-statusline/internal/input"
	"github.com/frank-bee/claude-statusline/internal/render"
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
		Action: promptAction,
		Commands: []*ucli.Command{
			promptCommand(),
			initCommand(),
			testCommand(),
			themesCommand(),
			planCommand(),
			refreshUsageCommand(),
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

// planCommand reports which account the current session is signed in as and
// what that plan meters. It exists because the two usage segments are not
// interchangeable - windows on Pro/Max/Team, a money budget on a usage-based
// Enterprise seat - and an empty segment gives no clue which one applies.
func planCommand() *ucli.Command {
	return &ucli.Command{
		Name:  "plan",
		Usage: "Show the current account, its plan, and what that plan meters",
		Action: func(cmd *ucli.Context) error {
			profile, err := anthropic.LoadProfile()
			if err != nil {
				fmt.Fprintln(cmd.App.ErrWriter, "cannot read profile:", err)

				return nil
			}

			plan := profile.Plan()

			meters := "nothing this tool can read"

			switch {
			case plan.MetersBudget():
				meters = "a money budget  ->  $credits"
			case plan.MetersWindows():
				meters = "rate-limit windows  ->  $usage"
			}

			fmt.Fprintf(cmd.App.Writer, "account       %s\n", profile.Account.Email)
			fmt.Fprintf(cmd.App.Writer, "organization  %s\n", profile.Organization.Name)
			fmt.Fprintf(cmd.App.Writer, "plan          %s\n", plan)

			if profile.Organization.SeatTier != "" {
				fmt.Fprintf(cmd.App.Writer, "seat          %s\n", profile.Organization.SeatTier)
			}

			fmt.Fprintf(cmd.App.Writer, "meters        %s\n", meters)

			usage, err := anthropic.Load()
			if err == nil && usage.Spend.Enabled && usage.Spend.Limit != nil {
				fmt.Fprintf(cmd.App.Writer, "budget        %.2f / %.2f %s (%.0f%%)\n",
					usage.Spend.Used.Major(), usage.Spend.Limit.Major(),
					usage.Spend.Used.Currency, usage.Spend.Percent)
			}

			return nil
		},
	}
}

// refreshUsageCommand refreshes the cached Anthropic usage reading. The
// credits module spawns it detached when its cache goes stale, so the render
// path never waits on an HTTP call; it is hidden because nobody runs it by hand.
func refreshUsageCommand() *ucli.Command {
	return &ucli.Command{
		Name:   "refresh-usage",
		Usage:  "Refresh the cached Anthropic usage reading",
		Hidden: true,
		Action: func(cmd *ucli.Context) error {
			err := anthropic.Refresh()
			if err != nil {
				fmt.Fprintln(cmd.App.ErrWriter, "refresh error:", err)
			}

			return nil
		},
	}
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

			return previewModules(writer, data)
		},
	}
}

// previewModules renders the modules that are off by default, so \`themes\` shows
// what enabling them looks like rather than only what the presets ship with.
//
// windows and credits read the usage cache, so the preview points XDG_STATE_HOME
// at a throwaway directory holding a mock reading. That exercises the real render
// path - no preview-only seam in the shipped code - and a cache written just now
// is fresh, so nothing triggers a network refresh.
func previewModules(writer io.Writer, data input.Data) error {
	restore, err := mockUsageCache()
	if err != nil {
		// A preview is not worth failing the command over: the presets above
		// have already rendered.
		return nil
	}
	defer restore()

	_, _ = fmt.Fprintf(writer, "optional modules (off by default, mock readings):\n\n")

	for _, variant := range moduleVariants() {
		output, renderErr := render.Render(variant.cfg, data)
		if renderErr != nil {
			return fmt.Errorf("rendering %s: %w", variant.name, renderErr)
		}

		_, _ = fmt.Fprintf(writer, "%s:\n  %s\n\n", variant.name, output)
	}

	return nil
}

type moduleVariant struct {
	name string
	cfg  config.Config
}

func moduleVariants() []moduleVariant {
	usageCfg := config.Default()
	usageCfg.Format = "$directory  $git_branch  $model  $context  $usage"
	usageCfg.Usage.Disabled = false

	windowsCfg := config.Default()
	windowsCfg.Format = "$directory  $model  $windows"
	windowsCfg.Windows.Disabled = false

	creditsCfg := config.Default()
	creditsCfg.Format = "$directory  $model  $credits"
	creditsCfg.Credits.Disabled = false

	bothCfg := config.Default()
	bothCfg.Format = "$directory  $model  $windows  $credits"
	bothCfg.Windows.Disabled = false
	bothCfg.Credits.Disabled = false

	return []moduleVariant{
		{"usage (from the Claude Code payload)", usageCfg},
		{"windows (from Anthropic)", windowsCfg},
		{"credits (from Anthropic)", creditsCfg},
		{"windows + credits", bothCfg},
	}
}

const (
	cacheDirPerms  = 0o700
	cacheFilePerms = 0o600
)

// mockUsageCache writes a fake reading into a temporary XDG_STATE_HOME and
// returns a function restoring the previous environment.
func mockUsageCache() (func(), error) {
	dir, err := os.MkdirTemp("", "claude-statusline-preview")
	if err != nil {
		return nil, err
	}

	// The environment is redirected before the reading is written, because the
	// cache layout is per-account: only the package can say where the file for
	// the current account goes, and it answers relative to XDG_STATE_HOME.
	previous, had := os.LookupEnv("XDG_STATE_HOME")

	err = os.Setenv("XDG_STATE_HOME", dir)
	if err != nil {
		return nil, err
	}

	path, err := anthropic.CachePath()
	if err != nil {
		return nil, err
	}

	err = os.WriteFile(path, []byte(mockUsageJSON), cacheFilePerms)
	if err != nil {
		return nil, err
	}

	return func() {
		if had {
			_ = os.Setenv("XDG_STATE_HOME", previous)
		} else {
			_ = os.Unsetenv("XDG_STATE_HOME")
		}

		_ = os.RemoveAll(dir)
	}, nil
}

const mockUsageJSON = `{
  "limits": [
    {"kind": "session", "percent": 42, "severity": "normal"},
    {"kind": "weekly_all", "percent": 63, "severity": "warning"}
  ],
  "spend": {
    "enabled": true,
    "percent": 58,
    "used": {"amount_minor": 11600, "currency": "USD", "exponent": 2},
    "limit": {"amount_minor": 20000, "currency": "USD", "exponent": 2}
  }
}`

//nolint:mnd // mock data uses literal values by design
func mockInput() input.Data {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "/tmp/project"
	}

	return input.Data{
		SessionID:      "test-session",
		TranscriptPath: "/tmp/claude/transcript.jsonl",
		Version:        "1.0.0",
		Model: input.Model{
			ID:          "claude-opus-4-20250514",
			DisplayName: "Claude Opus 4",
		},
		Cwd: cwd,
		Workspace: input.Workspace{
			CurrentDir: cwd,
			ProjectDir: cwd,
		},
		Cost: input.Cost{
			TotalCostUSD:       0.42,
			TotalDurationMs:    180000,
			TotalAPIDurationMs: 12000,
			TotalLinesAdded:    42,
			TotalLinesRemoved:  7,
		},
		ContextWindow: input.ContextWindow{
			TotalInputTokens:    15234,
			TotalOutputTokens:   4521,
			UsedPercentage:      42.5,
			RemainingPercentage: 57.5,
			ContextWindowSize:   200000,
			CurrentUsage: &input.CurrentUsage{
				InputTokens:              8500,
				OutputTokens:             1200,
				CacheCreationInputTokens: 5000,
				CacheReadInputTokens:     2000,
			},
		},
		OutputStyle: input.OutputStyle{
			Name: "default",
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
