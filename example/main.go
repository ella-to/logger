package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync/atomic"

	"ella.to/logger"
	"ella.to/logger/example/pkg/util"
)

func createGen() func() string {
	var id atomic.Uint64
	return func() string {
		return fmt.Sprintf("id-%d", id.Add(1))
	}
}

func main() {
	gen := createGen()

	logger.Setup(
		logger.WithSetDefault(),
		logger.WithJsonHandler(os.Stdout, logger.Debug, false),
		logger.WithCustomAggregateIdGen(gen),
		logger.WithExporter("http://localhost:2022"),
	)

	ctx := logger.SetAggregateIdToContext(context.Background(), gen)

	util.GetInfo(ctx)
	slog.InfoContext(ctx, "Cool")
}
