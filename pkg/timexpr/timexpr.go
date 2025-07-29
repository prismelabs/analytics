package timexpr

import (
	"errors"
	"regexp"
	"strconv"
	"time"
	"unicode"
)

var (
	ErrSyntax     = errors.New("invalid time expression syntax")
	durationRegex = regexp.MustCompile(`^\d+(y|M|d|h|m|s)$`)
)

// Parse parse a datetime expression such as 'now-7d' or
// '2025-07-10T22:00:00.000Z'.
func Parse(expr string) (t time.Time, err error) {
	if expr == "" {
		return time.Time{}, ErrSyntax
	}

	var sign int64 = 1
	var skip = 0

	for i, r := range expr {
		if skip > 0 {
			skip--
			continue
		}

		if unicode.IsLetter(r) {
			if !t.Equal(time.Time{}) {
				return time.Time{}, ErrSyntax
			}

			var read int
			t, read, err = parseRef(expr[i:])
			if err != nil {
				return t, err
			}
			skip = read - 1
			continue
		}
		if r == '-' || r == '+' {
			if t.Equal(time.Time{}) {
				return time.Time{}, ErrSyntax
			}
			if r == '-' {
				sign = -1
			}
			continue
		}

		if unicode.IsDigit(r) {
			if t.Equal(time.Time{}) {
				t, err = time.Parse(time.RFC3339, expr[i:])
				if err != nil {
					t, err = time.Parse(time.DateOnly, expr[i:])
				}
				return
			} else if durationRegex.MatchString(expr[i:]) {
				var n int64
				n, err = strconv.ParseInt(expr[i:len(expr)-1], 10, 64)
				if err != nil {
					return t, err
				}
				i := int(sign * n)
				d := time.Duration(sign * n)

				switch expr[len(expr)-1] {
				case 'y':
					t = t.AddDate(i, 0, 0)
				case 'M':
					t = t.AddDate(0, i, 0)
				case 'd':
					t = t.AddDate(0, 0, i)
				case 'h':
					t = t.Add(d * time.Hour)
				case 'm':
					t = t.Add(d * time.Minute)
				case 's':
					t = t.Add(d * time.Second)
				default:
					return time.Time{}, ErrSyntax
				}
				return
			}
		}

		return time.Time{}, ErrSyntax
	}

	return t, nil
}

func parseRef(expr string) (time.Time, int, error) {
	var ref string

	for i, r := range expr {
		if !unicode.IsLetter(r) {
			break
		}

		ref = expr[:i+1]
	}

	switch ref {
	case "now":
		return time.Now(), len(ref), nil
	default:
		return time.Time{}, 0, ErrSyntax
	}
}
