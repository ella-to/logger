package logger

import (
	"context"
	"log/slog"
	"os"
	"testing"
)

type testLogIds struct {
	logId       string
	logParentId string
}

type testHandler struct {
	inner slog.Handler
	ids   []testLogIds
}

var _ slog.Handler = (*testHandler)(nil)

func (h *testHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h *testHandler) Handle(ctx context.Context, record slog.Record) error {
	logId := getLogId(ctx)
	parentId := getLogParentId(ctx)

	attrs := []slog.Attr{
		slog.String("log_id", logId),
		slog.String("log_parent_id", parentId),
	}

	h.ids = append(h.ids, testLogIds{
		logId:       logId,
		logParentId: parentId,
	})

	record.AddAttrs(attrs...)

	err := h.inner.Handle(ctx, record)
	if err != nil {
		return err
	}

	return nil
}

func (h *testHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h.inner.WithAttrs(attrs)
}

func (h *testHandler) WithGroup(name string) slog.Handler {
	return h.inner.WithGroup(name)
}

func newTestHandler(inner slog.Handler) *testHandler {
	return &testHandler{inner: inner}
}

func TestLoggerContext(t *testing.T) {
	testHandler := newTestHandler(slog.NewJSONHandler(os.Stdout, nil))

	slog.SetDefault(
		slog.New(
			testHandler,
		),
	)

	ctx := context.Background()

	ctx = Info(ctx, "info message 1")

	ctx = Info(ctx, "info message 2")

	Info(ctx, "info message 3")

	if len(testHandler.ids) != 3 {
		t.Fatalf("expected 3 log ids, got %d", len(testHandler.ids))
	}

	if testHandler.ids[0].logParentId != "" {
		t.Fatalf("expected empty parent id, got %s", testHandler.ids[0].logParentId)
	}

	if testHandler.ids[1].logParentId != testHandler.ids[0].logId {
		t.Fatalf("expected parent id %s, got %s", testHandler.ids[0].logId, testHandler.ids[1].logParentId)
	}

	if testHandler.ids[2].logParentId != testHandler.ids[1].logId {
		t.Fatalf("expected parent id %s, got %s", testHandler.ids[1].logId, testHandler.ids[2].logParentId)
	}
}
