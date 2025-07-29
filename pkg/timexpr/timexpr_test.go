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
		err      error
	}
	testCases := []testCase{
		{expr: "now", expected: time.Now()},
		{expr: "now-120s", expected: time.Now().Add(-2 * time.Minute)},
		{expr: "now+120s", expected: time.Now().Add(2 * time.Minute)},
		{expr: "now-2h", expected: time.Now().Add(-2 * time.Hour)},
		{expr: "now-7d", expected: time.Now().AddDate(0, 0, -7)},
		{expr: "now-3M", expected: time.Now().AddDate(0, -3, 0)},
		{expr: "now-4y", expected: time.Now().AddDate(-4, 0, 0)},
		{expr: "now-7", expected: time.Time{}, err: ErrSyntax},
		{expr: "now-7h-7d", expected: time.Time{}, err: ErrSyntax},
		{expr: "", expected: time.Time{}, err: ErrSyntax},
		{expr: "¨¨", expected: time.Time{}, err: ErrSyntax},
		{
			expr: "2025-07-10T22:00:02.000Z",
			expected: testutils.Must2(time.Parse)(
				time.RFC3339,
				"2025-07-10T22:00:02.000Z",
			),
		},
		{
			expr: "2025-07-10",
			expected: testutils.Must2(time.Parse)(
				time.DateOnly,
				"2025-07-10",
			),
		},
	}

	for _, tcase := range testCases {
		t.Run(tcase.expr, func(t *testing.T) {
			ti, err := Parse(tcase.expr)
			if tcase.err != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tcase.err)
				require.Equal(t, tcase.expected, ti)
			} else {
				require.NoError(t, err)
				require.WithinDuration(t, tcase.expected, ti, time.Second)
			}
		})
	}
}
