package tracer

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"entgo.io/ent"
)

// EntHook defines the "mutation middleware". A function that gets a Mutator
// and returns a Mutator.
func EntHook() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			// Extract dynamic labels from mutation.
			var (
				labels      = fmt.Sprintf("%s.%s", m.Type(), m.Op().String())
				cx, span, l = StartSpanLogTrace(ctx, labels)
				start       = time.Now()
			)
			defer span.End()
			// Before mutation, start measuring time.
			v, err := next.Mutate(cx, m)
			if err != nil {
				l.Error().Err(err).Msg("error hook tracer")
			}
			// Stop time measure.
			var duration = time.Since(start)
			l.Info().Dur("duration", duration).
				Str("table", m.Type()).
				Str("operation", m.Op().String()).
				Msg("The latency of calls in milliseconds")
			return v, err
		})
	}
}

// EntInterceptor defines an execution middleware for various types of Ent queries.
func EntInterceptor() ent.Interceptor {
	return ent.InterceptFunc(func(next ent.Querier) ent.Querier {
		return ent.QuerierFunc(func(ctx context.Context, query ent.Query) (ent.Value, error) {
			var (
				labels      = strings.ToLower(reflect.TypeOf(query).String())
				cx, span, l = StartSpanLogTrace(ctx, fmt.Sprintf("query.%s", labels))
				start       = time.Now()
			)
			if strings.HasSuffix(labels, "query") {
				defer span.End()
				labels = strings.ReplaceAll(labels, "query", "")
				labels = strings.ReplaceAll(labels, "*ent.", "")
			}
			// Before mutation, start measuring time.
			v, err := next.Query(cx, query)
			if err != nil {
				l.Error().Err(err).Msg("error interceptor tracer")
			}
			// Stop time measure.
			var duration = time.Since(start)
			l.Info().Dur("duration", duration).
				Str("table", labels).
				Str("operation", "query").
				Msg("The latency of calls in milliseconds")
			return v, err
		})
	})
}
