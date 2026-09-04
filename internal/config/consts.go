package config

import (
	"errors"
	"path/filepath"

	"github.com/adrg/xdg"
)

var ConfigDir = filepath.Join(xdg.ConfigHome, "mise")
var DataDir = filepath.Join(xdg.DataHome, "mise")

var ErrInvalidConfig = errors.New("configuration error")
