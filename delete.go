package eloq

import "strings"

type DeleteBuilder struct {
	*commonBuilder

	table  string
	wheres []whereClause
	joins  []joinClause
	using  []string
}

func Delete(table string) *DeleteBuilder {
	return &DeleteBuilder{
		commonBuilder: &commonBuilder{
			placeholder: Question,
			quoteStyle:  Backtick,
			comments:    []string{},
			queryName:   "",
			meta:        map[string]string{},
		},
		table: table,
	}
}

func (b *commonBuilder) Delete(table string) *DeleteBuilder {
	return &DeleteBuilder{
		commonBuilder: b.clone(),
		table:         table,
	}
}

// WHERE
func (b *DeleteBuilder) Where(column string, args ...interface{}) *DeleteBuilder {
	b.wheres = b.commonBuilder.addWhere(b.wheres, false, column, args...)
	return b
}

func (b *DeleteBuilder) OrWhere(column string, args ...interface{}) *DeleteBuilder {
	b.wheres = b.commonBuilder.addWhere(b.wheres, true, column, args...)
	return b
}

func (b *DeleteBuilder) WhereIn(column string, values ...interface{}) *DeleteBuilder {
	b.wheres = b.commonBuilder.addWhereIn(b.wheres, column, values, false)
	return b
}

func (b *DeleteBuilder) WhereNotIn(column string, values ...interface{}) *DeleteBuilder {
	b.wheres = b.commonBuilder.addWhereIn(b.wheres, column, values, true)
	return b
}

func (b *DeleteBuilder) WhereNull(column string) *DeleteBuilder {
	b.wheres = b.commonBuilder.addWhereNull(b.wheres, column, false)
	return b
}

func (b *DeleteBuilder) WhereNotNull(column string) *DeleteBuilder {
	b.wheres = b.commonBuilder.addWhereNull(b.wheres, column, true)
	return b
}

func (b *DeleteBuilder) WhereBetween(column string, from, to interface{}) *DeleteBuilder {
	b.wheres = b.commonBuilder.addWhereBetween(b.wheres, column, from, to, false)
	return b
}

func (b *DeleteBuilder) WhereNotBetween(column string, from, to interface{}) *DeleteBuilder {
	b.wheres = b.commonBuilder.addWhereBetween(b.wheres, column, from, to, true)
	return b
}

func (b *DeleteBuilder) When(condition bool, thenFunc func(*DeleteBuilder) *DeleteBuilder, elseFunc ...func(*DeleteBuilder) *DeleteBuilder) *DeleteBuilder {
	if condition {
		if thenFunc != nil {
			thenFunc(b)
		}
	} else if len(elseFunc) > 0 && elseFunc[0] != nil {
		elseFunc[0](b)
	}

	return b
}

func (b *DeleteBuilder) addWhereNested(fn func(*DeleteBuilder) *DeleteBuilder, isOr bool) *DeleteBuilder {
	nestedBuilder := &DeleteBuilder{
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

func (b *DeleteBuilder) WhereNested(fn func(*DeleteBuilder) *DeleteBuilder) *DeleteBuilder {
	b.addWhereNested(fn, false)
	return b
}

func (b *DeleteBuilder) OrWhereNested(fn func(*DeleteBuilder) *DeleteBuilder) *DeleteBuilder {
	b.addWhereNested(fn, true)
	return b
}

// Suffix\Prefix
func (b *DeleteBuilder) Suffix(sql string, args ...interface{}) *DeleteBuilder {
	b.commonBuilder.Suffix(sql, args...)

	return b
}

func (b *DeleteBuilder) Prefix(sql string, args ...interface{}) *DeleteBuilder {
	b.commonBuilder.Prefix(sql, args...)

	return b
}

// JOIN
func (b *DeleteBuilder) Join(table, left, operator, right string) *DeleteBuilder {
	return b.addJoin("INNER", table, left, operator, right)
}

func (b *DeleteBuilder) LeftJoin(table, left, operator, right string) *DeleteBuilder {
	return b.addJoin("LEFT", table, left, operator, right)
}

func (b *DeleteBuilder) RightJoin(table, left, operator, right string) *DeleteBuilder {
	return b.addJoin("RIGHT", table, left, operator, right)
}

func (b *DeleteBuilder) JoinWith(table string, fn func(*JoinBuilder)) *DeleteBuilder {
	return b.addJoinWith("INNER", table, fn)
}

func (b *DeleteBuilder) LeftJoinWith(table string, fn func(*JoinBuilder)) *DeleteBuilder {
	return b.addJoinWith("LEFT", table, fn)
}

func (b *DeleteBuilder) RightJoinWith(table string, fn func(*JoinBuilder)) *DeleteBuilder {
	return b.addJoinWith("RIGHT", table, fn)
}

func (b *DeleteBuilder) JoinRaw(sql string) *DeleteBuilder {
	b.joins = append(b.joins, joinClause{
		raw:    true,
		rawSQL: sql,
	})
	return b
}

func (b *DeleteBuilder) addJoin(joinType, table, left, operator, right string) *DeleteBuilder {
	j := joinClause{
		joinType: joinType,
		table:    table,
		ons: []joinOnClause{
			{
				left:     left,
				operator: operator,
				right:    right,
			},
		},
	}

	b.joins = append(b.joins, j)
	return b
}

func (b *DeleteBuilder) addJoinWith(joinType, table string, fn func(*JoinBuilder)) *DeleteBuilder {
	jb := &JoinBuilder{}
	fn(jb)

	j := joinClause{
		joinType: joinType,
		table:    table,
		ons:      jb.ons,
		wheres:   jb.wheres,
	}

	b.joins = append(b.joins, j)
	return b
}

// USING
func (b *DeleteBuilder) Using(tables ...string) *DeleteBuilder {
	b.using = append(b.using, tables...)
	return b
}

func (b *DeleteBuilder) renderUsing(sql *strings.Builder) error {
	if len(b.using) == 0 {
		return nil
	}

	sql.WriteString(" USING ")

	var tables []string
	for _, t := range b.using {
		q, err := b.quoteIdentifier(t)
		if err != nil {
			return err
		}
		tables = append(tables, q)
	}

	sql.WriteString(strings.Join(tables, ","))
	return nil
}

// COMMON
func (b *DeleteBuilder) PlaceholderFormat(f PlaceholderFormat) *DeleteBuilder {
	b.commonBuilder.PlaceholderFormat(f)
	return b
}

func (b *DeleteBuilder) QuoteWith(q QuoteStyle) *DeleteBuilder {
	b.commonBuilder.QuoteWith(q)
	return b
}

func (b *DeleteBuilder) ToSql() (string, []interface{}, error) {
	var sql strings.Builder
	args := []interface{}{}

	// comments
	b.renderComments(&sql)

	// prefixes (EXPLAIN, etc)
	b.renderPrefixes(&sql, &args)

	sql.WriteString("DELETE FROM ")

	tbl, err := b.quoteIdentifier(b.table)
	if err != nil {
		return "", []interface{}{}, err
	}
	sql.WriteString(tbl)

	// USING (NEW)
	if err := b.renderUsing(&sql); err != nil {
		return "", nil, err
	}

	phIndex := 1

	// JOIN
	joinSql, joinBindings, nextIndex, joinErr := b.commonBuilder.renderJoins(b.joins, phIndex)
	if joinErr != nil {
		return "", []interface{}{}, joinErr
	} else if joinSql != "" {
		sql.WriteString(joinSql)
		args = append(args, joinBindings...)
		phIndex = nextIndex
	}

	// WHERE
	if len(b.wheres) > 0 {
		sql.WriteString(" WHERE ")

		var err error
		whereSQL, whereBindings, nextIndex, err := b.commonBuilder.renderWheres(b.wheres, phIndex)
		if err != nil {
			return "", []interface{}{}, err
		}

		if whereSQL != "" {
			sql.WriteString(whereSQL)

			args = append(args, whereBindings...)
			phIndex = nextIndex
		}
	}

	// suffixes (RETURNING later, FOR UPDATE, etc)
	b.renderSuffixes(&sql, &args)

	return sql.String(), args, nil
}
