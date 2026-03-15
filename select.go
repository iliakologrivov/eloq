package eloq

import (
	"fmt"
	"strings"
)

type orderByClause struct {
	column    string
	direction string
	raw       bool
}

type selectColumn struct {
	value string
	raw   bool
}

type unionClause struct {
	builder *SelectBuilder
	all     bool
}

type SelectBuilder struct {
	*commonBuilder
	from         string
	fromSubquery *SelectBuilder
	fromAlias    string
	columns      []selectColumn
	joins        []joinClause
	wheres       []whereClause
	bindings     []interface{}
	groupBys     []string
	havings      []whereClause
	orders       []orderByClause
	limit        *uint64
	offset       *uint64
	unions       []unionClause
}

var ErrEmptyTable = fmt.Errorf("no table specified in FROM")
var ErrEmptySubqueryAlias = fmt.Errorf("subquery alias is required")

func Select(columns ...string) *SelectBuilder {
	b := &SelectBuilder{
		commonBuilder: &commonBuilder{
			placeholder: Question,
			quoteStyle:  Backtick,
			comments:    []string{},
			queryName:   "",
			meta:        map[string]string{},
		},
		columns:  make([]selectColumn, 0),
		bindings: make([]interface{}, 0),
	}
	if len(columns) > 0 {
		b.columns = nil
		b.Select(columns...)
	}
	return b
}

func SelectRaw(columns ...string) *SelectBuilder {
	b := &SelectBuilder{
		commonBuilder: &commonBuilder{
			placeholder: Question,
			quoteStyle:  Backtick,
			comments:    []string{},
			queryName:   "",
			meta:        map[string]string{},
		},
		columns:  make([]selectColumn, 0),
		bindings: make([]interface{}, 0),
	}
	if len(columns) > 0 {
		b.columns = nil
		b.SelectRaw(columns...)
	}
	return b
}

func (cb *commonBuilder) Select(columns ...string) *SelectBuilder {
	b := &SelectBuilder{
		commonBuilder: cb.clone(),
		columns:       make([]selectColumn, 0),
		bindings:      make([]interface{}, 0),
	}
	if len(columns) > 0 {
		b.columns = nil
		b.Select(columns...)
	}
	return b
}

func (cb *commonBuilder) SelectRaw(columns ...string) *SelectBuilder {
	b := &SelectBuilder{
		commonBuilder: cb.clone(),
		columns:       make([]selectColumn, 0),
		bindings:      make([]interface{}, 0),
	}
	if len(columns) > 0 {
		b.columns = nil
		b.SelectRaw(columns...)
	}
	return b
}

func (b *SelectBuilder) Select(columns ...string) *SelectBuilder {
	b.columns = nil
	for _, col := range columns {
		b.columns = append(b.columns, selectColumn{
			value: col,
			raw:   false,
		})
	}
	return b
}

func (b *SelectBuilder) SelectRaw(expr ...string) *SelectBuilder {
	b.columns = nil
	for _, e := range expr {
		b.columns = append(b.columns, selectColumn{
			value: e,
			raw:   true,
		})
	}
	return b
}

func (b *SelectBuilder) AddSelect(columns ...string) *SelectBuilder {
	for _, col := range columns {
		b.columns = append(b.columns, selectColumn{
			value: col,
			raw:   false,
		})
	}
	return b
}

func (b *SelectBuilder) AddSelectRaw(expr ...string) *SelectBuilder {
	for _, e := range expr {
		b.columns = append(b.columns, selectColumn{
			value: e,
			raw:   true,
		})
	}
	return b
}

func (b *SelectBuilder) From(from string) *SelectBuilder {
	b.from = from
	b.fromSubquery = nil
	b.fromAlias = ""
	return b
}

func (b *SelectBuilder) FromSub(query *SelectBuilder, alias string) *SelectBuilder {
	b.from = ""
	b.fromSubquery = query
	b.fromAlias = alias
	return b
}

func (b *SelectBuilder) Table(table string) *SelectBuilder {
	b.From(table)
	return b
}

// WHERE
func (b *SelectBuilder) Where(column string, args ...interface{}) *SelectBuilder {
	b.wheres = b.commonBuilder.addWhere(b.wheres, false, column, args...)
	return b
}

func (b *SelectBuilder) OrWhere(column string, args ...interface{}) *SelectBuilder {
	b.wheres = b.commonBuilder.addWhere(b.wheres, true, column, args...)
	return b
}

func (b *SelectBuilder) WhereIn(column string, values ...interface{}) *SelectBuilder {
	b.wheres = b.commonBuilder.addWhereIn(b.wheres, column, values, false)
	return b
}

func (b *SelectBuilder) WhereNotIn(column string, values ...interface{}) *SelectBuilder {
	b.wheres = b.commonBuilder.addWhereIn(b.wheres, column, values, true)
	return b
}

func (b *SelectBuilder) WhereInSub(column string, query *SelectBuilder) *SelectBuilder {
	b.wheres = append(b.wheres, whereClause{
		column:   column,
		operator: "IN",
		value: Subquery{
			builder: query,
		},
		isNot: false,
	})

	return b
}

func (b *SelectBuilder) WhereNotInSub(column string, query *SelectBuilder) *SelectBuilder {
	b.wheres = append(b.wheres, whereClause{
		column:   column,
		operator: "IN",
		value: Subquery{
			builder: query,
		},
		isNot: true,
	})

	return b
}

func (b *SelectBuilder) WhereNull(column string) *SelectBuilder {
	b.wheres = b.commonBuilder.addWhereNull(b.wheres, column, false)
	return b
}

func (b *SelectBuilder) WhereNotNull(column string) *SelectBuilder {
	b.wheres = b.commonBuilder.addWhereNull(b.wheres, column, true)
	return b
}

func (b *SelectBuilder) WhereBetween(column string, from, to interface{}) *SelectBuilder {
	b.wheres = b.commonBuilder.addWhereBetween(b.wheres, column, from, to, false)
	return b
}

func (b *SelectBuilder) WhereNotBetween(column string, from, to interface{}) *SelectBuilder {
	b.wheres = b.commonBuilder.addWhereBetween(b.wheres, column, from, to, true)
	return b
}

func (b *SelectBuilder) When(condition bool, thenFunc func(*SelectBuilder) *SelectBuilder, elseFunc ...func(*SelectBuilder) *SelectBuilder) *SelectBuilder {
	if condition {
		if thenFunc != nil {
			thenFunc(b)
		}
	} else if len(elseFunc) > 0 && elseFunc[0] != nil {
		elseFunc[0](b)
	}

	return b
}

func (b *SelectBuilder) addWhereNested(fn func(*SelectBuilder) *SelectBuilder, isOr bool) *SelectBuilder {
	nestedBuilder := &SelectBuilder{
		commonBuilder: &commonBuilder{
			placeholder: b.placeholder,
			quoteStyle:  b.quoteStyle,
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

func (b *SelectBuilder) WhereNested(fn func(*SelectBuilder) *SelectBuilder) *SelectBuilder {
	b.addWhereNested(fn, false)
	return b
}

func (b *SelectBuilder) OrWhereNested(fn func(*SelectBuilder) *SelectBuilder) *SelectBuilder {
	b.addWhereNested(fn, true)
	return b
}

func (b *SelectBuilder) WhereExists(fn func(*SelectBuilder)) *SelectBuilder {
	sb := (&commonBuilder{
		placeholder: b.placeholder,
		quoteStyle:  b.quoteStyle,
		comments:    []string{},
		queryName:   "",
		meta:        map[string]string{},
	}).SelectRaw("1")
	fn(sb)

	b.wheres = append(b.wheres, whereClause{
		isOr: false,
		value: Subquery{
			builder: sb,
		},
		operator: "EXISTS",
	})

	return b
}

func (b *SelectBuilder) OrWhereExists(fn func(*SelectBuilder)) *SelectBuilder {
	sb := (&commonBuilder{
		placeholder: b.placeholder,
		quoteStyle:  b.quoteStyle,
		comments:    []string{},
		queryName:   "",
		meta:        map[string]string{},
	}).SelectRaw("1")
	fn(sb)

	b.wheres = append(b.wheres, whereClause{
		isOr: true,
		value: Subquery{
			builder: sb,
		},
		operator: "EXISTS",
	})

	return b
}

// ORDER
func (b *SelectBuilder) OrderByRaw(value string) *SelectBuilder {
	b.orders = append(b.orders, orderByClause{
		column:    value,
		direction: "",
		raw:       true,
	})

	return b
}

func (b *SelectBuilder) OrderBy(column string, direction ...string) *SelectBuilder {
	dir := "ASC"
	if len(direction) > 0 && strings.ToUpper(direction[0]) == "DESC" {
		dir = "DESC"
	}

	b.orders = append(b.orders, orderByClause{
		column:    column,
		direction: dir,
		raw:       false,
	})

	return b
}

func (b *SelectBuilder) OrderByDesc(column string) *SelectBuilder {
	b.OrderBy(column, "DESC")

	return b
}

func (b *SelectBuilder) OrderByAsc(column string) *SelectBuilder {
	b.OrderBy(column, "ASC")

	return b
}

// LIMIT\OFFSET
func (b *SelectBuilder) Limit(n uint64) *SelectBuilder {
	b.limit = &n
	return b
}

func (b *SelectBuilder) Offset(n uint64) *SelectBuilder {
	b.offset = &n
	return b
}

// GROUP
func (b *SelectBuilder) GroupBy(columns ...string) *SelectBuilder {
	b.groupBys = append(b.groupBys, columns...)
	return b
}

func (b *SelectBuilder) GroupByRaw(expr string) *SelectBuilder {
	b.groupBys = append(b.groupBys, expr)
	return b
}

// UNION
func (b *SelectBuilder) Union(query *SelectBuilder) *SelectBuilder {
	if query == nil {
		return b
	}

	b.unions = append(b.unions, unionClause{
		builder: query,
		all:     false,
	})
	return b
}

func (b *SelectBuilder) UnionAll(query *SelectBuilder) *SelectBuilder {
	if query == nil {
		return b
	}

	b.unions = append(b.unions, unionClause{
		builder: query,
		all:     true,
	})
	return b
}

// Suffix\Prefix
func (b *SelectBuilder) Suffix(sql string, args ...interface{}) *SelectBuilder {
	b.commonBuilder.Suffix(sql, args...)

	return b
}

func (b *SelectBuilder) Prefix(sql string, args ...interface{}) *SelectBuilder {
	b.commonBuilder.Prefix(sql, args...)

	return b
}

// JOIN
func (b *SelectBuilder) Join(table, left, operator, right string) *SelectBuilder {
	return b.addJoin("INNER", table, left, operator, right)
}

func (b *SelectBuilder) LeftJoin(table, left, operator, right string) *SelectBuilder {
	return b.addJoin("LEFT", table, left, operator, right)
}

func (b *SelectBuilder) RightJoin(table, left, operator, right string) *SelectBuilder {
	return b.addJoin("RIGHT", table, left, operator, right)
}

func (b *SelectBuilder) JoinWith(table string, fn func(*JoinBuilder)) *SelectBuilder {
	return b.addJoinWith("INNER", table, fn)
}

func (b *SelectBuilder) LeftJoinWith(table string, fn func(*JoinBuilder)) *SelectBuilder {
	return b.addJoinWith("LEFT", table, fn)
}

func (b *SelectBuilder) RightJoinWith(table string, fn func(*JoinBuilder)) *SelectBuilder {
	return b.addJoinWith("RIGHT", table, fn)
}

func (b *SelectBuilder) JoinRaw(sql string) *SelectBuilder {
	b.joins = append(b.joins, joinClause{
		raw:    true,
		rawSQL: sql,
	})
	return b
}

func (b *SelectBuilder) addJoin(joinType, table, left, operator, right string) *SelectBuilder {
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

func (b *SelectBuilder) addJoinWith(joinType, table string, fn func(*JoinBuilder)) *SelectBuilder {
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

// COMMON
func (b *SelectBuilder) PlaceholderFormat(f PlaceholderFormat) *SelectBuilder {
	b.commonBuilder.PlaceholderFormat(f)
	return b
}

func (b *SelectBuilder) QuoteWith(q QuoteStyle) *SelectBuilder {
	b.commonBuilder.QuoteWith(q)
	return b
}

func (b *SelectBuilder) ToSql() (string, []interface{}, error) {
	if b.from == "" && b.fromSubquery == nil {
		return "", []interface{}{}, ErrEmptyTable
	}

	var sql strings.Builder
	// COMMENT
	b.renderComments(&sql)

	b.bindings = []interface{}{}
	b.renderPrefixes(&sql, &b.bindings)

	phIndex := 1

	sql.WriteString("SELECT ")
	if len(b.columns) == 0 {
		sql.WriteString("*")
	} else {
		var parts []string
		for _, col := range b.columns {
			if col.raw {
				parts = append(parts, col.value)
				continue
			}
			q, err := b.quoteIdentifier(col.value)
			if err != nil {
				return "", []interface{}{}, err
			}
			parts = append(parts, q)
		}

		sql.WriteString(strings.Join(parts, ", "))
	}

	if b.fromSubquery != nil {
		if strings.TrimSpace(b.fromAlias) == "" {
			return "", []interface{}{}, ErrEmptySubqueryAlias
		}

		sql.WriteString(" FROM ")
		fromSubSQL, fromSubArgs, err := b.fromSubquery.ToSql()
		if err != nil {
			return "", []interface{}{}, err
		}

		if b.placeholder == Dollar && phIndex > 1 {
			fromSubSQL = shiftDollarPlaceholders(fromSubSQL, phIndex-1)
		}

		alias, err := b.quoteIdentifier(b.fromAlias)
		if err != nil {
			return "", []interface{}{}, err
		}

		sql.WriteString("(")
		sql.WriteString(fromSubSQL)
		sql.WriteString(") AS ")
		sql.WriteString(alias)

		b.bindings = append(b.bindings, fromSubArgs...)
		phIndex += len(fromSubArgs)
	} else if b.from != "" {
		sql.WriteString(" FROM ")
		table, err := b.quoteIdentifier(b.from)
		if err != nil {
			return "", []interface{}{}, err
		}

		sql.WriteString(table)
	}

	// JOINS
	joinSql, joinBindings, nextIndex, joinErr := b.commonBuilder.renderJoins(b.joins, phIndex)
	if joinErr != nil {
		return "", []interface{}{}, joinErr
	} else if joinSql != "" {
		sql.WriteString(joinSql)
		b.bindings = append(b.bindings, joinBindings...)
		phIndex = nextIndex
	}

	if len(b.wheres) > 0 {
		whereSQL, whereBindings, nextIndex, err := b.renderWheres(b.wheres, phIndex)
		if err != nil {
			return "", []interface{}{}, err
		} else if whereSQL != "" {
			sql.WriteString(" WHERE ")
			sql.WriteString(whereSQL)

			b.bindings = append(b.bindings, whereBindings...)
			phIndex = nextIndex
		}
	}

	// GROUP BY
	if len(b.groupBys) > 0 {
		sql.WriteString(" GROUP BY ")

		parts := make([]string, 0, len(b.groupBys))
		for _, col := range b.groupBys {
			quoted, err := b.quoteIdentifier(col)
			if err != nil {
				return "", []interface{}{}, err
			}
			parts = append(parts, quoted)
		}

		sql.WriteString(strings.Join(parts, ", "))
	}

	// HAVING
	if len(b.havings) > 0 {
		havingSQL, havingBindings, next, err := b.renderWheres(b.havings, phIndex)
		if err != nil {
			return "", []interface{}{}, err
		}

		if havingSQL != "" {
			sql.WriteString(" HAVING ")
			sql.WriteString(havingSQL)

			b.bindings = append(b.bindings, havingBindings...)
			phIndex = next
		}
	}

	if len(b.unions) > 0 {
		unionSQL, unionBindings, nextIndex, err := b.renderUnions(phIndex)
		if err != nil {
			return "", []interface{}{}, err
		}

		if unionSQL != "" {
			sql.WriteString(unionSQL)
			b.bindings = append(b.bindings, unionBindings...)
			phIndex = nextIndex
		}
	}

	if len(b.orders) > 0 {
		sql.WriteString(" ORDER BY ")

		var parts []string
		for _, ob := range b.orders {
			if ob.raw {
				parts = append(parts, ob.column)
				continue
			}
			col, err := b.quoteIdentifier(ob.column)
			if err != nil {
				return "", []interface{}{}, err
			}
			parts = append(parts, col+" "+ob.direction)
		}
		sql.WriteString(strings.Join(parts, ", "))
	}

	if b.limit != nil {
		sql.WriteString(fmt.Sprintf(" LIMIT %d", *b.limit))
	}
	if b.offset != nil {
		sql.WriteString(fmt.Sprintf(" OFFSET %d", *b.offset))
	}

	b.renderSuffixes(&sql, &b.bindings)

	return sql.String(), b.bindings, nil
}

func (b *SelectBuilder) renderUnions(startIndex int) (string, []interface{}, int, error) {
	if len(b.unions) == 0 {
		return "", nil, startIndex, nil
	}

	var sql strings.Builder
	args := make([]interface{}, 0)
	index := startIndex

	for _, u := range b.unions {
		if u.builder == nil {
			continue
		}

		unionSQL, unionArgs, err := u.builder.ToSql()
		if err != nil {
			return "", nil, startIndex, err
		}

		if b.placeholder == Dollar && index > 1 {
			unionSQL = shiftDollarPlaceholders(unionSQL, index-1)
		}

		if u.all {
			sql.WriteString(" UNION ALL ")
		} else {
			sql.WriteString(" UNION ")
		}
		sql.WriteString("(")
		sql.WriteString(unionSQL)
		sql.WriteString(")")

		args = append(args, unionArgs...)
		index += len(unionArgs)
	}

	return sql.String(), args, index, nil
}
