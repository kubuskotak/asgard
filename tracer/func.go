package tracer

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// StartSpanLogTrace returns a copy of parent with span set as the current Span and logging.
func StartSpanLogTrace(ctx context.Context, spanName string) (
	newCtx context.Context, span trace.Span, logger zerolog.Logger,
) {
	sc := trace.SpanContextFromContext(ctx)
	if sc.IsValid() {
		tr := otel.Tracer("fn." + spanName)
		newCtx, span = tr.Start(ctx, spanName)
	} else {
		newCtx = ctx
		span = trace.SpanFromContext(newCtx)
	}
	return trace.ContextWithSpan(ctx, span), span, log.Hook(TraceContextHook(newCtx))
}
