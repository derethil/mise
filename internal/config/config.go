// Package config provides configuration management for the application
package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
	"github.com/urfave/cli/v3"
)

type TandoorConfig struct {
	Token   string `key:"token" usage:"Tandoor API token"`
	BaseURL string `key:"base_url" usage:"Tandoor base URL"`
}

type Config struct {
	Tandoor TandoorConfig `key:"tandoor"`
}

func Load(cmd *cli.Command) (Config, error) {
	k := koanf.New(".")

	if err := k.Load(structs.Provider(defaultConfig, "key"), nil); err != nil {
		return Config{}, err
	}

	configPath := filepath.Join(ConfigDir, "config.toml")
	if _, err := os.Stat(configPath); err == nil {
		if err := k.Load(file.Provider(configPath), toml.Parser()); err != nil {
			return Config{}, err
		}
	}

	if err := k.Load(env.Provider("MISE_", ".", envKeyLookup()), nil); err != nil {
		return Config{}, err
	}

	if err := k.Load(confmap.Provider(flagValues(cmd), "."), nil); err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := k.UnmarshalWithConf("", &cfg, koanf.UnmarshalConf{Tag: "key"}); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func envKeyLookup() func(string) string {
	names := map[string]string{}

	walkSchema(reflect.TypeFor[Config](), "", func(f schemaField) {
		envName := "MISE_" + strings.ToUpper(strings.ReplaceAll(f.Key, ".", "_"))
		names[envName] = f.Key
	})

	return func(envName string) string {
		return names[envName]
	}
}

func flagValues(cmd *cli.Command) map[string]any {
	values := map[string]any{}

	walkSchema(reflect.TypeFor[Config](), "", func(f schemaField) {
		if cmd.IsSet(f.Key) {
			values[f.Key] = cmd.String(f.Key)
		}
	})

	return values
}
