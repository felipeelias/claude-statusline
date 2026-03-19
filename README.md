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
| `git_branch` | on | Current git branch (with worktree indicator) |
| `model` | on | Model display name |
| `cost` | on | Session cost in USD |
| `context` | on | Context window usage with progress bar |
| `session_timer` | off | Session elapsed time |
| `lines_changed` | off | Lines added/removed |
| `usage` | off | Plan usage limits (5-hour block and weekly) |

### Enabling modules

To enable a disabled module, set `disabled = false` and add it to the format string:

```toml
format = "$directory | $git_branch | $model | $cost | $context | $session_timer"

[session_timer]
disabled = false
```

### Usage module

The `usage` module shows your Claude plan usage limits by querying the Anthropic OAuth API. It requires OAuth credentials:

- **macOS**: Reads from the system Keychain (set up automatically by `claude login`)
- **Linux**: Reads from `~/.claude/.credentials.json`

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

To only show usage when it exceeds a threshold (e.g. session above 70%, weekly above 80%):

```toml
[usage]
disabled = false
format = '{{if ge .BlockPct 70.0}}{{.BlockBar}} {{printf "%.0f" .BlockPct}}%{{end}}{{if ge .WeeklyPct 80.0}} W:{{printf "%.0f" .WeeklyPct}}%{{end}}'
```

API responses are cached to `~/.cache/claude-statusline/usage.json` (default TTL: 120 seconds). The module renders empty if credentials are unavailable.

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

## License

MIT
