package logger

import (
	"context"
	"io"
	"net/http"
	"time"
)

const (
	headerKeyLogParentId = "X-Log-Parent-Id"
)

func HttpMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			logParentId := r.Header.Get(headerKeyLogParentId)
			if logParentId == "" {
				logParentId = newId()
			}

			ctx := setLogParentId(r.Context(), logParentId)
			Info(ctx, "received http request")

			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)

			Info(ctx, "finished http request", "method", r.Method, "url", r.URL.String(), "duration", time.Since(start))
		})
	}
}

func NewHttpRequest(ctx context.Context, method string, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	parentId := getLogParentId(ctx)
	req.Header.Set(headerKeyLogParentId, parentId)

	return req, nil
}
