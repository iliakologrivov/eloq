package eloq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnion_Psql(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("id").
		From("users").
		Where("active", true).
		Union(
			getPsqlBuilder().
				Select("id").
				From("admins").
				Where("enabled", true),
		).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(
		t,
		`SELECT "id" FROM "users" WHERE "active" = $1 UNION (SELECT "id" FROM "admins" WHERE "enabled" = $2)`,
		sql,
	)
	assert.Equal(t, []interface{}{true, true}, args)
}

func TestUnionAll_OrderLimit_Psql(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("id").
		From("users").
		UnionAll(
			getPsqlBuilder().
				Select("id").
				From("guests"),
		).
		OrderByDesc("id").
		Limit(10).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(
		t,
		`SELECT "id" FROM "users" UNION ALL (SELECT "id" FROM "guests") ORDER BY "id" DESC LIMIT 10`,
		sql,
	)
	assert.Empty(t, args)
}

func TestUnion_MysqlArgsOrder(t *testing.T) {
	sql, args, err := getMysqlBuilder().
		Select("id").
		From("users").
		Where("active", true).
		Union(
			getMysqlBuilder().
				Select("id").
				From("admins").
				Where("enabled", false),
		).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(
		t,
		"SELECT `id` FROM `users` WHERE `active` = ? UNION (SELECT `id` FROM `admins` WHERE `enabled` = ?)",
		sql,
	)
	assert.Equal(t, []interface{}{true, false}, args)
}

func TestUnion_ErrorFromNestedQuery(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("id").
		From("users").
		Union(
			getPsqlBuilder().
				Select("id"),
		).
		ToSql()

	assert.Error(t, err)
	assert.Empty(t, sql)
	assert.Empty(t, args)
}
