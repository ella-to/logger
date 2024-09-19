package logger

import (
	"context"
	"log/slog"
	"runtime"
	"strings"

	"github.com/rs/xid"
)

func Debug(ctx context.Context, msg string, args ...any) context.Context {
	id := newId()
	slog.DebugContext(setLogId(ctx, id), msg, args...)
	return setLogParentId(ctx, id)
}

func Info(ctx context.Context, msg string, args ...any) context.Context {
	id := newId()
	slog.InfoContext(setLogId(ctx, id), msg, args...)
	return setLogParentId(ctx, id)
}

func Warn(ctx context.Context, msg string, args ...any) context.Context {
	id := newId()
	slog.WarnContext(setLogId(ctx, id), msg, args...)
	return setLogParentId(ctx, id)
}

func Error(ctx context.Context, msg string, args ...any) context.Context {
	id := newId()
	slog.ErrorContext(setLogId(ctx, id), msg, args...)
	return setLogParentId(ctx, id)
}

// context

type ctxKey string

const (
	ctxKeyLogId       ctxKey = "log_id"
	ctxKeyLogParentId ctxKey = "log_parent_id"
)

func setLogParentId(ctx context.Context, logParentId string) context.Context {
	return context.WithValue(ctx, ctxKeyLogParentId, logParentId)
}

func getLogParentId(ctx context.Context) string {
	if v, ok := ctx.Value(ctxKeyLogParentId).(string); ok {
		return v
	}
	return ""
}

func setLogId(ctx context.Context, logId string) context.Context {
	return context.WithValue(ctx, ctxKeyLogId, logId)
}

func getLogId(ctx context.Context) string {
	if v, ok := ctx.Value(ctxKeyLogId).(string); ok {
		return v
	}
	return ""
}

func newId() string {
	return xid.New().String()
}

func getPkgFunc(skip int) (pkg, fn string) {
	pc, _, _, ok := runtime.Caller(skip)
	if !ok {
		return
	}
	info := runtime.FuncForPC(pc)
	funcName := info.Name()

	idx := strings.LastIndex(funcName, ".")
	if idx == -1 {
		return
	}

	pkg = funcName[:idx]
	fn = funcName[idx+1:]
	return
}
