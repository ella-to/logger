package format

import (
	"context"
	"log/slog"
)

func GetCurrency(ctx context.Context) string {
	slog.DebugContext(ctx, "Getting currency")
	return "USD"
}
