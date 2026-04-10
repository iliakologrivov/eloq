package eloq

import (
	"context"
	"fmt"
	"maps"
	"strconv"
	"strings"
)

type Config struct {
	placeholder  PlaceholderFormat
	quoteStyle   QuoteStyle
	requireWhere bool
}

type queryState struct {
	comments  []string
	queryName string
	meta      map[string]string
	prefixes  []sqlPart
	suffixes  []sqlPart
}

type baseBuilder struct {
	Config
	*queryState
}

func newBaseBuilder() baseBuilder {
	return baseBuilder{
		Config: Config{
			placeholder: Question,
			quoteStyle:  Backtick,
		},
		queryState: newQueryState(),
	}
}

func newBaseBuilderWithConfig(cfg Config) baseBuilder {
	return baseBuilder{
		Config:     cfg,
		queryState: newQueryState(),
	}
}

func newQueryState() *queryState {
	return &queryState{
		comments: []string{},
		meta:     map[string]string{},
	}
}

func (b *baseBuilder) clone() baseBuilder {
	return baseBuilder{
		Config:     b.Config,
		queryState: b.queryState.clone(),
	}
}

func (qs *queryState) clone() *queryState {
	return &queryState{
		comments:  append([]string{}, qs.comments...),
		queryName: qs.queryName,
		meta:      maps.Clone(qs.meta),
		prefixes:  append([]sqlPart{}, qs.prefixes...),
		suffixes:  append([]sqlPart{}, qs.suffixes...),
	}
}

func (b *baseBuilder) quoteIdentifier(id string) (string, error) {
	dotIdx := strings.IndexByte(id, '.')
	if dotIdx == -1 {
		p := strings.TrimSpace(id)
		if p == "*" {
			return p, nil
		}
		if !isValidIdentifier(p) {
			return "", fmt.Errorf("invalid identifier: %s", p)
		}
		quote := string(b.quoteStyle)
		return quote + p + quote, nil
	}

	quote := string(b.quoteStyle)
	parts := strings.Split(id, ".")
	var sb strings.Builder
	sb.Grow(len(id) + len(parts)*2)

	for i, p := range parts {
		p = strings.TrimSpace(p)
		if i > 0 {
			sb.WriteByte('.')
		}
		if p == "*" {
			sb.WriteString(p)
			continue
		}
		if !isValidIdentifier(p) {
			return "", fmt.Errorf("invalid identifier: %s", p)
		}
		sb.WriteString(quote)
		sb.WriteString(p)
		sb.WriteString(quote)
	}
	return sb.String(), nil
}

func isValidIdentifier(s string) bool {
	for _, r := range s {
		if !(r == '_' || r == '.' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9') {
			return false
		}
	}
	return true
}

func (b *baseBuilder) formatPlaceholder(n int) string {
	if b.placeholder == Dollar {
		return "$" + strconv.Itoa(n)
	}
	return "?"
}

func (b *baseBuilder) PlaceholderFormat(f PlaceholderFormat) {
	b.placeholder = f
}

func (b *baseBuilder) QuoteWith(q QuoteStyle) {
	b.quoteStyle = q
}

func (b *baseBuilder) Comment(text string) {
	if strings.TrimSpace(text) != "" {
		b.comments = append(b.comments, text)
	}
}

func (b *baseBuilder) CommentKV(kv ...interface{}) {
	if len(kv)%2 != 0 {
		return
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
}

func (b *baseBuilder) Name(name string) {
	name = strings.TrimSpace(name)
	if name == "" {
		return
	}

	// sanitize
	name = strings.ReplaceAll(name, "/*", "")
	name = strings.ReplaceAll(name, "*/", "")

	b.queryName = name
}

func (b *baseBuilder) Namef(format string, args ...interface{}) {
	b.Name(fmt.Sprintf(format, args...))
}

func (b *baseBuilder) AddMeta(key string, value interface{}) {
	key = strings.TrimSpace(key)
	if key == "" {
		return
	}

	if b.meta == nil {
		b.meta = make(map[string]string)
	}

	val := strings.TrimSpace(fmt.Sprint(value))
	if val == "" {
		return
	}

	// sanitize
	val = strings.ReplaceAll(val, "/*", "")
	val = strings.ReplaceAll(val, "*/", "")

	b.meta[key] = val
}

func (b *baseBuilder) WithMeta(m map[string]string) {
	if len(m) == 0 {
		return
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
}

func (b *baseBuilder) WithContext(ctx context.Context) {
	if ctx == nil {
		return
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
}

func (b *baseBuilder) renderComments(sql *strings.Builder) {
	// Быстрая проверка: есть ли что рендерить
	if b.queryName == "" && len(b.meta) == 0 && len(b.comments) == 0 {
		return
	}

	// Предварительно санитизируем комментарии и проверяем есть ли что выводить
	var sanitizedComments []string
	for _, c := range b.comments {
		c = strings.ReplaceAll(c, "/*", "")
		c = strings.ReplaceAll(c, "*/", "")
		c = strings.TrimSpace(c)
		if c != "" {
			sanitizedComments = append(sanitizedComments, c)
		}
	}

	// Проверяем: если всё пустое после санитизации — не выводим комментарий
	if b.queryName == "" && len(b.meta) == 0 && len(sanitizedComments) == 0 {
		return
	}

	sql.WriteString("/* ")

	first := true

	if b.queryName != "" {
		sql.WriteString("name=")
		sql.WriteString(b.queryName)
		first = false
	}

	for k, v := range b.meta {
		if !first {
			sql.WriteByte(' ')
		}
		sql.WriteString(k)
		sql.WriteByte('=')
		sql.WriteString(v)
		first = false
	}

	for _, c := range sanitizedComments {
		if !first {
			sql.WriteString(" | ")
		}
		sql.WriteString(c)
		first = false
	}

	sql.WriteString(" */ ")
}

func (b *baseBuilder) Prefix(sql string, args ...interface{}) {
	if sql == "" {
		return
	}

	b.prefixes = append(b.prefixes, sqlPart{
		sql:  sql,
		args: args,
	})
}

func (b *baseBuilder) renderPrefixes(sql *strings.Builder, args *[]interface{}, startIndex int) int {
	index := startIndex
	for _, p := range b.prefixes {
		renderedSQL := b.renderPlaceholderInSQL(p.sql, index)
		sql.WriteString(renderedSQL)
		sql.WriteByte(' ')
		*args = append(*args, p.args...)
		index += len(p.args)
	}
	return index
}

func (b *baseBuilder) renderPlaceholderInSQL(sql string, startIndex int) string {
	if b.placeholder == Question {
		return sql
	}

	var sb strings.Builder
	sb.Grow(len(sql) + 10)

	for _, ch := range sql {
		if ch == '?' {
			sb.WriteByte('$')
			sb.WriteString(strconv.Itoa(startIndex))
			startIndex++
		} else {
			sb.WriteRune(ch)
		}
	}
	return sb.String()
}

func (b *baseBuilder) Suffix(sql string, args ...interface{}) {
	if sql == "" {
		return
	}

	b.suffixes = append(b.suffixes, sqlPart{
		sql:  sql,
		args: args,
	})
}

func (b *baseBuilder) renderSuffixes(sql *strings.Builder, args *[]interface{}, startIndex int) int {
	index := startIndex
	for _, s := range b.suffixes {
		sql.WriteByte(' ')
		renderedSQL := b.renderPlaceholderInSQL(s.sql, index)
		sql.WriteString(renderedSQL)
		*args = append(*args, s.args...)
		index += len(s.args)
	}
	return index
}
