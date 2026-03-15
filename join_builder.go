package eloq

type JoinBuilder struct {
	ons    []joinOnClause
	wheres []whereClause
}

func (j *JoinBuilder) On(left, operator, right string) *JoinBuilder {
	j.ons = append(j.ons, joinOnClause{
		left:     left,
		operator: operator,
		right:    right,
	})
	return j
}

func (j *JoinBuilder) OrOn(left, operator, right string) *JoinBuilder {
	j.ons = append(j.ons, joinOnClause{
		left:     left,
		operator: operator,
		right:    right,
		isOr:     true,
	})
	return j
}

func (j *JoinBuilder) Where(column string, args ...interface{}) *JoinBuilder {
	if len(args) == 0 {
		return j
	}

	j.wheres = append(j.wheres, whereClause{
		column:   column,
		operator: "=",
		value:    args[0],
	})
	return j
}

func (j *JoinBuilder) OrWhere(column string, args ...interface{}) *JoinBuilder {
	if len(args) == 0 {
		return j
	}

	j.wheres = append(j.wheres, whereClause{
		column:   column,
		operator: "=",
		value:    args[0],
		isOr:     true,
	})
	return j
}
