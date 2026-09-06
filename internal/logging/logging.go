// Package logging provides logging functionality for the application.
package logging

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/derethil/mise/internal/config"
	"gopkg.in/natefinch/lumberjack.v2"
)

func Init(verbose bool) error {
	if err := os.MkdirAll(config.StateDir, 0o755); err != nil {
		return err
	}

	fileWriter := &lumberjack.Logger{
		Filename:   filepath.Join(config.StateDir, "mise.log"),
		MaxSize:    10, // megabytes
		MaxBackups: 5,
		MaxAge:     30, // days
		Compress:   true,
	}

	stdoutLevel := slog.LevelInfo
	if verbose {
		stdoutLevel = slog.LevelDebug
	}

	handler := contextHandler{next: slog.NewMultiHandler(
		slog.NewJSONHandler(fileWriter, &slog.HandlerOptions{Level: slog.LevelDebug}),
		newMessageHandler(os.Stdout, stdoutLevel),
	)}

	slog.SetDefault(slog.New(handler))
	return nil
}
