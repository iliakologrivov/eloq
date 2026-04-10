package eloq

func (b *SelectBuilder) Having(column string, args ...interface{}) *SelectBuilder {
	b.havings = b.baseBuilder.addWhere(b.havings, false, column, args...)
	return b
}

func (b *SelectBuilder) OrHaving(column string, args ...interface{}) *SelectBuilder {
	b.havings = b.baseBuilder.addWhere(b.havings, true, column, args...)
	return b
}

func (b *SelectBuilder) HavingNested(fn func(*SelectBuilder)) *SelectBuilder {
	nestedBuilder := &SelectBuilder{
		baseBuilder: baseBuilder{
			Config:     b.Config,
			queryState: newQueryState(),
		},
	}

	fn(nestedBuilder)

	if len(nestedBuilder.havings) == 0 {
		return b
	}

	b.havings = append(b.havings, whereClause{
		nested: nestedBuilder.havings,
	})

	return b
}
