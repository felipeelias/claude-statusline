#!/bin/bash
# Renders assets/preview.png: the status line in a handful of presets, with
# fake data, so the README image can be regenerated rather than re-screenshotted.
#
# Needs: a built claude-statusline, ImageMagick, and a JetBrains Mono Nerd Font
#   brew install --cask font-jetbrains-mono-nerd-font
set -euo pipefail

here=$(cd "$(dirname "$0")/.." && pwd)
bin=${CLAUDE_STATUSLINE:-$here/claude-statusline}
[ -x "$bin" ] || bin=$(command -v claude-statusline)

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
mkdir -p "$work/state/claude-statusline" "$work/home/code"

# A fake repository, so $git_branch has something to show.
git init -q "$work/home/code/widget-api"
git -C "$work/home/code/widget-api" config user.email preview@example.com
git -C "$work/home/code/widget-api" config user.name Preview
git -C "$work/home/code/widget-api" checkout -q -b feat/checkout-flow
git -C "$work/home/code/widget-api" commit -q --allow-empty -m init

cat > "$work/state/claude-statusline/usage.json" <<'JSON'
{
  "limits": [
    {"kind": "session", "percent": 42, "severity": "normal"},
    {"kind": "weekly_all", "percent": 63, "severity": "warning"}
  ],
  "spend": {
    "enabled": true, "percent": 58,
    "used": {"amount_minor": 11600, "currency": "USD", "exponent": 2},
    "limit": {"amount_minor": 20000, "currency": "USD", "exponent": 2}
  }
}
JSON

cat > "$work/payload.json" <<JSON
{"cwd":"$work/home/code/widget-api",
 "workspace":{"current_dir":"$work/home/code/widget-api","project_dir":"$work/home/code/widget-api"},
 "model":{"display_name":"Claude Opus 5","id":"claude-opus-5"},
 "session_id":"preview","version":"2.1.272","transcript_path":"/tmp/t.jsonl",
 "cost":{"total_cost_usd":1.27},
 "context_window":{"used_percentage":68,"context_window_size":200000},
 "rate_limits":{"five_hour":{"used_percentage":42},"seven_day":{"used_percentage":63}}}
JSON

render() {
  cat > "$work/config.toml"
  (cd "$work/home/code/widget-api" &&
    HOME="$work/home" XDG_STATE_HOME="$work/state" \
      "$bin" -c "$work/config.toml" < "$work/payload.json")
  echo
}

{
render <<'TOML'
preset = "minimal"
format = "$directory  $git_branch  $model  $context  $usage"
[context]
format = '{{.Bar}} {{printf "%.0f" .UsedPct}}%'
bar_style = "line"
bar_width = 10
thresholds = [ { above = 50, style = "yellow" }, { above = 90, style = "red" } ]
[usage]
format = '{{.BlockBar}} 5h {{printf "%.0f" .BlockPct}}% · {{.WeeklyBar}} wk {{printf "%.0f" .WeeklyPct}}%'
style = "blue"
bar_style = "line"
bar_width = 10
TOML
for p in catppuccin gruvbox-rainbow; do
  render <<TOML
preset = "$p"
[context]
bar_style = "line"
[usage]
bar_style = "line"
TOML
done
} > "$work/preview.ansi"

python3 "$here/scripts/ansi2png.py" "$here/assets/preview.png" < "$work/preview.ansi"
echo "wrote assets/preview.png"
