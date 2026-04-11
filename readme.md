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

## Statement Caching

Пакет предоставляет кеширование prepared statements для улучшения производительности при многократном выполнении одинаковых SQL-запросов.

### Зачем нужен кеш

При каждом выполнении SQL-запроса база данных:
1. Парсит SQL
2. Строит план выполнения
3. Выполняет запрос

Prepared statement позволяет сделать шаги 1-2 один раз, а потом только выполнять с разными аргументами. Это даёт прирост 10-30% на простых запросах и до 50% на сложных.

### Использование

```go
package main

import (
	"context"
	"database/sql"
	
	"github.com/iliakologrivov/eloq"
	_ "github.com/lib/pq"
)

func main() {
	db, _ := sql.Open("postgres", dsn)
	defer db.Close()

	// Создаём Runner с кешем
	runner := eloq.NewRunner(db)
	defer runner.Close()

	ctx := context.Background()

	// Первый вызов - создаёт prepared statement
	result, _ := runner.Exec(ctx, 
		"INSERT INTO users (name, email) VALUES (?, ?)",
		"Alice", "alice@example.com",
	)

	// Второй вызов - использует кешированный statement (быстрее)
	result, _ = runner.Exec(ctx,
		"INSERT INTO users (name, email) VALUES (?, ?)",
		"Bob", "bob@example.com",
	)
}
```

### С builder'ами

```go
runner := eloq.NewRunner(db)
defer runner.Close()

ctx := context.Background()

// SELECT через builder
sb := eloq.Select("*").From("users").Where("id", 1)
row, _ := eloq.QueryRowBuilder(ctx, runner, sb)

var name string
row.Scan(&name)

// INSERT через builder
ib := eloq.Insert("users").Values(map[string]interface{}{
	"name": "Alice",
})
result, _ := eloq.ExecBuilder(ctx, runner, ib)

// UPDATE через builder
ub := eloq.Update("users").Set("name", "Bob").Where("id", 1)
result, _ = eloq.ExecBuilder(ctx, runner, ub)

// DELETE через builder
dbb := eloq.Delete("users").Where("id", 1)
result, _ = eloq.ExecBuilder(ctx, runner, dbb)
```

### StmtCacher

Для низкоуровневого контроля:

```go
cache := eloq.NewStmtCache(db)

// Получить или создать prepared statement
stmt, _ := cache.Prepare(ctx, "SELECT * FROM users WHERE id = ?")

// Выполнить
rows, _ := stmt.Query(1)

// Проверить размер кеша
size := cache.Size()

// Закрыть все statements
cache.Close()
```

### Методы Runner

| Метод | Описание |
|-------|----------|
| `Prepare(ctx, sql)` | Возвращает кешированный `*sql.Stmt` |
| `Exec(ctx, sql, args...)` | Выполняет INSERT/UPDATE/DELETE |
| `Query(ctx, sql, args...)` | Выполняет SELECT, возвращает `*sql.Rows` |
| `QueryRow(ctx, sql, args...)` | Выполняет SELECT, возвращает одну строку |
| `Close()` | Закрывает все кешированные statements |
| `Size()` | Возвращает количество statements в кеше |

### Потокобезопасность

`StmtCacher` и `Runner` потокобезопасны. Используют `sync.RWMutex` для конкурентного доступа. Можно безопасно использовать из разных горутин.

### Важно

- Вызывайте `Close()` при завершении работы приложения
- Кеш хранит statements по точному совпадению SQL-строки
- Для PostgreSQL с `$1, $2...` placeholders кеш работает корректно

## Тесты

```bash
go test ./...
```
