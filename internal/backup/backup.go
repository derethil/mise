// Package backup stores timestamped copies of arbitrary byte data on disk.
package backup

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

var ErrNoBackups = errors.New("no backups found")

const timeLayout = "20060102T150405Z"

type Entry struct {
	Path string
	Time time.Time
}

type Store struct {
	dir    string
	keep   int
	now    func() time.Time
	remove func(string) error
}

func NewStore(dir string, keep int) *Store {
	return &Store{dir: dir, keep: keep, now: time.Now, remove: os.Remove}
}

func (s *Store) Save(id int, data []byte) (Entry, error) {
	dir := s.subdir(id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return Entry{}, err
	}

	at := s.now().UTC()
	path := filepath.Join(dir, at.Format(timeLayout)+".json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return Entry{}, err
	}

	if s.keep > 0 {
		if err := s.clearStaleEntries(id, s.keep); err != nil {
			slog.Warn(fmt.Sprintf("failed to clear stale backup entries for recipe %d: %v", id, err), slog.Int("recipe_id", id), slog.Any("error", err))
		}
	}

	return Entry{Path: path, Time: at}, nil
}

func (s *Store) Load(id int, n ...int) ([]byte, error) {
	ago := 0
	if len(n) > 0 {
		ago = n[0]
	}

	if len(n) > 1 {
		return nil, errors.New("too many arguments")
	}

	entries, err := s.List(id)
	if err != nil {
		return nil, err
	}

	if ago < 0 || ago >= len(entries) {
		return nil, ErrNoBackups
	}

	return os.ReadFile(entries[ago].Path)
}

func (s *Store) clearStaleEntries(id int, keep int) error {
	entries, err := s.List(id)
	if err != nil {
		return err
	}

	var delete []Entry
	if len(entries) > keep {
		delete = entries[keep:]
	} else {
		return nil
	}

	slog.Warn(fmt.Sprintf("deleting %d stale backup entries for recipe %d", len(delete), id), slog.Int("count", len(delete)), slog.Int("recipe_id", id))

	for _, e := range delete {
		err = s.remove(e.Path)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) List(id int) ([]Entry, error) {
	dirEntries, err := os.ReadDir(s.subdir(id))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var entries []Entry
	for _, e := range dirEntries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}

		at, err := time.Parse(timeLayout, strings.TrimSuffix(e.Name(), ".json"))
		if err != nil {
			continue
		}

		entries = append(entries, Entry{
			Path: filepath.Join(s.subdir(id), e.Name()),
			Time: at,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Time.After(entries[j].Time)
	})

	return entries, nil
}

func (s *Store) subdir(id int) string {
	return filepath.Join(s.dir, strconv.Itoa(id))
}
