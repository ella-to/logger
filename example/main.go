package main

import (
	"log/slog"
	"os"

	"ella.to/logger"
	"ella.to/logger/example/pkg/util"
)

func main() {
	logger.Setup(
		logger.WithSetDefault(),
		logger.WithJsonHandler(os.Stdout, logger.Debug, false),
		logger.WithCustomAggregateIdGen(func() string {
			return "custom-id"
		}),
		logger.WithFilter(
			logger.FindFunc("Get.*"),
		),
	)

	util.GetInfo()
	slog.Info("Cool")
}
