package config

import "context"

type contextKey struct{}

func NewContext(ctx context.Context, cfg Config) context.Context {
	return context.WithValue(ctx, contextKey{}, cfg)
}

func FromContext(ctx context.Context) Config {
	return ctx.Value(contextKey{}).(Config)
}
