// Package config provides configuration management for the application
package config

import "path/filepath"

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

type ProvidersConfig map[string]ProviderConfig

type ModelSize string

const (
	ModelSmall ModelSize = "small"
	ModelLarge ModelSize = "large"
)

type ModelsConfig struct {
	Small string `key:"small" flag:"-" usage:"Model for simpler tasks, as provider/model"`
	Large string `key:"large" flag:"-" usage:"Model for harder tasks, as provider/model"`
}

func (m ModelsConfig) Get(size ModelSize) string {
	if size == ModelLarge {
		return m.Large
	}

	return m.Small
}

type KeywordsConfig struct {
	SchemaFile string   `key:"schema_file" usage:"Path to the file describing your keyword schema"`
	Ignore     []string `key:"ignore" usage:"Keywords to protect from removal when using recipe keyword --replace"`
}

type Config struct {
	Tandoor   TandoorConfig   `key:"tandoor" category:"TANDOOR OPTIONS"`
	Backup    BackupConfig    `key:"backup"`
	Providers ProvidersConfig `key:"providers" category:"PROVIDER OPTIONS"`
	Models    ModelsConfig    `key:"models"`
	Keywords  KeywordsConfig  `key:"keywords" command:"recipe keyword"`
}

var defaultConfig = Config{
	Tandoor: TandoorConfig{
		BaseURL:   "https://tandoor.dev",
		BackupDir: filepath.Join(DataDir, "tandoor_backups"),
	},
	Backup: BackupConfig{
		Keep: 0,
	},
	Providers: ProvidersConfig{
		"ollama": {
			BaseURL: "http://localhost:11434",
			Timeout: 600,
		},
	},
	Keywords: KeywordsConfig{
		SchemaFile: filepath.Join(ConfigDir, "keyword_schema.md"),
	},
}
