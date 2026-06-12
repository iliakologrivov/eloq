package eloq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJoinSimple(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		Join("posts", "posts.user_id", "=", "users.id").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT * FROM "users" INNER JOIN "posts" ON "posts"."user_id" = "users"."id"`, sql)
	assert.Empty(t, args)
}

func TestLeftJoin(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		LeftJoin("profiles", "profiles.user_id", "=", "users.id").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT * FROM "users" LEFT JOIN "profiles" ON "profiles"."user_id" = "users"."id"`, sql)
	assert.Empty(t, args)
}

func TestRightJoin(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		RightJoin("orders", "orders.user_id", "=", "users.id").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT * FROM "users" RIGHT JOIN "orders" ON "orders"."user_id" = "users"."id"`, sql)
	assert.Empty(t, args)
}

func TestJoinWithCallback(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		JoinWith("orders", func(j *JoinBuilder) {
			j.On("orders.user_id", "=", "users.id").
				Where("orders.status", "paid")
		}).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT * FROM "users" INNER JOIN "orders" ON "orders"."user_id" = "users"."id" AND ("orders"."status" = $1)`, sql)
	assert.Equal(t, []interface{}{"paid"}, args)
}

func TestJoinWithMultipleOns(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("a").
		JoinWith("b", func(j *JoinBuilder) {
			j.On("a.id", "=", "b.a_id").
				OrOn("a.type", "=", "b.type")
		}).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT * FROM "a" INNER JOIN "b" ON "a"."id" = "b"."a_id" OR "a"."type" = "b"."type"`, sql)
	assert.Empty(t, args)
}

func TestJoinWithWhereOrWhere(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		JoinWith("orders", func(j *JoinBuilder) {
			j.On("orders.user_id", "=", "users.id").
				Where("orders.status", "paid").
				OrWhere("orders.type", "vip")
		}).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(
		t,
		`SELECT * FROM "users" INNER JOIN "orders" ON "orders"."user_id" = "users"."id" AND ("orders"."status" = $1 OR "orders"."type" = $2)`,
		sql,
	)
	assert.Equal(t, []interface{}{"paid", "vip"}, args)
}

func TestJoinRaw(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		JoinRaw("JOIN complex_view v ON v.user_id = users.id").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT * FROM "users" JOIN complex_view v ON v.user_id = users.id`, sql)
	assert.Empty(t, args)
}

func TestJoinInvalidIdentifier(t *testing.T) {
	_, _, err := getPsqlBuilder().
		Select("*").
		From("users").
		Join("posts", "posts.user_id;", "=", "users.id").
		ToSql()

	assert.Error(t, err)
}

func TestJoinWithArgs(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		JoinWith("orders", func(j *JoinBuilder) {
			j.On("orders.user_id", "=", "users.id").
				Where("orders.status", "paid")
		}).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(
		t,
		`SELECT * FROM "users" INNER JOIN "orders" ON "orders"."user_id" = "users"."id" AND ("orders"."status" = $1)`,
		sql,
	)
	assert.Equal(t, []interface{}{"paid"}, args)
}

func TestJoinAndWhereArgsOrder(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		JoinWith("orders", func(j *JoinBuilder) {
			j.On("orders.user_id", "=", "users.id").
				Where("orders.status", "paid")
		}).
		Where("users.active", true).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(
		t,
		`SELECT * FROM "users" INNER JOIN "orders" ON "orders"."user_id" = "users"."id" AND ("orders"."status" = $1) WHERE "users"."active" = $2`,
		sql,
	)
	assert.Equal(t, []interface{}{"paid", true}, args)
}

func TestMultipleJoinsArgs(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		JoinWith("orders", func(j *JoinBuilder) {
			j.On("orders.user_id", "=", "users.id").
				Where("orders.status", "paid")
		}).
		JoinWith("profiles", func(j *JoinBuilder) {
			j.On("profiles.user_id", "=", "users.id").
				Where("profiles.active", true)
		}).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(
		t,
		`SELECT * FROM "users" INNER JOIN "orders" ON "orders"."user_id" = "users"."id" AND ("orders"."status" = $1) INNER JOIN "profiles" ON "profiles"."user_id" = "users"."id" AND ("profiles"."active" = $2)`,
		sql,
	)
	assert.Equal(t, []interface{}{"paid", true}, args)
}

func TestJoinWithAndWithAlias_Psql(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("u.id", "p.*").
		From("users", "u").
		JoinWith("profiles", func(j *JoinBuilder) {
			j.As("p").On("p.user_id", "=", "u.id")
		}).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT "u"."id", "p".* FROM "users" AS "u" INNER JOIN "profiles" AS "p" ON "p"."user_id" = "u"."id"`, sql)
	assert.Empty(t, args)

	sql, args, err = getMysqlBuilder().
		Select("u.id", "p.*").
		From("users", "u").
		JoinWith("profiles", func(j *JoinBuilder) {
			j.As("p").On("p.user_id", "=", "u.id")
		}).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, "SELECT `u`.`id`, `p`.* FROM `users` AS `u` INNER JOIN `profiles` AS `p` ON `p`.`user_id` = `u`.`id`", sql)
	assert.Empty(t, args)
}

func TestComplexQueryWithAliases_Psql(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("u.id", "u.name", "u.status", "u.created_at", "o.*").
		From("users", "u").
		JoinWith("profiles", func(j *JoinBuilder) {
			j.As("p").On("p.user_id", "=", "u.id")
		}).
		JoinWith("orders", func(j *JoinBuilder) {
			j.As("o").On("o.profile_id", "=", "p.id")
		}).
		WhereNull("p.deleted_at").
		Where("p.status", "=", "active").
		Where("u.status", "=", "active").
		Where("o.status", "=", "active").
		OrderBy("p.priority", "ASC").
		OrderBy("u.id", "ASC").
		Limit(5).
		Suffix("FOR UPDATE OF u SKIP LOCKED").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t,
		`SELECT "u"."id", "u"."name", "u"."status", "u"."created_at", "o".* FROM "users" AS "u" INNER JOIN "profiles" AS "p" ON "p"."user_id" = "u"."id" INNER JOIN "orders" AS "o" ON "o"."profile_id" = "p"."id" WHERE "p"."deleted_at" IS NULL AND "p"."status" = $1 AND "u"."status" = $2 AND "o"."status" = $3 ORDER BY "p"."priority" ASC, "u"."id" ASC LIMIT 5 FOR UPDATE OF u SKIP LOCKED`,
		sql,
	)
	assert.Equal(t, []interface{}{"active", "active", "active"}, args)

	sql, args, err = getMysqlBuilder().
		Select("u.id", "u.name", "u.status", "u.created_at", "o.*").
		From("users", "u").
		JoinWith("profiles", func(j *JoinBuilder) {
			j.As("p").On("p.user_id", "=", "u.id")
		}).
		JoinWith("orders", func(j *JoinBuilder) {
			j.As("o").On("o.profile_id", "=", "p.id")
		}).
		WhereNull("p.deleted_at").
		Where("p.status", "=", "active").
		Where("u.status", "=", "active").
		Where("o.status", "=", "active").
		OrderBy("p.priority", "ASC").
		OrderBy("u.id", "ASC").
		Limit(5).
		Suffix("FOR UPDATE OF u SKIP LOCKED").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t,
		"SELECT `u`.`id`, `u`.`name`, `u`.`status`, `u`.`created_at`, `o`.* FROM `users` AS `u` INNER JOIN `profiles` AS `p` ON `p`.`user_id` = `u`.`id` INNER JOIN `orders` AS `o` ON `o`.`profile_id` = `p`.`id` WHERE `p`.`deleted_at` IS NULL AND `p`.`status` = ? AND `u`.`status` = ? AND `o`.`status` = ? ORDER BY `p`.`priority` ASC, `u`.`id` ASC LIMIT 5 FOR UPDATE OF u SKIP LOCKED",
		sql,
	)
	assert.Equal(t, []interface{}{"active", "active", "active"}, args)
}

func TestJoinWithoutAlias_Psql(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		Join("posts", "posts.user_id", "=", "users.id").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT * FROM "users" INNER JOIN "posts" ON "posts"."user_id" = "users"."id"`, sql)
	assert.Empty(t, args)
}

func TestMultipleJoinsAndWhere(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		JoinWith("orders", func(j *JoinBuilder) {
			j.On("orders.user_id", "=", "users.id").
				Where("orders.status", "paid")
		}).
		RightJoin("accounts", "accounts.user_id", "=", "users.id").
		LeftJoinWith("profiles", func(j *JoinBuilder) {
			j.On("profiles.user_id", "=", "users.id").
				Where("profiles.active", true)
		}).
		Where("users.age", ">=", 10).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(
		t,
		`SELECT * FROM "users" INNER JOIN "orders" ON "orders"."user_id" = "users"."id" AND ("orders"."status" = $1) RIGHT JOIN "accounts" ON "accounts"."user_id" = "users"."id" LEFT JOIN "profiles" ON "profiles"."user_id" = "users"."id" AND ("profiles"."active" = $2) WHERE "users"."age" >= $3`,
		sql,
	)
	assert.Equal(t, []interface{}{"paid", true, 10}, args)
}
