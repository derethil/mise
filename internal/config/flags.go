package config

import (
	"reflect"

	"github.com/urfave/cli/v3"
)

func Flags() []cli.Flag {
	return flagsFor("")
}

func FlagsForCommand(command string) []cli.Flag {
	return flagsFor(command)
}

func flagsFor(command string) []cli.Flag {
	var flags []cli.Flag

	walkSchema(reflect.TypeFor[Config](), "", func(f schemaField) {
		if !f.Flag || f.Command != command {
			return
		}

		flags = append(flags, newFlag(f))
	})

	return flags
}

func newFlag(f schemaField) cli.Flag {
	switch f.Type {
	case fieldTypeStrings:
		return &cli.StringSliceFlag{Name: f.FlagName, Usage: f.Usage, Category: f.Category}
	case fieldTypeInt:
		return &cli.IntFlag{Name: f.FlagName, Usage: f.Usage, Category: f.Category}
	case fieldTypeBool:
		return &cli.BoolFlag{Name: f.FlagName, Usage: f.Usage, Category: f.Category}
	default:
		return &cli.StringFlag{Name: f.FlagName, Usage: f.Usage, Category: f.Category}
	}
}
