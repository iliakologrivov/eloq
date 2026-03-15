package eloq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrderDefault_Psql(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		OrderBy("id").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT * FROM "users" ORDER BY "id" ASC`, sql)
	assert.Equal(t, []interface{}{}, args)
}

func TestOrderAsc_Psql(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		OrderByAsc("id").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT * FROM "users" ORDER BY "id" ASC`, sql)
	assert.Equal(t, []interface{}{}, args)
}

func TestOrderDesc_Psql(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		OrderByDesc("id").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT * FROM "users" ORDER BY "id" DESC`, sql)
	assert.Equal(t, []interface{}{}, args)
}

func TestOrderRandom_Psql(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		OrderByRaw("RANDOM()").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT * FROM "users" ORDER BY RANDOM()`, sql)
	assert.Equal(t, []interface{}{}, args)
}

func TestOrderDefault_Mysql(t *testing.T) {
	sql, args, err := getMysqlBuilder().
		Select("*").
		From("users").
		OrderBy("id").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, "SELECT * FROM `users` ORDER BY `id` ASC", sql)
	assert.Equal(t, []interface{}{}, args)
}

func TestOrderAsc_Mysql(t *testing.T) {
	sql, args, err := getMysqlBuilder().
		Select("*").
		From("users").
		OrderByAsc("id").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, "SELECT * FROM `users` ORDER BY `id` ASC", sql)
	assert.Equal(t, []interface{}{}, args)
}

func TestOrderDesc_Mysql(t *testing.T) {
	sql, args, err := getMysqlBuilder().
		Select("*").
		From("users").
		OrderByDesc("id").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, "SELECT * FROM `users` ORDER BY `id` DESC", sql)
	assert.Equal(t, []interface{}{}, args)
}

func TestOrderRandom_Mysql(t *testing.T) {
	sql, args, err := getMysqlBuilder().
		Select("*").
		From("users").
		OrderByRaw("RANDOM()").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, "SELECT * FROM `users` ORDER BY RANDOM()", sql)
	assert.Equal(t, []interface{}{}, args)
}
