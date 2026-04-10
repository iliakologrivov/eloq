package eloq

import "errors"

var ErrRequireWhere = errors.New("DELETE/UPDATE requires WHERE clause (RequireWhere is enabled)")

type PlaceholderFormat string

const (
	Question PlaceholderFormat = "?"
	Dollar   PlaceholderFormat = "$"
)

type QuoteStyle string

const (
	Backtick    QuoteStyle = "`"
	DoubleQuote QuoteStyle = `"`
)

type ContextKey string

const (
	ContextTraceID   ContextKey = "trace_id"
	ContextSpanID    ContextKey = "span_id"
	ContextRequestID ContextKey = "request_id"
	ContextUserID    ContextKey = "user_id"
)

type rawSqlRef string

func Raw(name string) rawSqlRef {
	return rawSqlRef(name)
}

type sqlPart struct {
	sql  string
	args []interface{}
}

func ToAnySlice[T interface{}](in []T) []interface{} {
	out := make([]interface{}, len(in))
	for i, v := range in {
		out[i] = v
	}
	return out
}
