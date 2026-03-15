# eloq

SQL query builder для Go в стиле Laravel Eloquent.

Цель пакета: упростить переход с PHP на Go за счёт знакомого fluent API.

## Установка

```bash
go get github.com/iliakologrivov/eloq
```

## Быстрый старт

```go
package main

import (
	"fmt"

	"github.com/iliakologrivov/eloq"
)

func main() {
	sql, args, err := eloq.NewBuilder().
		PlaceholderFormat(eloq.Dollar).
		QuoteWith(eloq.DoubleQuote).
		Select("id", "email").
		From("users").
		Where("active", true).
		OrderByDesc("id").
		Limit(10).
		ToSql()
	if err != nil {
		panic(err)
	}

	fmt.Println(sql)
	fmt.Println(args)
}
```

Результат:

```sql
SELECT "id", "email" FROM "users" WHERE "active" = $1 ORDER BY "id" DESC LIMIT 10
```

```go
[]interface{}{true}
```

## Конфигурация билдеров

`NewBuilder()` создаёт базовый билдер. На нём можно настроить:

- `PlaceholderFormat(eloq.Question)` -> `?` (MySQL-стиль)
- `PlaceholderFormat(eloq.Dollar)` -> `$1, $2 ...` (PostgreSQL-стиль)
- `QuoteWith(eloq.Backtick)` -> `` `col` ``
- `QuoteWith(eloq.DoubleQuote)` -> `"col"`

Обычно удобно завести переменные-конфиги:

```go
psql := eloq.NewBuilder().
	PlaceholderFormat(eloq.Dollar).
	QuoteWith(eloq.DoubleQuote)

mysql := eloq.NewBuilder().
	PlaceholderFormat(eloq.Question).
	QuoteWith(eloq.Backtick)
```

## SELECT

### Базовый select

```go
sql, args, _ := eloq.NewBuilder().
	PlaceholderFormat(eloq.Dollar).
	QuoteWith(eloq.DoubleQuote).
	Select("id", "name").
	From("users").
	ToSql()
```

### SelectRaw / AddSelect

```go
sql, args, _ := eloq.NewBuilder().
	PlaceholderFormat(eloq.Dollar).
	QuoteWith(eloq.DoubleQuote).
	Select("id").
	AddSelectRaw(`COUNT(*) OVER() AS "total"`).
	From("users").
	ToSql()
```

### WHERE

Поддерживаются:

- `Where("col", value)` (оператор `=`)
- `Where("col", ">", value)`
- `OrWhere(...)`
- `WhereNull`, `WhereNotNull`
- `WhereIn`, `WhereNotIn`
- `WhereBetween`, `WhereNotBetween`
- `WhereNested`, `OrWhereNested`
- `When(condition, then, else?)`
- `WhereExists`, `OrWhereExists`
- сравнение с подзапросом через обычный `Where`:
  - `Where("id", subQuery)`
  - `Where("id", "=", subQuery)`

### IN с подзапросом

```go
sub := eloq.NewBuilder().
	PlaceholderFormat(eloq.Dollar).
	QuoteWith(eloq.DoubleQuote).
	Select("user_id").
	From("orders").
	Where("status", "done").
	Where("amount", ">=", 100)

sql, args, _ := eloq.NewBuilder().
	PlaceholderFormat(eloq.Dollar).
	QuoteWith(eloq.DoubleQuote).
	Select("*").
	From("users").
	WhereInSub("id", sub).
	OrderByDesc("id").
	Limit(10).
	ToSql()
```

### JOIN

Поддерживаются:

- `Join`, `LeftJoin`, `RightJoin`
- `JoinWith` / `LeftJoinWith` / `RightJoinWith` (через `JoinBuilder`)
- `JoinRaw`

`JoinBuilder`:

- `On`, `OrOn`
- `Where`, `OrWhere` (условия внутри `ON (...)`)

### GROUP / HAVING / ORDER / LIMIT / OFFSET

- `GroupBy`, `GroupByRaw`
- `Having`, `OrHaving`, `HavingNested`
- `OrderBy`, `OrderByAsc`, `OrderByDesc`, `OrderByRaw`
- `Limit`, `Offset`

### UNION / UNION ALL

```go
sql, args, _ := eloq.NewBuilder().
	PlaceholderFormat(eloq.Dollar).
	QuoteWith(eloq.DoubleQuote).
	Select("id").
	From("users").
	UnionAll(
		eloq.NewBuilder().
			PlaceholderFormat(eloq.Dollar).
			QuoteWith(eloq.DoubleQuote).
			Select("id").
			From("admins"),
	).
	OrderByDesc("id").
	ToSql()
```

### FROM (subquery) AS alias

```go
sub := eloq.NewBuilder().
	PlaceholderFormat(eloq.Dollar).
	QuoteWith(eloq.DoubleQuote).
	SelectRaw(`id`, `email AS "name"`).
	From("user").
	UnionAll(
		eloq.NewBuilder().
			PlaceholderFormat(eloq.Dollar).
			QuoteWith(eloq.DoubleQuote).
			Select("id", "name").
			From("admins"),
	)

sql, args, _ := eloq.NewBuilder().
	PlaceholderFormat(eloq.Dollar).
	QuoteWith(eloq.DoubleQuote).
	Select("id", "name").
	FromSub(sub, "u").
	OrderByDesc("id").
	Limit(10).
	ToSql()
```

## INSERT

```go
sql, args, _ := eloq.NewBuilder().
	PlaceholderFormat(eloq.Dollar).
	QuoteWith(eloq.DoubleQuote).
	Insert("users").
	Values(map[string]interface{}{
		"email": "a@b.com",
		"name":  "John",
	}).
	Returning("id").
	ToSql()
```

Поддерживаются:

- `Values(...)` (можно вызывать несколько раз для multi-row insert)
- `OnConflict("col1", ...)`
- `DoNothing()`
- `DoUpdate(map[string]interface{}{"col": value})`
- `Returning(...)`

## UPDATE

```go
sql, args, _ := eloq.NewBuilder().
	PlaceholderFormat(eloq.Dollar).
	QuoteWith(eloq.DoubleQuote).
	Update("users").
	Set("name", "John").
	Set("updated_at", eloq.Raw("NOW()")).
	Where("id", 1).
	ToSql()
```

Поддерживаются:

- `Set`, `SetMap`
- все основные `WHERE`-методы (`Where`, `OrWhere`, `WhereIn`, `WhereNested`, `When`, ...)
- `Join` / `JoinWith` и их варианты

## DELETE

```go
sql, args, _ := eloq.NewBuilder().
	PlaceholderFormat(eloq.Dollar).
	QuoteWith(eloq.DoubleQuote).
	Delete("users").
	Where("id", 10).
	ToSql()
```

Поддерживаются:

- все основные `WHERE`-методы
- `Join` / `JoinWith` и их варианты
- `Using("orders")` (PostgreSQL-style `DELETE ... USING ...`)

## Комментарии, имя запроса и мета

Можно добавлять SQL-комментарий перед запросом:

- `Name("load_users")`, `Namef(...)`
- `Comment("text")`
- `CommentKV("k1", v1, "k2", v2)`
- `AddMeta("tenant", "eu")`
- `WithMeta(map[string]string{...})`
- `WithContext(ctx)` (`trace_id`, `span_id`, `request_id`, `user_id`)

Пример:

```go
sql, args, _ := eloq.NewBuilder().
	PlaceholderFormat(eloq.Dollar).
	QuoteWith(eloq.DoubleQuote).
	Name("load_users").
	AddMeta("tenant", "eu").
	Comment("dashboard").
	Select("*").
	From("users").
	ToSql()
```

## Prefix / Suffix

Для префиксов и суффиксов SQL:

- `Prefix("EXPLAIN")`
- `Suffix("FOR UPDATE")`

Они работают для `SELECT`, `INSERT`, `DELETE`.

## Raw

`Raw(...)` нужен для безопасного встраивания выражений без квотирования идентификатора:

- `Set("updated_at", Raw("NOW()"))`
- `Where("orders.user_id", "=", Raw(\`"users"."id"\`))`

Важно: `Raw` вставляется как есть. Используйте только доверенные строки.

## Ошибки и ограничения

- `Select` без `From`/`FromSub` -> `ErrEmptyTable`
- `FromSub(..., "")` -> `ErrEmptySubqueryAlias`
- пустой `IN` -> ошибка
- `Insert` без строк -> ошибка
- `Update` без `Set` -> ошибка
- идентификаторы валидируются (защита от SQL injection через имена таблиц/колонок)

## Тесты

```bash
go test ./...
```
