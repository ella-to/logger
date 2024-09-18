package logger

import (
	"context"
	"net/http"
)

type ctxKey string

const (
	aggregateIdKey ctxKey = "logger_aggregate_id"
)

const (
	aggregateIdHeaderKey = "X-Logger-Aggregate-Id"
)

type Header interface {
	Set(key, value string)
	Get(key string) string
}

func InjectAggregateId() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			foundInHeader := true
			foundInContext := true

			ctxAggregateId := GetAggregateIdFromContext(ctx)
			if ctxAggregateId == "" {
				foundInContext = false
			}

			headerAggregateId := r.Header.Get(aggregateIdHeaderKey)
			if headerAggregateId == "" {
				foundInHeader = false
			}

			var aggregateId string

			if !foundInHeader && !foundInContext {
				aggregateId = genId()
			} else if foundInContext {
				aggregateId = ctxAggregateId
			} else {
				aggregateId = headerAggregateId
			}

			if !foundInContext {
				r = r.WithContext(context.WithValue(r.Context(), aggregateIdKey, aggregateId))
			}

			w.Header().Set(aggregateIdHeaderKey, aggregateId)

			next.ServeHTTP(w, r)
		})
	}
}

func SetAggregateIdToContext(ctx context.Context) context.Context {
	aggregateId := ctx.Value(aggregateIdKey)
	if aggregateId == nil {
		aggregateId = genId()
	}

	return context.WithValue(ctx, aggregateIdKey, aggregateId)
}

func GetAggregateIdFromContext(ctx context.Context) string {
	aggregateId := ctx.Value(aggregateIdKey)
	if aggregateId == nil {
		return ""
	}

	return aggregateId.(string)
}
