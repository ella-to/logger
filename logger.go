package logger

import (
	"io"
	"log/slog"
	"os"
)

type Logger struct {
	*slog.Logger
}

type options struct {
	level     slog.Level
	addSource bool
	writer    io.Writer
}

type optFunc func(*options)

func WithLevel(level slog.Level) optFunc {
	return func(opt *options) {
		opt.level = level
	}
}

func WithWriter(writer io.Writer) optFunc {
	return func(opt *options) {
		opt.writer = writer
	}
}

func WithSource() func(opt *options) {
	return func(opt *options) {
		opt.addSource = true
	}
}

func Setup(fns ...optFunc) {
	opts := &options{
		level:     slog.LevelInfo,
		addSource: false,
		writer:    os.Stdout,
	}
	for _, fn := range fns {
		fn(opts)
	}

	logger := slog.New(slog.NewTextHandler(opts.writer, &slog.HandlerOptions{
		Level:     opts.level,
		AddSource: opts.addSource,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			return a
		},
	}))

	slog.SetDefault(logger)
}
