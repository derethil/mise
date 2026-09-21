// Package config provides configuration management for the application
package config

import (
	"path/filepath"

	"github.com/derethil/mise/internal/config/section"
)

type Config struct {
	Tandoor   section.TandoorConfig   `key:"tandoor" category:"TANDOOR OPTIONS"`
	Backup    section.BackupConfig    `key:"backup"`
	Providers section.ProvidersConfig `key:"providers" category:"PROVIDER OPTIONS"`
	Models    section.ModelsConfig    `key:"models"`
	Keywords  section.KeywordsConfig  `key:"keywords" command:"recipe keyword"`
}

var defaultConfig = Config{
	Tandoor: section.TandoorConfig{
		BaseURL:   "https://tandoor.dev",
		BackupDir: filepath.Join(DataDir, "tandoor_backups"),
	},
	Backup: section.BackupConfig{
		Keep: 0,
	},
	Providers: section.ProvidersConfig{
		Ollama: section.OllamaConfig{
			BaseURL:   "http://localhost:11434",
			Timeout:   600,
			Autostart: false,
		},
	},
	Keywords: section.KeywordsConfig{
		SchemaFile: filepath.Join(ConfigDir, "keyword_schema.md"),
	},
}
