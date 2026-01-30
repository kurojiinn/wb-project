package sl

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

func Traced(ctx context.Context) slog.Attr {
	spanContext := trace.SpanFromContext(ctx).SpanContext()
	if spanContext.HasTraceID() {
		return slog.String("trace_id", spanContext.TraceID().String())
	}

	return slog.Any("trace_id", nil)
}
