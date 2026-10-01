package video

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const workdirFileMode = 0o644

const workdirHashLen = 16

type WorkDir struct {
	Path string
}

func newWorkDir(url string) (*WorkDir, error) {
	path := filepath.Join(os.TempDir(), "mise-import", hashURL(url))

	if _, err := os.Stat(path); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	if err := os.MkdirAll(path, 0o755); err != nil {
		return nil, err
	}

	return &WorkDir{Path: path}, nil
}

func hashURL(url string) string {
	sum := sha256.Sum256([]byte(url))
	return hex.EncodeToString(sum[:])[:workdirHashLen]
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
