package eloq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeleteBasic(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Delete("users").
		Where("id", 11).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `DELETE FROM "users" WHERE "id" = $1`, sql)
	assert.Equal(t, []interface{}{11}, args)
}

func TestDeleteWithoutWhere(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Delete("users").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `DELETE FROM "users"`, sql)
	assert.Empty(t, args)
}

func TestDeleteWhereOr(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Delete("users").
		Where("id", 1).
		OrWhere("id", 2).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `DELETE FROM "users" WHERE "id" = $1 OR "id" = $2`, sql)
	assert.Equal(t, []interface{}{1, 2}, args)
}

func TestDeleteWhereIn(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Delete("users").
		WhereIn("id", 1, 2).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `DELETE FROM "users" WHERE "id" IN ($1, $2)`, sql)
	assert.Equal(t, []interface{}{1, 2}, args)
}

func TestDeleteWhereInEmpty(t *testing.T) {
	ids := []interface{}{}
	sql, args, err := getPsqlBuilder().
		Delete("users").
		WhereIn("id", ids...).
		ToSql()

	assert.Error(t, err)
	assert.Empty(t, sql)
	assert.Empty(t, args)
}

func TestDeleteWhereLikeStringBasic(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Delete("users").
		Where("name", "like", "an%").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `DELETE FROM "users" WHERE "name" LIKE $1`, sql)
	assert.Equal(t, []interface{}{"an%"}, args)
}

func TestDeleteWhereInIntEmpty(t *testing.T) {
	ids := []int{}
	sql, args, err := getPsqlBuilder().
		Delete("users").
		WhereIn("id", ToAnySlice(ids)...).
		ToSql()

	assert.Error(t, err)
	assert.Empty(t, sql)
	assert.Empty(t, args)
}

func TestDeleteWhereNotIn(t *testing.T) {
	ids := []interface{}{1, 2, 3, 4, 5}
	sql, args, err := getPsqlBuilder().
		Delete("users").
		WhereNotIn("id", ids...).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `DELETE FROM "users" WHERE "id" NOT IN ($1, $2, $3, $4, $5)`, sql)
	assert.Equal(t, ids, args)
}

func TestDeleteWithExplain(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Delete("users").
		Prefix("EXPLAIN").
		Where("id", 1).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `EXPLAIN DELETE FROM "users" WHERE "id" = $1`, sql)
	assert.Equal(t, []interface{}{1}, args)
}

func TestDeleteWithJoin(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Delete("users").
		JoinWith("orders", func(j *JoinBuilder) {
			j.On("orders.user_id", "=", "users.id")
		}).
		Where("orders.status", "paid").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `DELETE FROM "users" INNER JOIN "orders" ON "orders"."user_id" = "users"."id" WHERE "orders"."status" = $1`, sql)
	assert.Equal(t, []interface{}{"paid"}, args)
}

func TestDeleteUsing(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Delete("users").
		Using("orders").
		Where("orders.user_id", "=", Raw(`"users"."id"`)).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `DELETE FROM "users" USING "orders" WHERE "orders"."user_id" = "users"."id"`, sql)
	assert.Empty(t, args)
}

func TestDeleteRequireWhere(t *testing.T) {
	builder := NewStatementBuilder().
		PlaceholderFormat(Dollar).
		QuoteWith(DoubleQuote).
		RequireWhere(true)

	sql, args, err := builder.Delete("users").ToSql()

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrRequireWhere)
	assert.Empty(t, sql)
	assert.Nil(t, args)
}

func TestDeleteRequireWhere_WithWhere(t *testing.T) {
	builder := NewStatementBuilder().
		PlaceholderFormat(Dollar).
		QuoteWith(DoubleQuote).
		RequireWhere(true)

	sql, args, err := builder.Delete("users").Where("id", 1).ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `DELETE FROM "users" WHERE "id" = $1`, sql)
	assert.Equal(t, []interface{}{1}, args)
}

func TestDeleteRequireWhere_WithExplicitAll(t *testing.T) {
	builder := NewStatementBuilder().
		PlaceholderFormat(Dollar).
		QuoteWith(DoubleQuote).
		RequireWhere(true)

	sql, args, err := builder.Delete("users").Where("1", "=", Raw("1")).ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `DELETE FROM "users" WHERE "1" = 1`, sql)
	assert.Empty(t, args)
}
