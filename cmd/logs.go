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
	Usage:    "Print the log file path for use with a viewer or other command",
	Metadata: cliutil.GlobalFlagMetadata(),
	Action: func(ctx context.Context, cmd *cli.Command) error {
		fmt.Println(logging.LogPath)
		return nil
	},
}
