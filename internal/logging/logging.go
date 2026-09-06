// Package logging provides logging functionality for the application.
package logging

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/derethil/mise/internal/config"
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
		slog.Int64("duration_ms", time.Since(inv.start).Milliseconds()),
	}
}

func Init(verbose bool) error {
	if err := os.MkdirAll(config.StateDir, 0o755); err != nil {
		return err
	}

	f, err := os.OpenFile(filepath.Join(config.StateDir, "mise.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}

	stdoutLevel := slog.LevelInfo
	if verbose {
		stdoutLevel = slog.LevelDebug
	}

	handler := contextHandler{next: slog.NewMultiHandler(
		slog.NewJSONHandler(f, &slog.HandlerOptions{Level: slog.LevelDebug}),
		newMessageHandler(os.Stdout, stdoutLevel),
	)}

	slog.SetDefault(slog.New(handler))
	return nil
}
