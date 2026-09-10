package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
)

func Print(cfg Config) error {
	b, err := marshalConfig(cfg)
	if err != nil {
		return err
	}

	fmt.Println(string(b))

	return nil
}

func Save(cfg Config, configPath string) error {
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return err
	}

	b, err := marshalConfig(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, b, 0o644)
}

func marshalConfig(cfg Config) ([]byte, error) {
	k := koanf.New(".")
	if err := k.Load(structs.Provider(cfg, "key"), nil); err != nil {
		return nil, err
	}

	b, err := k.Marshal(toml.Parser())
	if err != nil {
		return nil, err
	}

	return b, nil
}
