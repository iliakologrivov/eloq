package eloq

import (
	"errors"
	"sort"
	"strings"
)

type InsertBuilder struct {
	baseBuilder
	table      string
	values     []map[string]interface{}
	returning  []string
	onConflict *onConflictClause
}

type onConflictClause struct {
	target     string
	doNothing  bool
	doUpdates  []string
	doUpdateKV map[string]interface{}
}

type OnConflictBuilder struct {
	insertBuilder *InsertBuilder
	target        string
}

func (ocb *OnConflictBuilder) DoNothing() *InsertBuilder {
	ocb.insertBuilder.onConflict = &onConflictClause{
		target:    ocb.target,
		doNothing: true,
	}
	return ocb.insertBuilder
}

func (ocb *OnConflictBuilder) DoUpdate(updates map[string]interface{}) *InsertBuilder {
	ocb.insertBuilder.onConflict = &onConflictClause{
		target:     ocb.target,
		doNothing:  false,
		doUpdateKV: updates,
	}
	return ocb.insertBuilder
}

var ErrNoTable = errors.New("eloq: insert has no table")
var ErrNoValues = errors.New("eloq: insert has no values")
var ErrEmptyValues = errors.New("eloq: insert has empty values")

func Insert(table string) *InsertBuilder {
	return &InsertBuilder{
		baseBuilder: newBaseBuilder(),
		table:       table,
	}
}

func (b *InsertBuilder) Values(values map[string]interface{}) *InsertBuilder {
	b.values = append(b.values, values)
	return b
}

func (b *InsertBuilder) Returning(columns ...string) *InsertBuilder {
	b.returning = append(b.returning, columns...)
	return b
}

func (b *InsertBuilder) OnConflict(target ...string) *OnConflictBuilder {
	targetCol := ""
	if len(target) > 0 {
		targetCol = target[0]
	}
	return &OnConflictBuilder{
		insertBuilder: b,
		target:        targetCol,
	}
}

func (b *InsertBuilder) OnConflictDoNothing(target ...string) *InsertBuilder {
	targetCol := ""
	if len(target) > 0 {
		targetCol = target[0]
	}

	b.onConflict = &onConflictClause{
		target:    targetCol,
		doNothing: true,
	}

	return b
}

func (b *InsertBuilder) OnConflictDoUpdate(target string, updateColumns ...string) *InsertBuilder {
	b.onConflict = &onConflictClause{
		target:    target,
		doNothing: false,
		doUpdates: updateColumns,
	}

	return b
}

func (b *InsertBuilder) Prefix(sql string, args ...interface{}) *InsertBuilder {
	b.baseBuilder.Prefix(sql, args...)
	return b
}

func (b *InsertBuilder) Suffix(sql string, args ...interface{}) *InsertBuilder {
	b.baseBuilder.Suffix(sql, args...)
	return b
}

func (b *InsertBuilder) PlaceholderFormat(f PlaceholderFormat) *InsertBuilder {
	b.baseBuilder.PlaceholderFormat(f)
	return b
}

func (b *InsertBuilder) QuoteWith(q QuoteStyle) *InsertBuilder {
	b.baseBuilder.QuoteWith(q)
	return b
}

func (b *InsertBuilder) Comment(text string) *InsertBuilder {
	b.baseBuilder.Comment(text)
	return b
}

func (b *InsertBuilder) CommentKV(kv ...interface{}) *InsertBuilder {
	b.baseBuilder.CommentKV(kv...)
	return b
}

func (b *InsertBuilder) Name(name string) *InsertBuilder {
	b.baseBuilder.Name(name)
	return b
}

func (b *InsertBuilder) Namef(format string, args ...interface{}) *InsertBuilder {
	b.baseBuilder.Namef(format, args...)
	return b
}

func (b *InsertBuilder) AddMeta(key string, value interface{}) *InsertBuilder {
	b.baseBuilder.AddMeta(key, value)
	return b
}

func (b *InsertBuilder) WithMeta(m map[string]string) *InsertBuilder {
	b.baseBuilder.WithMeta(m)
	return b
}

func (b *InsertBuilder) ToSql() (string, []interface{}, error) {
	if b.table == "" {
		return "", nil, ErrNoTable
	}

	if len(b.values) == 0 {
		return "", nil, ErrNoValues
	}

	columnSet := make(map[string]bool)
	for _, row := range b.values {
		for col := range row {
			columnSet[col] = true
		}
	}

	if len(columnSet) == 0 {
		return "", nil, ErrEmptyValues
	}

	columns := make([]string, 0, len(columnSet))
	for col := range columnSet {
		columns = append(columns, col)
	}
	sort.Strings(columns)

	var sql strings.Builder
	var args []interface{}

	b.renderComments(&sql)
	phIndex := b.renderPrefixes(&sql, &args, 1)

	// INSERT INTO table (columns...)
	sql.WriteString("INSERT INTO ")
	tbl, err := b.quoteIdentifier(b.table)
	if err != nil {
		return "", nil, err
	}
	sql.WriteString(tbl)
	sql.WriteString(" (")

	for i, col := range columns {
		if i > 0 {
			sql.WriteString(", ")
		}
		q, err := b.quoteIdentifier(col)
		if err != nil {
			return "", nil, err
		}
		sql.WriteString(q)
	}
	sql.WriteString(") VALUES ")

	// VALUES (...)
	for rowIdx, row := range b.values {
		if rowIdx > 0 {
			sql.WriteString(", ")
		}
		sql.WriteString("(")
		for i, col := range columns {
			if i > 0 {
				sql.WriteString(", ")
			}
			val, ok := row[col]
			if !ok {
				sql.WriteString("DEFAULT")
			} else {
				sql.WriteString(b.formatPlaceholder(phIndex))
				args = append(args, val)
				phIndex++
			}
		}
		sql.WriteString(")")
	}

	// ON CONFLICT
	if b.onConflict != nil {
		sql.WriteString(" ON CONFLICT")
		if b.onConflict.target != "" {
			sql.WriteString(" (")
			q, err := b.quoteIdentifier(b.onConflict.target)
			if err != nil {
				return "", nil, err
			}
			sql.WriteString(q)
			sql.WriteString(")")
		}
		if b.onConflict.doNothing {
			sql.WriteString(" DO NOTHING")
		} else if len(b.onConflict.doUpdateKV) > 0 {
			// Handle map-based updates with values
			updateCols := make([]string, 0, len(b.onConflict.doUpdateKV))
			for col := range b.onConflict.doUpdateKV {
				updateCols = append(updateCols, col)
			}
			sort.Strings(updateCols)

			sql.WriteString(" DO UPDATE SET ")
			for i, col := range updateCols {
				if i > 0 {
					sql.WriteString(", ")
				}
				q, err := b.quoteIdentifier(col)
				if err != nil {
					return "", nil, err
				}
				sql.WriteString(q)
				sql.WriteString(" = ")
				sql.WriteString(b.formatPlaceholder(phIndex))
				args = append(args, b.onConflict.doUpdateKV[col])
				phIndex++
			}
		} else if len(b.onConflict.doUpdates) > 0 {
			sql.WriteString(" DO UPDATE SET ")
			for i, col := range b.onConflict.doUpdates {
				if i > 0 {
					sql.WriteString(", ")
				}
				q, err := b.quoteIdentifier(col)
				if err != nil {
					return "", nil, err
				}
				sql.WriteString(q)
				sql.WriteString(" = EXCLUDED.")
				sql.WriteString(q)
			}
		}
	}

	// RETURNING
	if len(b.returning) > 0 {
		sql.WriteString(" RETURNING ")
		for i, col := range b.returning {
			if i > 0 {
				sql.WriteString(", ")
			}
			if col == "*" {
				sql.WriteString(col)
			} else {
				q, err := b.quoteIdentifier(col)
				if err != nil {
					return "", nil, err
				}
				sql.WriteString(q)
			}
		}
	}

	b.renderSuffixes(&sql, &args, phIndex)

	return sql.String(), args, nil
}
