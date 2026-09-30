package config

import (
	"fmt"
	"reflect"
	"strings"

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
		if f.FlagName == "" || f.Command != command {
			return
		}

		flags = append(flags, newFlag(f))
	})

	return flags
}

func newFlag(f schemaField) cli.Flag {
	defaultText := flagDefaultText(f.Key)
	var aliases []string
	if f.Alias != "" {
		aliases = []string{f.Alias}
	}

	switch f.Type {
	case fieldTypeStrings:
		return &stringSliceFlag{StringSliceFlag: &cli.StringSliceFlag{
			Name: f.FlagName, Aliases: aliases, Usage: f.Usage, Category: f.Category,
			DefaultText: defaultText, OnlyOnce: true,
		}}
	case fieldTypeInt:
		return &cli.IntFlag{Name: f.FlagName, Aliases: aliases, Usage: f.Usage, Category: f.Category, DefaultText: defaultText}
	case fieldTypeBool:
		return &cli.BoolFlag{Name: f.FlagName, Aliases: aliases, Usage: f.Usage, Category: f.Category, DefaultText: defaultText}
	default:
		return &cli.StringFlag{Name: f.FlagName, Aliases: aliases, Usage: f.Usage, Category: f.Category, DefaultText: defaultText}
	}
}

// stringSliceFlag accepts a comma-separated list in one occurrence. urfave/cli
// otherwise presents every slice as a repeatable multi-value flag in help.
type stringSliceFlag struct {
	*cli.StringSliceFlag
}

func (f *stringSliceFlag) String() string {
	return cli.FlagStringer(f)
}

func (f *stringSliceFlag) IsMultiValueFlag() bool {
	return false
}

func flagDefaultText(key string) string {
	value := reflect.ValueOf(defaultConfig)

	for _, part := range strings.Split(key, ".") {
		typeOfValue := value.Type()
		for i := range typeOfValue.NumField() {
			if typeOfValue.Field(i).Tag.Get("key") == part {
				value = value.Field(i)
				break
			}
		}
	}

	if value.Kind() == reflect.Slice && value.Len() == 0 {
		return ""
	}

	return fmt.Sprint(value.Interface())
}
