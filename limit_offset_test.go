package eloq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelectLimit_Psql(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		Limit(50).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT * FROM "users" LIMIT 50`, sql)
	assert.Equal(t, []interface{}{}, args)
}

func TestSelectOffset_Psql(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		Offset(50).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT * FROM "users" OFFSET 50`, sql)
	assert.Equal(t, []interface{}{}, args)
}

func TestSelectLimitOffset_Psql(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		Limit(50).
		Offset(150).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT * FROM "users" LIMIT 50 OFFSET 150`, sql)
	assert.Equal(t, []interface{}{}, args)
}

func TestSelectLimit_Mysql(t *testing.T) {
	sql, args, err := getMysqlBuilder().
		Select("*").
		From("users").
		Limit(50).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, "SELECT * FROM `users` LIMIT 50", sql)
	assert.Equal(t, []interface{}{}, args)
}

func TestSelectOffset_Mysql(t *testing.T) {
	sql, args, err := getMysqlBuilder().
		Select("*").
		From("users").
		Offset(50).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, "SELECT * FROM `users` OFFSET 50", sql)
	assert.Equal(t, []interface{}{}, args)
}

func TestSelectLimitOffset_Mysql(t *testing.T) {
	sql, args, err := getMysqlBuilder().
		Select("*").
		From("users").
		Limit(50).
		Offset(150).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, "SELECT * FROM `users` LIMIT 50 OFFSET 150", sql)
	assert.Equal(t, []interface{}{}, args)
}
