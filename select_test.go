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

func TestSelectError_Psql(t *testing.T) {
	sql, args, err := getMysqlBuilder().
		Select("id; DROP TABLE users").
		Table("users").
		ToSql()

	assert.ErrorContains(t, err, "invalid identifier:")
	assert.Empty(t, sql)
	assert.Empty(t, args)
}
