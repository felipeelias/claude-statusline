# Roadmap

Feature ideas for claude-statusline. Each section describes the feature in detail, including what it does, how it should work from a user's perspective, the config surface, and implementation notes.

## Priority: High

### 1. Git status indicators

The current `git_branch` module only shows the branch name and a worktree indicator. It should show richer git state: dirty/clean status, ahead/behind remote counts, and staged/unstaged/untracked file counts.

**New template fields:**

| Field            | Type | Description                                                                |
| ---------------- | ---- | -------------------------------------------------------------------------- |
| `{{.Staged}}`    | int  | Number of staged files                                                     |
| `{{.Modified}}`  | int  | Number of modified (unstaged) files                                        |
| `{{.Untracked}}` | int  | Number of untracked files                                                  |
| `{{.Ahead}}`     | int  | Commits ahead of upstream                                                  |
| `{{.Behind}}`    | int  | Commits behind upstream                                                    |
| `{{.IsDirty}}`   | bool | True if there are any uncommitted changes (staged, modified, or untracked) |
| `{{.IsClean}}`   | bool | Inverse of IsDirty                                                         |
| `{{.Conflicts}}` | int  | Number of files with merge conflicts                                       |

**Default format (updated):**

```
 {{.Branch}}{{if .InWorktree}} {{end}}{{if .IsDirty}} *{{end}}{{if .Ahead}} {{.Ahead}}{{end}}{{if .Behind}} {{.Behind}}{{end}}
```

The default shows a `*` when dirty and up/down arrows for ahead/behind. Users who want verbose output can customize to show counts:

```toml
[git_branch]
format = ' {{.Branch}}{{if .Staged}} +{{.Staged}}{{end}}{{if .Modified}} !{{.Modified}}{{end}}{{if .Untracked}} ?{{.Untracked}}{{end}}'
```

**Implementation:**

Run `git status --porcelain=v2 --branch` in the working directory. This single command provides:

- Branch name (from `# branch.head`)
- Ahead/behind (from `# branch.ab +N -M`)
- File statuses (lines starting with `1`, `2`, `u`, `?`)

Parse the porcelain v2 output to populate all fields. This replaces the current `git rev-parse --abbrev-ref HEAD` call, so branch detection and status come from one command instead of two.

**Rename consideration:**

The module name stays `git_branch` and config key stays `[git_branch]` for backwards compatibility — even though it now shows more than the branch. The module struct can be renamed internally if needed.

---

### 3. Context bar display styles

The context module currently renders a progress bar using configurable `bar_fill` and `bar_empty` characters. Add a `bar_style` config field that selects from named presets, making it easy to switch visual styles without manually setting fill/empty characters.

**New config field:**

```toml
[context]
bar_style = "blocks"
```

**Built-in bar styles:**

| Name      | Fill | Empty | Example (60%)             |
| --------- | ---- | ----- | ------------------------- |
| `classic` | `█`  | `░`   | `███░░` (current default) |
| `blocks`  | `█`  | `▒`   | `███▒▒`                   |
| `dots`    | `⣿`  | `⣀`   | `⣿⣿⣿⣀⣀`                   |
| `line`    | `━`  | `─`   | `━━━──`                   |
| `squares` | `◼`  | `◻`   | `◼◼◼◻◻`                   |

**Behavior:**

- When `bar_style` is set, it overrides `bar_fill` and `bar_empty`.
- When `bar_fill` or `bar_empty` are explicitly set alongside `bar_style`, the explicit values win (user overrides take priority).
- Default: no `bar_style` set, uses `bar_fill`/`bar_empty` directly (preserves current behavior).

---

### 4. Directory display modes

The directory module currently does tilde substitution and first-character truncation of path segments beyond `truncation_length`. Add a `display_mode` config field that controls the truncation strategy.

**New config field:**

```toml
[directory]
display_mode = "truncate"
```

**Display modes:**

| Mode       | Description                                                    | Example (`~/code/projects/claude-statusline`) |
| ---------- | -------------------------------------------------------------- | --------------------------------------------- |
| `full`     | No truncation, only tilde substitution                         | `~/code/projects/claude-statusline`           |
| `truncate` | Current behavior: abbreviate early segments to first character | `~/c/p/claude-statusline`                     |
| `basename` | Only the last path segment                                     | `claude-statusline`                           |

**Behavior:**

- Default: `truncate` (preserves current behavior).
- `truncation_length` is only relevant for `truncate` mode. In `full` and `basename` modes it is ignored.
- Tilde substitution (`~` for `$HOME`) applies in all modes.

---

### 5. Custom command module

A new `custom_command` module that runs an arbitrary shell command and displays its output. The full Claude Code JSON payload is piped to the command via stdin, so the command can extract any field.

**Config:**

```toml
[custom_command]
disabled = false
command = "echo hello"
style = "dim"
timeout_ms = 500
```

| Field        | Type   | Default | Description                                   |
| ------------ | ------ | ------- | --------------------------------------------- |
| `command`    | string | `""`    | Shell command to run (via `sh -c`)            |
| `style`      | string | `""`    | ANSI style for the output                     |
| `disabled`   | bool   | `true`  | Disabled by default                           |
| `timeout_ms` | int    | `500`   | Kill the command if it takes longer than this |

**Behavior:**

- The module is referenced as `$custom_command` in the format string.
- The command receives the raw JSON on stdin and should write a single line to stdout.
- Trailing newlines are stripped.
- If the command fails, times out, or produces no output, the module renders empty (hidden).
- ANSI escape codes in command output are preserved (not stripped).
- The command runs with `cwd` set to the working directory from the JSON payload.

**Example: show Kubernetes context:**

```toml
[custom_command]
disabled = false
command = "kubectl config current-context 2>/dev/null"
style = "blue"
```

**Example: extract a custom field from the JSON payload:**

```toml
[custom_command]
disabled = false
command = "jq -r '.session_id[:8]'"
style = "dim"
```

---

### 6. Version module

A new `version` module that displays the Claude Code version string from the JSON payload. The `version` field already exists in `input.Data`.

**Config:**

```toml
[version]
format = "v{{.Version}}"
style = "dim"
disabled = true
```

| Field      | Type   | Default           | Description                     |
| ---------- | ------ | ----------------- | ------------------------------- |
| `format`   | string | `"v{{.Version}}"` | Go template with `{{.Version}}` |
| `style`    | string | `"dim"`           | ANSI style                      |
| `disabled` | bool   | `true`            | Disabled by default             |

**Template fields:**

| Field          | Type   | Description                                 |
| -------------- | ------ | ------------------------------------------- |
| `{{.Version}}` | string | Claude Code version string (e.g., `1.0.33`) |

**Behavior:**

- If the version string is empty, renders empty.
- Referenced as `$version` in the format string.

---

### 7. Model name formatting options

The model module currently exposes `{{.DisplayName}}` from the JSON payload. Add `{{.ID}}` (the raw model ID) and a `{{.Short}}` field that extracts a compact name from the model ID.

**New template fields on model module:**

| Field              | Type   | Description                                       |
| ------------------ | ------ | ------------------------------------------------- |
| `{{.ID}}`          | string | Raw model ID (e.g., `claude-sonnet-4-6-20250514`) |
| `{{.Short}}`       | string | Compact extracted name (e.g., `Sonnet 4.6`)       |
| `{{.DisplayName}}` | string | Display name from Claude Code (existing)          |

**Short name extraction:**

Parse the model ID with a regex pattern like `claude-(opus|sonnet|haiku)-(\d+)-(\d+)` to extract family and version. Map to `"Family X.Y"` format:

- `claude-sonnet-4-6-20250514` -> `Sonnet 4.6`
- `claude-opus-4-6-20250514` -> `Opus 4.6`
- `claude-haiku-4-5-20251001` -> `Haiku 4.5`

If the regex doesn't match (unknown model), fall back to `DisplayName`.

**Default format:** Keep `{{.DisplayName}}` as default. Users who prefer the compact name use:

```toml
[model]
format = "{{.Short}}"
```

---

### 8. Charset toggle (Nerd Font / text fallback)

Add a top-level `charset` config field that controls whether icon characters use Nerd Font glyphs or plain ASCII/text fallbacks. This affects the default format templates for modules that use icons (git_branch, powerline separators).

**Config:**

```toml
charset = "nerd-font"
```

**Values:**

| Value       | Description                             |
| ----------- | --------------------------------------- |
| `nerd-font` | Use Nerd Font glyphs (current behavior) |
| `text`      | Use ASCII/text fallbacks                |

**Icon mapping:**

| Icon                  | Nerd Font   | Text    |
| --------------------- | ----------- | ------- |
| Git branch            | `` (U+E0A0) | (empty) |
| Worktree              | `` (U+F0E8) | `[wt]`  |
| Powerline right arrow | `` (U+E0B0) | `>`     |
| Powerline left cap    | `` (U+E0B6) | `(`     |
| Powerline right cap   | `` (U+E0B4) | `)`     |

**Behavior:**

- Default: `nerd-font` (preserves current behavior).
- The `charset` field only affects the **default** format templates generated by presets. If a user explicitly sets a module's `format` field, their custom format is used as-is regardless of `charset`.
- Presets should resolve icon glyphs at config-build time based on the `charset` value.

---

### 9. Output style module

A new `output_style` module that shows Claude Code's current output style. The `output_style.name` field is confirmed present in the JSON payload.

**Config:**

```toml
[output_style]
format = "{{.Name}}"
style = "dim"
disabled = true
```

| Field      | Type   | Default         | Description         |
| ---------- | ------ | --------------- | ------------------- |
| `format`   | string | `"{{.Name}}"` | Go template         |
| `style`    | string | `"dim"`         | ANSI style          |
| `disabled` | bool   | `true`          | Disabled by default |

**Template fields:**

| Field       | Type   | Description                                        |
| ----------- | ------ | -------------------------------------------------- |
| `{{.Name}}` | string | Output style name (e.g., `default`, `Explanatory`) |

**Behavior:**

- Referenced as `$output_style` in the format string.
- If `output_style` is absent from the JSON or `name` is empty, renders empty.

---

### 10. Clickable hyperlinks (OSC 8)

Add OSC 8 terminal hyperlink support to modules where it makes sense:

- `git_branch`: link to the GitHub branch page
- `directory`: link to open the directory in an editor

OSC 8 format: `\033]8;;URL\033\\TEXT\033]8;;\033\\`

**Config:**

```toml
[git_branch]
hyperlink = true
hyperlink_base_url = ""  # auto-detected from git remote
```

```toml
[directory]
hyperlink = true
hyperlink_url_template = "vscode://file{{.AbsPath}}"
```

**git_branch hyperlink:**

- When `hyperlink = true`, wrap the branch name in an OSC 8 link.
- Auto-detect the base URL from `git remote get-url origin`, converting SSH URLs to HTTPS and appending `/tree/{branch}`.
- `hyperlink_base_url` allows manual override if the remote URL detection doesn't work (e.g., private GitLab instances).
- If no remote URL can be determined and no override is set, render the branch name without a link (graceful degradation).

**directory hyperlink:**

- When `hyperlink = true`, wrap the directory text in an OSC 8 link.
- `hyperlink_url_template` is a Go template with `{{.AbsPath}}` available, defaulting to `file://{{.AbsPath}}`.
- Users can set it to `vscode://file{{.AbsPath}}` to open in VS Code.

**Behavior:**

- Default: `hyperlink = false` on both modules (opt-in).
- Terminals that don't support OSC 8 will simply display the text without the link (the escape sequences are invisible in unsupported terminals).

---

### 11. Session name module

A new `session_name` module that shows the session's custom title (set via `/rename` in Claude Code). The session name is stored in the transcript JSONL file as a `custom-title` entry.

**How session names are stored:**

Claude Code stores session names in the transcript file (path available via `transcript_path` in the JSON payload) as a JSONL entry:

```json
{"type": "custom-title", "customTitle": "session check", "sessionId": "d2576725-..."}
```

**Config:**

```toml
[session_name]
format = "{{.Name}}"
style = "bold"
disabled = true
```

| Field      | Type   | Default        | Description         |
| ---------- | ------ | -------------- | ------------------- |
| `format`   | string | `"{{.Name}}"` | Go template         |
| `style`    | string | `"bold"`       | ANSI style          |
| `disabled` | bool   | `true`         | Disabled by default |

**Template fields:**

| Field       | Type   | Description                                    |
| ----------- | ------ | ---------------------------------------------- |
| `{{.Name}}` | string | Custom session title (e.g., `"session check"`) |

**Behavior:**

- Referenced as `$session_name` in the format string.
- Reads the transcript file at `transcript_path` to find the last `custom-title` entry.
- If no `transcript_path` is provided or no `custom-title` entry exists, renders empty.
- The transcript file is read on every render. Since status lines update after each assistant message (debounced at 300ms), this is acceptable. The file is local and the scan stops at the last match.

**Implementation:**

1. Parse `transcript_path` from the JSON payload (requires the prerequisite input expansion).
2. Read the transcript JSONL file, scanning for `{"type": "custom-title"}` entries.
3. Use the `customTitle` from the last such entry.
4. To keep it fast: read the file from the end (or scan all lines since the file is append-only and titles are rare — typically 0-1 entries per session).

---

### 12. Vim mode module

A new `vim_mode` module that shows the current vim editor mode when vim mode is enabled in Claude Code. The `vim.mode` field is in the JSON payload.

**Config:**

```toml
[vim_mode]
format = "{{.Mode}}"
style = "bold yellow"
disabled = true
```

| Field      | Type   | Default         | Description         |
| ---------- | ------ | --------------- | ------------------- |
| `format`   | string | `"{{.Mode}}"`  | Go template         |
| `style`    | string | `"bold yellow"` | ANSI style          |
| `disabled` | bool   | `true`          | Disabled by default |

**Template fields:**

| Field       | Type   | Description                      |
| ----------- | ------ | -------------------------------- |
| `{{.Mode}}` | string | Vim mode: `NORMAL` or `INSERT`   |

**Behavior:**

- Referenced as `$vim_mode` in the format string.
- If `vim` is absent from the JSON (vim mode not enabled), renders empty.
- Useful for users who enable vim mode and want a persistent mode indicator.

---

### 13. Agent name module

A new `agent_name` module that shows the agent name when running with `--agent` or agent settings. The `agent.name` field is in the JSON payload.

**Config:**

```toml
[agent_name]
format = "{{.Name}}"
style = "bold magenta"
disabled = true
```

| Field      | Type   | Default          | Description         |
| ---------- | ------ | ---------------- | ------------------- |
| `format`   | string | `"{{.Name}}"`   | Go template         |
| `style`    | string | `"bold magenta"` | ANSI style          |
| `disabled` | bool   | `true`           | Disabled by default |

**Template fields:**

| Field       | Type   | Description                                   |
| ----------- | ------ | --------------------------------------------- |
| `{{.Name}}` | string | Agent name (e.g., `"security-reviewer"`)      |

**Behavior:**

- Referenced as `$agent_name` in the format string.
- If `agent` is absent from the JSON (not running as agent), renders empty.
- Useful for users running named agents who want to identify which agent is active.

---

### 14. Token counts module

A new `tokens` module that shows token usage statistics from the JSON payload. Exposes cumulative totals, current context usage, and cache metrics.

**Config:**

```toml
[tokens]
format = "{{.TotalInput}}in {{.TotalOutput}}out"
style = "dim"
disabled = true
```

| Field      | Type   | Default                                      | Description         |
| ---------- | ------ | -------------------------------------------- | ------------------- |
| `format`   | string | `"{{.TotalInput}}in {{.TotalOutput}}out"` | Go template         |
| `style`    | string | `"dim"`                                      | ANSI style          |
| `disabled` | bool   | `true`                                       | Disabled by default |

**Template fields:**

| Field                    | Type   | Description                                                 |
| ------------------------ | ------ | ----------------------------------------------------------- |
| `{{.TotalInput}}`        | string | Cumulative input tokens, human-readable (e.g., `15.2k`)    |
| `{{.TotalOutput}}`       | string | Cumulative output tokens, human-readable (e.g., `4.5k`)    |
| `{{.TotalInputRaw}}`     | int    | Cumulative input tokens, raw number                         |
| `{{.TotalOutputRaw}}`    | int    | Cumulative output tokens, raw number                        |
| `{{.CacheRead}}`         | string | Cache read tokens from current usage, human-readable        |
| `{{.CacheCreation}}`     | string | Cache creation tokens from current usage, human-readable    |
| `{{.CacheReadRaw}}`      | int    | Cache read tokens, raw number                               |
| `{{.CacheCreationRaw}}`  | int    | Cache creation tokens, raw number                           |
| `{{.CacheHitPct}}`       | float64| Cache hit percentage: `cache_read / (cache_read + cache_creation) * 100` |
| `{{.ContextSize}}`       | string | Context window size, human-readable (e.g., `200k`)         |

Human-readable formatting: `1234` -> `1.2k`, `1234567` -> `1.2M`.

**Example configs:**

```toml
# Show cache efficiency
[tokens]
disabled = false
format = "cache: {{printf \"%.0f\" .CacheHitPct}}%"

# Show full breakdown
[tokens]
disabled = false
format = "{{.TotalInput}}in {{.TotalOutput}}out | cache {{.CacheRead}}r {{.CacheCreation}}w"
```

**Behavior:**

- Referenced as `$tokens` in the format string.
- If `current_usage` is null (before first API call), cache fields render as 0.
- `CacheHitPct` is 0 when both cache_read and cache_creation are 0.

---

### 15. Worktree details module

Expand the current worktree support from a simple boolean indicator on `git_branch` to a dedicated `worktree` module with full details. The JSON payload includes `worktree.name`, `worktree.path`, `worktree.branch`, `worktree.original_cwd`, and `worktree.original_branch`.

**Config:**

```toml
[worktree]
format = "{{.Name}} (from {{.OriginalBranch}})"
style = "cyan"
disabled = true
```

| Field      | Type   | Default                                       | Description         |
| ---------- | ------ | --------------------------------------------- | ------------------- |
| `format`   | string | `"{{.Name}} (from {{.OriginalBranch}})"` | Go template         |
| `style`    | string | `"cyan"`                                      | ANSI style          |
| `disabled` | bool   | `true`                                        | Disabled by default |

**Template fields:**

| Field                  | Type   | Description                                                        |
| ---------------------- | ------ | ------------------------------------------------------------------ |
| `{{.Name}}`            | string | Worktree name (e.g., `"my-feature"`)                              |
| `{{.Path}}`            | string | Absolute path to worktree directory                                |
| `{{.Branch}}`          | string | Git branch in the worktree (e.g., `"worktree-my-feature"`)        |
| `{{.OriginalCwd}}`     | string | Directory before entering the worktree                             |
| `{{.OriginalBranch}}`  | string | Branch checked out before entering the worktree                    |

**Behavior:**

- Referenced as `$worktree` in the format string.
- If `worktree` is absent from the JSON (not in a worktree session), renders empty.
- `Branch` and `OriginalBranch` may be empty for hook-based worktrees.
- The existing `{{.InWorktree}}` boolean on `git_branch` continues to work for simple use cases. This module is for users who want detailed worktree context.

---

### 16. PR links module

A new `pr` module that shows PRs created during the current session. PR data is stored in the transcript JSONL file as `pr-link` entries.

**How PR links are stored:**

In the transcript file (path via `transcript_path`):

```json
{
  "type": "pr-link",
  "sessionId": "d2576725-...",
  "prNumber": 10,
  "prUrl": "https://github.com/felipeelias/claude-statusline/pull/10",
  "prRepository": "felipeelias/claude-statusline",
  "timestamp": "2026-03-17T09:35:35.070Z"
}
```

**Config:**

```toml
[pr]
format = "{{.Count}} PRs"
style = "green"
disabled = true
```

| Field      | Type   | Default             | Description         |
| ---------- | ------ | ------------------- | ------------------- |
| `format`   | string | `"{{.Count}} PRs"` | Go template         |
| `style`    | string | `"green"`           | ANSI style          |
| `disabled` | bool   | `true`              | Disabled by default |
| `hyperlink`| bool   | `false`             | Wrap last PR in OSC 8 link |

**Template fields:**

| Field              | Type   | Description                                                   |
| ------------------ | ------ | ------------------------------------------------------------- |
| `{{.Count}}`       | int    | Number of PRs created in this session                         |
| `{{.LastNumber}}`  | int    | PR number of the most recent PR (e.g., `10`)                 |
| `{{.LastURL}}`     | string | Full URL of the most recent PR                                |
| `{{.LastRepo}}`    | string | Repository of the most recent PR (e.g., `owner/repo`)       |

**Example configs:**

```toml
# Show last PR as clickable link
[pr]
disabled = false
format = "#{{.LastNumber}}"
style = "green"
hyperlink = true

# Show count
[pr]
disabled = false
format = "{{if .Count}}{{.Count}} PRs{{end}}"
```

**Behavior:**

- Referenced as `$pr` in the format string.
- Reads `transcript_path` and scans for `pr-link` entries.
- When `hyperlink = true`, wraps the rendered text in an OSC 8 link pointing to `LastURL`.
- If no PRs exist in the session, renders empty.

---

### 17. Project directory module

A new `project_dir` module that shows the project directory where Claude Code was launched. This differs from `cwd` when the working directory changes during a session (e.g., via `cd` or worktree).

**Config:**

```toml
[project_dir]
format = "{{.Dir}}"
style = "dim"
disabled = true
```

| Field      | Type   | Default        | Description         |
| ---------- | ------ | -------------- | ------------------- |
| `format`   | string | `"{{.Dir}}"` | Go template         |
| `style`    | string | `"dim"`        | ANSI style          |
| `disabled` | bool   | `true`         | Disabled by default |

**Template fields:**

| Field              | Type   | Description                                                    |
| ------------------ | ------ | -------------------------------------------------------------- |
| `{{.Dir}}`         | string | Project directory, tilde-collapsed (same logic as directory)  |
| `{{.AbsPath}}`     | string | Full absolute path                                             |
| `{{.DiffersFromCwd}}` | bool | True when project_dir != cwd                                |

**Behavior:**

- Referenced as `$project_dir` in the format string.
- Uses `workspace.project_dir` from the JSON payload.
- Applies the same tilde-substitution as the `directory` module.
- Most useful in combination with conditional templates: `{{if .DiffersFromCwd}}proj: {{.Dir}}{{end}}`.

---

## Priority: Medium

### 18. Timeout for git subprocess calls

The `git_branch` module runs `git status --porcelain=v2 --branch` (detailed mode) or `git rev-parse --abbrev-ref HEAD` (simple mode) without a timeout. On network-mounted repos or hung git processes, this could block the status line indefinitely.

**Change:**

Replace `exec.Command` with `exec.CommandContext` using a `context.WithTimeout` (e.g., 5 seconds). Apply to both `gitBranchSimple` and `gitStatusDetailed` in `internal/modules/gitbranch.go`. On timeout, return empty output (same as a non-git directory).

---

### 19. Multi-line layout

Allow the format string to define multiple lines using `\n` as a line separator. Claude Code's status line natively supports multiple output lines — each `echo`/line in the output becomes a separate row.

**Config:**

```toml
format = "$directory | $git_branch | $model\n$cost | $context | $session_timer"
```

This renders two status lines:

```
Line 1: ~/c/p/claude-statusline |  main | Sonnet 4.6
Line 2: $0.42 | ███░░ 60% | 05m23s
```

**Behavior:**

- The format string is split on literal `\n` sequences.
- Each line is rendered independently using the same module rendering pipeline.
- Empty lines (where all modules in a line render empty) are omitted.
- Claude Code receives the multi-line output and displays each line as a separate row. This is confirmed supported by Claude Code docs.

---

### 20. Flex separator

A special token `$fill` in the format string that expands to fill available terminal width, enabling right-aligned segments.

**Config:**

```toml
format = "$directory | $git_branch $fill $cost | $context"
```

This would render:

```
~/c/p/claude-statusline |  main                    $0.42 | ███░░ 60%
```

**Behavior:**

- `$fill` is replaced by spaces to fill the remaining terminal width.
- Terminal width is obtained from the `COLUMNS` environment variable, or defaults to 80 if not set.
- If the content on both sides of `$fill` already exceeds the terminal width, `$fill` renders as a single space.
- Only one `$fill` per line is supported. If multiple `$fill` tokens appear, only the first is expanded; the rest render as a single space.
- Computing the fill width requires knowing the visible width of the rendered text (excluding ANSI escape codes). Use a function that strips ANSI sequences before measuring string length.

---

### 21. Message count module

A new `messages` module that shows the number of user and assistant messages in the current session. Reads from the transcript JSONL file.

**Config:**

```toml
[messages]
format = "{{.Total}} msgs"
style = "dim"
disabled = true
```

| Field      | Type   | Default             | Description         |
| ---------- | ------ | ------------------- | ------------------- |
| `format`   | string | `"{{.Total}} msgs"` | Go template        |
| `style`    | string | `"dim"`             | ANSI style          |
| `disabled` | bool   | `true`              | Disabled by default |

**Template fields:**

| Field            | Type | Description                     |
| ---------------- | ---- | ------------------------------- |
| `{{.Total}}`     | int  | Total messages (user+assistant) |
| `{{.User}}`      | int  | Number of user messages         |
| `{{.Assistant}}`  | int  | Number of assistant messages    |

**Behavior:**

- Referenced as `$messages` in the format string.
- Reads `transcript_path` and counts `user` and `assistant` type entries.
- If no transcript path or file doesn't exist, renders empty.

---

### 22. Exceeds 200k indicator

Add a `{{.Exceeds200k}}` boolean field to the context module that is true when `exceeds_200k_tokens` is true in the JSON payload. This warns when the last API response exceeded 200k total tokens.

**New template field on context module:**

| Field              | Type | Description                                       |
| ------------------ | ---- | ------------------------------------------------- |
| `{{.Exceeds200k}}` | bool | True when last response exceeded 200k total tokens |

**Example config:**

```toml
[context]
format = '{{.Bar}} {{printf "%.0f" .UsedPct}}%{{if .Exceeds200k}} LARGE{{end}}'
```

**Default format:** Unchanged. The field is available for users who want it.

---

## Priority: Low

### 23. Usage API integration

Query the Anthropic OAuth API (`api.anthropic.com/api/oauth/usage`) to show real-time 5-hour block and 7-day usage percentages.

**Config:**

```toml
[usage]
disabled = true
display = "block"  # "block" (5-hour), "weekly", or "both"
style = "dim"
cache_ttl_seconds = 300
```

**Behavior:**

- Read OAuth credentials from `~/.claude/.credentials.json` (or macOS Keychain on macOS).
- Cache responses locally (at `~/.cache/claude-statusline/usage.json`) to avoid hitting the API on every render.
- Show usage as a percentage (e.g., `72%` of 5-hour block used).
- Gracefully degrade: if credentials are missing or the API is unreachable, render empty.

**Why lower priority:**

- Requires HTTP calls (adds latency, even with caching).
- Requires reading credentials from the filesystem or OS keychain.
- Adds complexity (caching, error handling, credential discovery).
- The data is useful but not essential for most users.

---

### 24. Skills / hooks tracking

Track which Claude Code tools/skills are invoked during a session by integrating with Claude Code hooks.

**Config:**

```toml
[skills]
disabled = true
display = "last"  # "last", "count", or "list"
style = "dim"
```

**Behavior:**

- Register a Claude Code hook (`PostToolUse`) that writes tool invocations to a session-scoped file.
- The module reads the session file and displays the data based on `display` mode:
  - `last`: show the most recently used tool name
  - `count`: show total tool invocations count
  - `list`: show a deduplicated list of tool names used
- Session files are stored in `~/.cache/claude-statusline/skills/` and keyed by session ID.

**Why lower priority:**

- Requires Claude Code hook integration (separate setup step for users).
- Requires file-based state (session-scoped files on disk).
- The information is interesting but not actionable for most workflows.
