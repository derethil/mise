package recipe

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/derethil/mise/internal/config"
)

var failedIDsPath = filepath.Join(config.DataDir, "recipe-clean-failed.json")

func loadFailedIDs(path string) ([]int, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var ids []int
	if err := json.Unmarshal(data, &ids); err != nil {
		return nil, err
	}

	return ids, nil
}

func saveFailedIDs(path string, ids []int) error {
	if ids == nil {
		ids = []int{}
	}

	data, err := json.Marshal(ids)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}
