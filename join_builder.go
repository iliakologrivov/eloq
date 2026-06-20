package eloq

type JoinBuilder struct {
	baseBuilder
	ons    []joinOnClause
	wheres []whereClause
	alias  string
}

func newJoinBuilder(cfg Config) *JoinBuilder {
	return &JoinBuilder{
		baseBuilder: baseBuilder{
			Config:     cfg,
			queryState: newQueryState(),
		},
	}
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
	j.wheres = j.addWhere(j.wheres, false, column, args...)
	return j
}

func (j *JoinBuilder) OrWhere(column string, args ...interface{}) *JoinBuilder {
	j.wheres = j.addWhere(j.wheres, true, column, args...)
	return j
}

func (j *JoinBuilder) As(alias string) *JoinBuilder {
	j.alias = alias

	return j
}
