// Package config provides configuration management for the application
package config

import "path/filepath"

type TandoorConfig struct {
	Token     string `key:"token" usage:"Tandoor API token"`
	BaseURL   string `key:"base_url" usage:"Tandoor base URL"`
	BackupDir string `key:"backup_dir" flag:"-" usage:"Directory to store recipe backups"`
}

type Config struct {
	Tandoor TandoorConfig `key:"tandoor"`
}

var defaultConfig = Config{
	Tandoor: TandoorConfig{
		BaseURL:   "https://tandoor.dev/api",
		BackupDir: filepath.Join(DataDir, "tandoor_backups"),
	},
}
