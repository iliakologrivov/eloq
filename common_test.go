package eloq

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getPsqlBuilder() *StatementBuilder {
	return NewStatementBuilder().
		PlaceholderFormat(Dollar).
		QuoteWith(DoubleQuote)
}

func getMysqlBuilder() *StatementBuilder {
	return NewStatementBuilder().
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
		Select("*").
		Comment("hello").
		CommentKV("a", 1).
		From("users").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, []interface{}{}, args)
	assert.Equal(t, `/* hello | a=1 */ SELECT * FROM "users"`, sql)
}

func TestRenderCommentVK(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("*").
		Comment("hello").
		CommentKV("a", 1, "k").
		From("users").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, []interface{}{}, args)
	assert.Equal(t, `/* hello */ SELECT * FROM "users"`, sql)
}

func TestRenderCommentSanitize(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("*").
		Comment("*/").
		From("users").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, []interface{}{}, args)
	assert.Equal(t, `SELECT * FROM "users"`, sql)
}

func TestQueryName(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("*").
		Name("load_users").
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
		Select("*").
		Name("").
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
		Select("*").
		AddMeta("a", 1).
		AddMeta("b", "x").
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
		Select("*").
		Namef("user:%d", 42).
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
		Select("*").
		WithMeta(map[string]string{
			"tenant": "eu",
			"user":   "42",
		}).
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
		Select("*").
		WithContext(ctx).
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
		Select("*").
		WithContext(ctx).
		Namef("user:%d", 7).
		AddMeta("tenant", "eu").
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
		Select("*").
		Prefix("EXPLAIN").
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
		Select("*").
		Name("explain_users").
		Prefix("EXPLAIN ANALYZE").
		From("users").
		Suffix("FOR UPDATE").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `/* name=explain_users */ EXPLAIN ANALYZE SELECT * FROM "users" FOR UPDATE`, sql)
	assert.Empty(t, args)
}

func TestPrefixWithArgs(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("*").
		Prefix("SET LOCAL statement_timeout = ?", 1000).
		From("users").
		Where("id", "=", 1).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SET LOCAL statement_timeout = $1 SELECT * FROM "users" WHERE "id" = $2`, sql)
	assert.Equal(t, []interface{}{1000, 1}, args)
}
