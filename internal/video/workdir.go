package video

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const workdirFileMode = 0o644

type WorkDir struct {
	Path string
}

func newWorkDir() (*WorkDir, error) {
	path, err := os.MkdirTemp("", "mise-import-*")
	if err != nil {
		return nil, err
	}

	return &WorkDir{Path: path}, nil
}

func (w *WorkDir) Join(elem ...string) string {
	return filepath.Join(append([]string{w.Path}, elem...)...)
}

func (w *WorkDir) WriteFile(name string, data []byte) error {
	return os.WriteFile(w.Join(name), data, workdirFileMode)
}

func (w *WorkDir) WriteJSON(name string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal %s: %w", name, err)
	}

	return w.WriteFile(name, data)
}

func (w *WorkDir) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(w.Join(name))
}

func (w *WorkDir) Exists(name string) bool {
	_, err := os.Stat(w.Join(name))
	return err == nil
}

func (w *WorkDir) Glob(pattern string) ([]string, error) {
	return filepath.Glob(w.Join(pattern))
}

func (w *WorkDir) Cleanup() error {
	return os.RemoveAll(w.Path)
}
