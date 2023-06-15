package tracer

import (
	"bytes"
	"context"
	"net/http"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
)

func ExampleTraceContextHook() {
	handleRequest := func(w http.ResponseWriter, req *http.Request) {
		logger := zerolog.Ctx(req.Context()).Hook(TraceContextHook(req.Context()))
		logger.Error().Msg("message")
	}
	http.HandleFunc("/", handleRequest)
}

func TestTraceContextHookNothing(t *testing.T) {
	var buf bytes.Buffer
	writer := &ZeroWriter{MinLevel: zerolog.DebugLevel}
	log.Logger = zerolog.New(
		zerolog.MultiLevelWriter(&buf, writer)).
		With().Caller().Logger()

	l := log.Hook(TraceContextHook(context.Background()))
	l.Info().Msg("test")
	require.Equal(t, "{\"level\":\"info\",\"caller\":\"C:/Users/nanan/Documents/Projects/Telkom/repos/odin/pkg/shared/tracer/log_test.go:30\",\"message\":\"test\"}\n", buf.String())
}
