package eloq

import (
	"fmt"
	"strconv"
	"strings"
)

func shiftDollarPlaceholders(sql string, offset int) string {
	if offset == 0 {
		return sql
	}

	var sb strings.Builder
	sb.Grow(len(sql) + len(sql)/4)

	i := 0
	for i < len(sql) {
		if sql[i] == '$' && i+1 < len(sql) && sql[i+1] >= '0' && sql[i+1] <= '9' {
			j := i + 1
			for j < len(sql) && sql[j] >= '0' && sql[j] <= '9' {
				j++
			}
			numStr := sql[i+1 : j]
			if n, err := strconv.Atoi(numStr); err == nil {
				sb.WriteByte('$')
				sb.WriteString(strconv.Itoa(n + offset))
			} else {
				sb.WriteString(sql[i:j])
			}
			i = j
		} else {
			sb.WriteByte(sql[i])
			i++
		}
	}
	return sb.String()
}

// renderSubqueryValue пишет результат напрямую в builder
func (b *baseBuilder) renderSubqueryValue(
	sql *strings.Builder,
	value interface{},
	startIndex int,
) ([]interface{}, int, bool, error) {
	sub, ok := value.(Subquery)
	if !ok {
		return nil, startIndex, false, nil
	}

	subSQL, subArgs, err := sub.builder.ToSql()
	if err != nil {
		return nil, startIndex, true, err
	}

	if b.placeholder == Dollar && startIndex > 1 {
		subSQL = shiftDollarPlaceholders(subSQL, startIndex-1)
	}

	sql.WriteString(subSQL)
	return subArgs, startIndex + len(subArgs), true, nil
}

// renderWheres пишет результат напрямую в builder
// Возвращает (bindings, nextIndex, error)
func (b *baseBuilder) renderWheres(
	sql *strings.Builder,
	wheres []whereClause,
	startIndex int,
) ([]interface{}, int, error) {
	if len(wheres) == 0 {
		return nil, startIndex, nil
	}

	var allBindings []interface{}
	index := startIndex

	for i, w := range wheres {
		// Для вложенных условий получаем временный builder
		if w.nested != nil {
			var nestedSQL strings.Builder
			bindings, next, err := b.renderWheres(&nestedSQL, w.nested, index)
			if err != nil {
				return nil, startIndex, err
			}
			index = next

			nestedStr := nestedSQL.String()
			if nestedStr == "" {
				continue
			}

			if i > 0 {
				if w.isOr {
					sql.WriteString(" OR ")
				} else {
					sql.WriteString(" AND ")
				}
			}
			sql.WriteByte('(')
			sql.WriteString(nestedStr)
			sql.WriteByte(')')
			allBindings = append(allBindings, bindings...)
			continue
		}

		// Обычное условие
		beforeLen := sql.Len()
		bindings, next, err := b.renderWhere(sql, w, index)
		if err != nil {
			return nil, startIndex, err
		}
		index = next

		// Если ничего не добавлено
		if sql.Len() == beforeLen {
			continue
		}

		// Добавляем AND/OR перед условием (кроме первого)
		if i > 0 {
			prefix := " AND "
			if w.isOr {
				prefix = " OR "
			}
			// Вставляем prefix в начало
			result := sql.String()
			before := result[:beforeLen]
			after := result[beforeLen:]
			sql.Reset()
			sql.WriteString(before)
			sql.WriteString(prefix)
			sql.WriteString(after)
		}

		allBindings = append(allBindings, bindings...)
	}

	return allBindings, index, nil
}

// renderWhere пишет одно условие WHERE напрямую в builder
// Возвращает (bindings, nextIndex, error)
func (b *baseBuilder) renderWhere(
	sql *strings.Builder,
	w whereClause,
	startIndex int,
) ([]interface{}, int, error) {
	col, err := b.quoteIdentifier(w.column)
	if err != nil {
		return nil, startIndex, err
	}

	// NULL handling
	if w.value == nil || w.operator == "IS NULL" {
		sql.WriteString(col)
		isNot := w.isNot

		switch w.operator {
		case "=", "IS", "IS NULL", "":
			if isNot {
				sql.WriteString(" IS NOT NULL")
			} else {
				sql.WriteString(" IS NULL")
			}
			return nil, startIndex, nil

		case "!=", "<>", "IS NOT NULL", "IS NOT":
			if isNot {
				sql.WriteString(" IS NULL")
			} else {
				sql.WriteString(" IS NOT NULL")
			}
			return nil, startIndex, nil

		default:
			return nil, startIndex, fmt.Errorf("eloq: invalid operator %q for NULL", w.operator)
		}
	}

	switch w.operator {
	case "EXISTS":
		sub, ok := w.value.(Subquery)
		if !ok {
			return nil, startIndex, fmt.Errorf("eloq: EXISTS operator requires a subquery")
		}

		subSQL, subArgs, err := sub.builder.ToSql()
		if err != nil {
			return nil, startIndex, err
		}

		if b.placeholder == Dollar && startIndex > 1 {
			subSQL = shiftDollarPlaceholders(subSQL, startIndex-1)
		}

		sql.WriteString(w.operator)
		sql.WriteString(" (")
		sql.WriteString(subSQL)
		sql.WriteByte(')')
		return subArgs, startIndex + len(subArgs), nil

	case "IN":
		op := w.operator
		if w.isNot {
			op = "NOT " + op
		}

		// Проверяем subquery
		var tmpSQL strings.Builder
		subArgs, nextIndex, ok, err := b.renderSubqueryValue(&tmpSQL, w.value, startIndex)
		if err != nil {
			return nil, startIndex, err
		}
		if ok {
			sql.WriteString(col)
			sql.WriteByte(' ')
			sql.WriteString(op)
			sql.WriteString(" (")
			sql.WriteString(tmpSQL.String())
			sql.WriteByte(')')
			return subArgs, nextIndex, nil
		}

		values, ok := w.value.([]interface{})
		if !ok || len(values) == 0 {
			return nil, startIndex, fmt.Errorf("eloq: empty IN condition for %q", w.column)
		}

		sql.WriteString(col)
		sql.WriteByte(' ')
		sql.WriteString(op)
		sql.WriteString(" (")

		idx := startIndex
		for i := range values {
			if i > 0 {
				sql.WriteString(", ")
			}
			sql.WriteString(b.formatPlaceholder(idx))
			idx++
		}
		sql.WriteByte(')')
		return values, idx, nil

	case "BETWEEN":
		values, ok := w.value.([]interface{})
		if !ok || len(values) != 2 {
			return nil, startIndex, fmt.Errorf("eloq: BETWEEN expects 2 values for %q", w.column)
		}

		ph1 := b.formatPlaceholder(startIndex)
		ph2 := b.formatPlaceholder(startIndex + 1)

		sql.WriteString(col)
		if w.isNot {
			sql.WriteString(" NOT BETWEEN ")
		} else {
			sql.WriteString(" BETWEEN ")
		}
		sql.WriteString(ph1)
		sql.WriteString(" AND ")
		sql.WriteString(ph2)
		return values, startIndex + 2, nil

	default:
		// Проверяем subquery
		var tmpSQL strings.Builder
		subArgs, nextIndex, ok, err := b.renderSubqueryValue(&tmpSQL, w.value, startIndex)
		if err != nil {
			return nil, startIndex, err
		}
		if ok {
			sql.WriteString(col)
			sql.WriteByte(' ')
			sql.WriteString(w.operator)
			sql.WriteString(" (")
			sql.WriteString(tmpSQL.String())
			sql.WriteByte(')')
			return subArgs, nextIndex, nil
		}

		if rawVal, ok := w.value.(rawSqlRef); ok {
			sql.WriteString(col)
			sql.WriteByte(' ')
			sql.WriteString(w.operator)
			sql.WriteByte(' ')
			sql.WriteString(string(rawVal))
			return nil, startIndex, nil
		}

		ph := b.formatPlaceholder(startIndex)
		sql.WriteString(col)
		sql.WriteByte(' ')
		sql.WriteString(w.operator)
		sql.WriteByte(' ')
		sql.WriteString(ph)
		return []interface{}{w.value}, startIndex + 1, nil
	}
}
