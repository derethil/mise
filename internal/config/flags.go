package config

import (
	"reflect"

	"github.com/urfave/cli/v3"
)

func Flags() []cli.Flag {
	var flags []cli.Flag

	walkSchema(reflect.TypeFor[Config](), "", func(f schemaField) {
		flags = append(flags, &cli.StringFlag{
			Name:  f.Key,
			Usage: f.Usage,
		})
	})

	return flags
}
