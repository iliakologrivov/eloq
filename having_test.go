package eloq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHavingBasic(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("user_id").
		AddSelectRaw("COUNT(*) as cnt").
		From("orders").
		GroupBy("user_id").
		Having("cnt", ">", 5).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t,
		`SELECT "user_id", COUNT(*) as cnt FROM "orders" GROUP BY "user_id" HAVING "cnt" > $1`,
		sql,
	)
	assert.Equal(t, []interface{}{5}, args)
}

func TestHavingOr(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("user_id").
		AddSelectRaw("COUNT(*) as cnt").
		From("orders").
		GroupBy("user_id").
		Having("cnt", ">", 5).
		OrHaving("cnt", "<", 2).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t,
		`SELECT "user_id", COUNT(*) as cnt FROM "orders" GROUP BY "user_id" HAVING "cnt" > $1 OR "cnt" < $2`,
		sql,
	)
	assert.Equal(t, []interface{}{5, 2}, args)
}

func TestHavingNested(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("user_id").
		AddSelectRaw("COUNT(*) as cnt").
		From("orders").
		GroupBy("user_id").
		HavingNested(func(q *SelectBuilder) {
			q.Having("cnt", ">", 5).
				OrHaving("cnt", "<", 2)
		}).
		OrHaving("cnt", "=", 10).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t,
		`SELECT "user_id", COUNT(*) as cnt FROM "orders" GROUP BY "user_id" HAVING ("cnt" > $1 OR "cnt" < $2) OR "cnt" = $3`,
		sql,
	)
	assert.Equal(t, []interface{}{5, 2, 10}, args)
}

func TestHavingNestedEmpty(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("user_id").
		AddSelectRaw("COUNT(*) as cnt").
		From("orders").
		GroupBy("user_id").
		HavingNested(func(q *SelectBuilder) {
		}).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t,
		`SELECT "user_id", COUNT(*) as cnt FROM "orders" GROUP BY "user_id"`,
		sql,
	)
	assert.Equal(t, []interface{}{}, args)
}

func TestGroupByBasic(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("user_id").
		AddSelectRaw("COUNT(*) as cnt").
		From("orders").
		GroupBy("user_id").
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t,
		`SELECT "user_id", COUNT(*) as cnt FROM "orders" GROUP BY "user_id"`,
		sql,
	)
	assert.Empty(t, args)
}

func TestGroupByHaving(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("user_id").
		AddSelectRaw("COUNT(*) as cnt").
		From("orders").
		GroupBy("user_id").
		Having("cnt", ">", 5).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t,
		`SELECT "user_id", COUNT(*) as cnt FROM "orders" GROUP BY "user_id" HAVING "cnt" > $1`,
		sql,
	)
	assert.Equal(t, []interface{}{5}, args)
}

func TestGroupByHavingTwoOperators(t *testing.T) {
	sql, args, err := getPsqlBuilder().
		Select("user_id").
		AddSelectRaw("COUNT(*) as cnt", "SUM(amount) AS total_amount").
		From("orders").
		GroupBy("user_id").
		Having("cnt", ">", 5).
		Having("total_amount", ">=", 10000).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t,
		`SELECT "user_id", COUNT(*) as cnt, SUM(amount) AS total_amount FROM "orders" GROUP BY "user_id" HAVING "cnt" > $1 AND "total_amount" >= $2`,
		sql,
	)
	assert.Equal(t, []interface{}{5, 10000}, args)
}
