# claude-statusline

Configurable status line for [Claude Code](https://docs.anthropic.com/en/docs/claude-code).

```
~/project |  main | Claude Opus 4 | $0.42 | ██░░░ 42%
```

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
| `themes` | Preview all built-in themes with mock data |

Global flags: `--config / -c` to override config path, `--version`.

## Configuration

Config file location: `~/.config/claude-statusline/config.toml`

Works with zero config. The default format is:

```toml
format = "$directory | $git_branch | $model | $cost | $context"
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

### Enabling modules

To enable a disabled module, set `disabled = false` and add it to the format string:

```toml
format = "$directory | $git_branch | $model | $cost | $context | $session_timer"

[session_timer]
disabled = false
```

## Style system

Modules support a `style` field that accepts several formats:

- **Named:** `red`, `green`, `cyan`, `bold`, `dim`, `italic`
- **Hex:** `fg:#ff5500`, `bg:#333333`
- **256-color:** `208`
- **Combined:** `fg:#aaa bg:#333 bold`
- **Palette:** `palette:accent`

## Themes

Four palettes are built in: `default`, `tokyo-night`, `gruvbox`, and `catppuccin`. Switch with a single line:

```toml
palette = "tokyo-night"
```

Preview all themes: `claude-statusline themes`

You can also define your own palette:

```toml
palette = "my-theme"

[palettes.my-theme]
accent = "#ff5500"
cost_ok = "green"
cost_warn = "yellow"
cost_high = "red"
ctx_ok = "green"
ctx_warn = "yellow"
ctx_high = "red"
```

## Powerline

Opt into powerline-style separators. Requires a [Nerd Font](https://www.nerdfonts.com/).

The format string uses styled text groups for segment transitions:

- `` (start cap) with `fg:` matching the first segment background
- `` (arrow) with `fg:prev_bg bg:next_bg` for transitions between segments
- Each module's `style` must include a matching `bg:` color
- Each module's `format` should include padding spaces

```toml
format = "[](fg:blue)$directory[](fg:blue bg:green)$git_branch[](fg:green bg:magenta)$model[](fg:magenta)"

[directory]
format = " {{.Dir}} "
style = "fg:black bg:blue"

[git_branch]
disabled = false
format = "  {{.Branch}} "
style = "fg:black bg:green"

[model]
format = " {{.DisplayName}} "
style = "fg:black bg:magenta bold"
```

### Catppuccin Mocha Powerline

A complete powerline theme using [Catppuccin Mocha](https://catppuccin.com/) colors:

```toml
palette = "catppuccin-mocha"

format = "[](fg:#89b4fa)$directory[](fg:#89b4fa bg:#a6e3a1)$git_branch[](fg:#a6e3a1 bg:#cba6f7)$model[](fg:#cba6f7 bg:#45475a)$cost[](fg:#45475a bg:#313244)$context[](fg:#313244)"

[palettes.catppuccin-mocha]
accent = "#89b4fa"
cost_ok = "#a6e3a1"
cost_warn = "#f9e2af"
cost_high = "#f38ba8"
ctx_ok = "#a6e3a1"
ctx_warn = "#f9e2af"
ctx_high = "#f38ba8"

[directory]
format = " {{.Dir}} "
style = "fg:#1e1e2e bg:#89b4fa"

[git_branch]
disabled = false
format = "  {{.Branch}}{{if .InWorktree}} {{end}} "
style = "fg:#1e1e2e bg:#a6e3a1"

[model]
format = " {{.DisplayName}} "
style = "fg:#1e1e2e bg:#cba6f7 bold"

[cost]
format = " ${{printf \"%.2f\" .TotalCostUSD}} "
style = "fg:#a6e3a1 bg:#45475a"
thresholds = [
  { above = 1.0, style = "fg:#f9e2af bg:#45475a" },
  { above = 5.0, style = "fg:#f38ba8 bg:#45475a" },
]

[context]
format = " {{.Bar}} {{printf \"%.0f\" .UsedPct}}% "
style = "fg:#a6e3a1 bg:#313244"
bar_width = 5
bar_fill = "█"
bar_empty = "░"
thresholds = [
  { above = 50, style = "fg:#f9e2af bg:#313244" },
  { above = 70, style = "fg:#fab387 bg:#313244" },
  { above = 90, style = "fg:#f38ba8 bg:#313244" },
]
```

## License

MIT
