package eloq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInsertBasic(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Insert("users").
		Values(map[string]interface{}{
			"email": "a@b.com",
			"age":   42,
		}).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t,
		`INSERT INTO "users" ("age", "email") VALUES ($1, $2)`,
		sql,
	)
	assert.Equal(t, []interface{}{42, "a@b.com"}, args)
}

func TestInsertMany(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Insert("users").
		Values(map[string]interface{}{
			"email": "a@b.com",
			"age":   42,
		}).
		Values(map[string]interface{}{
			"email": "v@b.com",
			"age":   3,
		}).
		Values(map[string]interface{}{
			"email": "e@b.com",
			"age":   55,
		}).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `INSERT INTO "users" ("age", "email") VALUES ($1, $2), ($3, $4), ($5, $6)`, sql)
	assert.Equal(t, []interface{}{42, "a@b.com", 3, "v@b.com", 55, "e@b.com"}, args)
}

func TestInsertEmpty(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Insert("users").
		ToSql()

	assert.Error(t, err)
	assert.Empty(t, sql)
	assert.Nil(t, args)
}

func TestInsertEmptyValues(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Insert("users").
		Values(map[string]interface{}{}).
		ToSql()

	assert.ErrorContains(t, err, "eloq: insert has empty values")
	assert.Empty(t, sql)
	assert.Empty(t, args)
}

func TestInsertWithMissingColumns(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Insert("users").
		Values(map[string]interface{}{"a": 1}).
		Values(map[string]interface{}{"b": 2}).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `INSERT INTO "users" ("a", "b") VALUES ($1, DEFAULT), (DEFAULT, $2)`, sql)
	assert.Equal(t, []interface{}{1, 2}, args)
}

func TestInsertReturning(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Insert("users").
		Values(map[string]interface{}{"email": "a@b.com"}).
		Returning("id").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t,
		`INSERT INTO "users" ("email") VALUES ($1) RETURNING "id"`,
		sql,
	)
	assert.Equal(t, []interface{}{"a@b.com"}, args)
}

func TestInsertReturningMysql(t *testing.T) {
	sql, args, err := getMysqlBuilder().
		Insert("users").
		Values(map[string]interface{}{"email": "a@b.com"}).
		Returning("id").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, "INSERT INTO `users` (`email`) VALUES (?) RETURNING `id`", sql)
	assert.Equal(t, []interface{}{"a@b.com"}, args)
}

func TestInsertOnConflictUpdate(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Insert("users").
		Values(map[string]interface{}{
			"email": "a@b.com",
			"name":  "John",
		}).
		OnConflict("email").
		DoUpdate(map[string]interface{}{
			"name": "John",
		}).
		Returning("email").
		ToSql()

	assert.NoError(t, err)
	assert.Contains(t, sql, `INSERT INTO "users" ("email", "name") VALUES ($1, $2) ON CONFLICT ("email") DO UPDATE SET "name" = $3 RETURNING "email"`)
	assert.Equal(t, []interface{}{"a@b.com", "John", "John"}, args)
}

func TestInsertOnConflictUpdate_DeterministicOrder(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Insert("users").
		Values(map[string]interface{}{
			"email": "a@b.com",
			"name":  "John",
		}).
		OnConflict("email").
		DoUpdate(map[string]interface{}{
			"name": "John Updated",
			"role": "admin",
		}).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(
		t,
		`INSERT INTO "users" ("email", "name") VALUES ($1, $2) ON CONFLICT ("email") DO UPDATE SET "name" = $3, "role" = $4`,
		sql,
	)
	assert.Equal(t, []interface{}{"a@b.com", "John", "John Updated", "admin"}, args)
}
