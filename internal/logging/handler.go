package logging

import (
	"context"
	"fmt"
	"io"
	"log/slog"
)

type contextHandler struct {
	next slog.Handler
}

func (h contextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h contextHandler) Handle(ctx context.Context, r slog.Record) error {
	if attrs := invocationAttrs(ctx); len(attrs) > 0 {
		r.AddAttrs(attrs...)
	}
	return h.next.Handle(ctx, r)
}

func (h contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return contextHandler{next: h.next.WithAttrs(attrs)}
}

func (h contextHandler) WithGroup(name string) slog.Handler {
	return contextHandler{next: h.next.WithGroup(name)}
}

type messageHandler struct {
	w     io.Writer
	level slog.Level
}

func newMessageHandler(w io.Writer, level slog.Level) messageHandler {
	return messageHandler{w: w, level: level}
}

func (h messageHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h messageHandler) Handle(_ context.Context, r slog.Record) error {
	_, err := fmt.Fprintln(h.w, r.Message)
	return err
}

func (h messageHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

func (h messageHandler) WithGroup(_ string) slog.Handler {
	return h
}
