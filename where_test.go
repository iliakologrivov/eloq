package eloq

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWhereEqualStringBasic1_Mysql(t *testing.T) {
	sql, args, err := getMysqlBuilder().
		Select("*").
		From("users").
		Where("name", "admin").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, "SELECT * FROM `users` WHERE `name` = ?", sql)
	assert.Equal(t, []interface{}{"admin"}, args)
}

func TestWhereEqualStringBasic2_Mysql(t *testing.T) {
	sql, args, err := getMysqlBuilder().
		Select("*").
		From("users").
		Where("name", "=", "admin").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, "SELECT * FROM `users` WHERE `name` = ?", sql)
	assert.Equal(t, []interface{}{"admin"}, args)
}

func TestWhereEqualIntBasic1_Mysql(t *testing.T) {
	sql, args, err := getMysqlBuilder().
		Select("*").
		From("users").
		Where("id", 1).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, "SELECT * FROM `users` WHERE `id` = ?", sql)
	assert.Equal(t, []interface{}{1}, args)
}

func TestWhereEqualIntBasic2_Mysql(t *testing.T) {
	sql, args, err := getMysqlBuilder().
		Select("*").
		From("users").
		Where("id", "=", 1).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, "SELECT * FROM `users` WHERE `id` = ?", sql)
	assert.Equal(t, []interface{}{1}, args)
}

func TestWhereEqualStringBasic1_Psql(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		Where("name", "admin").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT * FROM "users" WHERE "name" = $1`, sql)
	assert.Equal(t, []interface{}{"admin"}, args)
}

func TestWhereEqualStringBasic2_Psql(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		Where("name", "=", "admin").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT * FROM "users" WHERE "name" = $1`, sql)
	assert.Equal(t, []interface{}{"admin"}, args)
}

func TestWhereEqualIntBasic1_Psql(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		Where("id", 1).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT * FROM "users" WHERE "id" = $1`, sql)
	assert.Equal(t, []interface{}{1}, args)
}

func TestWhereEqualIntBasic2_Psql(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		Where("id", "=", 1).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT * FROM "users" WHERE "id" = $1`, sql)
	assert.Equal(t, []interface{}{1}, args)
}

func TestWhereEqualNilBasic1_Psql(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		Where("deleted", nil).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT * FROM "users" WHERE "deleted" IS NULL`, sql)
	assert.Equal(t, []interface{}{}, args)
}

func TestWhereEqualNilBasic2_Psql(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		Where("deleted", "=", nil).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT * FROM "users" WHERE "deleted" IS NULL`, sql)
	assert.Equal(t, []interface{}{}, args)
}

func TestWhereNotEqualNilBasic1_Psql(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		Where("deleted", "!=", nil).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT * FROM "users" WHERE "deleted" IS NOT NULL`, sql)
	assert.Equal(t, []interface{}{}, args)
}

func TestWhereNotEqualNilBasic2_Psql(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		Where("deleted", "<>", nil).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT * FROM "users" WHERE "deleted" IS NOT NULL`, sql)
	assert.Equal(t, []interface{}{}, args)
}

func TestWhereGreaterIntBasic2_Psql(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		Where("id", ">", 1).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT * FROM "users" WHERE "id" > $1`, sql)
	assert.Equal(t, []interface{}{1}, args)
}

func TestWhereGreaterOrEqualIntBasic2_Psql(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		Where("id", ">=", 1).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT * FROM "users" WHERE "id" >= $1`, sql)
	assert.Equal(t, []interface{}{1}, args)
}

func TestWhereLessIntBasic2_Psql(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		Where("id", "<", 1).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT * FROM "users" WHERE "id" < $1`, sql)
	assert.Equal(t, []interface{}{1}, args)
}

func TestWhereLessOrEqualIntBasic2_Psql(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		Where("id", "<=", 1).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT * FROM "users" WHERE "id" <= $1`, sql)
	assert.Equal(t, []interface{}{1}, args)
}

func TestWhereLikeStringBasic_Psql(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		Where("name", "like", "an%").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT * FROM "users" WHERE "name" LIKE $1`, sql)
	assert.Equal(t, []interface{}{"an%"}, args)
}

func TestWhereIn_Psql(t *testing.T) {
	ids := []interface{}{1, 2, 3, 4, 5}
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		WhereIn("id", 1, 2, 3, 4, 5).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT * FROM "users" WHERE "id" IN ($1, $2, $3, $4, $5)`, sql)
	assert.Equal(t, ids, args)
}

func TestWhereInEmpty_Psql(t *testing.T) {
	ids := []interface{}{}
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		WhereIn("id", ids...).
		ToSql()

	assert.Error(t, err)
	assert.Empty(t, sql)
	assert.Empty(t, args)
}

func TestWhereInInt_Psql(t *testing.T) {
	ids := []int{1, 2, 3, 4, 5}
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		WhereIn("id", ToAnySlice(ids)...).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT * FROM "users" WHERE "id" IN ($1, $2, $3, $4, $5)`, sql)
	assert.Equal(t, ToAnySlice(ids), args)
}

func TestWhereInIntEmpty_Psql(t *testing.T) {
	ids := []int{}
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		WhereIn("id", ToAnySlice(ids)...).
		ToSql()

	assert.Error(t, err)
	assert.Empty(t, sql)
	assert.Empty(t, args)
}

func TestWhereNotIn_Psql(t *testing.T) {
	ids := []interface{}{1, 2, 3, 4, 5}
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		WhereNotIn("id", ids...).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT * FROM "users" WHERE "id" NOT IN ($1, $2, $3, $4, $5)`, sql)
	assert.Equal(t, ids, args)
}

func TestWhereNotInEmpty_Psql(t *testing.T) {
	ids := []interface{}{}
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		WhereNotIn("id", ids...).
		ToSql()

	assert.Error(t, err)
	assert.Empty(t, sql)
	assert.Empty(t, args)
}

func TestWhereNull_Psql(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		WhereNull("deleted").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT * FROM "users" WHERE "deleted" IS NULL`, sql)
	assert.Equal(t, []interface{}{}, args)
}

func TestWhereNotNull_Psql(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		WhereNotNull("deleted").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT * FROM "users" WHERE "deleted" IS NOT NULL`, sql)
	assert.Equal(t, []interface{}{}, args)
}

func TestWhereBetweenTime_Psql(t *testing.T) {
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 1, 1, 23, 59, 59, 1e9, time.UTC)
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		WhereBetween("deleted", start, end).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT * FROM "users" WHERE "deleted" BETWEEN $1 AND $2`, sql)
	assert.Equal(t, []interface{}{start, end}, args)
}

func TestWhereBetweenUint_Psql(t *testing.T) {
	var start uint = 0
	var end uint = 100
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		WhereBetween("id", start, end).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT * FROM "users" WHERE "id" BETWEEN $1 AND $2`, sql)
	assert.Equal(t, []interface{}{start, end}, args)
}

func TestWhereBetweenInt_Psql(t *testing.T) {
	var start int = -100
	var end int = 100
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		WhereBetween("id", start, end).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT * FROM "users" WHERE "id" BETWEEN $1 AND $2`, sql)
	assert.Equal(t, []interface{}{start, end}, args)
}

func TestWhereNotBetween_Psql(t *testing.T) {
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 1, 1, 23, 59, 59, 1e9, time.UTC)
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		WhereNotBetween("deleted", start, end).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT * FROM "users" WHERE "deleted" NOT BETWEEN $1 AND $2`, sql)
	assert.Equal(t, []interface{}{start, end}, args)
}

func TestWhereWhen1_Psql(t *testing.T) {
	age := 25
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		When(age >= 21, func(q *SelectBuilder) *SelectBuilder {
			return q.Where("adult", true)
		}, func(q *SelectBuilder) *SelectBuilder {
			return q.Where("adult", false).
				Where("age", "<=", age)
		}).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT * FROM "users" WHERE "adult" = $1`, sql)
	assert.Equal(t, []interface{}{true}, args)
}

func TestWhereWhen2_Psql(t *testing.T) {
	age := 12
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		When(age >= 21, func(q *SelectBuilder) *SelectBuilder {
			return q.Where("adult", true)
		}, func(q *SelectBuilder) *SelectBuilder {
			return q.Where("adult", false).
				Where("age", "<=", age)
		}).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT * FROM "users" WHERE "adult" = $1 AND "age" <= $2`, sql)
	assert.Equal(t, []interface{}{false, age}, args)
}

func TestWhereOr_Psql(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		Where("deleted", "!=", nil).
		OrWhere("trash", "=", true).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT * FROM "users" WHERE "deleted" IS NOT NULL OR "trash" = $1`, sql)
	assert.Equal(t, []interface{}{true}, args)
}

func TestWhereNested_Psql(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		WhereNested(func(q *SelectBuilder) *SelectBuilder {
			return q.Where("adult", false).
				Where("age", "<=", 10)
		}).
		OrWhere("trash", "=", true).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT * FROM "users" WHERE ("adult" = $1 AND "age" <= $2) OR "trash" = $3`, sql)
	assert.Equal(t, []interface{}{false, 10, true}, args)
}

func TestOrWhereNested_Psql(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		Where("trash", "=", true).
		OrWhereNested(func(q *SelectBuilder) *SelectBuilder {
			return q.Where("adult", false).
				Where("age", "<=", 10)
		}).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT * FROM "users" WHERE "trash" = $1 OR ("adult" = $2 AND "age" <= $3)`, sql)
	assert.Equal(t, []interface{}{true, false, 10}, args)
}

func TestWhereNestedEmpty_Psql(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		WhereNested(func(q *SelectBuilder) *SelectBuilder {
			return q
		}).
		OrWhere("trash", "=", true).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT * FROM "users" WHERE "trash" = $1`, sql)
	assert.Equal(t, []interface{}{true}, args)
}

func TestWhereExists(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		WhereExists(func(q *SelectBuilder) {
			q.SelectRaw("1").
				From("orders").
				Where("orders.user_id", "=", Raw(`"users"."id"`)).
				Where("orders.status", "paid")
		}).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, `SELECT * FROM "users" WHERE EXISTS (SELECT 1 FROM "orders" WHERE "orders"."user_id" = "users"."id" AND "orders"."status" = $1)`, sql)
	assert.Equal(t, []any{"paid"}, args)
}

func TestWhereExists_PlaceholderAndBooleanComposition(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		Where("group", 42).
		WhereExists(func(q *SelectBuilder) {
			q.From("orders").
				Where("orders.user_id", "=", Raw(`"users"."id"`)).
				Where("orders.status", "paid")
		}).
		OrWhere("is_admin", true).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(
		t,
		`SELECT * FROM "users" WHERE "group" = $1 AND EXISTS (SELECT 1 FROM "orders" WHERE "orders"."user_id" = "users"."id" AND "orders"."status" = $2) OR "is_admin" = $3`,
		sql,
	)
	assert.Equal(t, []any{42, "paid", true}, args)
}

func TestWhereExists_DoesNotInheritParentPrefixOrComment(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("*").
		Comment("outer").
		Prefix("EXPLAIN").
		From("users").
		WhereExists(func(q *SelectBuilder) {
			q.From("orders").
				Where("orders.user_id", "=", Raw(`"users"."id"`))
		}).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(
		t,
		`/* outer */ EXPLAIN SELECT * FROM "users" WHERE EXISTS (SELECT 1 FROM "orders" WHERE "orders"."user_id" = "users"."id")`,
		sql,
	)
	assert.Empty(t, args)
}

func TestWhereDirectSubqueryValue(t *testing.T) {
	sub := getPsqlBuilder().
		Select("id").
		From("users").
		OrderByAsc("id").
		Limit(1)

	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		Where("id", sub).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(
		t,
		`SELECT * FROM "users" WHERE "id" = (SELECT "id" FROM "users" ORDER BY "id" ASC LIMIT 1)`,
		sql,
	)
	assert.Empty(t, args)
}

func TestWhereDirectSubqueryValueWithOperatorAndPlaceholderOffset(t *testing.T) {
	sub := getPsqlBuilder().
		Select("id").
		From("users").
		Where("active", true).
		OrderByAsc("id").
		Limit(1)

	sql, args, err := getPsqlBuilder().
		Select("*").
		From("users").
		Where("group_id", 7).
		Where("id", "=", sub).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(
		t,
		`SELECT * FROM "users" WHERE "group_id" = $1 AND "id" = (SELECT "id" FROM "users" WHERE "active" = $2 ORDER BY "id" ASC LIMIT 1)`,
		sql,
	)
	assert.Equal(t, []interface{}{7, true}, args)
}
