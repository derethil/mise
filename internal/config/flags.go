package config

import (
	"reflect"

	"github.com/urfave/cli/v3"
)

func Flags() []cli.Flag {
	var flags []cli.Flag

	walkSchema(reflect.TypeFor[Config](), "", func(f schemaField) {
		if !f.Flag {
			return
		}

		flags = append(flags, newFlag(f))
	})

	return flags
}

func newFlag(f schemaField) cli.Flag {
	switch f.Type {
	case fieldTypeStrings:
		return &cli.StringSliceFlag{Name: f.Key, Usage: f.Usage}
	case fieldTypeInt:
		return &cli.IntFlag{Name: f.Key, Usage: f.Usage}
	case fieldTypeBool:
		return &cli.BoolFlag{Name: f.Key, Usage: f.Usage}
	default:
		return &cli.StringFlag{Name: f.Key, Usage: f.Usage}
	}
}
