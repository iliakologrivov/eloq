package eloq

import (
	"context"
	"fmt"
	"maps"
	"strings"
)

type PlaceholderFormat string

const (
	Question PlaceholderFormat = "?"
	Dollar   PlaceholderFormat = "$"
)

type QuoteStyle string

const (
	Backtick    QuoteStyle = "`"
	DoubleQuote QuoteStyle = `"`
)

type ContextKey string

const (
	ContextTraceID   ContextKey = "trace_id"
	ContextSpanID    ContextKey = "span_id"
	ContextRequestID ContextKey = "request_id"
	ContextUserID    ContextKey = "user_id"
)

type rawSqlRef string

func Raw(name string) rawSqlRef {
	return rawSqlRef(name)
}

type sqlPart struct {
	sql  string
	args []interface{}
}
type commonBuilder struct {
	placeholder PlaceholderFormat
	quoteStyle  QuoteStyle
	comments    []string
	queryName   string
	meta        map[string]string
	prefixes    []sqlPart
	suffixes    []sqlPart
}

func NewBuilder() *commonBuilder {
	return &commonBuilder{
		placeholder: Question,
		quoteStyle:  Backtick,
		comments:    []string{},
		queryName:   "",
		meta:        map[string]string{},
	}
}

func NewMysqlBuilder() *commonBuilder {
	return NewBuilder().
		PlaceholderFormat(Question).
		QuoteWith(Backtick)
}

func NewPsqlBuilder() *commonBuilder {
	return NewBuilder().
		PlaceholderFormat(Dollar).
		QuoteWith(DoubleQuote)
}

func (b *commonBuilder) quoteIdentifier(id string) (string, error) {
	quote := string(b.quoteStyle)
	parts := strings.Split(id, ".")
	for i, p := range parts {
		p = strings.TrimSpace(p)
		if p == "*" {
			parts[i] = p
			continue
		}
		if !isValidIdentifier(p) {
			return "", fmt.Errorf("invalid identifier: %s", p)
		}
		parts[i] = quote + p + quote
	}
	return strings.Join(parts, "."), nil
}

func isValidIdentifier(s string) bool {
	for _, r := range s {
		if !(r == '_' || r == '.' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9') {
			return false
		}
	}
	return true
}

func (b *commonBuilder) formatPlaceholder(n int) string {
	if b.placeholder == Dollar {
		return fmt.Sprintf("$%d", n)
	}
	return "?"
}

func (b *commonBuilder) PlaceholderFormat(f PlaceholderFormat) *commonBuilder {
	b.placeholder = f
	return b
}

func (b *commonBuilder) QuoteWith(q QuoteStyle) *commonBuilder {
	b.quoteStyle = q
	return b
}

func (b *commonBuilder) Comment(text string) *commonBuilder {
	if strings.TrimSpace(text) != "" {
		b.comments = append(b.comments, text)
	}
	return b
}

func (b *commonBuilder) CommentKV(kv ...interface{}) *commonBuilder {
	if len(kv)%2 != 0 {
		return b
	}

	var parts []string
	for i := 0; i < len(kv); i += 2 {
		k := fmt.Sprint(kv[i])
		v := fmt.Sprint(kv[i+1])
		parts = append(parts, k+"="+v)
	}

	if len(parts) > 0 {
		b.comments = append(b.comments, strings.Join(parts, " "))
	}

	return b
}

func (b *commonBuilder) Name(name string) *commonBuilder {
	name = strings.TrimSpace(name)
	if name == "" {
		return b
	}

	// sanitize
	name = strings.ReplaceAll(name, "/*", "")
	name = strings.ReplaceAll(name, "*/", "")

	b.queryName = name
	return b
}

func (b *commonBuilder) Namef(format string, args ...interface{}) *commonBuilder {
	return b.Name(fmt.Sprintf(format, args...))
}

func (b *commonBuilder) AddMeta(key string, value interface{}) *commonBuilder {
	key = strings.TrimSpace(key)
	if key == "" {
		return b
	}

	if b.meta == nil {
		b.meta = make(map[string]string)
	}

	val := strings.TrimSpace(fmt.Sprint(value))
	if val == "" {
		return b
	}

	// sanitize
	val = strings.ReplaceAll(val, "/*", "")
	val = strings.ReplaceAll(val, "*/", "")

	b.meta[key] = val
	return b
}

func (b *commonBuilder) WithMeta(m map[string]string) *commonBuilder {
	if len(m) == 0 {
		return b
	}

	if b.meta == nil {
		b.meta = make(map[string]string)
	}

	for k, v := range m {
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}

		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}

		v = strings.ReplaceAll(v, "/*", "")
		v = strings.ReplaceAll(v, "*/", "")

		b.meta[k] = v
	}

	return b
}

func (b *commonBuilder) WithContext(ctx context.Context) *commonBuilder {
	if ctx == nil {
		return b
	}

	if b.meta == nil {
		b.meta = map[string]string{}
	}

	if v := ctx.Value(ContextTraceID); v != nil {
		b.meta["trace_id"] = fmt.Sprint(v)
	}

	if v := ctx.Value(ContextSpanID); v != nil {
		b.meta["span_id"] = fmt.Sprint(v)
	}

	if v := ctx.Value(ContextRequestID); v != nil {
		b.meta["request_id"] = fmt.Sprint(v)
	}

	if v := ctx.Value(ContextUserID); v != nil {
		b.meta["user_id"] = fmt.Sprint(v)
	}

	return b
}

func (b *commonBuilder) renderComments(sql *strings.Builder) {
	var parts []string

	// name first
	if b.queryName != "" {
		parts = append(parts, "name="+b.queryName)
	}

	// meta (space separated)
	for k, v := range b.meta {
		parts = append(parts, k+"="+v)
	}

	// free comments (pipe separated)
	if len(b.comments) > 0 {
		var cs []string
		for _, c := range b.comments {
			c = strings.ReplaceAll(c, "/*", "")
			c = strings.ReplaceAll(c, "*/", "")
			c = strings.TrimSpace(c)

			if c != "" {
				cs = append(cs, c)
			}
		}
		if len(cs) > 0 {
			parts = append(parts, strings.Join(cs, " | "))
		}
	}

	if len(parts) == 0 {
		return
	}

	sql.WriteString("/* ")
	sql.WriteString(strings.Join(parts, " "))
	sql.WriteString(" */ ")
}

func (b *commonBuilder) Prefix(sql string, args ...interface{}) *commonBuilder {
	if sql == "" {
		return b
	}

	b.prefixes = append(b.prefixes, sqlPart{
		sql:  sql,
		args: args,
	})
	return b
}

func (b *commonBuilder) renderPrefixes(sql *strings.Builder, args *[]interface{}) {
	for _, p := range b.prefixes {
		sql.WriteString(p.sql)
		sql.WriteByte(' ')
		*args = append(*args, p.args...)
	}
}

func (b *commonBuilder) Suffix(sql string, args ...interface{}) *commonBuilder {
	if sql == "" {
		return b
	}

	b.suffixes = append(b.suffixes, sqlPart{
		sql:  sql,
		args: args,
	})
	return b
}

func (b *commonBuilder) renderSuffixes(sql *strings.Builder, args *[]interface{}) {
	for _, s := range b.suffixes {
		sql.WriteByte(' ')
		sql.WriteString(s.sql)
		*args = append(*args, s.args...)
	}
}

func ToAnySlice[T interface{}](in []T) []interface{} {
	out := make([]interface{}, len(in))
	for i, v := range in {
		out[i] = v
	}
	return out
}

func (b *commonBuilder) clone() *commonBuilder {
	return &commonBuilder{
		placeholder: b.placeholder,
		quoteStyle:  b.quoteStyle,
		comments:    append([]string{}, b.comments...),
		queryName:   b.queryName,
		meta:        maps.Clone(b.meta),
		prefixes:    append([]sqlPart{}, b.prefixes...),
		suffixes:    append([]sqlPart{}, b.suffixes...),
	}
}
