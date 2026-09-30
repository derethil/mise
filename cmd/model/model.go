// Package model implements mise's "models" command family: listing,
// pulling, and clearing Ollama models, plus managing the whisper.cpp
// transcription model.
package model

import (
	"github.com/derethil/mise/internal/cliutil"
	"github.com/urfave/cli/v3"
)

var Command = &cli.Command{
	Name:     "models",
	Usage:    "Manage your configured models and their availability",
	Metadata: cliutil.GlobalFlagMetadata(cliutil.ProviderOptions),
	Commands: []*cli.Command{
		listCmd,
		pullCmd,
		clearCmd,
		whisperCmd,
	},
}
