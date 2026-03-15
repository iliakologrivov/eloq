package eloq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateBasic(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Update("users").
		Set("name", "John").
		Set("age", 30).
		Where("id", 1).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `UPDATE "users" SET "age" = $1, "name" = $2 WHERE "id" = $3`, sql)
	assert.Equal(t, []interface{}{30, "John", 1}, args)
}

func TestUpdateRaw(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Update("users").
		Set("updated_at", Raw("NOW()")).
		Where("id", 1).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, sql, `UPDATE "users" SET "updated_at" = NOW() WHERE "id" = $1`)
	assert.Equal(t, []interface{}{1}, args)
}

func TestUpdateMap(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Update("users").
		SetMap(map[string]interface{}{
			"name":       "John",
			"age":        30,
			"updated_at": Raw("NOW()"),
		}).
		Where("id", 1).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, sql, `UPDATE "users" SET "age" = $1, "name" = $2, "updated_at" = NOW() WHERE "id" = $3`)
	assert.Equal(t, []interface{}{30, "John", 1}, args)
}

func TestUpdateEmpty(t *testing.T) {
	_, _, err := getPsqlBuilder().
		Update("users").
		ToSql()

	assert.Error(t, err)
}
