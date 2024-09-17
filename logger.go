package logger

import (
	"context"
	"io"
	"log/slog"
	"regexp"
	"runtime"
	"strings"
)

// These are temporary keys which never logged, they are used to store the
// package and function name of the caller to be used in the filters
const pkgAttrKey = "pkg"
const fnAttrKey = "fn"
const aggregateAttrKey = "aggregate_id"

var (
	Debug = slog.LevelDebug
	Info  = slog.LevelInfo
	Warn  = slog.LevelWarn
	Error = slog.LevelError
)

type Filter func(record slog.Record) bool
type Mapper func(groups []string, a slog.Attr) slog.Attr

type Handler struct {
	filters       []Filter
	inner         slog.Handler
	setDefault    bool
	aggregateIdFn func() string
}

var _ slog.Handler = (*Handler)(nil)

func (h *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h *Handler) Handle(ctx context.Context, record slog.Record) error {
	pkg, fn := getCurrentPkgFunc()
	parentPkg, parentFn := getParentPkgFunc()

	attrs := []slog.Attr{
		slog.String(pkgAttrKey, pkg),
		slog.String(fnAttrKey, fn),
		slog.String("parent_pkg", parentPkg),
		slog.String("parent_fn", parentFn),
	}

	aggregateId := GetAggregateIdFromContext(ctx)
	if aggregateId == "" && h.aggregateIdFn != nil {
		aggregateId = h.aggregateIdFn()
	}

	if aggregateId != "" {
		attrs = append(attrs, slog.String(aggregateAttrKey, aggregateId))
	}

	record.AddAttrs(attrs...)

	for _, filter := range h.filters {
		if !filter(record) {
			return nil // Skip the log entry if it does not pass the filter
		}
	}
	return h.inner.Handle(ctx, record)
}

func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &Handler{
		inner:   h.inner.WithAttrs(attrs),
		filters: h.filters,
	}
}

func (h *Handler) WithGroup(name string) slog.Handler {
	return &Handler{
		inner:   h.inner.WithGroup(name),
		filters: h.filters,
	}
}

// Options

type option interface {
	configureHandler(*Handler)
}

type handlerOptionFunc func(*Handler)

var _ option = (*handlerOptionFunc)(nil)

func (f handlerOptionFunc) configureHandler(h *Handler) {
	f(h)
}

func Setup(opts ...option) *slog.Logger {
	l := &Handler{}

	for _, opt := range opts {
		opt.configureHandler(l)
	}

	logger := slog.New(l)

	if l.setDefault {
		slog.SetDefault(logger)
	}

	return logger
}

func WithSetDefault() option {
	return handlerOptionFunc(func(h *Handler) {
		h.setDefault = true
	})
}

func WithTextHandler(w io.Writer, level slog.Leveler, addSource bool) option {
	return handlerOptionFunc(func(h *Handler) {
		h.inner = slog.NewTextHandler(w, &slog.HandlerOptions{
			Level:     level,
			AddSource: addSource,
		})
	})
}

func WithCustomAggregateIdGen(gen func() string) option {
	return handlerOptionFunc(func(h *Handler) {
		h.aggregateIdFn = gen
	})
}

func WithJsonHandler(w io.Writer, level slog.Leveler, addSource bool) option {
	return handlerOptionFunc(func(h *Handler) {
		h.inner = slog.NewJSONHandler(w, &slog.HandlerOptions{
			Level:     level,
			AddSource: addSource,
		})
	})
}

func WithExporter(addr string) option {
	return handlerOptionFunc(func(h *Handler) {
		h.inner = NewHttpExporter(addr, h.inner)
	})
}

func WithHandler(handler slog.Handler) option {
	return handlerOptionFunc(func(h *Handler) {
		h.inner = handler
	})
}

// WithFilter adds a filter to the log handler
// if any of the filters return false the log entry is skipped
func WithFilter(filters ...Filter) option {
	return handlerOptionFunc(func(h *Handler) {
		h.filters = append(h.filters, filters...)
	})
}

// FindPkg returns a filter that checks if the record has a pkg attribute
// that matches the given pattern. If the pattern is empty, the filter
// will always return true. The pattern must be a valid regular expression.
//
// Note: the filter will panics if the pattern is not a valid regular expression
//
// Example: FindPkg(".*pkg.*") will match all records that have a pkg
func FindPkg(pattern string) Filter {
	return func(record slog.Record) (result bool) {
		if pattern == "" {
			return true
		}

		record.Attrs(func(attr slog.Attr) bool {
			if attr.Key == pkgAttrKey {
				result = regexp.MustCompile(pattern).MatchString(attr.Value.String())
				return false
			}
			return true
		})

		return
	}
}

func FindFunc(pattern string) Filter {
	return func(record slog.Record) (result bool) {
		if pattern == "" {
			return true
		}

		record.Attrs(func(attr slog.Attr) bool {
			if attr.Key == fnAttrKey {
				result = regexp.MustCompile(pattern).MatchString(attr.Value.String())
				return false
			}
			return true
		})

		return
	}
}

// AttrHasAnyValues returns a filter that checks if the record has
// an attribute with the given key and any of the given values
func AttrHasAnyValues(key string, values ...any) Filter {
	return func(record slog.Record) (result bool) {
		record.Attrs(func(attr slog.Attr) bool {
			if attr.Key == key {
				for _, value := range values {
					result = attr.Value.Equal(slog.AnyValue(value))
					if result {
						break
					}
				}
				return false
			}
			return true
		})
		return
	}
}

func getCurrentPkgFunc() (pkg, fn string) {
	return getPkgFunc(5)
}

func getParentPkgFunc() (pkg, fn string) {
	return getPkgFunc(6)
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
