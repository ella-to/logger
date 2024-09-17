package util

import (
	"context"
	"log/slog"
	"sync"

	"ella.to/logger/example/pkg/format"
)

func GetInfo(ctx context.Context) {
	slog.DebugContext(ctx, "This is a log message from GetInfo")

	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()

		format.GetCurrency(ctx)
		slog.DebugContext(ctx, "This is a log message from a goroutine 1")

	}()

	go func() {
		defer wg.Done()

		slog.DebugContext(ctx, "This is a log message from a goroutine 2")
	}()

	wg.Wait()
}
