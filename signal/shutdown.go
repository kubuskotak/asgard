package signal

import (
	"context"
	"errors"
	"time"

	"github.com/rs/zerolog/log"
)

// Graceful is the loop main module.
func Graceful(
	timeout time.Duration,
	stopCh <-chan struct{},
	errCh <-chan error,
	fn func(ctx context.Context)) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	for { // wait for SIGTERM or SIGINT
		select {
		case <-stopCh:
			fn(ctx)
			log.Error().Err(errors.New("interrupt received, shutting down")).
				Msg("Server interrupted through context")
			return nil
		case err := <-errCh:
			fn(ctx)
			log.Error().Err(err).Msg("Server interrupted through error channel")
			return err
		}
	}
}
