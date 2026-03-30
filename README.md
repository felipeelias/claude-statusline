# claude-statusline

Configurable status line for [Claude Code](https://docs.anthropic.com/en/docs/claude-code).

![claude-statusline](assets/screenshot.webp)

## Installation

With Homebrew:

```bash
brew install felipeelias/tap/claude-statusline
```

Or with Go:

```bash
go install github.com/felipeelias/claude-statusline@latest
```

## Setup

Add to your Claude Code settings (`.claude/settings.json` or global settings):

```json
{
  "statusLine": {
    "type": "command",
    "command": "claude-statusline prompt"
  }
}
```

Generate a starter config:

```bash
claude-statusline init
```

Preview with mock data:

```bash
claude-statusline test
claude-statusline themes
```

## Commands

| Command | Description |
|---------|-------------|
| `prompt` | Render the status line (also the default when no command is given) |
| `init` | Create default config at `~/.config/claude-statusline/config.toml` |
| `test` | Render with your config and mock data (for config iteration) |
| `themes` | Preview all built-in presets with mock data |

Global flags: `--config / -c` to override config path, `--version`.

## Configuration

Config file location: `~/.config/claude-statusline/config.toml`

Works with zero config. The default format is:

```toml
format = "$directory | $git_branch | $model | $cost | $context"
```

## Presets

Presets are inspired by [Starship presets](https://starship.rs/presets/). Each preset defines the layout, separators, colors, and module configuration.

```toml
preset = "catppuccin"
```

Preview all presets: `claude-statusline themes`

### Built-in presets

| Preset | Description | Nerd Font |
|--------|-------------|-----------|
| `default` | Flat with `\|` pipes, standard colors | No |
| `minimal` | Clean spacing, no separators | No |
| `pastel-powerline` | Pastel powerline arrows (pink/peach/blue/teal) | Yes |
| `tokyo-night` | Dark blues rounded powerline with gradient | Yes |
| `gruvbox-rainbow` | Earthy rainbow powerline | Yes |
| `catppuccin` | Catppuccin Mocha powerline | Yes |

### Overriding preset defaults

Presets set the format string and module configs, but you can override any field:

```toml
preset = "catppuccin"

# Override just one module
[model]
format = " {{.DisplayName}} "
style = "fg:#11111b bg:#cba6f7 bold"
```

## Modules

| Module | Default | Description |
|--------|---------|-------------|
| `directory` | on | Current directory (tilde-collapsed, truncated) |
| `git_branch` | on | Git branch with status indicators (dirty, ahead/behind, worktree) |
| `model` | on | Model name (display name, short name, or raw ID) |
| `cost` | on | Session cost in USD |
| `context` | on | Context window usage with progress bar |
| `session_timer` | off | Session elapsed time |
| `lines_changed` | off | Lines added/removed |
| `usage` | off | Plan usage limits (5-hour block and weekly) |
| `vim_mode` | off | Vim mode indicator (NORMAL, INSERT, etc.) |

### Enabling modules

To enable a disabled module, set `disabled = false` and add it to the format string:

```toml
format = "$directory | $git_branch | $model | $cost | $context | $session_timer"

[session_timer]
disabled = false
```

### Model module

Template fields:

| Field | Description | Example |
|-------|-------------|---------|
| `{{.DisplayName}}` | Display name from Claude Code (default) | `Claude Sonnet 4.6` |
| `{{.Short}}` | Compact name extracted from model ID | `Sonnet 4.6` |
| `{{.ID}}` | Raw model ID | `claude-sonnet-4-6-20250514` |

```toml
[model]
format = "{{.Short}}"
style = "bold"
```

### Usage module

The `usage` module shows your Claude plan usage limits (5-hour rolling window and 7-day). Requires Claude Code 2.1.80+ which provides `rate_limits` in the status line payload.

```toml
format = "$directory | $git_branch | $model | $cost | $context | $usage"

[usage]
disabled = false
```

Template fields:

| Field | Description |
|-------|-------------|
| `{{.BlockPct}}` | 5-hour rolling window usage (0-100) |
| `{{.WeeklyPct}}` | 7-day usage (0-100) |
| `{{.BlockBar}}` | Progress bar for 5-hour window |
| `{{.WeeklyBar}}` | Progress bar for 7-day window |
| `{{.BlockResets}}` | Time until 5-hour reset (e.g. "2h13m") |
| `{{.WeeklyResets}}` | Time until 7-day reset (e.g. "3d2h") |

To only show usage when it exceeds a threshold (e.g. 5-hour block above 70%, weekly above 80%):

```toml
[usage]
disabled = false
format = '{{if ge .BlockPct 70.0}}{{.BlockBar}} {{printf "%.0f" .BlockPct}}%{{end}}{{if ge .WeeklyPct 80.0}} W:{{printf "%.0f" .WeeklyPct}}%{{end}}'
```

The module renders empty if `rate_limits` is not present in the Claude Code payload (older versions).

### Vim mode module

The `vim_mode` module shows the current vim editor mode when vim mode is enabled in Claude Code.

```toml
format = "$vim_mode | $directory | $git_branch | $model | $cost | $context"

[vim_mode]
disabled = false
```

Template fields:

| Field | Description |
|-------|-------------|
| `{{.Mode}}` | Current vim mode (e.g. `NORMAL`, `INSERT`) |

The module renders empty if vim mode is not enabled or the mode string is empty.

## Style system

Modules support a `style` field that accepts several formats:

| Format | Example |
|--------|---------|
| Named | `red`, `green`, `cyan`, `bold`, `dim`, `italic` |
| Hex | `fg:#ff5500`, `bg:#333333` |
| 256-color | `208`, `fg:208`, `bg:238` |
| Combined | `fg:#aabbcc bg:#333333 bold` |

## Alternatives

Other statusline tools from the [awesome-claude-code](https://github.com/hesreallyhim/awesome-claude-code) list:

- [claude-powerline](https://github.com/Owloops/claude-powerline)
- [CCometixLine](https://github.com/Haleclipse/CCometixLine)
- [claudia-statusline](https://github.com/hagan/claudia-statusline)
- [ccstatusline](https://github.com/sirmalloc/ccstatusline)

## Contributors

- [@sammcj](https://github.com/sammcj)

## License

MIT
