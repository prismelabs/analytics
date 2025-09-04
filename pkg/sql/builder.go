package sql

import (
	"fmt"
	"strings"
)

// Builder is an enhanced string builder to build SQL queries with placeholder
// parameters. It won't prevent you from building invalid queries as it
// performs 0 validations.
//
// The zero value is ready to use. Do not copy a non-zero Builder.
type Builder struct {
	builder strings.Builder
	args    []any
}

// Str adds a single part with it's arguments to the query.
func (b *Builder) Str(part string, args ...any) *Builder {
	b.space()
	_, _ = b.builder.WriteString(part)
	b.args = append(b.args, args...)

	return b
}

// Fmt format string and adds it to the query. Do not use this function to add
// untrusted data, instead use Str() with parameters.
func (b *Builder) Fmt(format string, args ...any) *Builder {
	b.space()
	_, _ = fmt.Fprintf(&b.builder, format, args...)
	return b
}

// Strs adds multiple string parts to the query but doesn't support placeholders.
// If you need placeholders, call Builder.Str in a loop.
func (b *Builder) Strs(parts ...string) *Builder {
	for _, str := range parts {
		b.space()
		_, _ = b.builder.WriteString(str)
	}

	return b
}

// Call calls provided function with this builder and `args` as argument.
func (b *Builder) Call(fn func(*Builder, ...any), args ...any) *Builder {
	b.space()
	fn(b, args...)
	return b
}

// Finish returns final query and placeholders values.
func (b *Builder) Finish() (string, []any) {
	query := b.builder.String()
	args := b.args

	return query, args
}

// Reset resets the builder.
func (b *Builder) Reset() *Builder {
	b.builder = strings.Builder{}
	b.args = nil
	return b
}

func (b *Builder) space() {
	if b.builder.Len() > 0 {
		b.builder.WriteRune(' ')
	}
}
