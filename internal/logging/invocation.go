package logging

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"time"
)

type invocationKey struct{}

type invocation struct {
	id      string
	command string
	args    []string
	start   time.Time
}

func NewInvocation(ctx context.Context, command string, args []string) context.Context {
	id := make([]byte, 4)
	_, _ = rand.Read(id)

	inv := &invocation{
		id:      hex.EncodeToString(id),
		command: command,
		args:    args,
		start:   time.Now(),
	}
	return context.WithValue(ctx, invocationKey{}, inv)
}

func invocationAttrs(ctx context.Context) []slog.Attr {
	inv, ok := ctx.Value(invocationKey{}).(*invocation)
	if !ok {
		return nil
	}

	return []slog.Attr{
		slog.String("invocation_id", inv.id),
		slog.String("command", inv.command),
		slog.Any("args", inv.args),
		slog.Int64("elapsed_ms", time.Since(inv.start).Milliseconds()),
	}
}
