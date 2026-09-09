package config

import (
	"os"
	"path/filepath"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
)

func Save(cfg Config, configPath string) error {
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return err
	}

	k := koanf.New(".")
	if err := k.Load(structs.Provider(cfg, "key"), nil); err != nil {
		return err
	}

	b, err := k.Marshal(toml.Parser())
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, b, 0o644)
}
