# claude-statusline

A configurable status line for Claude Code, written in Go. Claude Code pipes a JSON payload on
stdin; this renders one line of ANSI-styled text on stdout.

A fork of [felipeelias/claude-statusline](https://github.com/felipeelias/claude-statusline)
(MIT). Upstream stays the source of truth for everything except the `windows` and `credits`
modules and the fixes listed in `git log upstream/main..main`.

## Layout

```
main.go              wiring only
internal/input/      the JSON Claude Code sends on stdin
internal/config/     TOML config, defaults, and the presets
internal/render/     parses the format string, resolves $tokens
internal/modules/    one file per module, the bulk of the code
internal/style/      ANSI colours, hex parsing, named attributes
internal/anthropic/  the usage API client and its file cache
internal/cli/        commands: prompt, init, test, themes, refresh-usage
```

A module implements `Name()` and `Render()`; register it in `internal/render` and give it a
config struct with defaults in `internal/config`. Every module has a `_test.go` beside it.

## Paths

| | |
|---|---|
| Config | `~/.config/claude-statusline/config.toml` |
| Cache | `~/.local/state/claude-statusline/usage.json` |

Identical to upstream on purpose — this is a drop-in replacement, so a user swapping between
the two keeps their config.

## Working on it

```bash
go build ./... && go test ./... && golangci-lint run
claude-statusline test      # render with mock data
claude-statusline themes    # preview every preset
```

Tests must not touch the developer's real `$HOME`: `main_test.go` and
`internal/cli/cli_test.go` point `HOME`, `XDG_CONFIG_HOME` and `XDG_STATE_HOME` at
`t.TempDir()`. Keep that invariant when adding tests that read config or cache.

`internal/anthropic` spawns a detached refresh process. It must never re-exec itself during a
test run — see the guard in `usage.go` and `export_test.go`.

## Contributing back upstream

Changes that are not fork-specific go upstream: open an issue on
`felipeelias/claude-statusline`, then a PR from a branch cut at `upstream/main`, so the diff
carries none of this fork's renames.

```bash
git checkout -b fix/thing upstream/main
gh pr create --repo felipeelias/claude-statusline --base main
```

`gh` defaults to the parent repository for a fork, which is what you want here — but pass
`--repo frank-bee/claude-statusline` explicitly when you mean this one.

## Releases

release-please raises the release PR from conventional commits; merging it tags, and goreleaser
builds the archives and pushes the formula to `frank-bee/homebrew-tap`. That push needs the
repository secret `HOMEBREW_TAP_TOKEN` — a fine-grained PAT with `contents: write` on the tap
repository and nothing else.
