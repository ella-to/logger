package logger

import (
	"context"
	"net/http"
)

type CtxKey string

const (
	AggregateIdKey CtxKey = "logger_aggregate_id"
)

const (
	AggregateIdHeaderKey = "X-Logger-Aggregate-Id"
)

type Header interface {
	Set(key, value string)
	Get(key string) string
}

func EnhanacedHandler(gen func() string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			foundInHeader := true
			foundInContext := true

			ctxAggregateId := FromContext(ctx)
			if ctxAggregateId == "" {
				foundInContext = false
			}

			headerAggregateId := FromHeader(r.Header)
			if headerAggregateId == "" {
				foundInHeader = false
			}

			var aggregateId string

			if !foundInHeader && !foundInContext {
				aggregateId = gen()
			} else if foundInContext {
				aggregateId = ctxAggregateId
			} else {
				aggregateId = headerAggregateId
			}

			if !foundInContext {
				r = r.WithContext(context.WithValue(r.Context(), AggregateIdKey, aggregateId))
			}

			w.Header().Set(AggregateIdHeaderKey, aggregateId)

			next.ServeHTTP(w, r)
		})
	}
}

func EnhancedHeader(header Header, gen func() string) Header {
	aggregateId := header.Get(AggregateIdHeaderKey)
	if aggregateId == "" {
		aggregateId = gen()
	}

	header.Set(AggregateIdHeaderKey, aggregateId)

	return header
}

func EnahancedContext(ctx context.Context, gen func() string) context.Context {
	aggregateId := ctx.Value(AggregateIdKey)
	if aggregateId == nil {
		aggregateId = gen()
	}

	return context.WithValue(ctx, AggregateIdKey, aggregateId)
}

func FromContext(ctx context.Context) string {
	aggregateId := ctx.Value(AggregateIdKey)
	if aggregateId == nil {
		return ""
	}

	return aggregateId.(string)
}

func FromHeader(header Header) string {
	return header.Get(AggregateIdHeaderKey)
}
