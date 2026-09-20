// Package runs stores the latest result of recipe runs.
package runs

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/derethil/mise/internal/config"
)

type Result string

const (
	ResultOK      Result = "ok"
	ResultFailure Result = "failure"
)

type entry struct {
	ID     int       `json:"id"`
	RanAt  time.Time `json:"ran_at"`
	Result Result    `json:"result"`
}

type Store struct {
	mu      sync.Mutex
	entries map[int]entry
	path    string
}

func OpenDefault(command string) (*Store, error) {
	return Open(config.CacheDir, command)
}

func Open(dir, command string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}

	path := filepath.Join(dir, command+".json")

	entries, err := loadEntries(path)
	if err != nil {
		return nil, err
	}

	return &Store{entries: entries, path: path}, nil
}

func (s *Store) Has(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.entries[id]
	return ok
}

func (s *Store) Failed() []int {
	s.mu.Lock()
	defer s.mu.Unlock()

	var ids []int
	for id, e := range s.entries {
		if e.Result == ResultFailure {
			ids = append(ids, id)
		}
	}

	sort.Ints(ids)
	return ids
}

func (s *Store) Mark(id int, result Result) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	e := entry{ID: id, RanAt: time.Now().UTC(), Result: result}

	previous, hadPrevious := s.entries[id]
	s.entries[id] = e
	if err := writeEntries(s.path, s.entries); err != nil {
		if hadPrevious {
			s.entries[id] = previous
		} else {
			delete(s.entries, id)
		}
		return err
	}

	return nil
}

func (s *Store) MarkOK(ctx context.Context, id int) {
	s.markResult(ctx, id, ResultOK)
}

func (s *Store) MarkFailure(ctx context.Context, id int) {
	s.markResult(ctx, id, ResultFailure)
}

func (s *Store) markResult(ctx context.Context, id int, result Result) {
	if err := s.Mark(id, result); err != nil {
		slog.WarnContext(ctx, "Could not update run state", slog.Int("recipe_id", id), slog.Any("error", err))
		return
	}

	slog.DebugContext(ctx, "recorded recipe run result", slog.Int("recipe_id", id), slog.String("result", string(result)))
}

func (s *Store) Close() error {
	return nil
}

func loadEntries(path string) (map[int]entry, error) {
	f, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return make(map[int]entry), nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	entries := make(map[int]entry)
	if err := json.NewDecoder(f).Decode(&entries); err != nil {
		return nil, err
	}

	return entries, nil
}

func writeEntries(path string, entries map[int]entry) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".runs-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if err := tmp.Chmod(0o644); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := json.NewEncoder(tmp).Encode(entries); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	return os.Rename(tmpPath, path)
}
