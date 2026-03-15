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

func (b *commonBuilder) renderJoins(
	joinClauses []joinClause,
	startIndex int,
) (string, []interface{}, int, error) {
	if len(joinClauses) == 0 {
		return "", []interface{}{}, startIndex, nil
	}

	var sql strings.Builder

	bindings := []interface{}{}
	for _, j := range joinClauses {
		if j.raw {
			sql.WriteString(" ")
			sql.WriteString(j.rawSQL)
			continue
		}

		sql.WriteString(" ")
		sql.WriteString(j.joinType)
		sql.WriteString(" JOIN ")

		tbl, err := b.quoteIdentifier(j.table)
		if err != nil {
			return "", []interface{}{}, startIndex, err
		}
		sql.WriteString(tbl)

		if len(j.ons) == 0 && len(j.wheres) == 0 {
			continue
		}

		sql.WriteString(" ON ")

		var parts []string

		// ON conditions (column-to-column)
		for _, on := range j.ons {
			left, err := b.quoteIdentifier(on.left)
			if err != nil {
				return "", []interface{}{}, startIndex, err
			}

			right, err := b.quoteIdentifier(on.right)
			if err != nil {
				return "", []interface{}{}, startIndex, err
			}

			expr := left + " " + on.operator + " " + right

			if len(parts) > 0 {
				if on.isOr {
					expr = "OR " + expr
				} else {
					expr = "AND " + expr
				}
			}

			parts = append(parts, expr)
		}

		// WHERE inside JOIN
		if len(j.wheres) > 0 {
			whereSQL, whereBindings, next, err := b.renderWheres(j.wheres, startIndex)
			if err != nil {
				return "", []interface{}{}, startIndex, err
			}

			if whereSQL != "" {
				if len(parts) > 0 {
					whereSQL = "AND (" + whereSQL + ")"
				}
				parts = append(parts, whereSQL)

				bindings = append(bindings, whereBindings...)
				startIndex = next
			}
		}

		sql.WriteString(strings.Join(parts, " "))
	}

	return sql.String(), bindings, startIndex, nil
}
