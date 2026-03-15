package eloq

import (
	"errors"
	"sort"
	"strings"
)

type UpdateBuilder struct {
	*commonBuilder

	table  string
	values map[string]any
	wheres []whereClause
	joins  []joinClause
}

func Update(table string) *UpdateBuilder {
	b := &UpdateBuilder{
		commonBuilder: &commonBuilder{
			placeholder: Question,
			quoteStyle:  Backtick,
			comments:    []string{},
			queryName:   "",
			meta:        map[string]string{},
		},
		table:  table,
		values: map[string]interface{}{},
	}

	return b
}

func (cb *commonBuilder) Update(table string) *UpdateBuilder {
	return &UpdateBuilder{
		commonBuilder: cb.clone(),
		table:         table,
		values:        map[string]interface{}{},
	}
}

// WHERE
func (b *UpdateBuilder) Where(column string, args ...interface{}) *UpdateBuilder {
	b.wheres = b.commonBuilder.addWhere(b.wheres, false, column, args...)
	return b
}

func (b *UpdateBuilder) OrWhere(column string, args ...interface{}) *UpdateBuilder {
	b.wheres = b.commonBuilder.addWhere(b.wheres, true, column, args...)
	return b
}

func (b *UpdateBuilder) WhereIn(column string, values ...interface{}) *UpdateBuilder {
	b.wheres = b.commonBuilder.addWhereIn(b.wheres, column, values, false)
	return b
}

func (b *UpdateBuilder) WhereNotIn(column string, values ...interface{}) *UpdateBuilder {
	b.wheres = b.commonBuilder.addWhereIn(b.wheres, column, values, true)
	return b
}

func (b *UpdateBuilder) WhereNull(column string) *UpdateBuilder {
	b.wheres = b.commonBuilder.addWhereNull(b.wheres, column, false)
	return b
}

func (b *UpdateBuilder) WhereNotNull(column string) *UpdateBuilder {
	b.wheres = b.commonBuilder.addWhereNull(b.wheres, column, true)
	return b
}

func (b *UpdateBuilder) WhereBetween(column string, from, to interface{}) *UpdateBuilder {
	b.wheres = b.commonBuilder.addWhereBetween(b.wheres, column, from, to, false)
	return b
}

func (b *UpdateBuilder) WhereNotBetween(column string, from, to interface{}) *UpdateBuilder {
	b.wheres = b.commonBuilder.addWhereBetween(b.wheres, column, from, to, true)
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
		commonBuilder: b.commonBuilder.clone(),
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

func (b *UpdateBuilder) ToSql() (string, []interface{}, error) {
	var sql strings.Builder
	var args []interface{}

	b.renderComments(&sql)
	b.renderPrefixes(&sql, &args)

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

	phIndex := 1

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

	b.renderSuffixes(&sql, &args)

	return sql.String(), args, nil
}
