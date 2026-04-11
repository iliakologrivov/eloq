package eloq

import (
	"context"
	"database/sql"
	"sync"
)

// StmtCacher caches prepared statements by SQL string.
// Thread-safe. Uses sync.RWMutex for concurrent access.
//
// Example usage:
//
//	db, _ := sql.Open("postgres", dsn)
//	cache := eloq.NewStmtCache(db)
//
//	// Builder will use cached statements
//	sb := eloq.Select("*").From("users").Where("id", 1)
//	sql, args, _ := sb.ToSql()
//
//	stmt, _ := cache.Prepare(context.Background(), sql)
//	rows, _ := stmt.Query(args...)
type StmtCacher struct {
	db    *sql.DB
	stmts map[string]*sql.Stmt
	mu    sync.RWMutex
}

// NewStmtCache creates a new prepared statements cache.
func NewStmtCache(db *sql.DB) *StmtCacher {
	return &StmtCacher{
		db:    db,
		stmts: make(map[string]*sql.Stmt),
	}
}

// Prepare returns a cached *sql.Stmt for the given SQL.
// If the statement doesn't exist, it creates and caches it.
// If the statement already exists, it returns it from cache.
func (c *StmtCacher) Prepare(ctx context.Context, sql string) (*sql.Stmt, error) {
	// Fast path: read without full lock
	c.mu.RLock()
	stmt, ok := c.stmts[sql]
	c.mu.RUnlock()
	if ok {
		return stmt, nil
	}

	// Slow path: create with write lock
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check: could have been created while waiting for lock
	if stmt, ok := c.stmts[sql]; ok {
		return stmt, nil
	}

	stmt, err := c.db.PrepareContext(ctx, sql)
	if err != nil {
		return nil, err
	}

	c.stmts[sql] = stmt
	return stmt, nil
}

// Exec executes SQL with arguments using a cached statement.
func (c *StmtCacher) Exec(ctx context.Context, sql string, args ...interface{}) (sql.Result, error) {
	stmt, err := c.Prepare(ctx, sql)
	if err != nil {
		return nil, err
	}
	return stmt.ExecContext(ctx, args...)
}

// Query executes a SQL query with arguments using a cached statement.
func (c *StmtCacher) Query(ctx context.Context, sql string, args ...interface{}) (*sql.Rows, error) {
	stmt, err := c.Prepare(ctx, sql)
	if err != nil {
		return nil, err
	}
	return stmt.QueryContext(ctx, args...)
}

// QueryRow executes a SQL query expecting a single row.
// Returns *sql.Row and a statement preparation error (if any).
func (c *StmtCacher) QueryRow(ctx context.Context, sql string, args ...interface{}) (*sql.Row, error) {
	stmt, err := c.Prepare(ctx, sql)
	if err != nil {
		return nil, err
	}
	return stmt.QueryRowContext(ctx, args...), nil
}

// Close closes all cached statements and clears the cache.
// Should be called when the application shuts down.
func (c *StmtCacher) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var firstErr error
	for sql, stmt := range c.stmts {
		if err := stmt.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
		delete(c.stmts, sql)
	}
	return firstErr
}

// Size returns the number of cached statements.
func (c *StmtCacher) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.stmts)
}

// Runner combines SQL execution capabilities through a cache.
// Implements Execable, Queryable interfaces for convenient integration.
type Runner struct {
	cache *StmtCacher
}

// NewRunner creates a Runner with statement cache.
func NewRunner(db *sql.DB) *Runner {
	return &Runner{
		cache: NewStmtCache(db),
	}
}

// Cache returns the underlying StmtCacher.
func (r *Runner) Cache() *StmtCacher {
	return r.cache
}

// Exec executes SQL through a cached statement.
func (r *Runner) Exec(ctx context.Context, sql string, args ...interface{}) (sql.Result, error) {
	return r.cache.Exec(ctx, sql, args...)
}

// Query executes a SQL query through a cached statement.
func (r *Runner) Query(ctx context.Context, sql string, args ...interface{}) (*sql.Rows, error) {
	return r.cache.Query(ctx, sql, args...)
}

// QueryRow executes a query expecting a single row.
func (r *Runner) QueryRow(ctx context.Context, sql string, args ...interface{}) (*sql.Row, error) {
	return r.cache.QueryRow(ctx, sql, args...)
}

// Close closes the statement cache.
func (r *Runner) Close() error {
	return r.cache.Close()
}

// ExecBuilder executes any builder implementing ToSql().
func ExecBuilder(ctx context.Context, runner *Runner, builder SqlBuilder) (sql.Result, error) {
	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}
	return runner.Exec(ctx, sql, args...)
}

// QueryBuilder executes a SELECT builder through the cache.
func QueryBuilder(ctx context.Context, runner *Runner, builder SqlBuilder) (*sql.Rows, error) {
	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}
	return runner.Query(ctx, sql, args...)
}

// QueryRowBuilder executes a SELECT builder expecting a single row.
func QueryRowBuilder(ctx context.Context, runner *Runner, builder SqlBuilder) (*sql.Row, error) {
	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}
	return runner.QueryRow(ctx, sql, args...)
}

// SqlBuilder is the interface for all eloq builders.
type SqlBuilder interface {
	ToSql() (string, []interface{}, error)
}
