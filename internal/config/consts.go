package config

import (
	"path/filepath"

	"github.com/adrg/xdg"
)

var ConfigDir = filepath.Join(xdg.ConfigHome, "mise")
var DataDir = filepath.Join(xdg.DataHome, "mise")
var BackupDir = filepath.Join(DataDir, "tandoor_backups")
