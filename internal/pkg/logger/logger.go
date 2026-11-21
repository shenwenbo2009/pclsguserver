package logger

import (
	"context"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type ctxKey struct{}

var requestIDKey = ctxKey{}

func Init(level string) {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		lvl = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(lvl)
	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
}

func L() *zerolog.Logger {
	return &log.Logger
}

func WithRequestId(ctx context.Context, requestId string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestId)
}

func FromCtx(ctx context.Context) *zerolog.Logger {
	l := log.Logger
	if rid, ok := ctx.Value(requestIDKey).(string); ok && rid != "" {
		l = l.With().Str("requestId", rid).Logger()
	}
	return &l
}

func RequestIdFromCtx(ctx context.Context) string {
	if rid, ok := ctx.Value(requestIDKey).(string); ok {
		return rid
	}
	return ""
}
