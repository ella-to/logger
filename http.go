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
			ctx = Info(ctx, "received http request", "method", r.Method, "url", r.URL.String())

			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)

			Info(ctx, "finished http request", "duration", time.Since(start).String())
		})
	}
}

func NewHttpRequest(ctx context.Context, method string, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	HttpUpdateRequest(req)

	return req, nil
}

// HttpUpdateRequest updates the http request with the log parent id added to the header
func HttpUpdateRequest(r *http.Request) {
	parentId := getLogParentId(r.Context())
	r.Header.Set(headerKeyLogParentId, parentId)
}
