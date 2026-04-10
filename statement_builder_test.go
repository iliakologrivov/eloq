package eloq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatementBuilder_Select(t *testing.T) {
	sb := NewStatementBuilder().
		PlaceholderFormat(Dollar).
		QuoteWith(DoubleQuote)

	sql, args, err := sb.Select("*").
		From("users").
		Where("id", "=", 1).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT * FROM "users" WHERE "id" = $1`, sql)
	assert.Equal(t, []interface{}{1}, args)
}

func TestStatementBuilder_SelectRaw(t *testing.T) {
	sb := NewStatementBuilder().
		PlaceholderFormat(Dollar).
		QuoteWith(DoubleQuote)

	sql, args, err := sb.SelectRaw("COUNT(*) as cnt").
		From("users").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT COUNT(*) as cnt FROM "users"`, sql)
	assert.Empty(t, args)
}

func TestStatementBuilder_Insert(t *testing.T) {
	sb := NewStatementBuilder().
		PlaceholderFormat(Dollar).
		QuoteWith(DoubleQuote)

	sql, args, err := sb.Insert("users").
		Values(map[string]interface{}{"name": "John", "email": "john@example.com"}).
		ToSql()

	assert.NoError(t, err)
	assert.Contains(t, sql, `INSERT INTO "users"`)
	assert.Contains(t, sql, `"name"`)
	assert.Contains(t, sql, `"email"`)
	assert.Len(t, args, 2)
}

func TestStatementBuilder_Update(t *testing.T) {
	sb := NewStatementBuilder().
		PlaceholderFormat(Dollar).
		QuoteWith(DoubleQuote)

	sql, args, err := sb.Update("users").
		Set("name", "Jane").
		Where("id", "=", 1).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `UPDATE "users" SET "name" = $1 WHERE "id" = $2`, sql)
	assert.Equal(t, []interface{}{"Jane", 1}, args)
}

func TestStatementBuilder_Delete(t *testing.T) {
	sb := NewStatementBuilder().
		PlaceholderFormat(Dollar).
		QuoteWith(DoubleQuote)

	sql, args, err := sb.Delete("users").
		Where("id", "=", 1).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `DELETE FROM "users" WHERE "id" = $1`, sql)
	assert.Equal(t, []interface{}{1}, args)
}

func TestStatementBuilder_MySQL(t *testing.T) {
	sb := NewStatementBuilder().
		PlaceholderFormat(Question).
		QuoteWith(Backtick)

	sql, args, err := sb.Select("*").
		From("users").
		Where("id", "=", 1).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, "SELECT * FROM `users` WHERE `id` = ?", sql)
	assert.Equal(t, []interface{}{1}, args)
}

func TestDefaultStatementBuilder(t *testing.T) {
	sql, args, err := DefaultStatementBuilder.Select("*").
		From("users").
		Where("id", "=", 1).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, "SELECT * FROM `users` WHERE `id` = ?", sql)
	assert.Equal(t, []interface{}{1}, args)
}

func TestStatementBuilder_PrefixSuffix(t *testing.T) {
	sb := NewStatementBuilder().
		PlaceholderFormat(Dollar).
		QuoteWith(DoubleQuote)

	sql, args, err := sb.Select("*").
		From("users").
		Prefix("EXPLAIN").
		Suffix("FOR UPDATE").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `EXPLAIN SELECT * FROM "users" FOR UPDATE`, sql)
	assert.Empty(t, args)
}

func TestStatementBuilder_PrefixWithArgs(t *testing.T) {
	sb := NewStatementBuilder().
		PlaceholderFormat(Dollar).
		QuoteWith(DoubleQuote)

	sql, args, err := sb.Select("*").
		From("users").
		Where("id", "=", 2).
		Prefix("SET LOCAL statement_timeout = ?", 1000).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SET LOCAL statement_timeout = $1 SELECT * FROM "users" WHERE "id" = $2`, sql)
	assert.Equal(t, []interface{}{1000, 2}, args)
}

func TestStatementBuilder_SuffixWithArgs(t *testing.T) {
	sb := NewStatementBuilder().
		PlaceholderFormat(Dollar).
		QuoteWith(DoubleQuote)

	sql, args, err := sb.Select("*").
		From("users").
		Where("id", "=", 1).
		Suffix("LIMIT ?", 10).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT * FROM "users" WHERE "id" = $1 LIMIT $2`, sql)
	assert.Equal(t, []interface{}{1, 10}, args)
}

func TestRepositoryPattern(t *testing.T) {
	type Repository struct {
		psql StatementBuilderType
	}

	repo := &Repository{
		psql: NewStatementBuilder().
			PlaceholderFormat(Dollar).
			QuoteWith(DoubleQuote),
	}

	sql, args, err := repo.psql.Select("*").
		From("users").
		Where("id", "=", 42).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT * FROM "users" WHERE "id" = $1`, sql)
	assert.Equal(t, []interface{}{42}, args)
}
