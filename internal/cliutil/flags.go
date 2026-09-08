package cliutil

import "github.com/urfave/cli/v3"

type GlobalFlag string

const (
	GlobalFlagModel   GlobalFlag = "model"
	GlobalFlagVerbose GlobalFlag = "verbose"
)

func ResolveFlag(cmd *cli.Command, flag GlobalFlag, fallback string) string {
	if v := cmd.String(string(flag)); v != "" {
		return v
	}

	return fallback
}
