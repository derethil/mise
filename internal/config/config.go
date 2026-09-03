// Package config provides configuration management for the application
package config

import "path/filepath"

type TandoorConfig struct {
	Token     string `key:"token" usage:"Tandoor API token"`
	BaseURL   string `key:"base_url" usage:"Tandoor base URL"`
	BackupDir string `key:"backup_dir" flag:"-" usage:"Directory to store recipe backups"`
}

type ProviderConfig struct {
	BaseURL string `key:"base_url" usage:"Base URL of the provider's API"`
	APIKey  string `key:"api_key" usage:"API key for the provider"`
}

type ProvidersConfig struct {
	Ollama ProviderConfig `key:"ollama"`
}

type ModelsConfig struct {
	Small string `key:"small" flag:"-" usage:"Model for simpler tasks, as provider/model"`
	Large string `key:"large" flag:"-" usage:"Model for harder tasks, as provider/model"`
}

type Config struct {
	Tandoor   TandoorConfig   `key:"tandoor"`
	Providers ProvidersConfig `key:"providers"`
	Models    ModelsConfig    `key:"models"`
}

var defaultConfig = Config{
	Tandoor: TandoorConfig{
		BaseURL:   "https://tandoor.dev/api",
		BackupDir: filepath.Join(DataDir, "tandoor_backups"),
	},
	Providers: ProvidersConfig{
		Ollama: ProviderConfig{
			BaseURL: "http://localhost:11434",
		},
	},
	Models: ModelsConfig{
		Small: "ollama/qwen2.5:7b",
		Large: "ollama/qwen2.5:14b",
	},
}
