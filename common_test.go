package eloq

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getPsqlBuilder() *commonBuilder {
	return NewBuilder().
		PlaceholderFormat(Dollar).
		QuoteWith(DoubleQuote)
}

func getMysqlBuilder() *commonBuilder {
	return NewBuilder().
		PlaceholderFormat(Question).
		QuoteWith(Backtick)
}

func TestCommonBasic_Psql(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT * FROM "users"`, sql)
	assert.Equal(t, []interface{}{}, args)
}

func TestCommonBasic_MySql(t *testing.T) {
	sql, args, err := getMysqlBuilder().
		Select("*").
		From("users").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, "SELECT * FROM `users`", sql)
	assert.Equal(t, []interface{}{}, args)
}

func TestRenderComments(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Comment("hello").
		CommentKV("a", 1).
		Select("*").
		From("users").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, []interface{}{}, args)
	assert.Equal(t, `/* hello | a=1 */ SELECT * FROM "users"`, sql)
}

func TestRenderCommentVK(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Comment("hello").
		CommentKV("a", 1, "k").
		Select("*").
		From("users").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, []interface{}{}, args)
	assert.Equal(t, `/* hello */ SELECT * FROM "users"`, sql)
}

func TestRenderCommentSanitize(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Comment("*/").
		Select("*").
		From("users").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, []interface{}{}, args)
	assert.Equal(t, `SELECT * FROM "users"`, sql)
}

func TestQueryName(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Name("load_users").
		Select("*").
		From("users").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, []interface{}{}, args)
	assert.Equal(t,
		`/* name=load_users */ SELECT * FROM "users"`,
		sql,
	)
}

func TestEmptyQueryName(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Name("").
		Select("*").
		From("users").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, []interface{}{}, args)
	assert.Equal(t,
		`SELECT * FROM "users"`,
		sql,
	)
}

func TestMeta(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		AddMeta("a", 1).
		AddMeta("b", "x").
		Select("*").
		From("users").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, []interface{}{}, args)
	assert.Contains(t, sql, "a=1")
	assert.Contains(t, sql, "b=x")
	assert.Contains(t, sql, `SELECT * FROM "users"`)
}

func TestNamef(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Namef("user:%d", 42).
		Select("*").
		From("users").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, []interface{}{}, args)
	assert.Equal(t,
		`/* name=user:42 */ SELECT * FROM "users"`,
		sql,
	)
}

func TestWithMeta(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		WithMeta(map[string]string{
			"tenant": "eu",
			"user":   "42",
		}).
		Select("*").
		From("users").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, []interface{}{}, args)
	assert.Contains(t, sql, `tenant=eu`)
	assert.Contains(t, sql, `user=42`)
	assert.Contains(t, sql, `SELECT * FROM "users"`)
}

func TestWithContext(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, ContextTraceID, "trace123")
	ctx = context.WithValue(ctx, ContextRequestID, "req456")

	sql, args, err := getPsqlBuilder().
		WithContext(ctx).
		Select("*").
		From("users").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, []interface{}{}, args)
	assert.Contains(t, sql, `trace_id=trace123`)
	assert.Contains(t, sql, `request_id=req456`)
	assert.Contains(t, sql, `SELECT * FROM "users"`)
}

func TestNameMetaContextTogether(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, ContextTraceID, "abc")

	sql, args, err := getPsqlBuilder().
		WithContext(ctx).
		Namef("user:%d", 7).
		AddMeta("tenant", "eu").
		Select("*").
		From("users").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, []interface{}{}, args)
	assert.Contains(t, sql, `name=user:7`)
	assert.Contains(t, sql, `trace_id=abc`)
	assert.Contains(t, sql, `tenant=eu`)
	assert.Contains(t, sql, `SELECT * FROM "users"`)
}

func TestPrefixBasic(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Prefix("EXPLAIN").
		Select("*").
		From("users").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `EXPLAIN SELECT * FROM "users"`, sql)
	assert.Empty(t, args)
}

func TestSuffixBasic(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		Suffix("FOR UPDATE").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT * FROM "users" FOR UPDATE`, sql)
	assert.Empty(t, args)
}

func TestExplainSelect(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Name("explain_users").
		Prefix("EXPLAIN ANALYZE").
		Select("*").
		From("users").
		Suffix("FOR UPDATE").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `/* name=explain_users */ EXPLAIN ANALYZE SELECT * FROM "users" FOR UPDATE`, sql)
	assert.Empty(t, args)
}

func TestPrefixWithArgs(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Prefix("SET LOCAL statement_timeout = ?", 1000).
		Select("*").
		From("users").
		Where("id", "=", 1).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SET LOCAL statement_timeout = ? SELECT * FROM "users" WHERE "id" = $1`, sql)
	assert.Equal(t, []interface{}{1000, 1}, args)
}
