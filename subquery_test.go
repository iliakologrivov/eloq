package eloq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelectFromSubqueryWithUnionAll(t *testing.T) {
	sub := getPsqlBuilder().
		SelectRaw(`id`, `email AS "name"`).
		From("user").
		UnionAll(
			getPsqlBuilder().
				Select("id", "name").
				From("admins"),
		)

	sql, args, err := getPsqlBuilder().
		Select("id", "name").
		FromSub(sub, "u").
		OrderByDesc("id").
		Limit(10).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(
		t,
		`SELECT "id", "name" FROM (SELECT id, email AS "name" FROM "user" UNION ALL (SELECT "id", "name" FROM "admins")) AS "u" ORDER BY "id" DESC LIMIT 10`,
		sql,
	)
	assert.Empty(t, args)
}

func TestWhereInSubquery(t *testing.T) {
	sub := getPsqlBuilder().
		Select("user_id").
		From("orders").
		Where("status", "done").
		Where("amount", ">=", 100)

	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		WhereInSub("id", sub).
		OrderByDesc("id").
		Limit(10).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(
		t,
		`SELECT * FROM "users" WHERE "id" IN (SELECT "user_id" FROM "orders" WHERE "status" = $1 AND "amount" >= $2) ORDER BY "id" DESC LIMIT 10`,
		sql,
	)
	assert.Equal(t, []interface{}{"done", 100}, args)
}

func TestWhereInSubquery_PlaceholderOffset(t *testing.T) {
	sub := getPsqlBuilder().
		Select("user_id").
		From("orders").
		Where("status", "done")

	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		Where("tenant_id", 7).
		WhereInSub("id", sub).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(
		t,
		`SELECT * FROM "users" WHERE "tenant_id" = $1 AND "id" IN (SELECT "user_id" FROM "orders" WHERE "status" = $2)`,
		sql,
	)
	assert.Equal(t, []interface{}{7, "done"}, args)
}

func TestFromSubqueryAliasRequired(t *testing.T) {
	sub := getPsqlBuilder().
		Select("id").
		From("users")

	sql, args, err := getPsqlBuilder().
		Select("id").
		FromSub(sub, "").
		ToSql()

	assert.ErrorIs(t, err, ErrEmptySubqueryAlias)
	assert.Empty(t, sql)
	assert.Empty(t, args)
}

func TestWhereSubquery(t *testing.T) {
	sub := getPsqlBuilder().
		Select("user_id").
		From("orders").
		Where("status", "done")

	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		Where("tenant_id", 7).
		Where("id", "in", sub).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT * FROM "users" WHERE "tenant_id" = $1 AND "id" IN (SELECT "user_id" FROM "orders" WHERE "status" = $2)`, sql)
	assert.Equal(t, []interface{}{7, "done"}, args)
}
