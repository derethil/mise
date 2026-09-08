// Package logging provides logging functionality for the application.
package logging

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/derethil/mise/internal/config"
	genkitlogger "github.com/firebase/genkit/go/core/logger"
	"gopkg.in/natefinch/lumberjack.v2"
)

var LogPath string

func Init(ctx context.Context, verbosity int) (context.Context, error) {
	if err := os.MkdirAll(config.StateDir, 0o755); err != nil {
		return ctx, err
	}

	LogPath = filepath.Join(config.StateDir, "mise.log")

	fileWriter := &lumberjack.Logger{
		Filename:   LogPath,
		MaxSize:    10, // megabytes
		MaxBackups: 5,
		MaxAge:     30, // days
		Compress:   true,
	}

	newHandler := func(stdoutLevel slog.Level) contextHandler {
		return contextHandler{next: slog.NewMultiHandler(
			slog.NewJSONHandler(fileWriter, &slog.HandlerOptions{Level: slog.LevelDebug}),
			newMessageHandler(os.Stdout, stdoutLevel),
		)}
	}

	stdoutLevel := slog.LevelInfo
	if verbosity >= 1 {
		stdoutLevel = slog.LevelDebug
	}
	slog.SetDefault(slog.New(newHandler(stdoutLevel)))

	genkitLevel := slog.LevelInfo
	if verbosity >= 2 {
		genkitLevel = slog.LevelDebug
	}
	ctx = genkitlogger.WithContext(ctx, slog.New(newHandler(genkitLevel)))

	return ctx, nil
}
