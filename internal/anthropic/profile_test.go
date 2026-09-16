package anthropic_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/frank-bee/claude-statusline/internal/anthropic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// enterpriseProfile is a real /oauth/profile response, trimmed to the fields
// this tool reads. The shape is the contract; the values are a live Enterprise
// usage-based seat.
const enterpriseProfile = `{
	"account": {
		"email": "someone@example.com",
		"has_claude_pro": false,
		"has_claude_max": false
	},
	"organization": {
		"name": "example-org",
		"organization_type": "claude_enterprise",
		"seat_tier": "enterprise_usage_based"
	}
}`

func TestProfilePlan(t *testing.T) {
	tests := map[string]struct {
		orgType  string
		pro, max bool
		expected anthropic.Plan
	}{
		"an enterprise organization":      {orgType: "claude_enterprise", expected: anthropic.PlanEnterprise},
		"a team organization":             {orgType: "claude_team", expected: anthropic.PlanTeam},
		"a personal Max subscription":     {max: true, expected: anthropic.PlanMax},
		"a personal Pro subscription":     {pro: true, expected: anthropic.PlanPro},
		"Max wins when both flags are on": {pro: true, max: true, expected: anthropic.PlanMax},
		"no subscription at all":          {expected: anthropic.PlanFree},
		"an organization type we do not know": {
			orgType: "claude_something_new", expected: anthropic.PlanUnknown,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var profile anthropic.Profile

			profile.Organization.Type = test.orgType
			profile.Account.HasClaudePro = test.pro
			profile.Account.HasClaudeMax = test.max

			assert.Equal(t, test.expected, profile.Plan())
		})
	}
}

func TestPlanMetering(t *testing.T) {
	t.Run("an enterprise seat meters a budget, not windows", func(t *testing.T) {
		// This is the whole reason the plan is worth knowing: on this seat
		// every window in /oauth/usage is null and Claude Code sends no
		// rate_limits, so a windows-only status line renders nothing.
		assert.True(t, anthropic.PlanEnterprise.MetersBudget())
		assert.False(t, anthropic.PlanEnterprise.MetersWindows())
	})

	t.Run("pro, max and team meter windows, not a budget", func(t *testing.T) {
		for _, plan := range []anthropic.Plan{anthropic.PlanPro, anthropic.PlanMax, anthropic.PlanTeam} {
			assert.True(t, plan.MetersWindows(), plan)
			assert.False(t, plan.MetersBudget(), plan)
		}
	})

	t.Run("an unknown plan claims neither", func(t *testing.T) {
		assert.False(t, anthropic.PlanUnknown.MetersWindows())
		assert.False(t, anthropic.PlanUnknown.MetersBudget())
	})
}

func TestProfileUsageBased(t *testing.T) {
	var profile anthropic.Profile

	assert.False(t, profile.UsageBased(), "a seat tier we were not given is not usage-based")

	profile.Organization.SeatTier = "enterprise_usage_based"
	assert.True(t, profile.UsageBased())
}

func TestLoadProfile(t *testing.T) {
	// A temporary HOME and an empty CLAUDE_CONFIG_DIR, so the suite never
	// reaches the credentials of whoever is running it.
	setup := func(t *testing.T, body string) {
		t.Helper()

		anthropic.SuppressKeychain(t)

		home := t.TempDir()
		t.Setenv("HOME", home)
		t.Setenv("CLAUDE_CONFIG_DIR", home)
		t.Setenv("XDG_STATE_HOME", t.TempDir())

		require.NoError(t, os.WriteFile(
			filepath.Join(home, ".credentials.json"),
			[]byte(`{"claudeAiOauth":{"accessToken":"test-token"}}`), 0o600))

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/profile", r.URL.Path)
			_, _ = w.Write([]byte(body))
		}))
		t.Cleanup(server.Close)

		anthropic.SetBaseURL(t, server.URL)
	}

	t.Run("fetches and classifies a real enterprise response", func(t *testing.T) {
		setup(t, enterpriseProfile)

		profile, err := anthropic.LoadProfile()

		require.NoError(t, err)
		assert.Equal(t, anthropic.PlanEnterprise, profile.Plan())
		assert.Equal(t, "example-org", profile.Organization.Name)
		assert.Equal(t, "someone@example.com", profile.Account.Email)
		assert.True(t, profile.UsageBased())
	})

	t.Run("a corrupt cache errors rather than panicking", func(t *testing.T) {
		setup(t, "{not json")

		_, err := anthropic.LoadProfile()

		require.ErrorContains(t, err, "parsing profile cache")
	})
}
