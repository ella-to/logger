package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"ella.to/logger"
)

func main() {

	slog.SetDefault(
		slog.New(
			logger.NewHttpExporter(
				"http://localhost:2022", // logger server address
				slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
					Level: slog.LevelDebug,
				}),
			),
		),
	)

	ctx := context.Background()

	ctx = logger.Info(ctx, "app started")

	fn1(ctx)

	time.Sleep(2 * time.Second)
}

func fn1(ctx context.Context) {
	logger.Info(ctx, "fn1 started")
	defer logger.Debug(ctx, "fn1 finished")

	// time.Sleep(1 * time.Second)
}
