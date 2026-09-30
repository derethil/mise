package model

import (
	"github.com/derethil/mise/internal/cliutil"
	"github.com/urfave/cli/v3"
)

var whisperCmd = &cli.Command{
	Name:     "whisper",
	Usage:    "Manage the whisper.cpp model used for transcription",
	Metadata: cliutil.GlobalFlagMetadata(),
	Commands: []*cli.Command{
		whisperPullCmd,
	},
}
