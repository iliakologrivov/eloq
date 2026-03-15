package eloq

func (b *SelectBuilder) Having(column string, args ...interface{}) *SelectBuilder {
	b.havings = b.commonBuilder.addWhere(b.havings, false, column, args...)
	return b
}

func (b *SelectBuilder) OrHaving(column string, args ...interface{}) *SelectBuilder {
	b.havings = b.commonBuilder.addWhere(b.havings, true, column, args...)
	return b
}

func (b *SelectBuilder) HavingNested(fn func(*SelectBuilder)) *SelectBuilder {
	nestedBuilder := &SelectBuilder{
		commonBuilder: &commonBuilder{
			placeholder: b.placeholder,
			quoteStyle:  b.quoteStyle,
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
