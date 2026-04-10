package eloq

import "strings"

type joinClause struct {
	joinType string // INNER, LEFT, RIGHT
	table    string

	ons    []joinOnClause
	wheres []whereClause

	raw    bool
	rawSQL string
}

type joinOnClause struct {
	left     string
	operator string
	right    string
	isOr     bool
}

// renderJoins пишет результат напрямую в builder
// Возвращает (bindings, nextIndex, error)
func (b *baseBuilder) renderJoins(
	sql *strings.Builder,
	joinClauses []joinClause,
	startIndex int,
) ([]interface{}, int, error) {
	if len(joinClauses) == 0 {
		return nil, startIndex, nil
	}

	bindings := []interface{}{}
	firstOn := true

	for _, j := range joinClauses {
		if j.raw {
			sql.WriteByte(' ')
			sql.WriteString(j.rawSQL)
			continue
		}

		sql.WriteByte(' ')
		sql.WriteString(j.joinType)
		sql.WriteString(" JOIN ")

		tbl, err := b.quoteIdentifier(j.table)
		if err != nil {
			return nil, startIndex, err
		}
		sql.WriteString(tbl)

		if len(j.ons) == 0 && len(j.wheres) == 0 {
			continue
		}

		sql.WriteString(" ON ")
		firstOn = true

		for _, on := range j.ons {
			left, err := b.quoteIdentifier(on.left)
			if err != nil {
				return nil, startIndex, err
			}

			right, err := b.quoteIdentifier(on.right)
			if err != nil {
				return nil, startIndex, err
			}

			if !firstOn {
				if on.isOr {
					sql.WriteString(" OR ")
				} else {
					sql.WriteString(" AND ")
				}
			}
			sql.WriteString(left)
			sql.WriteByte(' ')
			sql.WriteString(on.operator)
			sql.WriteByte(' ')
			sql.WriteString(right)
			firstOn = false
		}

		// WHERE inside JOIN
		if len(j.wheres) > 0 {
			if !firstOn {
				sql.WriteString(" AND (")
			} else {
				sql.WriteByte('(')
			}

			whereBindings, next, err := b.renderWheres(sql, j.wheres, startIndex)
			if err != nil {
				return nil, startIndex, err
			}

			sql.WriteByte(')')

			bindings = append(bindings, whereBindings...)
			startIndex = next
		}
	}

	return bindings, startIndex, nil
}
