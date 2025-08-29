package timexpr

import (
	"errors"
	"strconv"
	"time"
	"unicode"
)

var (
	ErrSyntax = errors.New("invalid time expression syntax")
)

func Parse(expr string, floor bool) (time.Time, error) {
	p := parser{
		expr:   expr,
		floor:  floor,
		cursor: 0,
		time:   time.Time{},
	}
	err := p.parse()
	if err != nil {
		return time.Time{}, err
	}
	return p.time, nil
}

type parser struct {
	expr   string
	floor  bool
	cursor int
	time   time.Time
}

// Peek a byte.
func (p *parser) peek() rune {
	if p.cursor >= len(p.expr) {
		return 0
	}

	return rune(p.expr[p.cursor])
}

// Returns next byte and move cursor.
func (p *parser) next() rune {
	p.cursor = min(p.cursor+1, len(p.expr))
	return p.peek()
}

func (p *parser) parse() error {
	for r := p.peek(); r != 0; r = p.peek() {
		if unicode.IsSpace(r) {
			_ = p.next()
			continue
		}

		if unicode.IsLetter(r) {
			err := p.parseRef()
			if err != nil {
				return err
			}
			continue
		}

		if unicode.IsDigit(r) {
			err := p.parseDate()
			if err != nil {
				return err
			}
			continue
		}

		if r == '-' || r == '+' {
			err := p.parseOffset()
			if err != nil {
				return err
			}
			continue
		}

		if r == '/' {
			err := p.parseRoundingFactor()
			if err != nil {
				return err
			}
			break
		}

		return ErrSyntax
	}

	if p.time.Equal(time.Time{}) {
		return ErrSyntax
	}

	return nil
}

func (p *parser) parseRef() error {
	if !p.time.Equal(time.Time{}) {
		return ErrSyntax
	}

	start := p.cursor
	for unicode.IsLetter(p.next()) {
	}

	ref := p.expr[start:p.cursor]
	if ref == "now" {
		p.time = time.Now()
		return nil
	}

	return ErrSyntax
}

func (p *parser) parseOffset() error {
	if p.time.Equal(time.Time{}) {
		return ErrSyntax
	}

	start := p.cursor
	for unicode.IsDigit(p.next()) {
	}

	n, err := strconv.ParseInt(p.expr[start:p.cursor], 10, 64)
	if err != nil {
		return err
	}
	i := int(n)
	d := time.Duration(n)

	switch p.peek() {
	case 'y': // Year.
		p.time = p.time.AddDate(i, 0, 0)
	case 'Q': // Quarter.
		p.time = p.time.AddDate(0, 3*i, 0)
	case 'M': // Month.
		p.time = p.time.AddDate(0, i, 0)
	case 'w': // Week.
		p.time = p.time.AddDate(0, 0, 7*i)
	case 'd': // Day.
		p.time = p.time.AddDate(0, 0, i)
	case 'h': // Hour.
		p.time = p.time.Add(d * time.Hour)
	case 'm': // Minute.
		p.time = p.time.Add(d * time.Minute)
	case 's': // Second.
		p.time = p.time.Add(d * time.Second)
	default:
		return ErrSyntax
	}
	_ = p.next()

	return nil
}

func (p *parser) parseDate() (err error) {
	if !p.time.Equal(time.Time{}) {
		return ErrSyntax
	}

	start := p.cursor
	for {
		r := p.next()
		if !unicode.IsDigit(r) && r != 'T' && r != 'Z' && r != '-' && r != ':' && r != '.' {
			break
		}
	}
	p.time, err = time.Parse(time.RFC3339, p.expr[start:p.cursor])
	if err != nil {
		p.time, err = time.Parse(time.DateOnly, p.expr[start:p.cursor])
	}

	return
}

func (p *parser) parseRoundingFactor() error {
	if p.time.Equal(time.Time{}) {
		return ErrSyntax
	}

	_ = p.next()
	t := p.time

	switch p.expr[p.cursor:] {
	case "y": // Year.
		if p.floor {
			p.time = time.Date(t.Year(), time.January, 1, 0, 0, 0, 0, time.UTC)
		} else {
			p.time = time.Date(t.Year(), time.December, 31, 23, 59, 59, 999999999, time.UTC)
		}
	case "Q": // Quarter.
		if p.floor {
			p.time = time.Date(t.Year(), t.Month()-t.Month()%3, 1, 0, 0, 0, 0, time.UTC)
		} else {
			p.time = time.Date(t.Year(), t.Month()+3-t.Month()%3, 0, 23, 59, 59, 999999999, time.UTC)
		}
	case "fQ": // Fiscal quarter.
		if p.floor {
			p.time = time.Date(t.Year(), ((t.Month()-1)/3)*3, 1, 0, 0, 0, 0, time.UTC)
		} else {
			p.time = time.Date(t.Year(), ((t.Month()+2)/3)*3+1, 0, 23, 59, 59, 999999999, time.UTC)
		}
	case "M": // Month.
		if p.floor {
			p.time = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
		} else {
			p.time = time.Date(t.Year(), t.Month()+1, 0, 23, 59, 59, 999999999, time.UTC)
		}
	case "w": // Week.
		if p.floor {
			p.time = time.Date(t.Year(), t.Month(), t.Day()-int(t.Weekday()), 0, 0, 0, 0, time.UTC)
		} else {
			p.time = time.Date(t.Year(), t.Month(), t.Day()+int(time.Saturday)-int(t.Weekday()), 23, 59, 59, 999999999, time.UTC)
		}
	case "d": // Day.
		if p.floor {
			p.time = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
		} else {
			p.time = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, time.UTC)
		}
	case "h": // Hour.
		if p.floor {
			p.time = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, time.UTC)
		} else {
			p.time = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 59, 59, 999999999, time.UTC)
		}
	case "m": // Minute.
		if p.floor {
			p.time = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, time.UTC)
		} else {
			p.time = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 59, 999999999, time.UTC)
		}
	case "s": // Second.
		if p.floor {
			p.time = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, time.UTC)
		} else {
			p.time = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 999999999, time.UTC)
		}
	default:
		return ErrSyntax
	}

	return nil
}
