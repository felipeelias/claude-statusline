package anthropic

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// What the plan is, and why the tool has to ask rather than assume.
//
// GET /api/oauth/profile says which kind of subscription the current token
// belongs to. This matters because the two things worth putting in a status
// line are metered differently and are mutually exclusive:
//
//	Pro, Max, Team   rate-limit windows (a 5-hour block, a weekly pool)
//	usage-based      a credit pool in money, and no windows at all
//
// On an Enterprise usage-based seat every window in /oauth/usage comes back
// null and Claude Code sends no rate_limits in the status line payload, so a
// windows-only status line renders nothing and looks broken. On Pro the credit
// pool reports spend.enabled false and the budget renders nothing instead.
// Knowing the plan is what lets one configuration be right on both.

// ProfileTTL is how long a cached profile is served. The plan changes when a
// seat is reassigned, which is rare - unlike usage, which moves constantly.
const ProfileTTL = 24 * time.Hour

// Plan is the kind of subscription the current credentials belong to.
type Plan string

// The plans this distinguishes. Unknown covers a response whose shape is new
// to us; callers treat it as "show what the payload offers and nothing more".
const (
	PlanUnknown    Plan = "unknown"
	PlanFree       Plan = "free"
	PlanPro        Plan = "pro"
	PlanMax        Plan = "max"
	PlanTeam       Plan = "team"
	PlanEnterprise Plan = "enterprise"
)

// MetersBudget reports whether this plan meters a credit pool in money, which
// is what the credits module renders.
func (p Plan) MetersBudget() bool { return p == PlanEnterprise }

// MetersWindows reports whether this plan meters rate-limit windows, which is
// what the usage module renders from the status line payload.
func (p Plan) MetersWindows() bool {
	return p == PlanPro || p == PlanMax || p == PlanTeam || p == PlanFree
}

// Profile is the part of the profile response this tool reads.
type Profile struct {
	Account struct {
		Email        string `json:"email"`
		HasClaudePro bool   `json:"has_claude_pro"`
		HasClaudeMax bool   `json:"has_claude_max"`
	} `json:"account"`
	Organization struct {
		Name string `json:"name"`
		// Type is claude_enterprise, claude_team and so on.
		Type string `json:"organization_type"`
		// SeatTier distinguishes a usage-based seat from a fixed one within
		// the same organization type, which is what decides budget metering.
		SeatTier string `json:"seat_tier"`
	} `json:"organization"`

	// Age is how long ago the cache this came from was written. Not part of
	// the API response.
	Age time.Duration `json:"-"`
}

// Plan classifies the profile. Organization type is checked before the
// personal has_claude_* flags, because an Enterprise seat reports both of
// those false while still being the plan in force.
func (p Profile) Plan() Plan {
	switch p.Organization.Type {
	case "claude_enterprise":
		return PlanEnterprise
	case "claude_team":
		return PlanTeam
	}

	switch {
	case p.Account.HasClaudeMax:
		return PlanMax
	case p.Account.HasClaudePro:
		return PlanPro
	case p.Organization.Type != "":
		return PlanUnknown
	}

	return PlanFree
}

// UsageBased reports whether the seat bills by usage against a credit pool,
// as opposed to a fixed allowance.
func (p Profile) UsageBased() bool {
	return p.Organization.SeatTier == "enterprise_usage_based"
}

func profilePath() (string, error) {
	dir, err := accountDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "profile.json"), nil
}

// LoadProfile returns the cached profile, fetching it when there is none or it
// has gone stale. Unlike usage this is allowed to block: it is called by the
// plan command, not on the render path.
func LoadProfile() (*Profile, error) {
	path, err := profilePath()
	if err != nil {
		return nil, err
	}

	info, statErr := os.Stat(path)
	if statErr != nil || time.Since(info.ModTime()) > ProfileTTL {
		err = RefreshProfile()
		if err != nil {
			return nil, err
		}

		info, statErr = os.Stat(path)
		if statErr != nil {
			return nil, fmt.Errorf("no cached profile: %w", statErr)
		}
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading profile cache: %w", err)
	}

	var profile Profile

	err = json.Unmarshal(raw, &profile)
	if err != nil {
		return nil, fmt.Errorf("parsing profile cache: %w", err)
	}

	profile.Age = time.Since(info.ModTime())

	return &profile, nil
}

// RefreshProfile fetches the profile from Anthropic and writes the cache.
func RefreshProfile() error {
	token, err := accessToken()
	if err != nil {
		return err
	}

	body, retry, err := get(token, "profile")
	if err != nil {
		setBackoff(retry)

		return err
	}

	path, err := profilePath()
	if err != nil {
		return err
	}

	// Write-then-rename, so a reader never sees a half-written cache.
	tmp := path + ".tmp"

	err = os.WriteFile(tmp, body, filePerms)
	if err != nil {
		return fmt.Errorf("writing profile cache: %w", err)
	}

	return os.Rename(tmp, path)
}
