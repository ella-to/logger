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
		logger.WithTextHandler(os.Stdout, logger.Debug, false),
		logger.WithFilter(
			logger.FindFunc("Get.*"),
		),
	)

	util.GetInfo()
	slog.Info("Cool")
}
