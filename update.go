package eloq

import (
	"errors"
	"sort"
	"strings"
)

type UpdateBuilder struct {
	baseBuilder

	table  string
	values map[string]any
	wheres []whereClause
	joins  []joinClause
}

func Update(table string) *UpdateBuilder {
	b := &UpdateBuilder{
		baseBuilder: newBaseBuilder(),
		table:       table,
		values:      map[string]interface{}{},
	}

	return b
}

// WHERE
func (b *UpdateBuilder) Where(column string, args ...interface{}) *UpdateBuilder {
	b.wheres = b.addWhere(b.wheres, false, column, args...)
	return b
}

func (b *UpdateBuilder) OrWhere(column string, args ...interface{}) *UpdateBuilder {
	b.wheres = b.addWhere(b.wheres, true, column, args...)
	return b
}

func (b *UpdateBuilder) WhereIn(column string, values ...interface{}) *UpdateBuilder {
	b.wheres = b.addWhereIn(b.wheres, column, values, false)
	return b
}

func (b *UpdateBuilder) WhereNotIn(column string, values ...interface{}) *UpdateBuilder {
	b.wheres = b.addWhereIn(b.wheres, column, values, true)
	return b
}

func (b *UpdateBuilder) WhereNull(column string) *UpdateBuilder {
	b.wheres = b.addWhereNull(b.wheres, column, false)
	return b
}

func (b *UpdateBuilder) WhereNotNull(column string) *UpdateBuilder {
	b.wheres = b.addWhereNull(b.wheres, column, true)
	return b
}

func (b *UpdateBuilder) WhereBetween(column string, from, to interface{}) *UpdateBuilder {
	b.wheres = b.addWhereBetween(b.wheres, column, from, to, false)
	return b
}

func (b *UpdateBuilder) WhereNotBetween(column string, from, to interface{}) *UpdateBuilder {
	b.wheres = b.addWhereBetween(b.wheres, column, from, to, true)
	return b
}

func (b *UpdateBuilder) When(condition bool, thenFunc func(*UpdateBuilder) *UpdateBuilder, elseFunc ...func(*UpdateBuilder) *UpdateBuilder) *UpdateBuilder {
	if condition {
		if thenFunc != nil {
			thenFunc(b)
		}
	} else if len(elseFunc) > 0 && elseFunc[0] != nil {
		elseFunc[0](b)
	}

	return b
}

func (b *UpdateBuilder) addWhereNested(fn func(*UpdateBuilder) *UpdateBuilder, isOr bool) *UpdateBuilder {
	nestedBuilder := &UpdateBuilder{
		baseBuilder: baseBuilder{
			Config:     b.Config,
			queryState: newQueryState(),
		},
	}

	fn(nestedBuilder)

	if len(nestedBuilder.wheres) == 0 {
		return b
	}

	b.wheres = append(b.wheres, whereClause{
		nested: nestedBuilder.wheres,
		isOr:   isOr,
	})

	return b
}

func (b *UpdateBuilder) WhereNested(fn func(*UpdateBuilder) *UpdateBuilder) *UpdateBuilder {
	b.addWhereNested(fn, false)
	return b
}

func (b *UpdateBuilder) OrWhereNested(fn func(*UpdateBuilder) *UpdateBuilder) *UpdateBuilder {
	b.addWhereNested(fn, true)
	return b
}

// COMMON
func (b *UpdateBuilder) Set(column string, value interface{}) *UpdateBuilder {
	b.values[column] = value
	return b
}

func (b *UpdateBuilder) SetMap(values map[string]interface{}) *UpdateBuilder {
	for k, v := range values {
		b.values[k] = v
	}
	return b
}

func (b *UpdateBuilder) Prefix(sql string, args ...interface{}) *UpdateBuilder {
	b.baseBuilder.Prefix(sql, args...)
	return b
}

func (b *UpdateBuilder) Suffix(sql string, args ...interface{}) *UpdateBuilder {
	b.baseBuilder.Suffix(sql, args...)
	return b
}

func (b *UpdateBuilder) PlaceholderFormat(f PlaceholderFormat) *UpdateBuilder {
	b.baseBuilder.PlaceholderFormat(f)
	return b
}

func (b *UpdateBuilder) QuoteWith(q QuoteStyle) *UpdateBuilder {
	b.baseBuilder.QuoteWith(q)
	return b
}

func (b *UpdateBuilder) Comment(text string) *UpdateBuilder {
	b.baseBuilder.Comment(text)
	return b
}

func (b *UpdateBuilder) CommentKV(kv ...interface{}) *UpdateBuilder {
	b.baseBuilder.CommentKV(kv...)
	return b
}

func (b *UpdateBuilder) Name(name string) *UpdateBuilder {
	b.baseBuilder.Name(name)
	return b
}

func (b *UpdateBuilder) Namef(format string, args ...interface{}) *UpdateBuilder {
	b.baseBuilder.Namef(format, args...)
	return b
}

func (b *UpdateBuilder) AddMeta(key string, value interface{}) *UpdateBuilder {
	b.baseBuilder.AddMeta(key, value)
	return b
}

func (b *UpdateBuilder) WithMeta(m map[string]string) *UpdateBuilder {
	b.baseBuilder.WithMeta(m)
	return b
}

func (b *UpdateBuilder) ToSql() (string, []interface{}, error) {
	var sql strings.Builder
	var args []interface{}

	if b.requireWhere && len(b.wheres) == 0 {
		return "", nil, ErrRequireWhere
	}

	b.renderComments(&sql)
	phIndex := b.renderPrefixes(&sql, &args, 1)

	if len(b.values) == 0 {
		return "", nil, errors.New("eloq: update has no values")
	}

	sql.WriteString("UPDATE ")

	tbl, err := b.quoteIdentifier(b.table)
	if err != nil {
		return "", nil, err
	}
	sql.WriteString(tbl)

	sql.WriteString(" SET ")

	cols := make([]string, 0, len(b.values))
	for k := range b.values {
		cols = append(cols, k)
	}
	sort.Strings(cols)

	var sets []string
	for _, col := range cols {
		q, err := b.quoteIdentifier(col)
		if err != nil {
			return "", nil, err
		}

		switch v := b.values[col].(type) {

		case rawSqlRef:
			sets = append(sets, q+" = "+string(v))

		default:
			sets = append(sets, q+" = "+b.formatPlaceholder(phIndex))
			args = append(args, v)
			phIndex++
		}
	}

	sql.WriteString(strings.Join(sets, ", "))

	// JOIN
	if len(b.joins) > 0 {
		joinSql, joinBindings, nextIndex, joinErr := b.renderJoins(b.joins, phIndex)
		if joinErr != nil {
			return "", []interface{}{}, joinErr
		} else if joinSql != "" {
			sql.WriteString(joinSql)
			args = append(args, joinBindings...)
			phIndex = nextIndex
		}
	}

	// WHERE
	if len(b.wheres) > 0 {
		sql.WriteString(" WHERE ")

		var err error
		whereSQL, whereBindings, nextIndex, err := b.renderWheres(b.wheres, phIndex)
		if err != nil {
			return "", []interface{}{}, err
		}

		if whereSQL != "" {
			sql.WriteString(whereSQL)

			args = append(args, whereBindings...)
			phIndex = nextIndex
		}
	}

	b.renderSuffixes(&sql, &args, phIndex)

	return sql.String(), args, nil
}
