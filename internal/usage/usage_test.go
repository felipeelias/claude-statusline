package usage

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockFetcher(t *testing.T) {
	data := &UsageData{
		FiveHour: Window{Utilisation: 42},
		SevenDay: Window{Utilisation: 15},
	}

	f := MockFetcher{Data: data}
	got, err := f.GetUsage(120 * time.Second)

	require.NoError(t, err)
	assert.InDelta(t, 42, got.FiveHour.Utilisation, 0.1)
	assert.InDelta(t, 15, got.SevenDay.Utilisation, 0.1)
}

func TestMockFetcher_NilData(t *testing.T) {
	f := MockFetcher{}
	got, err := f.GetUsage(120 * time.Second)

	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestErrorBackoff(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected time.Duration
	}{
		{
			name:     "token expired",
			err:      errTokenExpired,
			expected: tokenExpiredBackoff,
		},
		{
			name:     "rate limited",
			err:      &rateLimitError{retryAfter: 90 * time.Second},
			expected: 90 * time.Second,
		},
		{
			name:     "generic error",
			err:      assert.AnError,
			expected: networkErrorBackoff,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, errorBackoff(tt.err))
		})
	}
}
