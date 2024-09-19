package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type Record struct {
	Id        string         `json:"id"`
	ParentId  string         `json:"parent_id"`
	Level     string         `json:"level"`
	Message   string         `json:"message"`
	Meta      map[string]any `json:"meta"`
	Timestamp string         `json:"timestamp"`
}

type HttpExporter struct {
	handler slog.Handler
	buffer  *buffer[*Record]
}

var _ slog.Handler = (*HttpExporter)(nil)

func (e *HttpExporter) Enabled(ctx context.Context, level slog.Level) bool {
	return e.handler.Enabled(ctx, level)
}

func (e *HttpExporter) Handle(ctx context.Context, r slog.Record) error {
	err := e.handler.Handle(ctx, r)
	if err != nil {
		return err
	}

	logId := getLogId(ctx)
	parentId := getLogParentId(ctx)
	pkg, fn := getPkgFunc(5)

	meta := make(map[string]any)
	r.Attrs(func(a slog.Attr) bool {
		meta[a.Key] = a.Value.Any()
		return true
	})

	meta["pkg"] = pkg
	meta["fn"] = fn

	e.buffer.Add(&Record{
		Id:        logId,
		ParentId:  parentId,
		Level:     r.Level.String(),
		Message:   r.Message,
		Meta:      meta,
		Timestamp: r.Time.Format(time.RFC3339),
	})

	return nil
}

func (e *HttpExporter) WithAttrs(attrs []slog.Attr) slog.Handler {
	return e.handler.WithAttrs(attrs)
}

func (e *HttpExporter) WithGroup(name string) slog.Handler {
	return e.handler.WithGroup(name)
}

func (e *HttpExporter) Close() {
	e.buffer.Close()
}

type httpExporterOpts struct {
	size     int
	interval time.Duration
}

type httpExporterOptsFn func(*httpExporterOpts)

func WithBufferSize(size int) httpExporterOptsFn {
	return func(opts *httpExporterOpts) {
		opts.size = size
	}
}

func WithInterval(interval time.Duration) httpExporterOptsFn {
	return func(opts *httpExporterOpts) {
		opts.interval = interval
	}
}

func NewHttpExporter(addr string, handler slog.Handler, opts ...httpExporterOptsFn) *HttpExporter {
	httpOpts := &httpExporterOpts{
		size:     100,
		interval: 500 * time.Millisecond,
	}

	for _, opt := range opts {
		opt(httpOpts)
	}

	addr = fmt.Sprintf("%s/logs", addr)

	httpClient := &http.Client{
		Timeout: 5 * time.Second,
	}

	return &HttpExporter{
		handler: handler,
		buffer: newBuffer(httpOpts.size, httpOpts.interval, func(records []*Record) error {
			pr, pw := io.Pipe()
			go func() {
				var err error

				for i, record := range records {
					if i > 0 {
						_, err = io.WriteString(pw, "\n")
						if err != nil {
							break
						}
					}

					err := json.NewEncoder(pw).Encode(record)
					if err != nil {
						break
					}
				}

				if err != nil {
					pw.CloseWithError(err)
				} else {
					pw.Close()
				}
			}()

			req, err := http.NewRequest(http.MethodPost, addr, pr)
			if err != nil {
				return err
			}

			resp, err := httpClient.Do(req)
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
		}),
	}
}
