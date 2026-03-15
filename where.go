package eloq

import (
	"strings"
)

type whereClause struct {
	column   string
	operator string
	value    interface{}

	isOr  bool
	isNot bool

	// nested WHERE
	nested []whereClause
}

type Subquery struct {
	builder interface {
		ToSql() (string, []any, error)
	}
}

type subqueryBuilder interface {
	ToSql() (string, []any, error)
}

func normalizeWhereValue(value interface{}) interface{} {
	switch v := value.(type) {
	case Subquery:
		return v
	case subqueryBuilder:
		return Subquery{
			builder: v,
		}
	default:
		return value
	}
}

func (b *commonBuilder) addWhere(wheres []whereClause, isOr bool, column string, args ...interface{}) []whereClause {
	if len(args) == 0 {
		return wheres
	}

	var operator string
	var value interface{}

	switch len(args) {
	case 1:
		operator = "="
		value = args[0]
	case 2:
		op, ok := args[0].(string)
		if !ok {
			return wheres
		}
		operator = strings.ToUpper(op)
		value = args[1]
	default:
		return wheres
	}

	if value == nil {
		if operator == "=" || operator == "" {
			wheres = append(wheres, whereClause{
				column: column,
				isOr:   isOr,
				value:  nil,
				isNot:  false,
			})
		} else if operator == "!=" || operator == "<>" {
			wheres = append(wheres, whereClause{
				column: column,
				isOr:   isOr,
				value:  nil,
				isNot:  true,
			})
		}
		return wheres
	}

	return append(wheres, whereClause{
		column:   column,
		isOr:     isOr,
		value:    normalizeWhereValue(value),
		operator: operator,
	})
}

func (b *commonBuilder) addWhereIn(wheres []whereClause, column string, values []interface{}, isNot bool) []whereClause {
	return append(wheres, whereClause{
		column:   column,
		operator: "IN",
		value:    values,
		isNot:    isNot,
	})
}

func (b *commonBuilder) addWhereNull(wheres []whereClause, column string, isNot bool) []whereClause {
	return append(wheres, whereClause{
		column: column,
		isNot:  isNot,
		value:  nil,
	})
}

func (b *commonBuilder) addWhereBetween(wheres []whereClause, column string, from, to interface{}, isNot bool) []whereClause {
	wheres = append(wheres, whereClause{
		column:   column,
		operator: "BETWEEN",
		isNot:    isNot,
		value:    []interface{}{from, to},
	})

	return wheres
}
