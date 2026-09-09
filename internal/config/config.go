// Package config provides configuration management for the application
package config

import (
	"path/filepath"
	"reflect"
)

type TandoorConfig struct {
	Token     string `key:"token" usage:"Tandoor API token"`
	BaseURL   string `key:"base_url" usage:"Tandoor base URL"`
	BackupDir string `key:"backup_dir" flag:"-" usage:"Directory to store recipe backups"`
}

type BackupConfig struct {
	Keep int `key:"keep" flag:"-" usage:"Number of backups to keep per recipe (0 = keep all)"`
}

type ProviderConfig struct {
	BaseURL string `key:"base_url" usage:"Base URL of the provider's API"`
	APIKey  string `key:"api_key" usage:"API key for the provider"`
	Timeout int    `key:"timeout" flag:"-" usage:"Seconds to wait for a response from the provider"`
}

type ProvidersConfig struct {
	Ollama ProviderConfig `key:"ollama"`
}

func (p ProvidersConfig) Get(name string) (ProviderConfig, bool) {
	t := reflect.TypeOf(p)
	v := reflect.ValueOf(p)

	for field := range t.Fields() {
		if field.Tag.Get("key") != name {
			continue
		}

		cfg, ok := v.FieldByIndex(field.Index).Interface().(ProviderConfig)
		return cfg, ok
	}

	return ProviderConfig{}, false
}

func (p *ProvidersConfig) Set(name string, cfg ProviderConfig) bool {
	t := reflect.TypeOf(p).Elem()
	v := reflect.ValueOf(p).Elem()

	for field := range t.Fields() {
		if field.Tag.Get("key") != name {
			continue
		}

		v.FieldByIndex(field.Index).Set(reflect.ValueOf(cfg))
		return true
	}

	return false
}

type ModelsConfig struct {
	Small string `key:"small" flag:"-" usage:"Model for simpler tasks, as provider/model"`
	Large string `key:"large" flag:"-" usage:"Model for harder tasks, as provider/model"`
}

type Config struct {
	Tandoor   TandoorConfig   `key:"tandoor"`
	Backup    BackupConfig    `key:"backup"`
	Providers ProvidersConfig `key:"providers"`
	Models    ModelsConfig    `key:"models"`
}

var defaultConfig = Config{
	Tandoor: TandoorConfig{
		BaseURL:   "https://tandoor.dev/api",
		BackupDir: filepath.Join(DataDir, "tandoor_backups"),
	},
	Backup: BackupConfig{
		Keep: 0,
	},
	Providers: ProvidersConfig{
		Ollama: ProviderConfig{
			BaseURL: "http://localhost:11434",
			Timeout: 600,
		},
	},
	Models: ModelsConfig{
		Small: "ollama/qwen2.5:7b",
		Large: "ollama/qwen2.5:14b",
	},
}
