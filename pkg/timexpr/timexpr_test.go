package timexpr

import (
	"testing"
	"time"

	"github.com/prismelabs/analytics/pkg/testutils"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	type testCase struct {
		expr     string
		expected time.Time
	}
	testCases := []testCase{
		{expr: "now", expected: time.Now()},
		{expr: "now-2h", expected: time.Now().Add(-2 * time.Hour)},
		{expr: "now-7d", expected: time.Now().AddDate(0, 0, -7)},
		{
			expr: "2025-07-10T22:00:00.000Z",
			expected: testutils.Must2(time.Parse)(
				time.RFC3339,
				"2025-07-10T22:00:00.000Z",
			),
		},
	}

	for _, tcase := range testCases {
		t.Run(tcase.expr, func(t *testing.T) {
			ti, err := Parse(tcase.expr)
			require.NoError(t, err)
			require.WithinDuration(t, tcase.expected, ti, time.Second)
		})
	}
}
