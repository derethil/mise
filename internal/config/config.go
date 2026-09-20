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
	BaseURL string
	APIKey  string
	Timeout int
}

type OllamaConfig struct {
	BaseURL   string `key:"base_url" usage:"Base URL for the Ollama API, e.g. http://localhost:11434"`
	Timeout   int    `key:"timeout" flag:"-" usage:"Seconds to wait for a response from Ollama"`
	Autostart bool   `key:"autostart" usage:"Start Ollama automatically if not running, requires Ollama to be installed and in PATH"`
}

type ProvidersConfig struct {
	Ollama OllamaConfig `key:"ollama"`
}

func (p ProvidersConfig) Get(name string) (ProviderConfig, bool) {
	switch name {
	case "ollama":
		return ProviderConfig{
			BaseURL: p.Ollama.BaseURL,
			Timeout: p.Ollama.Timeout,
		}, true
	default:
		return ProviderConfig{}, false
	}
}

func (p *ProvidersConfig) Set(name string, cfg ProviderConfig) bool {
	switch name {
	case "ollama":
		p.Ollama.BaseURL = cfg.BaseURL
		p.Ollama.Timeout = cfg.Timeout
		return true
	default:
		return false
	}
}

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
		Ollama: OllamaConfig{
			BaseURL:   "http://localhost:11434",
			Timeout:   600,
			Autostart: false,
		},
	},
	Keywords: KeywordsConfig{
		SchemaFile: filepath.Join(ConfigDir, "keyword_schema.md"),
	},
}
