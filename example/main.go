package main

import (
	"context"
	"log/slog"
	"os"

	"ella.to/logger"
	"ella.to/logger/example/pkg/util"
)

func main() {
	logger.Setup(
		logger.WithSetDefault(),
		logger.WithJsonHandler(os.Stdout, logger.Debug, false),
		logger.WithExporter("http://localhost:2022"),
	)

	ctx := logger.SetAggregateIdToContext(context.Background())

	util.GetInfo(ctx)
	slog.InfoContext(ctx, "Cool")
}
