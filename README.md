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
| `themes` | Preview all built-in themes and palettes with mock data |

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
- **256-color:** `208`, `fg:208`, `bg:238`
- **Combined:** `fg:#aaa bg:#333 bold`
- **Palette:** `palette:accent`, `fg:palette:seg_fg bg:palette:dir_bg`

## Themes and palettes

**Themes** control the visual structure (separators, padding, icons). **Palettes** control the colors. Any theme works with any palette.

```toml
theme = "powerline"
palette = "catppuccin"
```

Preview all combinations: `claude-statusline themes`

### Built-in themes

| Theme | Look | Needs Nerd Font |
|-------|------|-----------------|
| `default` | Flat with `\|` pipes | No |
| `powerline` | Arrow segments with colored backgrounds | Yes |
| `rounded` | Capsule/pill segments with gaps | Yes |
| `minimal` | Clean spacing, no separators or icons | No |

### Built-in palettes

Four palettes are built in: `default`, `tokyo-night`, `gruvbox`, and `catppuccin`.

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
# Segment colors (used by powerline/rounded themes)
seg_fg = "black"
dir_bg = "blue"
git_bg = "green"
model_bg = "magenta"
cost_bg = "238"
ctx_bg = "236"
```

### Overriding theme defaults

Themes set the format string and module configs, but you can override any field:

```toml
theme = "powerline"
palette = "catppuccin"

# Override just one module
[model]
format = " {{.DisplayName}} "
style = "fg:palette:seg_fg bg:palette:model_bg bold"
```

## License

MIT
