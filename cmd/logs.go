package cmd

import (
	"context"
	"fmt"

	"github.com/derethil/mise/internal/cliutil"
	"github.com/derethil/mise/internal/logging"
	"github.com/urfave/cli/v3"
)

var logsCmd = &cli.Command{
	Name:     "logs",
	Usage:    "Print path to the mise logs file",
	Metadata: cliutil.GlobalFlagMetadata(),
	Action: func(ctx context.Context, cmd *cli.Command) error {
		fmt.Println(logging.LogPath)
		return nil
	},
}
