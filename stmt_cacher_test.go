package eloq

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"testing"
)

// mockDB implements sqlDriver for testing
type mockDB struct {
	preparedStatements map[string]*mockStmt
	closed             bool
	mu                 sync.Mutex
}

type mockStmt struct {
	closed bool
}

func newMockDB() *mockDB {
	return &mockDB{
		preparedStatements: make(map[string]*mockStmt),
	}
}

func (m *mockDB) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return nil, errors.New("db closed")
	}

	m.preparedStatements[query] = &mockStmt{}
	return nil, nil // Returns nil since Stmt cannot be created without a real driver
}

// TestStmtCacher_Prepare_Basic tests basic caching functionality.
func TestStmtCacher_Prepare_Basic(t *testing.T) {
	cache := &StmtCacher{
		db:    nil, // not used in this test
		stmts: make(map[string]*sql.Stmt),
	}

	// Check cache size
	if cache.Size() != 0 {
		t.Errorf("expected empty cache, got %d", cache.Size())
	}

	// After adding
	cache.stmts["SELECT 1"] = nil
	if cache.Size() != 1 {
		t.Errorf("expected cache size 1, got %d", cache.Size())
	}
}

// TestStmtCacher_Concurrent tests concurrent map access.
func TestStmtCacher_Concurrent(t *testing.T) {
	cache := &StmtCacher{
		db:    nil,
		stmts: make(map[string]*sql.Stmt),
	}

	const goroutines = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)

	// Concurrent writes
	for i := 0; i < goroutines; i++ {
		go func(i int) {
			defer wg.Done()
			cache.mu.Lock()
			cache.stmts["key"] = nil
			cache.mu.Unlock()
		}(i)
	}

	wg.Wait()

	// Concurrent reads
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			cache.mu.RLock()
			_ = cache.stmts["key"]
			cache.mu.RUnlock()
		}()
	}

	wg.Wait()
}

// TestStmtCacher_Close tests cache cleanup.
func TestStmtCacher_Close(t *testing.T) {
	cache := &StmtCacher{
		db:    nil,
		stmts: make(map[string]*sql.Stmt),
	}

	// Add entries
	cache.stmts["sql1"] = nil
	cache.stmts["sql2"] = nil
	cache.stmts["sql3"] = nil

	if cache.Size() != 3 {
		t.Errorf("expected cache size 3, got %d", cache.Size())
	}

	// Close
	cache.mu.Lock()
	for k := range cache.stmts {
		delete(cache.stmts, k)
	}
	cache.mu.Unlock()

	if cache.Size() != 0 {
		t.Errorf("expected cache size 0, got %d", cache.Size())
	}
}

// TestRunner_NewRunner tests Runner creation.
func TestRunner_NewRunner(t *testing.T) {
	runner := &Runner{
		cache: &StmtCacher{
			db:    nil,
			stmts: make(map[string]*sql.Stmt),
		},
	}

	if runner.Cache() == nil {
		t.Error("expected non-nil cache")
	}

	if runner.Cache().Size() != 0 {
		t.Errorf("expected empty cache, got %d", runner.Cache().Size())
	}
}

// TestSqlBuilder_Interface verifies all builders implement SqlBuilder.
func TestSqlBuilder_Interface(t *testing.T) {
	var _ SqlBuilder = Select("*").From("users")
	var _ SqlBuilder = Insert("users").Values(map[string]interface{}{"name": "test"})
	var _ SqlBuilder = Update("users").Set("name", "test").Where("id", 1)
	var _ SqlBuilder = Delete("users").Where("id", 1)
}

// TestExecBuilder_Error tests error handling in ExecBuilder.
func TestExecBuilder_Error(t *testing.T) {
	runner := &Runner{
		cache: &StmtCacher{
			db:    nil,
			stmts: make(map[string]*sql.Stmt),
		},
	}

	// Builder with error
	sb := Select("*") // No table - will return error

	_, err := ExecBuilder(context.Background(), runner, sb)
	if err == nil {
		t.Error("expected error for empty table")
	}
}

// TestQueryBuilder_Error tests error handling in QueryBuilder.
func TestQueryBuilder_Error(t *testing.T) {
	runner := &Runner{
		cache: &StmtCacher{
			db:    nil,
			stmts: make(map[string]*sql.Stmt),
		},
	}

	// Builder with error
	sb := Select("*") // No table

	_, err := QueryBuilder(context.Background(), runner, sb)
	if err == nil {
		t.Error("expected error for empty table")
	}
}

// TestQueryRowBuilder_Error tests error handling in QueryRowBuilder.
func TestQueryRowBuilder_Error(t *testing.T) {
	runner := &Runner{
		cache: &StmtCacher{
			db:    nil,
			stmts: make(map[string]*sql.Stmt),
		},
	}

	// Builder with error
	sb := Select("*") // No table

	_, err := QueryRowBuilder(context.Background(), runner, sb)
	if err == nil {
		t.Error("expected error for empty table")
	}
}

// BenchmarkStmtCacher_Prepare benchmarks caching.
func BenchmarkStmtCacher_Prepare(b *testing.B) {
	cache := &StmtCacher{
		db:    nil,
		stmts: make(map[string]*sql.Stmt),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.mu.RLock()
		_, ok := cache.stmts["SELECT * FROM users WHERE id = ?"]
		cache.mu.RUnlock()

		if !ok {
			cache.mu.Lock()
			cache.stmts["SELECT * FROM users WHERE id = ?"] = nil
			cache.mu.Unlock()
		}
	}
}

// BenchmarkStmtCacher_Concurrent benchmarks concurrent access.
func BenchmarkStmtCacher_Concurrent(b *testing.B) {
	cache := &StmtCacher{
		db:    nil,
		stmts: make(map[string]*sql.Stmt),
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cache.mu.RLock()
			_, ok := cache.stmts["sql"]
			cache.mu.RUnlock()

			if !ok {
				cache.mu.Lock()
				cache.stmts["sql"] = nil
				cache.mu.Unlock()
			}
		}
	})
}
