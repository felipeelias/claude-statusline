# claude-statusline

Configurable status line for [Claude Code](https://docs.anthropic.com/en/docs/claude-code).
Shows where you are, what you are running, and how much of your plan you have used — in one
line, with no API key and no configuration to get started.

![claude-statusline, plain and powerline](assets/preview.png)

## Installation

With Homebrew:

```bash
brew install frank-bee/tap/claude-statusline
brew upgrade claude-statusline          # later
```

Or with Go:

```bash
go install github.com/frank-bee/claude-statusline@latest
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

`themes` previews every preset and then `usage`, `windows` and `credits` against a mock
reading, so you can see what enabling them looks like before wiring up a real one. The mock
lives in a throwaway directory; your own config and cached reading are untouched.

## Commands

| Command | Description |
|---------|-------------|
| `prompt` | Render the status line (also the default when no command is given) |
| `init` | Create default config at `~/.config/claude-statusline/config.toml` |
| `test` | Render with your config and mock data (for config iteration) |
| `themes` | Preview all built-in presets, then the modules that are off by default, with mock data |
| `plan` | Which account this session is signed in as, its plan, and whether that plan meters windows or a money budget |

Global flags: `--config / -c` to override config path, `--version`.

## Configuration

Config file location: `~/.config/claude-statusline/config.toml`

Works with zero config. The default format is:

```toml
format = "$directory | $git_branch | $model | $cost | $context | $usage"
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

### Progress bars on powerline presets

`context`, `windows`, `credits` and `usage` draw a bar with `bar_fill` / `bar_empty`
(or a named `bar_style`). The default `░` empty cell is written in the segment's
foreground colour, so on the light and mid-tone pills of `catppuccin` and
`pastel-powerline` it reads as a solid block rather than as empty track. On those
presets either pick a thinner track:

```toml
[context]
bar_fill = "━"
bar_empty = "─"
```

or drop `{{.Bar}}` from the format and keep the percentage alone. The dark pills of
`tokyo-night` and `gruvbox-rainbow` render `█░` blocks fine.

Bar width is a resolution limit: at `bar_width = 5` each cell is 20%, so anything
under that fills a single cell. Widen the bar when you want small readings to be
legible rather than merely visible.

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
| `usage` | **on** | Plan usage limits (5-hour block and weekly), from the payload |
| `windows` | off | Same windows, read from Anthropic instead of the payload |
| `credits` | off | Credit spend on usage-based seats, read from Anthropic |
| `vim_mode` | off | Vim mode indicator (NORMAL, INSERT, etc.) |
| `agent_name` | off | Agent name when running with `--agent` |

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

### Where usage figures come from

There are two sources, and they answer different questions.

`usage` reads the `rate_limits` Claude Code puts in the status line payload. It is on by
default, costs no request, reads no credentials, and cannot be rate-limited. For the 5-hour
and weekly windows, this is all you need.

`windows` and `credits` ask Anthropic's usage API directly, using the OAuth token Claude Code
already holds. The one thing they add is **`credits`**: the credit pool metered on usage-based
seats, which the payload does not report at all. `windows` returns the same two windows
`usage` already shows, so enable it only if you specifically want the account's own accounting.

**On an Enterprise usage-based seat, `usage` has nothing to show.** There are no rate-limit
windows on such a seat: Claude Code sends no `rate_limits` in the payload, and every window in
the API response comes back `null`. `credits` carries the reading instead — the money budget
and how much of it is spent. On Pro the reverse holds: `spend.enabled` is `false` and
`credits` is the silent one.

Run `claude-statusline plan` to see which applies:

```
$ claude-statusline plan
account       someone@example.com
organization  example-org
plan          enterprise
seat          enterprise_usage_based
meters        a money budget  ->  $credits
budget        776.24 / 993.00 USD (78%)
```

| | `usage` | `windows` / `credits` |
|---|---|---|
| Source | status line payload | Anthropic's usage API |
| Credit pool | not reported | `credits` reports it |
| 5-hour and weekly windows | yes | yes, same figures |
| HTTP request | none | one per 5 minutes, can be rate-limited |
| Reads credentials | no | yes |

Each renders nothing when it does not apply, so `usage` and `credits` can both sit in one
format string across plans — on any given account exactly one of them has data. That is the
setup to reach for, and the one to keep if you switch between accounts, because switching then
needs no config change:

```toml
format = "$directory | $git_branch | $model | $context | $usage$credits"

[credits]
disabled = false
```

`windows` is the same figures as `usage` from the API instead of the payload; enable it only
if you want the account's own accounting:

```toml
[windows]
disabled = false
```

### Which account it reads

Claude Code isolates credentials per config directory, and keeps the live token in the macOS
Keychain rather than on disk:

| `CLAUDE_CONFIG_DIR` | Keychain service | File fallback |
|---|---|---|
| unset | `Claude Code-credentials` | `~/.claude/.credentials.json` |
| set | `Claude Code-credentials-<sha256(dir)[:8]>` | `$CLAUDE_CONFIG_DIR/.credentials.json` |

The Keychain is read first and the file second, and a candidate whose `expiresAt` has passed
is skipped in favour of a live one. Claude Code writes refreshed tokens to the Keychain, so a
`.credentials.json` can sit expired for days while the session it belongs to works fine —
reading only the file is how a budget segment goes blank with no explanation.

Because `CLAUDE_CONFIG_DIR` is exported into the status line process, this needs no knowledge
of whatever manages those directories. Tools that switch between several Claude accounts work
by pointing that variable at a profile, so following it follows the switch: an Enterprise
session shows the Enterprise budget and a Pro session shows Pro's windows, at the same time,
in two terminals, with one config file. Cached readings are kept per account for the same
reason — one shared cache would let whichever session refreshed last overwrite the other.

The HTTP request never happens while rendering. Both modules read a cached reading (5 minutes)
and, when it is stale, spawn a detached background process that refreshes it and outlives the
render. The cache is
`~/.local/state/claude-statusline/accounts/<config-dir-hash>/usage.json`; `claude-statusline
refresh-usage` forces a refresh in the foreground and prints why one failed, which is the way
to tell an expired login from a rate-limited endpoint.

When a refresh fails, the next one waits: Anthropic rate-limits this endpoint and says for how
long (`Retry-After`), and that is honoured, capped at an hour. One success ends the wait, and
`claude-statusline refresh-usage` ignores it — an explicit request is not a retry storm.

A reading older than 30 minutes — usually an expired login — is marked `⚠︎` rather than shown
as current, by both modules. Both expose it as `{{.Stale}}`, so a custom format can place it;
leave `{{.Stale}}` out of a `windows` format and the marker is still appended at the end, so
the warning cannot be lost by accident.

`windows` template fields:

| Field | Description |
|-------|-------------|
| `{{.Name}}` | Window label (`5h`, `wk`) |
| `{{.Pct}}` | Usage (0-100) |
| `{{.Bar}}` | Progress bar |
| `{{.Resets}}` | When it resets (clock time today, weekday otherwise) |
| `{{.Stale}}` | `⚠︎` when the reading is old, empty otherwise |

`credits` template fields:

| Field | Description |
|-------|-------------|
| `{{.Used}}` | Spend so far, in major units |
| `{{.Limit}}` | The ceiling |
| `{{.Pct}}` | Percentage of the pool used |
| `{{.Bar}}` | Progress bar |
| `{{.Currency}}` | Currency code |
| `{{.Stale}}` | `⚠︎` when the reading is old, empty otherwise |

### Usage module

Shows your plan usage: the 5-hour rolling window and the 7-day one. On by default, in every
preset. To turn it off:

```toml
[usage]
disabled = true
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

## Clickable hyperlinks (OSC 8)

Modules can wrap their output in [OSC 8 terminal hyperlinks](https://gist.github.com/egmontkob/eb114294efbcd5adb1944c9f3cb5feda), making text clickable in supported terminals.

### git_branch

When enabled, the branch name becomes a clickable link to the branch page on the remote. The base URL is auto-detected from `git remote get-url origin`, and the branch path pattern is selected based on the host:

- **GitHub** (default): `/tree/<branch>`
- **GitLab** (hosts containing "gitlab"): `/-/tree/<branch>`
- **Bitbucket** (hosts containing "bitbucket"): `/src/<branch>`

Branch names are percent-encoded so characters like `#` don't break the URL.

```toml
[git_branch]
hyperlink = true
# hyperlink_base_url = "https://github.com/owner/repo"  # override auto-detection
```

### directory

When enabled, the directory text links to the path using a configurable URL template. The default opens `file://` URLs with properly encoded paths; set `hyperlink_url_template` for VS Code or other editors.

Template fields:
- `{{.AbsPathEncoded}}` — percent-encoded absolute path (use for URLs)
- `{{.AbsPath}}` — raw absolute path (use for schemes that handle raw paths, like `vscode://`)

```toml
[directory]
hyperlink = true
# hyperlink_url_template = "file://{{.AbsPathEncoded}}"  # default
# hyperlink_url_template = "vscode://file{{.AbsPath}}"   # open in VS Code
```

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

MIT — see [LICENSE](LICENSE). Originally written by
[Felipe Philipp](https://github.com/felipeelias); maintained here by
[Frank Bernhardt](https://github.com/frank-bee).
