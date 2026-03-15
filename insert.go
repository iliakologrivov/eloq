package eloq

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type InsertBuilder struct {
	*commonBuilder

	table string
	rows  []map[string]interface{}

	conflictColumns []string
	updateValues    map[string]interface{}
	doNothing       bool
}

func Insert(table string) *InsertBuilder {
	b := &InsertBuilder{
		commonBuilder: &commonBuilder{
			placeholder: Question,
			quoteStyle:  Backtick,
			comments:    []string{},
			queryName:   "",
			meta:        map[string]string{},
		},
		table: table,
	}

	return b
}

func (b *commonBuilder) Insert(table string) *InsertBuilder {
	return &InsertBuilder{
		commonBuilder: b.clone(),
		table:         table,
	}
}

func (b *InsertBuilder) Values(row map[string]interface{}) *InsertBuilder {
	if len(row) > 0 {
		b.rows = append(b.rows, row)
	}
	return b
}

func (b *InsertBuilder) Returning(cols ...string) *InsertBuilder {
	var quoted []string
	for _, c := range cols {
		q, err := b.quoteIdentifier(c)
		if err != nil {
			return b
		}
		quoted = append(quoted, q)
	}

	b.Suffix("RETURNING " + strings.Join(quoted, ","))
	return b
}

func (b *InsertBuilder) OnConflict(cols ...string) *InsertBuilder {
	b.conflictColumns = append(b.conflictColumns, cols...)
	return b
}

func (b *InsertBuilder) DoNothing() *InsertBuilder {
	b.doNothing = true
	return b
}

func (b *InsertBuilder) DoUpdate(values map[string]interface{}) *InsertBuilder {
	if b.updateValues == nil {
		b.updateValues = map[string]interface{}{}
	}
	for k, v := range values {
		b.updateValues[k] = v
	}
	return b
}

// COMMON
func (b *InsertBuilder) Suffix(sql string, args ...interface{}) *InsertBuilder {
	b.commonBuilder.Suffix(sql, args...)

	return b
}

func (b *InsertBuilder) Prefix(sql string, args ...interface{}) *InsertBuilder {
	b.commonBuilder.Prefix(sql, args...)

	return b
}

func (b *InsertBuilder) ToSql() (string, []interface{}, error) {
	var sql strings.Builder

	// comments
	b.renderComments(&sql)
	var args []interface{}
	index := 1
	b.renderPrefixes(&sql, &args)

	if len(b.rows) == 0 {
		return "", []interface{}{}, errors.New("eloq: insert values are empty")
	}

	// build columns from first row
	var columns []string
	for k := range b.rows[0] {
		columns = append(columns, k)
	}
	sort.Strings(columns)

	// validate rows
	for i, row := range b.rows {
		if len(row) != len(columns) {
			return "", []interface{}{}, fmt.Errorf("eloq: inconsistent insert row %d", i+1)
		}

		for _, col := range columns {
			if _, ok := row[col]; !ok {
				return "", []interface{}{}, fmt.Errorf("eloq: missing column %q in insert row %d", col, i+1)
			}
		}
	}

	// INSERT INTO
	sql.WriteString("INSERT INTO ")

	tbl, err := b.quoteIdentifier(b.table)
	if err != nil {
		return "", []interface{}{}, err
	}
	sql.WriteString(tbl)

	// columns
	sql.WriteString(" (")

	var quotedCols []string
	for _, c := range columns {
		q, err := b.quoteIdentifier(c)
		if err != nil {
			return "", []interface{}{}, err
		}
		quotedCols = append(quotedCols, q)
	}

	sql.WriteString(strings.Join(quotedCols, ","))
	sql.WriteString(") VALUES ")

	// values
	var rowSQL []string

	for _, row := range b.rows {
		var ph []string
		for _, col := range columns {
			ph = append(ph, b.formatPlaceholder(index))
			args = append(args, row[col])
			index++
		}
		rowSQL = append(rowSQL, "("+strings.Join(ph, ",")+")")
	}

	sql.WriteString(strings.Join(rowSQL, ", "))

	// ON CONFLICT
	if len(b.conflictColumns) > 0 {
		sql.WriteString(" ON CONFLICT (")

		var cols []string
		for _, c := range b.conflictColumns {
			q, err := b.quoteIdentifier(c)
			if err != nil {
				return "", []interface{}{}, err
			}
			cols = append(cols, q)
		}

		sql.WriteString(strings.Join(cols, ","))
		sql.WriteString(") ")

		if b.doNothing {
			sql.WriteString("DO NOTHING")
		} else {
			if len(b.updateValues) == 0 {
				return "", []interface{}{}, errors.New("eloq: ON CONFLICT requires DoNothing or DoUpdate")
			}

			sql.WriteString("DO UPDATE SET ")

			var sets []string
			updateCols := make([]string, 0, len(b.updateValues))
			for col := range b.updateValues {
				updateCols = append(updateCols, col)
			}
			sort.Strings(updateCols)

			for _, col := range updateCols {
				val := b.updateValues[col]
				q, err := b.quoteIdentifier(col)
				if err != nil {
					return "", []interface{}{}, err
				}

				sets = append(sets, q+" = "+b.formatPlaceholder(index))
				args = append(args, val)
				index++
			}

			sql.WriteString(strings.Join(sets, ", "))
		}
	}

	b.renderSuffixes(&sql, &args)

	return sql.String(), args, nil
}
