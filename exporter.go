package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"
)

type HttpExporter struct {
	addr    string
	client  *http.Client
	handler slog.Handler
	counter atomic.Int64
}

var _ slog.Handler = (*HttpExporter)(nil)

func (e *HttpExporter) genId() string {
	return fmt.Sprintf("%d", e.counter.Add(1))
}

func (e *HttpExporter) Enabled(ctx context.Context, level slog.Level) bool {
	return e.handler.Enabled(ctx, level)
}

func (e *HttpExporter) Handle(ctx context.Context, r slog.Record) error {
	err := e.handler.Handle(ctx, r)
	if err != nil {
		return err
	}

	type Record struct {
		Id        string         `json:"id"`
		Level     string         `json:"level"`
		Message   string         `json:"message"`
		Meta      map[string]any `json:"meta"`
		Timestamp string         `json:"timestamp"`
	}

	meta := make(map[string]any)
	r.Attrs(func(a slog.Attr) bool {
		meta[a.Key] = a.Value.Any()
		return true
	})

	logRecord := Record{
		Id:        e.genId(),
		Level:     r.Level.String(),
		Message:   r.Message,
		Meta:      meta,
		Timestamp: r.Time.Format(time.RFC3339),
	}

	pr, pw := io.Pipe()
	go func() {
		err := json.NewEncoder(pw).Encode(logRecord)
		if err != nil {
			pw.CloseWithError(err)
		} else {
			pw.Close()
		}
	}()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.addr, pr)
	if err != nil {
		return err
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		errMsg, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, errMsg)
	}

	return nil
}

func (e *HttpExporter) WithAttrs(attrs []slog.Attr) slog.Handler {
	return e.handler.WithAttrs(attrs)
}

func (e *HttpExporter) WithGroup(name string) slog.Handler {
	return e.handler.WithGroup(name)
}

func NewHttpExporter(addr string, handler slog.Handler) *HttpExporter {
	return &HttpExporter{
		addr:    fmt.Sprintf("%s/logs", addr),
		client:  &http.Client{},
		handler: handler,
	}
}
