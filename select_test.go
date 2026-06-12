package eloq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelectRaw_Psql(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		SelectRaw("COUNT(*) AS total_users", "NOW() AS current_time").
		From("users").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, sql, `SELECT COUNT(*) AS total_users, NOW() AS current_time FROM "users"`)
	assert.Equal(t, []interface{}{}, args)
}

func TestSelectEmptyTable_Psql(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select().
		Table("").
		ToSql()

	assert.NotEmpty(t, err)
	assert.Empty(t, sql)
	assert.Equal(t, []interface{}{}, args)
}

func TestSelectAndSelectRaw_Psql(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("id", "name").
		AddSelectRaw("COUNT(*) AS total_users", "NOW() AS current_time").
		From("users").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, sql, `SELECT "id", "name", COUNT(*) AS total_users, NOW() AS current_time FROM "users"`)
	assert.Equal(t, []interface{}{}, args)
}

func TestSelectEmpty_Psql(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select().
		From("users").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, sql, `SELECT * FROM "users"`)
	assert.Equal(t, []interface{}{}, args)
}

func TestSelectRaw_Mysql(t *testing.T) {
	sql, args, err := getMysqlBuilder().
		SelectRaw("COUNT(*) AS total_users", "NOW() AS current_time").
		Table("users").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, sql, "SELECT COUNT(*) AS total_users, NOW() AS current_time FROM `users`")
	assert.Equal(t, []interface{}{}, args)
}

func TestSelectEmptyTable_Mysql(t *testing.T) {
	sql, args, err := getMysqlBuilder().
		SelectRaw("COUNT(*) AS total_users", "NOW() AS current_time").
		Table("").
		ToSql()

	assert.Error(t, err)
	assert.Empty(t, sql)
	assert.Equal(t, []interface{}{}, args)
}

func TestSelectEmpty_Mysql(t *testing.T) {
	sql, args, err := getMysqlBuilder().
		Select().
		From("users").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, sql, "SELECT * FROM `users`")
	assert.Equal(t, []interface{}{}, args)
}

func TestSelectAndSelectRaw_Mysql(t *testing.T) {
	sql, args, err := getMysqlBuilder().
		Select("id", "name").
		AddSelectRaw("COUNT(*) AS total_users", "NOW() AS current_time").
		Table("users").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, sql, "SELECT `id`, `name`, COUNT(*) AS total_users, NOW() AS current_time FROM `users`")
	assert.Empty(t, args)
}

func TestSelectFromAlias_Psql(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("t.id").
		From("table", "t").
		OrderByDesc("t.id").
		Limit(1).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT "t"."id" FROM "table" AS "t" ORDER BY "t"."id" DESC LIMIT 1`, sql)
	assert.Equal(t, []interface{}{}, args)
}

func TestSelectFromAlias_Mysql(t *testing.T) {
	sql, args, err := getMysqlBuilder().
		Select("t.id").
		From("table", "t").
		OrderByDesc("t.id").
		Limit(1).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, "SELECT `t`.`id` FROM `table` AS `t` ORDER BY `t`.`id` DESC LIMIT 1", sql)
	assert.Equal(t, []interface{}{}, args)
}

func TestSelectTableAlias_Psql(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("*").
		Table("users", "u").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT * FROM "users" AS "u"`, sql)
	assert.Equal(t, []interface{}{}, args)
}

func TestSelectTableAlias_Mysql(t *testing.T) {
	sql, args, err := getMysqlBuilder().
		Select("*").
		Table("users", "u").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, "SELECT * FROM `users` AS `u`", sql)
	assert.Equal(t, []interface{}{}, args)
}

func TestSelectFromNoAlias_Psql(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT * FROM "users"`, sql)
	assert.Equal(t, []interface{}{}, args)
}

func TestSelectFromEmptyAlias_Psql(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users", "").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT * FROM "users"`, sql)
	assert.Equal(t, []interface{}{}, args)
}

func TestSelectFromAliasWithWhere_Psql(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("t.id", "t.name").
		From("users", "t").
		Where("t.status", "=", "active").
		OrderByDesc("t.id").
		Limit(1).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT "t"."id", "t"."name" FROM "users" AS "t" WHERE "t"."status" = $1 ORDER BY "t"."id" DESC LIMIT 1`, sql)
	assert.Equal(t, []interface{}{"active"}, args)
}

func TestSelectError_Psql(t *testing.T) {
	sql, args, err := getMysqlBuilder().
		Select("id; DROP TABLE users").
		Table("users").
		ToSql()

	assert.ErrorContains(t, err, "invalid identifier:")
	assert.Empty(t, sql)
	assert.Empty(t, args)
}
