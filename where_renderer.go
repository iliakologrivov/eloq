package eloq

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var dollarPlaceholderRe = regexp.MustCompile(`\$(\d+)`)

func shiftDollarPlaceholders(sql string, offset int) string {
	if offset == 0 {
		return sql
	}

	return dollarPlaceholderRe.ReplaceAllStringFunc(sql, func(m string) string {
		n, err := strconv.Atoi(m[1:])
		if err != nil {
			return m
		}
		return fmt.Sprintf("$%d", n+offset)
	})
}

func (b *baseBuilder) renderSubqueryValue(
	value interface{},
	startIndex int,
) (string, []interface{}, int, bool, error) {
	sub, ok := value.(Subquery)
	if !ok {
		return "", nil, startIndex, false, nil
	}

	subSQL, subArgs, err := sub.builder.ToSql()
	if err != nil {
		return "", nil, startIndex, true, err
	}

	if b.placeholder == Dollar && startIndex > 1 {
		subSQL = shiftDollarPlaceholders(subSQL, startIndex-1)
	}

	return subSQL, subArgs, startIndex + len(subArgs), true, nil
}

// entry point
func (b *baseBuilder) renderWheres(
	wheres []whereClause,
	startIndex int,
) (string, []interface{}, int, error) {
	if len(wheres) == 0 {
		return "", nil, startIndex, nil
	}

	var parts []string
	var allBindings []interface{}
	index := startIndex

	for i, w := range wheres {
		sql, bindings, next, err := b.renderWhere(w, index)
		if err != nil {
			return "", nil, startIndex, err
		}

		index = next

		if sql == "" {
			continue
		}

		if i > 0 && len(parts) > 0 {
			if w.isOr {
				sql = "OR " + sql
			} else {
				sql = "AND " + sql
			}
		}

		parts = append(parts, sql)
		allBindings = append(allBindings, bindings...)
	}

	return strings.Join(parts, " "), allBindings, index, nil
}

func (b *baseBuilder) renderWhere(
	w whereClause,
	startIndex int,
) (string, []interface{}, int, error) {
	// Nested
	if w.nested != nil {
		sql, bindings, next, err := b.renderWheres(w.nested, startIndex)
		if err != nil {
			return "", nil, startIndex, err
		}
		if sql == "" {
			return "", nil, startIndex, nil
		}
		return "(" + sql + ")", bindings, next, nil
	}

	col, err := b.quoteIdentifier(w.column)
	if err != nil {
		return "", []interface{}{}, startIndex, err
	}

	// NULL handling
	if w.value == nil || w.operator == "IS NULL" {
		isNot := w.isNot

		switch w.operator {
		case "=", "IS", "IS NULL", "":
			if isNot {
				return fmt.Sprintf("%s IS NOT NULL", col), []interface{}{}, startIndex, nil
			}
			return fmt.Sprintf("%s IS NULL", col), []interface{}{}, startIndex, nil

		case "!=", "<>", "IS NOT NULL", "IS NOT":
			if isNot {
				return fmt.Sprintf("%s IS NULL", col), []interface{}{}, startIndex, nil
			}
			return fmt.Sprintf("%s IS NOT NULL", col), []interface{}{}, startIndex, nil

		default:
			return "", []interface{}{}, startIndex, fmt.Errorf("eloq: invalid operator %q for NULL", w.operator)
		}
	}

	switch w.operator {
	case "EXISTS":
		subSQL, subArgs, nextIndex, ok, err := b.renderSubqueryValue(w.value, startIndex)
		if err != nil {
			return "", nil, startIndex, err
		}
		if ok {
			sql := fmt.Sprintf("%s (%s)", w.operator, subSQL)
			return sql, subArgs, nextIndex, nil
		}

		return "", []interface{}{}, startIndex, fmt.Errorf("eloq: EXISTS operator requires a subquery")
	case "IN":
		op := w.operator
		if w.isNot {
			op = "NOT " + op
		}

		subSQL, subArgs, nextIndex, ok, err := b.renderSubqueryValue(w.value, startIndex)
		if err != nil {
			return "", nil, startIndex, err
		}
		if ok {
			sql := fmt.Sprintf("%s %s (%s)", col, op, subSQL)
			return sql, subArgs, nextIndex, nil
		}

		values, ok := w.value.([]interface{})
		if !ok || len(values) == 0 {
			return "", nil, startIndex, fmt.Errorf("eloq: empty IN condition for %q", w.column)
		}

		var ph []string
		idx := startIndex
		for range values {
			ph = append(ph, b.formatPlaceholder(idx))
			idx++
		}

		sql := fmt.Sprintf("%s %s (%s)", col, op, strings.Join(ph, ", "))
		return sql, values, idx, nil

	case "BETWEEN":
		values, ok := w.value.([]interface{})
		if !ok || len(values) != 2 {
			return "", []interface{}{}, startIndex, fmt.Errorf("eloq: BETWEEN expects 2 values for %q", w.column)
		}

		ph1 := b.formatPlaceholder(startIndex)
		ph2 := b.formatPlaceholder(startIndex + 1)

		if w.isNot {
			return fmt.Sprintf("%s NOT BETWEEN %s AND %s", col, ph1, ph2), values, startIndex + 2, nil
		}

		return fmt.Sprintf("%s BETWEEN %s AND %s", col, ph1, ph2), values, startIndex + 2, nil

	default:
		subSQL, subArgs, nextIndex, ok, err := b.renderSubqueryValue(w.value, startIndex)
		if err != nil {
			return "", nil, startIndex, err
		}
		if ok {
			sql := fmt.Sprintf("%s %s (%s)", col, w.operator, subSQL)
			return sql, subArgs, nextIndex, nil
		}

		if rawVal, ok := w.value.(rawSqlRef); ok {
			sql := fmt.Sprintf("%s %s %s", col, w.operator, rawVal)
			return sql, []interface{}{}, startIndex, nil
		}

		ph := b.formatPlaceholder(startIndex)
		sql := fmt.Sprintf("%s %s %s", col, w.operator, ph)
		return sql, []interface{}{w.value}, startIndex + 1, nil
	}
}
