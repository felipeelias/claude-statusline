# claude-statusline

A configurable status line for Claude Code, written in Go. Claude Code pipes a JSON payload on
stdin; this renders one line of ANSI-styled text on stdout.

A fork of [felipeelias/claude-statusline](https://github.com/felipeelias/claude-statusline)
(MIT). Upstream stays the source of truth for everything except the `windows` and `credits`
modules and the fixes listed in `git log upstream/main..main`.

`usage` is **on by default here**, in `Default()` and in every preset — upstream ships it off.
Two gates decide whether any module renders: `Disabled`, and whether the format string names it.
Setting one without the other renders nothing and looks like a broken module;
`TestEveryPresetShowsUsage` pins both for every preset.

**Which usage module to reach for.** Upstream's `usage` reads the `rate_limits` Claude Code
puts in the payload: live, no HTTP, nothing to rate-limit. Measured against a real payload,
`/api/oauth/usage` returns the *same* two windows, so `windows` is redundant with `usage`;
the finer API keys (`seven_day_opus`, `seven_day_sonnet`) come back `null`. Only `credits`
covers something the payload lacks — the credit pool on usage-based seats, and an account
without one (`spend.enabled: false`) renders nothing at all. Recommend `usage` by default and
treat the API path as the credit-pool case.

**The plan decides which one has data, and they are mutually exclusive.** On an Enterprise
usage-based seat there are no rate-limit windows at all: Claude Code sends no `rate_limits`
in the payload, and every window in `/api/oauth/usage` (`five_hour`, `seven_day` and the
fifteen others) comes back `null`. `usage` therefore renders nothing there, which looks like
a broken module rather than an empty one. On Pro the reverse holds — `spend.enabled` is
`false` and `credits` renders nothing.

So put **both** in the format string. Whichever plan the session is signed in as, one has
data and the other is silent, and switching accounts needs no config change.
`claude-statusline plan` prints which is live:

```
account       someone@example.com
organization  example-org
plan          enterprise
seat          enterprise_usage_based
meters        a money budget  ->  $credits
budget        776.24 / 993.00 USD (78%)
```

## Credentials

Claude Code isolates credentials **per config directory**, and the live token is in the
Keychain, not the file:

| `CLAUDE_CONFIG_DIR` | Keychain service | File |
|---|---|---|
| unset | `Claude Code-credentials` | `~/.claude/.credentials.json` |
| set | `Claude Code-credentials-<sha256(dir)[:8]>` | `$CLAUDE_CONFIG_DIR/.credentials.json` |

`internal/anthropic/credentials.go` resolves the Keychain item first and the file second,
skipping any candidate whose `expiresAt` has passed. The ordering is not arbitrary: Claude
Code writes refreshed access tokens to the Keychain, so a `.credentials.json` that is
symlinked or copied by an account switcher can sit expired for days while the session it
belongs to works fine. Reading `CLAUDE_CONFIG_DIR` from the environment — Claude Code exports
it into the status line child — is what keeps this independent of whatever manages those
directories: a switcher works by pointing that variable at a profile, so following it follows
the switch with no knowledge of the switcher.

The cache is keyed on the config directory for the same reason (`accounts/<hash>/` under the
state directory). One shared `usage.json` lets a Pro session — no credit pool, `spend.enabled`
false — silently blank the budget an Enterprise session is displaying. The key is the
directory rather than the resolved account because `Load` is on the render path, where
hashing an environment variable is free and reading the Keychain is a process spawn.

**Never let the suite reach real credentials.** Tests pin `CLAUDE_CONFIG_DIR` to `""` and call
`anthropic.SuppressKeychain(t)`. Without both, a test that sets `HOME` to a temporary
directory still resolves the *default* Keychain item and the developer's own live token, and
assertions pass or fail depending on whose machine they run on.

## Layout

```
main.go              wiring only
internal/input/      the JSON Claude Code sends on stdin
internal/config/     TOML config, defaults, and the presets
internal/render/     parses the format string, resolves $tokens
internal/modules/    one file per module, the bulk of the code
internal/style/      ANSI colours, hex parsing, named attributes
internal/anthropic/  the usage API client, credential resolution, file cache
internal/cli/        commands: prompt, init, test, themes, plan, refresh-usage
```

A module implements `Name()` and `Render()`; register it in `internal/render` and give it a
config struct with defaults in `internal/config`. Every module has a `_test.go` beside it.

## Paths

| | |
|---|---|
| Config | `~/.config/claude-statusline/config.toml` |
| Cache | `~/.local/state/claude-statusline/accounts/<config-dir-hash>/usage.json` |
| Profile cache | `~/.local/state/claude-statusline/accounts/<config-dir-hash>/profile.json` |

The config path is identical to upstream on purpose — this is a drop-in replacement, so a user
swapping between the two keeps their config. The cache is per-account rather than upstream's
flat `usage.json`, because two Claude Code sessions signed in as two different accounts would
otherwise overwrite each other's readings. Ask `anthropic.CachePath()` rather than rebuilding
the path; four places used to construct it by hand.

## Working on it

```bash
go build ./... && go test ./... && golangci-lint run
GOOS=windows go build ./...   # the detach path is per-platform; it breaks quietly here
claude-statusline test        # render with mock data
claude-statusline themes      # preview every preset, and the off-by-default modules
claude-statusline plan        # which account is live, and what its plan meters
```

**Check the linter actually ran.** A golangci-lint built against an older Go than the local
toolchain fails inside the standard library and reports typecheck noise instead of your code,
so `run` looks like it passed on nothing. If the output mentions `math/rand/v2` or files under
`libexec/src`, upgrade it before believing a clean run. CI pins its own version, so it will
catch what a stale local one waves through.

Tests must not touch the developer's real `$HOME`: `main_test.go` and
`internal/cli/cli_test.go` point `HOME`, `XDG_CONFIG_HOME` and `XDG_STATE_HOME` at
`t.TempDir()`. Keep that invariant when adding tests that read config or cache.

Two invariants in `internal/anthropic`:

- It spawns a **detached refresh process** by re-executing this binary, which must never happen
  during a test run — see the `testing.Testing()` guard in `usage.go`. `export_test.go` holds
  the seams (`SetBaseURL`, `SuppressRefresh`) that let tests avoid the real thing.
- A failed refresh **backs off** and renders respect it. Anthropic rate-limits the usage
  endpoint and answers 429 with a `Retry-After` of nearly an hour; the original code retried on
  every render whose cache had expired, which sustained the limit and froze the displayed
  figure for its whole duration. Do not reintroduce a retry-per-render path.

`themes` previews `windows` and `credits` by pointing `XDG_STATE_HOME` at a temporary
directory holding a mock reading, then restoring it. That keeps the preview on the real render
path with no preview-only seam in shipped code, and a cache written just now is fresh, so no
network refresh fires. `TestThemesLeavesTheRealCacheAlone` pins both halves: the user's reading
must survive, and the preview must not show it.

**Verifying a reading is right, not merely rendered.** That a module renders is no evidence its
number is correct — a frozen cache renders beautifully. Capture what Claude Code actually sends
and compare:

```bash
printf '#!/bin/sh\ntee /tmp/payload.json | claude-statusline\n' > /tmp/wrap.sh && chmod +x /tmp/wrap.sh
# point statusLine.command at /tmp/wrap.sh, wait for a render, then restore it
python3 -c "import json;print(json.load(open('/tmp/payload.json'))['rate_limits'])"
```

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

## The README image

`assets/preview.png` is generated, not screenshotted: `./scripts/render-preview.sh` renders the
status line in two presets against fake data (a throwaway repo, a mock usage cache) and pipes
the ANSI through `scripts/ansi2png.py`, which draws each coloured run at an exact cell position
with a Nerd Font. Regenerate it rather than taking a new screenshot, so the image never carries
a real path, branch or reading.

Needs ImageMagick and `brew install --cask font-jetbrains-mono-nerd-font`. Two traps worth
knowing: ImageMagick's SVG renderer ignores `font-family` and will silently typeset the whole
thing in a proportional italic, and `-draw text` drops glyphs that `-annotate` renders fine.
Checking whether a glyph exists with `label:` is useless - it falls back to another font and
reports success for a glyph the chosen font does not have.

## Releases

Fully automated. release-please raises the release PR from conventional commits; merging it
tags, and goreleaser builds the archives and pushes the formula into
`frank-bee/homebrew-tap` under `Formula/`.

The cross-repository push is authenticated by the **`claude-statusline-tap-publisher` GitHub
App** (app id `4949256`, installed on `homebrew-tap` alone with `contents: write` and
`metadata: read`). `actions/create-github-app-token` mints a token scoped to that one
repository, and it expires an hour later. `GITHUB_TOKEN` cannot write to another repository,
and a PAT would carry the whole account for the sake of one file.

Repository config this depends on:

| | |
|---|---|
| `TAP_APP_ID` | variable — the App's id |
| `TAP_APP_PRIVATE_KEY` | secret — the App's private key |
| Settings → Actions → Workflow permissions | "Allow GitHub Actions to create and approve pull requests" must be **on**, or release-please cannot open its PR |

If the formula push 403s, check the App's installation still covers `homebrew-tap` — narrowing
it to the wrong repository is the easy mistake, and the error says only "Resource not
accessible by integration".

`brew update` does not always pull a freshly changed tap. When a just-published formula is "not
found", run `git -C $(brew --repo frank-bee/tap) pull`.
