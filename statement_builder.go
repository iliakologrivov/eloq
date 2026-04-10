package eloq

type StatementBuilder struct {
	Config
}

type StatementBuilderType = *StatementBuilder

func NewStatementBuilder() *StatementBuilder {
	return &StatementBuilder{
		Config: Config{
			placeholder:  Question,
			quoteStyle:   Backtick,
			requireWhere: false,
		},
	}
}

var DefaultStatementBuilder = &StatementBuilder{
	Config: Config{
		placeholder:  Question,
		quoteStyle:   Backtick,
		requireWhere: false,
	},
}

func (sb *StatementBuilder) PlaceholderFormat(f PlaceholderFormat) *StatementBuilder {
	return &StatementBuilder{
		Config: Config{
			placeholder:  f,
			quoteStyle:   sb.quoteStyle,
			requireWhere: sb.requireWhere,
		},
	}
}

func (sb *StatementBuilder) QuoteWith(q QuoteStyle) *StatementBuilder {
	return &StatementBuilder{
		Config: Config{
			placeholder:  sb.placeholder,
			quoteStyle:   q,
			requireWhere: sb.requireWhere,
		},
	}
}

func (sb *StatementBuilder) RequireWhere(require bool) *StatementBuilder {
	return &StatementBuilder{
		Config: Config{
			placeholder:  sb.placeholder,
			quoteStyle:   sb.quoteStyle,
			requireWhere: require,
		},
	}
}

func (sb *StatementBuilder) Select(columns ...string) *SelectBuilder {
	b := &SelectBuilder{
		baseBuilder: newBaseBuilderWithConfig(sb.Config),
		columns:     make([]selectColumn, 0),
		bindings:    make([]interface{}, 0),
	}
	if len(columns) > 0 {
		b.columns = nil
		for _, col := range columns {
			b.columns = append(b.columns, selectColumn{value: col, raw: false})
		}
	}
	return b
}

func (sb *StatementBuilder) SelectRaw(columns ...string) *SelectBuilder {
	b := &SelectBuilder{
		baseBuilder: newBaseBuilderWithConfig(sb.Config),
		columns:     make([]selectColumn, 0),
		bindings:    make([]interface{}, 0),
	}
	if len(columns) > 0 {
		b.columns = nil
		for _, col := range columns {
			b.columns = append(b.columns, selectColumn{value: col, raw: true})
		}
	}
	return b
}

func (sb *StatementBuilder) Insert(table string) *InsertBuilder {
	return &InsertBuilder{
		baseBuilder: newBaseBuilderWithConfig(sb.Config),
		table:       table,
	}
}

func (sb *StatementBuilder) Update(table string) *UpdateBuilder {
	return &UpdateBuilder{
		baseBuilder: newBaseBuilderWithConfig(sb.Config),
		table:       table,
		values:      map[string]interface{}{},
	}
}

func (sb *StatementBuilder) Delete(table string) *DeleteBuilder {
	return &DeleteBuilder{
		baseBuilder: newBaseBuilderWithConfig(sb.Config),
		table:       table,
	}
}
